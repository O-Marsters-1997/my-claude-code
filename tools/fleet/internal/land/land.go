package land

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type Input struct {
	Dir       string
	Label     string
	Check     bool
	Plan      string
	Regen     map[string]string
	PrePush   []string
	Suite     []string
	Resolved  []string
	Started   time.Time
	CIPoll    time.Duration
	CITimeout time.Duration
}

type resolution struct {
	PR     int
	Failed bool
	Class  string
	Report string
}

func parseResolution(l string) (resolution, error) {
	bad := fmt.Errorf("bad --resolved line %q: want \"#<PR> resolved|failed append|rewrite <report path>\"", l)
	f := strings.SplitN(l, " ", 4)
	if len(f) != 4 || f[3] == "" {
		return resolution{}, bad
	}
	pr, err := strconv.Atoi(strings.TrimPrefix(f[0], "#"))
	if err != nil || !strings.HasPrefix(f[0], "#") || pr <= 0 {
		return resolution{}, bad
	}
	if (f[1] != "resolved" && f[1] != "failed") || (f[2] != "append" && f[2] != "rewrite") {
		return resolution{}, bad
	}
	return resolution{PR: pr, Failed: f[1] == "failed", Class: f[2], Report: f[3]}, nil
}

func Run(ctx context.Context, d deps.Deps, in Input) (err error) {
	if in.Check {
		return run(ctx, d, in, &metrics.Row{}, &time.Time{})
	}
	row := metrics.Row{Label: in.Label, Suite: "skipped"}
	started := in.Started
	defer func() {
		row.TS = time.Now().UTC()
		if !started.IsZero() {
			row.WallClockS = int(row.TS.Sub(started).Seconds())
		}
		if merr := metrics.Append(in.Dir, row); merr != nil {
			err = errors.Join(err, fmt.Errorf("record metrics: %w", merr))
			return
		}
		fmt.Fprintln(d.Log, row)
	}()
	return run(ctx, d, in, &row, &started)
}

func run(ctx context.Context, d deps.Deps, in Input, row *metrics.Row, started *time.Time) error {
	if in.Check {
		in.CITimeout = 0
	}
	resolutions := map[int]resolution{}
	var firstFailed int
	for _, l := range in.Resolved {
		r, err := parseResolution(l)
		if err != nil {
			return err
		}
		fmt.Fprintf(d.Log, "resolver: %s\n", l)
		row.AddResolution(r.Failed, r.Class)
		resolutions[r.PR] = r
		if r.Failed && firstFailed == 0 {
			firstFailed = r.PR
		}
	}
	if in.Plan == "" {
		in.Plan = PlanPath(in.Dir)
	}
	tip := "feat/" + strings.TrimPrefix(in.Label, "project:")
	if err := fetch(ctx, d, in.Dir); err != nil {
		return err
	}
	prs, err := d.GH.OpenPRs(ctx, tip)
	if err != nil {
		return err
	}
	issues, err := d.GH.OpenIssues(ctx, in.Label)
	if err != nil {
		return err
	}
	if err := requireWave(ctx, d, in, tip, prs, issues); err != nil {
		return err
	}
	entries := queue.Build(prs, issues)
	if started.IsZero() {
		*started = earliestCreated(prs)
	}

	nodes, err := analyse(ctx, d, in, tip, entries)
	if err != nil {
		return err
	}
	if in.Check {
		return line(d, "%s", recommend(nodes))
	}
	for _, n := range nodes {
		if len(n.Conflicts) > 0 || len(n.PullIn) > 0 || len(n.regenFrom) > 0 {
			row.TrialConflicts++
		}
	}
	fixed, ok, err := fixMechanical(ctx, d, in, row, nodes)
	if err != nil || !ok {
		return err
	}
	if fixed > 0 {
		if err := fetch(ctx, d, in.Dir); err != nil {
			return err
		}
		if nodes, err = analyse(ctx, d, in, tip, entries); err != nil {
			return err
		}
	}

	prev := readPlan(in.Plan, in.Label)
	assignStatus(nodes, resolutions, prev)
	for _, n := range nodes {
		if strings.HasPrefix(n.Status, "gate-failed") {
			row.GateFailures++
		}
		fmt.Fprintf(d.Out, "#%d %s ci=%s conflicts=%s pull-in=%s %s\n", n.Number, n.Branch, n.CI,
			orNone(strings.Join(n.Conflicts, ",")), orNone(pullInList(n.PullIn)), n.Status)
	}
	if err := writePlan(in.Plan, plan{Label: in.Label, Tip: tip, PRs: nodes}); err != nil {
		return fmt.Errorf("write plan: %w", err)
	}
	if pending := withStatus(nodes, "needs-work"); len(pending) > 0 {
		return line(d, "needs-work %s", strings.Join(pending, " "))
	}
	for _, n := range nodes {
		if strings.HasPrefix(n.Status, "gate-failed ") {
			return line(d, "gate-failed #%d %s", n.Number, strings.TrimPrefix(n.Status, "gate-failed "))
		}
	}
	for _, n := range nodes {
		if n.CI == "pending" {
			return line(d, "gate-failed #%d ci-pending", n.Number)
		}
	}
	if firstFailed != 0 {
		row.GateFailures++
		return line(d, "gate-failed #%d resolver-failed", firstFailed)
	}
	return verify(ctx, d, in, row, tip, nodes)
}

