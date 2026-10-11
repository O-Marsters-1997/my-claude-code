package board_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/board"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
)

type fakeGH struct {
	deps.GH
	issues   []gh.Issue
	prs      []gh.PR
	external map[int]bool
	base     string
}

func (f *fakeGH) Issues(context.Context, string) ([]gh.Issue, error) { return f.issues, nil }

func (f *fakeGH) PRs(_ context.Context, base string) ([]gh.PR, error) {
	f.base = base
	return f.prs, nil
}

func (f *fakeGH) IssueClosed(_ context.Context, n int) (bool, error) { return f.external[n], nil }

func (f *fakeGH) MoveStatus(_ context.Context, n int, from, to string) error {
	i := f.find(n)
	var kept []gh.Label
	for _, l := range f.issues[i].Labels {
		if l.Name != from {
			kept = append(kept, l)
		}
	}
	f.issues[i].Labels = append(kept, gh.Label{Name: to})
	return nil
}

func (f *fakeGH) CloseIssue(_ context.Context, n int) error {
	f.issues[f.find(n)].State = "CLOSED"
	return nil
}

func (f *fakeGH) find(n int) int {
	for i, issue := range f.issues {
		if issue.Number == n {
			return i
		}
	}
	panic("no such issue")
}

func (f *fakeGH) status(n int) string {
	issue := f.issues[f.find(n)]
	for _, l := range issue.Labels {
		if strings.HasPrefix(l.Name, "status:") {
			return strings.TrimPrefix(l.Name, "status:") + "/" + strings.ToLower(issue.State)
		}
	}
	return "none/" + strings.ToLower(issue.State)
}

func issue(n int, status, blockedBy string) gh.Issue {
	body := "## Blocked by\n\n- None\n"
	if blockedBy != "" {
		body = "## Blocked by\n\n- Blocked by " + blockedBy + "\n"
	}
	i := gh.Issue{Number: n, Title: "ticket " + string(rune('A'+n%26)), Body: body, State: "OPEN",
		Labels: []gh.Label{{Name: "project:demo"}}}
	switch status {
	case "closed":
		i.State = "CLOSED"
		i.Labels = append(i.Labels, gh.Label{Name: "status:done"})
	case "":
	default:
		i.Labels = append(i.Labels, gh.Label{Name: "status:" + status})
	}
	return i
}

func pr(n int, state, branch, body string) gh.PR {
	return gh.PR{Number: n, State: state, HeadRefName: branch, Body: body, URL: "https://x/pull/" + string(rune('0'+n%10))}
}

type result struct {
	out, log, state string
}

func run(t *testing.T, f *fakeGH) result {
	t.Helper()
	var out, log bytes.Buffer
	statePath := filepath.Join(t.TempDir(), "state.md")
	d := deps.Deps{GH: f, Out: &out, Log: &log}
	if err := board.Run(context.Background(), d, board.Input{Label: "project:demo", StatePath: statePath}); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	state, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	return result{out: out.String(), log: log.String(), state: string(state)}
}

