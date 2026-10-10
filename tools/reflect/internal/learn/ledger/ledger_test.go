package ledger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
)

func TestReadFoldsLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "learnings.jsonl")
	for _, l := range []ledger.Learning{
		{ID: "a", Kind: "fix", Text: "one", Status: "pending"},
		{ID: "b", Kind: "later", Status: "pending"},
		{ID: "a", Status: "promoted", Issue: "u"},
	} {
		if err := ledger.Append(path, l); err != nil {
			t.Fatal(err)
		}
	}
	want := []ledger.Learning{
		{ID: "a", Kind: "fix", Text: "one", Status: "promoted", Issue: "u"},
		{ID: "b", Kind: "later", Status: "pending"},
	}
	got, err := ledger.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Read mismatch (-want +got):\n%s", diff)
	}
}

func TestReadMissingIsEmpty(t *testing.T) {
	got, err := ledger.Read(filepath.Join(t.TempDir(), "none.jsonl"))
	if err != nil || len(got) != 0 {
		t.Errorf("Read(missing) = %v, %v, want empty", got, err)
	}
}

func TestReadRejectsGarbage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "l.jsonl")
	if err := os.WriteFile(path, []byte("nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Read(path); err == nil {
		t.Error("Read(garbage) error = nil, want error")
	}
}

func TestDefaultPathHonoursXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/x")
	if got, want := ledger.DefaultPath(), "/x/learn/learnings.jsonl"; got != want {
		t.Errorf("DefaultPath() = %q, want %q", got, want)
	}
}