func assignStatus(nodes []node, resolutions map[int]resolution, prev plan) {
	carried := map[int][]pullIn{}
	for _, p := range prev.PRs {
		carried[p.Number] = p.PullIn
	}
	broken := map[int]bool{}
	for i := range nodes {
		n := &nodes[i]
		r, resolved := resolutions[n.Number]
		if resolved && len(n.PullIn) == 0 {
			n.PullIn = carried[n.Number]
		}
		blocker := 0
		for _, p := range n.PullIn {
			if broken[p.PR] {
				blocker = p.PR
				break
			}
		}
		switch {
		case resolved && r.Failed:
			n.Status = "gate-failed resolver-failed"
		case blocker != 0 && !resolved:
			n.Status = fmt.Sprintf("blocked-by #%d", blocker)
		case resolved && n.NeedsWork && n.CI == "red":
			n.Status = "gate-failed ci-red"
		case resolved && n.NeedsWork:
			n.Status = "gate-failed conflicts-remain"
		case n.NeedsWork:
			n.Status = "needs-work"
		default:
			n.Status = "ok"
		}
		broken[n.Number] = n.Status != "ok" && n.Status != "needs-work"
	}
}

func fixMechanical(ctx context.Context, d deps.Deps, in Input, row *metrics.Row, nodes []node) (fixed int, ok bool, err error) {
	for _, n := range nodes {
		pr := gh.PR{Number: n.Number, HeadRefName: n.Branch}
		for _, f := range n.regenFrom {
			done, err := regenerate(ctx, d, in, row, f.from, pr, f.paths)
			if err != nil || !done {
				return fixed, false, err
			}
			row.RegenFixes++
			fixed++
		}
	}
	return fixed, true, nil
}

func verify(ctx context.Context, d deps.Deps, in Input, row *metrics.Row, tip string, nodes []node) error {
	if err := fetch(ctx, d, in.Dir); err != nil {
		return err
	}
	var chain []landedPR
	running := "origin/" + tip
	for _, n := range nodes {
		sha, paths, err := trial.Merge(ctx, d, in.Dir, running, "origin/"+n.Branch)
		if err != nil {
			return fmt.Errorf("simulate #%d: %w", n.Number, err)
		}
		if len(paths) > 0 {
			return line(d, "gate-failed #%d conflicts-remain", n.Number)
		}
		running = sha
		chain = append(chain, landedPR{number: n.Number, sha: sha})
	}

	if len(chain) > 0 && len(in.Regen) > 0 {
		stale, err := regenStale(ctx, d, in, running)
		if err != nil {
			return err
		}
		if len(stale) > 0 {
			return line(d, "regen-stale %s", strings.Join(stale, " "))
		}
	}
	if len(chain) > 0 {
		green, err := suiteGreen(ctx, d, in, running)
		if err != nil {
			return err
		}
		if len(in.Suite) > 0 {
			row.Suite = "green"
			if !green {
				row.Suite = "red"
				row.Bisects++
				culprit, err := bisect(ctx, d, in, chain)
				if err != nil {
					return err
				}
				return line(d, "suite-red #%d", culprit)
			}
		}
	}
	for _, n := range nodes {
		if len(n.PullIn) == 0 {
			continue
		}
		last := n.PullIn[len(n.PullIn)-1]
		if err := d.GH.CommentPR(ctx, n.Number, compareComment(n, last)); err != nil {
			return fmt.Errorf("comment on #%d: %w", n.Number, err)
		}
	}
	return line(d, "ready %d", len(nodes))
}

func compareComment(n node, last pullIn) string {
	repo, _, _ := strings.Cut(n.URL, "/pull/")
	return fmt.Sprintf("This branch pulled in earlier PRs to merge conflict-free. Its own diff, since #%d: %s/compare/%s...%s",
		last.PR, repo, last.Branch, n.Branch)
}

func withStatus(nodes []node, status string) []string {
	var out []string
	for _, n := range nodes {
		if n.Status == status {
			out = append(out, fmt.Sprintf("#%d", n.Number))
		}
	}
	return out
}

func pullInList(ps []pullIn) string {
	refs := make([]string, len(ps))
	for i, p := range ps {
		refs[i] = fmt.Sprintf("#%d", p.PR)
	}
	return strings.Join(refs, ",")
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

func earliestCreated(prs []gh.PR) time.Time {
	var earliest time.Time
	for _, p := range prs {
		if !p.CreatedAt.IsZero() && (earliest.IsZero() || p.CreatedAt.Before(earliest)) {
			earliest = p.CreatedAt
		}
	}
	return earliest
}

func line(d deps.Deps, format string, args ...any) error {
	_, err := fmt.Fprintf(d.Out, format+"\n", args...)
	return err
}

func fetch(ctx context.Context, d deps.Deps, dir string) error {
	if _, err := d.Git.Run(ctx, dir, nil, "fetch", "--prune", "origin"); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	return nil
}

func suiteGreen(ctx context.Context, d deps.Deps, in Input, ref string) (bool, error) {
	if len(in.Suite) == 0 {
		return true, nil
	}
	green := true
	err := onRef(ctx, d, in, ref, func(dir string) error {
		for _, cmd := range in.Suite {
			fmt.Fprintf(d.Log, "suite: %s\n", cmd)
			_, err := d.Sh.Run(ctx, dir, cmd)
			var exit *exec.ExitError
			switch {
			case errors.As(err, &exit):
				green = false
				return nil
			case err != nil:
				return fmt.Errorf("run suite %q: %w", cmd, err)
			}
		}
		return nil
	})
	return green, err
}

func onRef(ctx context.Context, d deps.Deps, in Input, ref string, fn func(dir string) error) error {
	tmp, err := os.MkdirTemp("", "fleet-ref-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if _, err := d.Git.Run(ctx, in.Dir, nil, "worktree", "add", "--detach", tmp, ref); err != nil {
		return fmt.Errorf("check out %s: %w", ref, err)
	}
	defer d.Git.Run(context.WithoutCancel(ctx), in.Dir, nil, "worktree", "remove", "--force", tmp)
	return fn(tmp)
}
