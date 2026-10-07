package hook_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
)

func TestRunRecordsEachSessionOnce(t *testing.T) {
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "AGENTS.md"), []byte("be good"), 0o644); err != nil {
		t.Fatal(err)
	}
	recordDir := t.TempDir()
	now := time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC)
	start := `{"hook_event_name":"SessionStart","session_id":"s1","source":"startup"}`

	for _, payload := range []string{start, start, `{"hook_event_name":"Stop","session_id":"s2"}`} {
		if err := hook.Run(strings.NewReader(payload), project, recordDir, now); err != nil {
			t.Fatalf("Run(%s) error = %v", payload, err)
		}
	}

	got, err := record.Read(recordDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("Read() = %d records, want 1", len(got))
	}
	r := got[0]
	if r.SessionID != "s1" || r.Cwd != project || !r.TS.Equal(now) || r.InstrHash == "" {
		t.Errorf("record = %+v, want sid s1, cwd %s, ts %v and an instr hash", r, project, now)
	}
	if _, ok := r.Files[filepath.Join(project, "AGENTS.md")]; !ok {
		t.Errorf("record Files = %v, want AGENTS.md hashed", r.Files)
	}
}
