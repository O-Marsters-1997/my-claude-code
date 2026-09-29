package hook_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

func newRepo(t *testing.T, enabled bool) (dir string, s logstore.Store) {
	t.Helper()
	dir = t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	s = logstore.New(dir)
	if enabled {
		if err := s.On(); err != nil {
			t.Fatal(err)
		}
	}
	return dir, s
}

func fire(t *testing.T, dir string, p map[string]any) {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	hook.Run(strings.NewReader(string(b)), dir)
}

func readEvents(t *testing.T, s logstore.Store) []logstore.Event {
	t.Helper()
	events, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func TestRunDisabledWritesNothing(t *testing.T) {
	dir, s := newRepo(t, false)
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, use tabs"})
	if _, err := os.Stat(s.Dir()); !os.IsNotExist(err) {
		t.Errorf("disabled repo created %s", s.Dir())
	}
}

func TestRunLogsFailureAndCorrection(t *testing.T) {
	dir, s := newRepo(t, true)
	fire(t, dir, map[string]any{
		"hook_event_name": "PostToolUseFailure", "session_id": "s", "agent_id": "a1",
		"tool_name": "Read", "tool_use_id": "toolu_1", "tool_input": map[string]string{"file_path": "/nope"},
		"error": "File does not exist.", "transcript_path": "/t/s.jsonl",
	})
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, use tabs"})
	events := readEvents(t, s)
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2: %+v", len(events), events)
	}
	if fail := events[0]; fail.Class != detect.PathMissing || fail.AgentID != "a1" || fail.Transcript != "/t/s/subagents/agent-a1.jsonl" {
		t.Errorf("failure event = %+v", fail)
	}
	if events[1].Kind != "correction" || events[1].Conf < 0.6 {
		t.Errorf("correction event = %+v", events[1])
	}
	if _, err := os.Stat(filepath.Join(s.Dir(), ".gitignore")); err != nil {
		t.Error("log dir is not self-ignoring")
	}
}

func TestRunResolvesWorktreeToMainCheckout(t *testing.T) {
	main, s := newRepo(t, true)
	wt := t.TempDir()
	gitdir := filepath.Join(main, ".git", "worktrees", "wt")
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: "+gitdir+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fire(t, wt, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, use tabs"})
	if n := len(readEvents(t, s)); n != 1 {
		t.Errorf("worktree event landed outside the main checkout log: %d events", n)
	}
}

func TestStopSweepRecoversToolUseErrorsOnce(t *testing.T) {
	dir, s := newRepo(t, true)
	tp := filepath.Join(t.TempDir(), "s.jsonl")
	lines := []string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tu1","name":"Edit","input":{"file_path":"/a.go","old_string":"x"}}]}}`,
		fmt.Sprintf(`{"type":"user","timestamp":"2026-01-02T03:04:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"tu1","is_error":true,"content":%q}]}}`,
			"<tool_use_error>String to replace not found in file.\nString: x</tool_use_error>"),
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"tu2","is_error":true,"content":"Exit code 1"}]}}`,
	}
	if err := os.WriteFile(tp, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stop := map[string]any{"hook_event_name": "Stop", "session_id": "s", "transcript_path": tp}
	fire(t, dir, stop)
	fire(t, dir, stop)
	events := readEvents(t, s)
	if len(events) != 1 || events[0].Class != detect.EditMiss || events[0].ToolUseID != "tu1" {
		t.Errorf("sweep events = %+v, want one edit_miss for tu1", events)
	}
}

func TestSessionEndCountsSubagents(t *testing.T) {
	dir, s := newRepo(t, true)
	root := t.TempDir()
	tp := filepath.Join(root, "s.jsonl")
	use := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t","name":"Bash","input":{}}]}}` + "\n"
	prompt := `{"type":"user","message":{"content":"hello there"}}` + "\n"
	if err := os.WriteFile(tp, []byte(prompt+use), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "s", "subagents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "s", "subagents", "agent-x.jsonl"), []byte(use+use), 0o644); err != nil {
		t.Fatal(err)
	}
	fire(t, dir, map[string]any{"hook_event_name": "SessionEnd", "session_id": "s", "transcript_path": tp})
	events := readEvents(t, s)
	last := events[len(events)-1]
	if last.Kind != "session_end" || last.ToolCalls != 3 || last.Prompts != 1 {
		t.Errorf("session_end = %+v, want 3 tool calls and 1 prompt", last)
	}
}

func TestRunIgnoresGarbageInput(t *testing.T) {
	dir, _ := newRepo(t, true)
	hook.Run(strings.NewReader("not json"), dir)
}
