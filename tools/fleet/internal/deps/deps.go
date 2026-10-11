package deps

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Git interface {
	Run(ctx context.Context, dir string, env []string, args ...string) ([]byte, error)
}

type GH interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

type Deps struct {
	Git Git
	GH  GH
	Out io.Writer
	Log io.Writer
}

func DefaultDeps(out, errw io.Writer) Deps {
	return Deps{Git: execGit{}, GH: execGH{}, Out: out, Log: errw}
}

type execGit struct{}

func (execGit) Run(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	return runCapture(cmd, "git", args)
}

type execGH struct{}

func (execGH) Run(ctx context.Context, args ...string) ([]byte, error) {
	return runCapture(exec.CommandContext(ctx, "gh", args...), "gh", args)
}

func runCapture(cmd *exec.Cmd, name string, args []string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("%s %v: %w: %s", name, args, err, bytes.TrimSpace(stderr.Bytes()))
	}
	return stdout.Bytes(), nil
}
