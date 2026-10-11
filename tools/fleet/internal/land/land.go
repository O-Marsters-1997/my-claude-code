package land

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type Input struct {
	Dir        string
	Label      string
	RegenGlobs []string
	Suite      []string
	Resolved   []string
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

func Run(ctx context.Context, d deps.Deps, in Input) error {
	var resolutions []resolution
	for _, l := range in.Resolved {
		r, err := parseResolution(l)
		if err != nil {
			return err
		}
		fmt.Fprintf(d.Log, "resolver: %s\n", l)
		resolutions = append(resolutions, r)
	}
	for _, r := range resolutions {
		if r.Failed {
			return line(d, "gate-failed #%d", r.PR)
		}
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
	ready, drafts := queue.Build(prs, issues)

	var landed []landedPR
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
			paths := strings.Join(strings.Fields(verdict)[1:], " ")
			return line(d, "needs-resolver #%d %s %s", e.PR.Number, e.PR.HeadRefName, paths)
		}
		if err := d.GH.MergePR(ctx, e.PR.Number); err != nil {
			return fmt.Errorf("merge #%d: %w", e.PR.Number, err)
		}
		fmt.Fprintf(d.Log, "landed #%d\n", e.PR.Number)
		if err := fetch(ctx, d, in.Dir); err != nil {
			return err
		}
		sha, err := d.Git.Run(ctx, in.Dir, nil, "rev-parse", "origin/"+tip)
		if err != nil {
			return fmt.Errorf("resolve %s after #%d: %w", tip, e.PR.Number, err)
		}
		landed = append(landed, landedPR{number: e.PR.Number, sha: strings.TrimSpace(string(sha))})
	}

	green := true
	if len(landed) > 0 {
		green, err = suiteGreen(ctx, d, in, "origin/"+tip)
		if err != nil {
			return err
		}
	}
	switch {
	case !green:
		culprit, err := bisect(ctx, d, in, landed)
		if err != nil {
			return err
		}
		return line(d, "suite-red #%d", culprit)
	case len(drafts) > 0:
		refs := make([]string, len(drafts))
		for i, p := range drafts {
			refs[i] = fmt.Sprintf("#%d", p.Number)
		}
		return line(d, "drafts %s", strings.Join(refs, " "))
	}
	return line(d, "done %d", len(landed))
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
	tmp, err := os.MkdirTemp("", "fleet-suite-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmp)
	if _, err := d.Git.Run(ctx, in.Dir, nil, "worktree", "add", "--detach", tmp, ref); err != nil {
		return false, fmt.Errorf("check out %s: %w", ref, err)
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
