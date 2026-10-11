package land

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
)

const commentLimit = 60000

func regenerate(ctx context.Context, d deps.Deps, in Input, tip string, pr gh.PR, paths []string) (bool, error) {
	at := func(command string) ([]byte, error) {
		return d.Sh.Run(ctx, in.Dir, "tp exec "+quote(pr.HeadRefName)+" sh -c "+quote(command))
	}
	origin := "origin/" + tip

	if _, err := at("git merge --no-commit --no-ff " + quote(origin) + " || true"); err != nil {
		return false, fmt.Errorf("merge %s into #%d: %w", tip, pr.Number, err)
	}
	committed := false
	defer func() {
		if !committed {
			at("git merge --abort")
		}
	}()

	for _, p := range paths {
		if _, err := at("git checkout --theirs -- " + quote(p) + " && git add -- " + quote(p)); err != nil {
			return false, fmt.Errorf("take %s's %s for #%d: %w", tip, p, pr.Number, err)
		}
	}
	for _, cmd := range regenCommands(in.Regen, paths) {
		fmt.Fprintf(d.Log, "regen #%d: %s\n", pr.Number, cmd)
		if _, err := at(cmd); err != nil {
			return false, fmt.Errorf("regen %q for #%d: %w", cmd, pr.Number, err)
		}
	}
	msg := fmt.Sprintf("Merge %s into %s, regenerating %s", origin, pr.HeadRefName, strings.Join(paths, " "))
	if _, err := at("git add -A && git commit -q -m " + quote(msg)); err != nil {
		return false, fmt.Errorf("commit regen for #%d: %w", pr.Number, err)
	}
	committed = true

	for _, cmd := range in.PrePush {
		fmt.Fprintf(d.Log, "pre_push #%d: %s\n", pr.Number, cmd)
		_, err := at(cmd)
		var exit *exec.ExitError
		switch {
		case errors.As(err, &exit):
			return false, line(d, "gate-failed #%d %s", pr.Number, cmd)
		case err != nil:
			return false, fmt.Errorf("run pre_push %q for #%d: %w", cmd, pr.Number, err)
		}
	}

	if _, err := at("git push origin HEAD:refs/heads/" + quote(pr.HeadRefName)); err != nil {
		return false, fmt.Errorf("push #%d: %w", pr.Number, err)
	}
	diff, err := at("git diff " + quote(origin) + " HEAD -- " + quoteAll(paths))
	if err != nil {
		return false, fmt.Errorf("diff regen for #%d: %w", pr.Number, err)
	}
	if err := d.GH.CommentPR(ctx, pr.Number, regenComment(tip, paths, string(diff))); err != nil {
		return false, fmt.Errorf("comment on #%d: %w", pr.Number, err)
	}
	return true, nil
}

func regenCommands(regen map[string]string, paths []string) []string {
	var cmds []string
	for _, glob := range sortedKeys(regen) {
		cmd := regen[glob]
		if slices.Contains(cmds, cmd) {
			continue
		}
		for _, p := range paths {
			if config.MatchAny([]string{glob}, p) {
				cmds = append(cmds, cmd)
				break
			}
		}
	}
	return cmds
}

func regenComment(tip string, paths []string, diff string) string {
	if len(diff) > commentLimit {
		diff = diff[:commentLimit] + "\n... truncated"
	}
	return fmt.Sprintf("Merged `%s` into this branch and regenerated %s:\n\n```diff\n%s\n```", tip, strings.Join(paths, ", "), diff)
}

func regenStale(ctx context.Context, d deps.Deps, in Input, tip string) ([]string, error) {
	var stale []string
	err := onTip(ctx, d, in, tip, func(dir string) error {
		for _, glob := range sortedKeys(in.Regen) {
			cmd := in.Regen[glob]
			fmt.Fprintf(d.Log, "regen check: %s\n", cmd)
			if _, err := d.Sh.Run(ctx, dir, cmd); err != nil {
				return fmt.Errorf("run regen %q: %w", cmd, err)
			}
		}
		out, err := d.Git.Run(ctx, dir, nil, "status", "--porcelain", "-z", "--untracked-files=all")
		if err != nil {
			return err
		}
		for _, entry := range strings.Split(string(out), "\x00") {
			if len(entry) > 3 {
				stale = append(stale, entry[3:])
			}
		}
		return nil
	})
	return stale, err
}

func onTip(ctx context.Context, d deps.Deps, in Input, tip string, fn func(dir string) error) error {
	if err := fetch(ctx, d, in.Dir); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "fleet-tip-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if _, err := d.Git.Run(ctx, in.Dir, nil, "worktree", "add", "--detach", tmp, "origin/"+tip); err != nil {
		return fmt.Errorf("check out %s: %w", tip, err)
	}
	defer d.Git.Run(context.WithoutCancel(ctx), in.Dir, nil, "worktree", "remove", "--force", tmp)
	return fn(tmp)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func quoteAll(ss []string) string {
	q := make([]string, len(ss))
	for i, s := range ss {
		q[i] = quote(s)
	}
	return strings.Join(q, " ")
}
