package deps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Git interface {
	Run(ctx context.Context, dir string, env []string, args ...string) ([]byte, error)
}

type Issue struct {
	Number int    `json:"number"`
	Body   string `json:"body"`
}

type GH interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
	ReadyIssues(ctx context.Context, label string) ([]Issue, error)
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

func (g execGH) ReadyIssues(ctx context.Context, label string) ([]Issue, error) {
	out, err := g.Run(ctx, "issue", "list", "--state", "open", "--label", label, "--label", "status:ready",
		"--limit", "200", "--json", "number,body")
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("decode gh issue list: %w", err)
	}
	return issues, nil
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
