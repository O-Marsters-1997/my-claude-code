package queue_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
)

func pr(number int, draft bool, closes ...int) gh.PR {
	body := "Summary\n"
	for _, c := range closes {
		body += fmt.Sprintf("Closes #%d\n", c)
	}
	return gh.PR{Number: number, IsDraft: draft, Body: body}
}

func blockedBy(number int, blockers ...int) gh.Issue {
	body := "## What to build\nsee #999\n\n## Blocked by\n\n"
	if len(blockers) == 0 {
		body += "None\n"
	}
	for _, b := range blockers {
		body += fmt.Sprintf("- Blocked by #%d\n", b)
	}
	return gh.Issue{Number: number, Body: body + "\n## User stories\n- #888\n"}
}

func numbers(entries []queue.Entry) []int {
	var out []int
	for _, e := range entries {
		out = append(out, e.PR.Number)
	}
	return out
}

func TestBuildKeepsDrafts(t *testing.T) {
	prs := []gh.PR{pr(30, false, 3), pr(20, true, 2), pr(10, false, 1)}
	if diff := cmp.Diff([]int{10, 20, 30}, numbers(queue.Build(prs, nil))); diff != "" {
		t.Errorf("Build() (-want +got):\n%s", diff)
	}
}

func TestBuildOrder(t *testing.T) {
	tests := []struct {
		name   string
		prs    []gh.PR
		issues []gh.Issue
		want   []int
	}{
		{
			name: "lowest issue number first",
			prs:  []gh.PR{pr(50, false, 5), pr(70, false, 3), pr(60, false, 4)},
			want: []int{70, 60, 50},
		},
		{
			name:   "blocker lands before the dependent despite a higher number",
			prs:    []gh.PR{pr(51, false, 1), pr(52, false, 2), pr(53, false, 3)},
			issues: []gh.Issue{blockedBy(1, 3), blockedBy(2), blockedBy(3)},
			want:   []int{52, 53, 51},
		},
		{
			name:   "blocker outside the queue is ignored",
			prs:    []gh.PR{pr(51, false, 1), pr(52, false, 2)},
			issues: []gh.Issue{blockedBy(1, 99)},
			want:   []int{51, 52},
		},
		{
			name: "bundle sorts by its lowest number",
			prs:  []gh.PR{pr(60, false, 4), pr(61, false, 9, 2, 7), pr(62, false, 3)},
			want: []int{61, 62, 60},
		},
		{
			name:   "bundle blocked by a later ticket waits for it",
			prs:    []gh.PR{pr(60, false, 2, 3), pr(61, false, 5)},
			issues: []gh.Issue{blockedBy(3, 5)},
			want:   []int{61, 60},
		},
		{
			name:   "cycle falls back to issue order",
			prs:    []gh.PR{pr(51, false, 1), pr(52, false, 2)},
			issues: []gh.Issue{blockedBy(1, 2), blockedBy(2, 1)},
			want:   []int{51, 52},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready := queue.Build(tt.prs, tt.issues)
			if diff := cmp.Diff(tt.want, numbers(ready)); diff != "" {
				t.Errorf("Build() order (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildBundleTickets(t *testing.T) {
	ready := queue.Build([]gh.PR{pr(61, false, 9, 2, 7)}, nil)
	if diff := cmp.Diff([]int{2, 7, 9}, ready[0].Tickets); diff != "" {
		t.Errorf("Tickets (-want +got):\n%s", diff)
	}
}
