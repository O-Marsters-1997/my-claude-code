package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
)

func TestLearnStatus(t *testing.T) {
	tests := []struct {
		name    string
		pending int
		done    int
		remind  bool
		want    string
	}{
		{"empty count", 0, 0, false, "0\n"},
		{"counts pending only", 3, 2, false, "3\n"},
		{"no reminder below threshold", 9, 0, true, ""},
		{"reminder at threshold", 10, 1, true, "10 pending learnings: run /triage-learnings\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "learnings.jsonl")
			for i := 0; i < tt.pending+tt.done; i++ {
				status := "pending"
				if i >= tt.pending {
					status = "promoted"
				}
				id := strings.Repeat("a", i+1)
				if err := ledger.Append(path, ledger.Learning{ID: id, Status: status}); err != nil {
					t.Fatal(err)
				}
			}
			got, err := learnStatus(path, tt.remind)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}
