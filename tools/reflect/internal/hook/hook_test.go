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

func newRepo(t *testing.T) (dir string, s logstore.Store) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	dir = t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, logstore.New(dir)
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

func startSession(t *testing.T, dir, sid string, extra map[string]any) {
	t.Helper()
	p := map[string]any{"hook_event_name": "SessionStart", "session_id": sid, "cwd": dir, "transcript_path": "/t/" + sid + ".jsonl", "source": "startup"}
	for k, v := range extra {
		p[k] = v
	}
	fire(t, dir, p)
}

func rawLines(t *testing.T, s logstore.Store) []string {
	t.Helper()
	b, err := os.ReadFile(s.LogPath())
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

func rawEvents(t *testing.T, s logstore.Store) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range rawLines(t, s) {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatal(err)
		}
		out = append(out, m)
	}
	return out
}

func kinds(events []logstore.Event, kind string) []logstore.Event {
	var out []logstore.Event
	for _, e := range events {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

func failure(sid, toolUseID, path, errText string) map[string]any {
	return map[string]any{
		"hook_event_name": "PostToolUseFailure", "session_id": sid, "tool_name": "Read", "tool_use_id": toolUseID,
		"tool_input": map[string]string{"file_path": path}, "error": errText,
	}
}

func TestRunLogsFailureAndCorrection(t *testing.T) {
	dir, s := newRepo(t)
	startSession(t, dir, "s", nil)
	fail := failure("s", "toolu_1", "/nope", "File does not exist.")
	fail["agent_id"] = "a1"
	fire(t, dir, fail)
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, use tabs"})
	events := readEvents(t, s)
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3: %+v", len(events), events)
	}
	if fail := events[1]; fail.Class != detect.PathMissing || fail.AgentID != "a1" || fail.Transcript != "/t/s/subagents/agent-a1.jsonl" {
		t.Errorf("failure event = %+v", fail)
	}
	if events[2].Kind != "correction" || events[2].Conf < 0.6 || events[2].Transcript != "/t/s.jsonl" {
		t.Errorf("correction event = %+v", events[2])
	}
	if _, err := os.Stat(filepath.Join(s.Dir(), ".gitignore")); err != nil {
		t.Error("log dir is not self-ignoring")
	}
}

func TestRunResolvesWorktreeToMainCheckout(t *testing.T) {
	main, s := newRepo(t)
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
	dir, s := newRepo(t)
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
	dir, s := newRepo(t)
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
	dir, _ := newRepo(t)
	hook.Run(strings.NewReader("not json"), dir)
}

func TestOnlySessionEventsCarryCwdAndTranscript(t *testing.T) {
	dir, s := newRepo(t)
	startSession(t, dir, "s", nil)
	fire(t, dir, failure("s", "toolu_1", "/nope", "File does not exist."))
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, use tabs"})
	fire(t, dir, map[string]any{"hook_event_name": "SessionEnd", "session_id": "s", "transcript_path": "/t/s.jsonl"})
	for _, m := range rawEvents(t, s) {
		_, hasCwd := m["cwd"]
		_, hasTranscript := m["transcript"]
		if isSession := m["kind"] == "session"; hasCwd != isSession || hasTranscript != isSession {
			t.Errorf("%v event: cwd=%v transcript=%v", m["kind"], hasCwd, hasTranscript)
		}
	}
}

func TestEveryEventKindStaysUnderLineBudget(t *testing.T) {
	const budget = 2048
	dir, s := newRepo(t)
	fat := strings.Repeat("x", 5000)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("rules"), 0o644); err != nil {
		t.Fatal(err)
	}
	startSession(t, dir, "s", nil)
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "rules v2")
	startSession(t, dir, "s", nil)
	bad := failure("s", "toolu_1", "/"+fat, fat)
	bad["agent_id"] = "agent-with-a-long-id-0123456789"
	fire(t, dir, bad)
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "prompt": "no, " + strings.Repeat("x", 490)})
	for range 2 {
		fire(t, dir, map[string]any{
			"hook_event_name": "PostToolUse", "session_id": "s", "tool_name": "Edit", "tool_use_id": "toolu_2",
			"tool_input": map[string]string{"file_path": "/" + strings.Repeat("d/", 40) + "f.go", "old_string": fat, "new_string": fat},
		})
	}
	fire(t, dir, map[string]any{"hook_event_name": "SessionEnd", "session_id": "s", "transcript_path": "/t/s.jsonl"})
	seen := map[string]bool{}
	for _, line := range rawLines(t, s) {
		var e logstore.Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		seen[e.Kind] = true
		if len(line) > budget {
			t.Errorf("%s line is %d bytes, budget %d", e.Kind, len(line), budget)
		}
	}
	for _, kind := range []string{"session", "manifest", "tool_error", "correction", "edit", "session_end"} {
		if !seen[kind] {
			t.Errorf("no %s event was exercised", kind)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCorrectionRecordsTheToolCallBeforeIt(t *testing.T) {
	dir, s := newRepo(t)
	tr := filepath.Join(t.TempDir(), "s.jsonl")
	writeFile(t, tr, strings.Join([]string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tu1","name":"Read","input":{}}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tu2","name":"Edit","input":{}}]}}`,
		`{"type":"user","message":{"content":"no, use tabs"}}`,
	}, "\n")+"\n")
	startSession(t, dir, "s", map[string]any{"transcript_path": tr})
	fire(t, dir, map[string]any{"hook_event_name": "UserPromptSubmit", "session_id": "s", "transcript_path": tr, "prompt": "no, use tabs"})
	got := kinds(readEvents(t, s), "correction")
	if len(got) != 1 || got[0].ToolUseID != "tu2" || got[0].Tool != "Edit" {
		t.Errorf("correction = %+v, want tool_use_id tu2 from Edit", got)
	}
}

func TestManifestIsWrittenOnlyWhenInstructionsChange(t *testing.T) {
	dir, s := newRepo(t)
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "v1")
	startSession(t, dir, "s1", nil)
	startSession(t, dir, "s2", nil)
	events := readEvents(t, s)
	if manifests := kinds(events, "manifest"); len(manifests) != 0 {
		t.Fatalf("baseline wrote %d manifests, want none", len(manifests))
	}
	baseline := kinds(events, "session")[0].InstrHash

	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "v2")
	startSession(t, dir, "s3", nil)
	manifests := kinds(readEvents(t, s), "manifest")
	if len(manifests) != 1 {
		t.Fatalf("got %d manifests after an edit, want 1", len(manifests))
	}
	second := manifests[0]
	if second.Prev != baseline || len(second.Changed) != 1 || len(second.Added) != 0 || len(second.Removed) != 0 {
		t.Errorf("manifest = %+v, want one changed file and prev %s", second, baseline)
	}

	if err := os.Remove(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	startSession(t, dir, "s4", nil)
	manifests = kinds(readEvents(t, s), "manifest")
	if last := manifests[len(manifests)-1]; len(last.Removed) != 1 {
		t.Errorf("manifest after deleting the file = %+v, want one removed", last)
	}
}

func TestBaselineWithManyInstructionFilesWritesNoManifest(t *testing.T) {
	dir, s := newRepo(t)
	for i := range 50 {
		writeFile(t, filepath.Join(dir, ".claude", "skills", fmt.Sprintf("s%d", i), "SKILL.md"), "x")
	}
	startSession(t, dir, "s", nil)
	if n := len(kinds(readEvents(t, s), "manifest")); n != 0 {
		t.Errorf("first session wrote %d manifests, want none", n)
	}
}

func TestSessionEventIsSmallAndHasNoFileList(t *testing.T) {
	dir, s := newRepo(t)
	for i := range 20 {
		writeFile(t, filepath.Join(dir, ".claude", "skills", fmt.Sprintf("s%d", i), "SKILL.md"), "x")
	}
	startSession(t, dir, "s", nil)
	lines := rawLines(t, s)
	for i, m := range rawEvents(t, s) {
		if m["kind"] != "session" {
			continue
		}
		if _, ok := m["files"]; ok || len(lines[i]) > 500 {
			t.Errorf("session line is %d bytes, files key present=%v", len(lines[i]), ok)
		}
	}
}

func TestMirroredSkillsCountOnce(t *testing.T) {
	dir, s := newRepo(t)
	writeFile(t, filepath.Join(dir, ".agents", "skills", "x", "SKILL.md"), "one")
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, ".agents", "skills", "x"), filepath.Join(dir, ".claude", "skills", "x")); err != nil {
		t.Fatal(err)
	}
	startSession(t, dir, "s", nil)
	writeFile(t, filepath.Join(dir, ".agents", "skills", "x", "SKILL.md"), "two")
	startSession(t, dir, "s2", nil)
	manifests := kinds(readEvents(t, s), "manifest")
	if len(manifests) != 1 || len(manifests[0].Changed) != 1 {
		t.Errorf("manifests = %+v, want the mirrored skill changed once", manifests)
	}
}

func TestCompactIsSkippedOnlyWhenNothingChanged(t *testing.T) {
	dir, s := newRepo(t)
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "v1")
	startSession(t, dir, "s", nil)
	startSession(t, dir, "s", map[string]any{"source": "compact"})
	if n := len(kinds(readEvents(t, s), "session")); n != 1 {
		t.Errorf("unchanged compact wrote a session event: %d session events", n)
	}
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "v2")
	startSession(t, dir, "s", map[string]any{"source": "compact"})
	if n := len(kinds(readEvents(t, s), "session")); n != 2 {
		t.Errorf("compact after an instruction change was skipped: %d session events", n)
	}
	startSession(t, dir, "other", map[string]any{"source": "clear"})
	if n := len(kinds(readEvents(t, s), "session")); n != 3 {
		t.Errorf("clear was skipped: %d session events", n)
	}
}

func TestSessionEndCountsErrors(t *testing.T) {
	dir, s := newRepo(t)
	startSession(t, dir, "s", nil)
	fire(t, dir, failure("s", "t1", "/a", "File does not exist."))
	fire(t, dir, failure("s", "t2", "/b", "File does not exist."))
	fire(t, dir, failure("other", "t3", "/c", "File does not exist."))
	fire(t, dir, map[string]any{"hook_event_name": "SessionEnd", "session_id": "s", "transcript_path": "/t/s.jsonl"})
	end := kinds(readEvents(t, s), "session_end")
	if len(end) != 1 || end[0].Errors != 2 {
		t.Errorf("session_end = %+v, want errors=2", end)
	}
}

func TestFailureSeenByHookAndSweepIsLoggedOnce(t *testing.T) {
	dir, s := newRepo(t)
	tp := filepath.Join(t.TempDir(), "s.jsonl")
	lines := []string{
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tu1","name":"Read","input":{"file_path":"/nope"}}]}}`,
		fmt.Sprintf(`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"tu1","is_error":true,"content":%q}]}}`,
			"<tool_use_error>File does not exist.</tool_use_error>"),
	}
	writeFile(t, tp, strings.Join(lines, "\n")+"\n")
	fire(t, dir, failure("s", "tu1", "/nope", "File does not exist."))
	fire(t, dir, map[string]any{"hook_event_name": "Stop", "session_id": "s", "transcript_path": tp})
	if n := len(kinds(readEvents(t, s), "tool_error")); n != 1 {
		t.Errorf("hook and sweep produced %d tool_error events for one failure, want 1", n)
	}
}

