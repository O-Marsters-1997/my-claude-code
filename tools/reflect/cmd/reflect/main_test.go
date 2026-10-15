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
		{"reminder at threshold", 10, 1, true, "10 pending reflect markers: run /triage-reflections\n"},
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

func TestAddObservation(t *testing.T) {
	tests := []struct {
		name, area, wantScope string
	}{
		{"area", "tools/reflect/", "area:tools/reflect"},
		{"whole repo", "", "repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "learnings.jsonl")
			if err := addObservation(path, "/r", tt.area, "package is too big"); err != nil {
				t.Fatal(err)
			}
			got, err := ledger.Read(path)
			if err != nil || len(got) != 1 {
				t.Fatalf("read = %v, %v", got, err)
			}
			l := got[0]
			if l.Scope != tt.wantScope || l.Text != "package is too big" || l.Status != "pending" {
				t.Errorf("record = %+v", l)
			}
			if l.File != "" || l.Line != 0 || l.TargetText != "" || l.After != "" {
				t.Errorf("code fields set: %+v", l)
			}
			if loc := location(l); loc != tt.wantScope {
				t.Errorf("location = %q", loc)
			}
		})
	}
	if err := addObservation(filepath.Join(t.TempDir(), "l.jsonl"), "/r", "", "  "); err == nil {
		t.Error("empty text accepted")
	}
}
