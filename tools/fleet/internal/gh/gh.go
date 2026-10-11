package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type Issue struct {
	Number int    `json:"number"`
	Body   string `json:"body"`
}

type PR struct {
	Number      int    `json:"number"`
	HeadRefName string `json:"headRefName"`
	IsDraft     bool   `json:"isDraft"`
	Body        string `json:"body"`
}

type Client struct{}

func (Client) OpenIssues(ctx context.Context, labels ...string) ([]Issue, error) {
	args := []string{"issue", "list", "--state", "open", "--limit", "200", "--json", "number,body"}
	for _, l := range labels {
		args = append(args, "--label", l)
	}
	var issues []Issue
	if err := decode(ctx, &issues, args); err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	return issues, nil
}

func (Client) OpenPRs(ctx context.Context, base string) ([]PR, error) {
	args := []string{"pr", "list", "--state", "open", "--base", base, "--limit", "200",
		"--json", "number,headRefName,isDraft,body"}
	var prs []PR
	if err := decode(ctx, &prs, args); err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	return prs, nil
}

func (Client) MergePR(ctx context.Context, number int) error {
	_, err := run(ctx, "pr", "merge", strconv.Itoa(number), "--squash", "--delete-branch")
	return err
}

func decode(ctx context.Context, v any, args []string) error {
	out, err := run(ctx, args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(out, v); err != nil {
		return fmt.Errorf("decode gh output: %w", err)
	}
	return nil
}

func run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("gh %v: %w: %s", args, err, bytes.TrimSpace(stderr.Bytes()))
	}
	return stdout.Bytes(), nil
}
