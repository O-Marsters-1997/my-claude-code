package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type Issue struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	State  string  `json:"state"`
	Labels []Label `json:"labels"`
}

type PR struct {
	Number      int       `json:"number"`
	HeadRefName string    `json:"headRefName"`
	IsDraft     bool      `json:"isDraft"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Client struct{}

func (Client) OpenIssues(ctx context.Context, labels ...string) ([]Issue, error) {
	args := []string{"issue", "list", "--state", "open", "--limit", "200", "--json", "number,title,body,state,labels"}
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
		"--json", "number,headRefName,isDraft,body,createdAt"}
	var prs []PR
	if err := decode(ctx, &prs, args); err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	return prs, nil
}

func (Client) Issues(ctx context.Context, label string) ([]Issue, error) {
	args := []string{"issue", "list", "--state", "all", "--label", label, "--limit", "200",
		"--json", "number,title,body,state,labels"}
	var issues []Issue
	if err := decode(ctx, &issues, args); err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	return issues, nil
}

func (Client) PRs(ctx context.Context, base string) ([]PR, error) {
	args := []string{"pr", "list", "--state", "all", "--base", base, "--limit", "200",
		"--json", "number,headRefName,isDraft,body,state,url"}
	var prs []PR
	if err := decode(ctx, &prs, args); err != nil {
		return nil, fmt.Errorf("list pull requests: %w", err)
	}
	return prs, nil
}

func (Client) IssueClosed(ctx context.Context, number int) (bool, error) {
	var issue Issue
	if err := decode(ctx, &issue, []string{"issue", "view", strconv.Itoa(number), "--json", "state"}); err != nil {
		return false, fmt.Errorf("view issue #%d: %w", number, err)
	}
	return issue.State == "CLOSED", nil
}

func (Client) MoveStatus(ctx context.Context, number int, from, to string) error {
	args := []string{"issue", "edit", strconv.Itoa(number), "--add-label", to}
	if from != "" {
		args = append(args, "--remove-label", from)
	}
	_, err := run(ctx, args...)
	return err
}

func (Client) CloseIssue(ctx context.Context, number int) error {
	_, err := run(ctx, "issue", "close", strconv.Itoa(number))
	return err
}

func (Client) MergePR(ctx context.Context, number int) error {
	_, err := run(ctx, "pr", "merge", strconv.Itoa(number), "--squash", "--delete-branch")
	return err
}

func (Client) CommentPR(ctx context.Context, number int, body string) error {
	_, err := run(ctx, "pr", "comment", strconv.Itoa(number), "--body", body)
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
