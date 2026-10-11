package land

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type Input struct {
	Dir        string
	Label      string
	RegenGlobs []string
	Suite      []string
	Resolved   []string
	Started    time.Time
}

func Run(ctx context.Context, d deps.Deps, in Input) (err error) {
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
	if rerr := row.ApplyResolved(in.Resolved); rerr != nil {
		return rerr
	}
	return run(ctx, d, in, &row, &started)
}

func run(ctx context.Context, d deps.Deps, in Input, row *metrics.Row, started *time.Time) error {
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
	ready, drafts := queue.Build(prs, issues)
	row.DraftsSkipped = len(drafts)
	if started.IsZero() {
		*started = earliestCreated(prs)
	}

	landed := 0
	for _, e := range ready {
		if err := fetch(ctx, d, in.Dir); err != nil {
			return err
		}
		verdict, err := trial.Classify(ctx, d, trial.Input{
			Dir:        in.Dir,
			Tip:        "origin/" + tip,
			Branch:     "origin/" + e.PR.HeadRefName,
			RegenGlobs: in.RegenGlobs,
		})
		if err != nil {
			return fmt.Errorf("trial #%d: %w", e.PR.Number, err)
		}
		if verdict != "clean" {
			row.TrialConflicts++
			paths := strings.Join(strings.Fields(verdict)[1:], " ")
			return line(d, "needs-resolver #%d %s %s", e.PR.Number, e.PR.HeadRefName, paths)
		}
		if err := d.GH.MergePR(ctx, e.PR.Number); err != nil {
			return fmt.Errorf("merge #%d: %w", e.PR.Number, err)
		}
		fmt.Fprintf(d.Log, "landed #%d\n", e.PR.Number)
		landed++
		row.PRsLanded = landed
	}

	green := true
	if landed > 0 {
		green, err = suiteGreen(ctx, d, in, tip)
		if err != nil {
			return err
		}
		if len(in.Suite) > 0 {
			row.Suite = "green"
			if !green {
				row.Suite = "red"
			}
		}
	}
	switch {
	case !green:
		return line(d, "suite-red")
	case len(drafts) > 0:
		refs := make([]string, len(drafts))
		for i, p := range drafts {
			refs[i] = fmt.Sprintf("#%d", p.Number)
		}
		return line(d, "drafts %s", strings.Join(refs, " "))
	}
	return line(d, "done %d", landed)
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

func suiteGreen(ctx context.Context, d deps.Deps, in Input, tip string) (bool, error) {
	if len(in.Suite) == 0 {
		return true, nil
	}
	if err := fetch(ctx, d, in.Dir); err != nil {
		return false, err
	}
	tmp, err := os.MkdirTemp("", "fleet-suite-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmp)
	if _, err := d.Git.Run(ctx, in.Dir, nil, "worktree", "add", "--detach", tmp, "origin/"+tip); err != nil {
		return false, fmt.Errorf("check out %s: %w", tip, err)
	}
	defer d.Git.Run(context.WithoutCancel(ctx), in.Dir, nil, "worktree", "remove", "--force", tmp)

	for _, cmd := range in.Suite {
		fmt.Fprintf(d.Log, "suite: %s\n", cmd)
		_, err := d.Sh.Run(ctx, tmp, cmd)
		var exit *exec.ExitError
		switch {
		case errors.As(err, &exit):
			return false, nil
		case err != nil:
			return false, fmt.Errorf("run suite %q: %w", cmd, err)
		}
	}
	return true, nil
}
