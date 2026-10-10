package ledger_test

import (
	"path/filepath"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
)

func TestMark(t *testing.T) {
	tests := []struct {
		name, id, status, issue, scope string
		wantErr                        bool
		wantStatus, wantIssue          string
	}{
		{name: "promoted with issue", id: "a", status: "promoted", issue: "https://x/1", wantStatus: "promoted", wantIssue: "https://x/1"},
		{name: "deferred", id: "a", status: "deferred", wantStatus: "deferred"},
		{name: "scope set", id: "a", status: "rejected", scope: "global", wantStatus: "rejected"},
		{name: "promoted without issue", id: "a", status: "promoted", wantErr: true},
		{name: "bad status", id: "a", status: "done", wantErr: true},
		{name: "bad scope", id: "a", status: "deferred", scope: "x", wantErr: true},
		{name: "unknown id", id: "zzz", status: "deferred", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "l.jsonl")
			if err := ledger.Append(path, ledger.Learning{ID: "a", Text: "keep", Status: "pending", Kind: "fix"}); err != nil {
				t.Fatal(err)
			}
			err := ledger.Mark(path, tt.id, tt.status, ledger.MarkFields{Issue: tt.issue, Scope: tt.scope})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := ledger.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0].Text != "keep" {
				t.Fatalf("fold lost fields: %+v", got)
			}
			want := tt.wantStatus
			if tt.wantErr {
				want = "pending"
			}
			if got[0].Status != want || got[0].Issue != tt.wantIssue {
				t.Errorf("got status=%q issue=%q", got[0].Status, got[0].Issue)
			}
			if tt.scope != "" && !tt.wantErr && got[0].Scope != tt.scope {
				t.Errorf("scope = %q", got[0].Scope)
			}
		})
	}
}

func TestMarkProposalKeepsIssue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "l.jsonl")
	if err := ledger.Append(path, ledger.Learning{ID: "a", Status: "pending"}); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Mark(path, "a", "promoted", ledger.MarkFields{Issue: "https://x/1"}); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Mark(path, "a", "promoted", ledger.MarkFields{Proposal: "https://x/2"}); err != nil {
		t.Fatal(err)
	}
	got, err := ledger.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Issue != "https://x/1" || got[0].Proposal != "https://x/2" {
		t.Errorf("issue=%q proposal=%q, want https://x/1 and https://x/2", got[0].Issue, got[0].Proposal)
	}
}