func TestToolErrorPayloadIsTrimmedAndGoesBareAfterThree(t *testing.T) {
	dir, s := newRepo(t)
	long := strings.Repeat("a", 1000) + "THE-END"
	for i := range 4 {
		fire(t, dir, failure("s", fmt.Sprintf("t%d", i), "/nope", long))
	}
	errs := kinds(readEvents(t, s), "tool_error")
	if len(errs) != 4 {
		t.Fatalf("got %d tool_error events, want 4", len(errs))
	}
	first := errs[0]
	if len(first.Input) > 150+len("…") {
		t.Errorf("input is %d bytes, want at most %d", len(first.Input), 150+len("…"))
	}
	if !strings.HasSuffix(first.Error, "THE-END") || len(first.Error) > 300+len("…") {
		t.Errorf("error is %d bytes ending %q, want the last 300 bytes", len(first.Error), first.Error[max(0, len(first.Error)-8):])
	}
	for i, e := range errs[3:] {
		if e.Input != "" || e.Error != "" || e.FP != first.FP || e.Class != first.Class {
			t.Errorf("event %d after the third = %+v, want bare with the same fp and class", i+4, e)
		}
	}
	if errs[2].Error == "" {
		t.Error("third event was already bare")
	}
}

func editPayload(sid, file, old, new string) map[string]any {
	return map[string]any{
		"hook_event_name": "PostToolUse", "session_id": sid, "tool_name": "Edit", "tool_use_id": "tu-" + new,
		"tool_input": map[string]string{"file_path": file, "old_string": old, "new_string": new},
	}
}

