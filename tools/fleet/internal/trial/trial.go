package trial

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
)

type Input struct {
	Dir        string
	Tip        string
	Branch     string
	RegenGlobs []string
}

// Run writes exactly one line to d.Out: clean, regen <path>... or conflict <path>...
// It returns an error only when git itself fails, never for a conflict.
func Run(ctx context.Context, d deps.Deps, in Input) error {
	paths, err := conflicts(ctx, d, in)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(d.Out, verdict(paths, in.RegenGlobs))
	return err
}

func conflicts(ctx context.Context, d deps.Deps, in Input) ([]string, error) {
	env := []string{"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1"}
	out, err := d.Git.Run(ctx, in.Dir, env, "merge-tree", "--write-tree", "--name-only", "-z", in.Tip, in.Branch)
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(out) == 0 {
			return nil, err
		}
	}
	fields := bytes.Split(out, []byte{0})
	var paths []string
	for _, f := range fields[1:] {
		if len(f) == 0 {
			break
		}
		paths = append(paths, string(f))
	}
	return paths, nil
}

func verdict(paths, globs []string) string {
	if len(paths) == 0 {
		return "clean"
	}
	kind := "regen"
	for _, p := range paths {
		if !config.MatchAny(globs, p) {
			kind = "conflict"
			break
		}
	}
	return kind + " " + strings.Join(paths, " ")
}
