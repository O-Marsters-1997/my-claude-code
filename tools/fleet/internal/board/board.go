package board

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
)

const statusPrefix = "status:"

type Input struct {
	Label     string
	StatePath string
}

var closesRef = regexp.MustCompile(`(?i)\bcloses?\s+#(\d+)`)

func Run(ctx context.Context, d deps.Deps, in Input) error {
	tip := "feat/" + strings.TrimPrefix(in.Label, "project:")
	issues, err := d.GH.Issues(ctx, in.Label)
	if err != nil {
		return err
	}
	prs, err := d.GH.PRs(ctx, tip)
	if err != nil {
		return err
	}
	b := &board{ctx: ctx, d: d, issues: issues, prs: prs}

	branches, err := b.closeMerged()
	if err != nil {
		return err
	}
	if err := b.promote(); err != nil {
		return err
	}
	if err := os.WriteFile(in.StatePath, []byte(b.state(in.Label, tip)), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", in.StatePath, err)
	}

	verdict := "unfinished"
	if b.openCount() == 0 {
		verdict = "finished"
	}
	_, err = fmt.Fprintln(d.Out, strings.Join(append([]string{verdict}, branches...), " "))
	return err
}

type board struct {
	ctx    context.Context
	d      deps.Deps
	issues []gh.Issue
	prs    []gh.PR
}

func (b *board) closeMerged() ([]string, error) {
	var branches []string
	seen := map[string]bool{}
	for i := range b.issues {
		issue := &b.issues[i]
		status := statusOf(*issue)
		if issue.State == "CLOSED" || (status != "in-progress" && status != "in-review") {
			continue
		}
		matched := b.prsFor(issue.Number)
		switch outcome := outcomeOf(matched); outcome {
		case "merged":
			if err := b.d.GH.MoveStatus(b.ctx, issue.Number, statusPrefix+status, statusPrefix+"done"); err != nil {
				return nil, fmt.Errorf("move #%d to done: %w", issue.Number, err)
			}
			if err := b.d.GH.CloseIssue(b.ctx, issue.Number); err != nil {
				return nil, fmt.Errorf("close #%d: %w", issue.Number, err)
			}
			setStatus(issue, "done")
			issue.State = "CLOSED"
			fmt.Fprintf(b.d.Log, "closed #%d\n", issue.Number)
			for _, pr := range matched {
				if pr.State == "MERGED" && !seen[pr.HeadRefName] {
					seen[pr.HeadRefName] = true
					branches = append(branches, pr.HeadRefName)
				}
			}
		case "closed-unmerged", "no-pr":
			fmt.Fprintf(b.d.Log, "%s #%d\n", outcome, issue.Number)
		}
	}
	sort.Strings(branches)
	return branches, nil
}

func (b *board) promote() error {
	blockedBy := queue.Blockers(b.issues)
	for i := range b.issues {
		issue := &b.issues[i]
		if issue.State == "CLOSED" || statusOf(*issue) != "backlog" {
			continue
		}
		clear, err := b.allClosed(blockedBy[issue.Number])
		if err != nil {
			return err
		}
		if !clear {
			continue
		}
		from := ""
		if hasLabel(*issue, statusPrefix+"backlog") {
			from = statusPrefix + "backlog"
		}
		if err := b.d.GH.MoveStatus(b.ctx, issue.Number, from, statusPrefix+"ready"); err != nil {
			return fmt.Errorf("move #%d to ready: %w", issue.Number, err)
		}
		setStatus(issue, "ready")
		fmt.Fprintf(b.d.Log, "promoted #%d\n", issue.Number)
	}
	return nil
}

func (b *board) allClosed(blockers []int) (bool, error) {
	for _, n := range blockers {
		if i := b.index(n); i >= 0 {
			if b.issues[i].State != "CLOSED" {
				return false, nil
			}
			continue
		}
		closed, err := b.d.GH.IssueClosed(b.ctx, n)
		if err != nil {
			return false, err
		}
		if !closed {
			return false, nil
		}
	}
	return true, nil
}

func (b *board) state(label, tip string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Label: %s\n", label)
	blockedBy := queue.Blockers(b.issues)
	for _, issue := range b.issues {
		if issue.State == "CLOSED" || statusOf(issue) != "ready" {
			continue
		}
		fmt.Fprintf(&sb, "\n## #%d %s\n\nBase: %s\nExample: %s\n", issue.Number, issue.Title, tip, b.example(blockedBy[issue.Number]))
	}
	return sb.String()
}

func (b *board) example(blockers []int) string {
	best := gh.PR{}
	for _, n := range blockers {
		for _, pr := range b.prsFor(n) {
			if pr.State == "MERGED" && pr.Number > best.Number {
				best = pr
			}
		}
	}
	if best.Number == 0 {
		for _, pr := range b.prs {
			if pr.State == "MERGED" && pr.Number > best.Number {
				best = pr
			}
		}
	}
	if best.Number == 0 {
		return "none"
	}
	return fmt.Sprintf("#%d %s", best.Number, best.URL)
}

func (b *board) openCount() int {
	n := 0
	for _, issue := range b.issues {
		if issue.State != "CLOSED" {
			n++
		}
	}
	return n
}

func (b *board) index(number int) int {
	for i, issue := range b.issues {
		if issue.Number == number {
			return i
		}
	}
	return -1
}

func (b *board) prsFor(number int) []gh.PR {
	branchPrefix := fmt.Sprintf("issue-%d/", number)
	var out []gh.PR
	for _, pr := range b.prs {
		if strings.HasPrefix(pr.HeadRefName, branchPrefix) || closes(pr.Body, number) {
			out = append(out, pr)
		}
	}
	return out
}

func closes(body string, number int) bool {
	for _, m := range closesRef.FindAllStringSubmatch(body, -1) {
		if n, _ := strconv.Atoi(m[1]); n == number {
			return true
		}
	}
	return false
}

func outcomeOf(prs []gh.PR) string {
	outcome := "no-pr"
	for _, pr := range prs {
		switch pr.State {
		case "MERGED":
			return "merged"
		case "OPEN":
			outcome = "open"
		case "CLOSED":
			if outcome == "no-pr" {
				outcome = "closed-unmerged"
			}
		}
	}
	return outcome
}

func statusOf(issue gh.Issue) string {
	for _, l := range issue.Labels {
		if s, ok := strings.CutPrefix(l.Name, statusPrefix); ok {
			return s
		}
	}
	return "backlog"
}

func hasLabel(issue gh.Issue, name string) bool {
	for _, l := range issue.Labels {
		if l.Name == name {
			return true
		}
	}
	return false
}

func setStatus(issue *gh.Issue, status string) {
	kept := issue.Labels[:0:0]
	for _, l := range issue.Labels {
		if !strings.HasPrefix(l.Name, statusPrefix) {
			kept = append(kept, l)
		}
	}
	issue.Labels = append(kept, gh.Label{Name: statusPrefix + status})
}
