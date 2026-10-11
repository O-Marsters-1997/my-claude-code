package deps

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
)

type Git interface {
	Run(ctx context.Context, dir string, env []string, args ...string) ([]byte, error)
}

type GH interface {
	OpenIssues(ctx context.Context, labels ...string) ([]gh.Issue, error)
	OpenPRs(ctx context.Context, base string) ([]gh.PR, error)
	MergePR(ctx context.Context, number int) error
	Issues(ctx context.Context, label string) ([]gh.Issue, error)
	PRs(ctx context.Context, base string) ([]gh.PR, error)
	IssueClosed(ctx context.Context, number int) (bool, error)
	MoveStatus(ctx context.Context, number int, from, to string) error
	CloseIssue(ctx context.Context, number int) error
}

type Shell interface {
	Run(ctx context.Context, dir, command string) ([]byte, error)
}

type Deps struct {
	Git Git
	GH  GH
	Sh  Shell
	Out io.Writer
	Log io.Writer
}

func DefaultDeps(out, errw io.Writer) Deps {
	return Deps{Git: execGit{}, GH: gh.Client{}, Sh: execShell{}, Out: out, Log: errw}
}

type execGit struct{}

func (execGit) Run(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	return runCapture(cmd, "git", args)
}

type execShell struct{}

func (execShell) Run(ctx context.Context, dir, command string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	return runCapture(cmd, "sh", []string{"-c", command})
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
