package land

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type Input struct {
	Dir     string
	Label   string
	Regen   map[string]string
	PrePush []string
	Suite   []string
}

func Run(ctx context.Context, d deps.Deps, in Input) error {
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

	landed := 0
	for _, e := range ready {
		if err := fetch(ctx, d, in.Dir); err != nil {
			return err
		}
		verdict, err := trial.Classify(ctx, d, trial.Input{
			Dir:        in.Dir,
			Tip:        "origin/" + tip,
			Branch:     "origin/" + e.PR.HeadRefName,
			RegenGlobs: config.Config{Regen: in.Regen}.RegenGlobs(),
		})
		if err != nil {
			return fmt.Errorf("trial #%d: %w", e.PR.Number, err)
		}
		kind, paths, _ := strings.Cut(verdict, " ")
		switch kind {
		case "conflict":
			return line(d, "needs-resolver #%d %s %s", e.PR.Number, e.PR.HeadRefName, paths)
		case "regen":
			ok, err := regenerate(ctx, d, in, tip, e.PR, strings.Fields(paths))
			if err != nil || !ok {
				return err
			}
		}
		if err := d.GH.MergePR(ctx, e.PR.Number); err != nil {
			return fmt.Errorf("merge #%d: %w", e.PR.Number, err)
		}
		fmt.Fprintf(d.Log, "landed #%d\n", e.PR.Number)
		landed++
	}

	if landed > 0 && len(in.Regen) > 0 {
		stale, err := regenStale(ctx, d, in, tip)
		if err != nil {
			return err
		}
		if len(stale) > 0 {
			return line(d, "regen-stale %s", strings.Join(stale, " "))
		}
	}

	green := true
	if landed > 0 {
		green, err = suiteGreen(ctx, d, in, tip)
		if err != nil {
			return err
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
	green := true
	err := onTip(ctx, d, in, tip, func(dir string) error {
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
