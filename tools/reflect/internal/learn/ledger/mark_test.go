package ledger_test

import (
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"path/filepath"
	"testing"
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
			err := ledger.Mark(path, tt.id, tt.status, tt.issue, tt.scope, "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ledger.Read(path)
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