func TestEditIsLoggedFromTheSecondEditToAFile(t *testing.T) {
	dir, s := newRepo(t)
	fire(t, dir, editPayload("s", "/a.go", "1", "2"))
	fire(t, dir, editPayload("s", "/b.go", "1", "2"))
	if n := len(kinds(readEvents(t, s), "edit")); n != 0 {
		t.Fatalf("single edits were logged: %d edit events", n)
	}
	fire(t, dir, editPayload("s", "/a.go", "2", "3"))
	edits := kinds(readEvents(t, s), "edit")
	if len(edits) != 2 || edits[0].NewHash == edits[1].NewHash || edits[0].File != "/a.go" {
		t.Fatalf("after a second edit to /a.go got %+v, want both edits of /a.go", edits)
	}
	if edits[0].TS.After(edits[1].TS) {
		t.Errorf("edit timestamps out of order: %v then %v", edits[0].TS, edits[1].TS)
	}
	fire(t, dir, editPayload("s", "/a.go", "3", "4"))
	if n := len(kinds(readEvents(t, s), "edit")); n != 3 {
		t.Errorf("third edit to /a.go: %d edit events, want 3", n)
	}
}

func TestEditsAreTrackedPerSessionAndCleanedAtEnd(t *testing.T) {
	dir, s := newRepo(t)
	fire(t, dir, editPayload("s1", "/a.go", "1", "2"))
	fire(t, dir, editPayload("s2", "/a.go", "2", "3"))
	if n := len(kinds(readEvents(t, s), "edit")); n != 0 {
		t.Errorf("edits from different sessions were paired: %d edit events", n)
	}
	fire(t, dir, map[string]any{"hook_event_name": "SessionEnd", "session_id": "s1", "transcript_path": "/t/s1.jsonl"})
	if _, err := os.Stat(filepath.Join(s.Dir(), "edits", "s1")); !os.IsNotExist(err) {
		t.Errorf("edit state for an ended session was kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir(), "edits", "s2")); err != nil {
		t.Errorf("edit state for a live session was removed: %v", err)
	}
}

func TestCompactIsNotConfusedByAnotherSessionStartingInBetween(t *testing.T) {
	dir, s := newRepo(t)
	writeFile(t, filepath.Join(dir, "CLAUDE.md"), "v1")
	startSession(t, dir, "a", nil)
	startSession(t, dir, "b", nil)
	startSession(t, dir, "a", map[string]any{"source": "compact"})
	if n := len(kinds(readEvents(t, s), "session")); n != 2 {
		t.Errorf("compact of session a was logged again after b started: %d session events, want 2", n)
	}
}