func TestMergedTicketIsClosedAndDependentPromoted(t *testing.T) {
	f := &fakeGH{
		issues: []gh.Issue{issue(1, "in-review", ""), issue(2, "backlog", "#1"), issue(3, "", "#1\n- Blocked by #2")},
		prs:    []gh.PR{pr(11, "MERGED", "issue-1/first", "Closes #1")},
	}
	res := run(t, f)

	got := []string{f.status(1), f.status(2), f.status(3)}
	want := []string{"done/closed", "ready/open", "none/open"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
	if want := "unfinished issue-1/first\n"; res.out != want {
		t.Errorf("stdout = %q, want %q", res.out, want)
	}
	if f.base != "feat/demo" {
		t.Errorf("PRs listed for base %q, want feat/demo", f.base)
	}
	for _, want := range []string{"Label: project:demo", "## #2 ", "Base: feat/demo", "Example: #11"} {
		if !strings.Contains(res.state, want) {
			t.Errorf("state.md lacks %q:\n%s", want, res.state)
		}
	}
}

func TestOpenAndMissingPRsAreLeftAlone(t *testing.T) {
	f := &fakeGH{
		issues: []gh.Issue{issue(1, "in-review", ""), issue(2, "in-progress", ""), issue(3, "backlog", "#1")},
		prs:    []gh.PR{pr(11, "OPEN", "issue-1/first", "Closes #1")},
	}
	res := run(t, f)

	got := []string{f.status(1), f.status(2), f.status(3)}
	want := []string{"in-review/open", "in-progress/open", "backlog/open"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
	if want := "unfinished\n"; res.out != want {
		t.Errorf("stdout = %q, want %q", res.out, want)
	}
	if !strings.Contains(res.log, "no-pr #2") {
		t.Errorf("log lacks no-pr #2:\n%s", res.log)
	}
}

func TestClosedUnmergedHoldsOnlyItsDependents(t *testing.T) {
	f := &fakeGH{
		issues: []gh.Issue{
			issue(1, "in-review", ""), issue(2, "in-review", ""),
			issue(3, "backlog", "#1"), issue(4, "backlog", "#2"),
		},
		prs: []gh.PR{
			pr(11, "CLOSED", "issue-1/first", "Closes #1"),
			pr(12, "MERGED", "issue-2/second", "Closes #2"),
		},
	}
	res := run(t, f)

	got := []string{f.status(1), f.status(2), f.status(3), f.status(4)}
	want := []string{"in-review/open", "done/closed", "backlog/open", "ready/open"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
	if !strings.Contains(res.log, "closed-unmerged #1") {
		t.Errorf("log lacks closed-unmerged #1:\n%s", res.log)
	}
}

func TestBundleAndBranchNameMatching(t *testing.T) {
	f := &fakeGH{
		issues: []gh.Issue{issue(5, "in-review", ""), issue(6, "in-review", ""), issue(7, "in-progress", "")},
		prs: []gh.PR{
			pr(20, "MERGED", "issue-5/tiny-bundle", "Closes #5\nCloses #6"),
			pr(21, "MERGED", "issue-7/by-branch", "no closing keyword"),
		},
	}
	res := run(t, f)

	got := []string{f.status(5), f.status(6), f.status(7)}
	want := []string{"done/closed", "done/closed", "done/closed"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
	if want := "finished issue-5/tiny-bundle issue-7/by-branch\n"; res.out != want {
		t.Errorf("stdout = %q, want %q", res.out, want)
	}
}

func TestBlockerOutsideLabelUsesItsOwnState(t *testing.T) {
	f := &fakeGH{
		issues:   []gh.Issue{issue(1, "backlog", "#90"), issue(2, "backlog", "#91")},
		external: map[int]bool{90: true},
	}
	run(t, f)

	got := []string{f.status(1), f.status(2)}
	want := []string{"ready/open", "backlog/open"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
}

func TestReadyTicketAppearsInStateWithoutBeingPromoted(t *testing.T) {
	f := &fakeGH{issues: []gh.Issue{issue(1, "ready", ""), issue(2, "closed", "")}}
	res := run(t, f)

	if !strings.Contains(res.state, "## #1 ") {
		t.Errorf("state.md lacks ready ticket #1:\n%s", res.state)
	}
	if strings.Contains(res.state, "## #2 ") {
		t.Errorf("state.md lists closed ticket #2:\n%s", res.state)
	}
}

type failingGH struct {
	deps.GH
}

func (failingGH) Issues(context.Context, string) ([]gh.Issue, error) {
	return nil, errBoom
}

var errBoom = errors.New("boom")

func TestIssueListingFailureIsReturned(t *testing.T) {
	d := deps.Deps{GH: failingGH{}, Out: &bytes.Buffer{}, Log: &bytes.Buffer{}}
	err := board.Run(context.Background(), d, board.Input{Label: "project:demo", StatePath: filepath.Join(t.TempDir(), "s.md")})
	if !errors.Is(err, errBoom) {
		t.Errorf("Run() = %v, want errBoom", err)
	}
}
