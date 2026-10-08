package scan_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/scan"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

func lines(t *testing.T, raws ...string) []transcript.Line {
	t.Helper()
	var out []transcript.Line
	for i, raw := range raws {
		l := transcript.Line{N: i + 1, Raw: []byte(raw)}
		if err := json.Unmarshal(l.Raw, &l.Entry); err != nil {
			t.Fatal(err)
		}
		out = append(out, l)
	}
	return out
}

func testSession(t *testing.T) session.Session {
	t.Helper()
	main := &session.Agent{ID: session.MainID, Type: session.MainID, Lines: lines(t,
		`{"type":"user","message":{"content":"export GITHUB_TOKEN=ghp_abcdefghijklmnopqrstuvwxyz0123456789AB please"}}`,
		`{"type":"mode","mode":"default"}`,
		`{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/x.go"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"File does not exist.","is_error":true}]}}`,
	)}
	sub := &session.Agent{ID: "sub1", Type: "Explore", Parent: session.MainID, Link: session.LinkToolUseID, Depth: 1}
	return session.Session{ID: "s1", Agents: []*session.Agent{main, sub}}
}

func TestWriteDigestsAndIndex(t *testing.T) {
	dir := t.TempDir()

	index, err := scan.Write(testSession(t), dir, 8)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	for _, id := range []string{session.MainID, "sub1"} {
		if _, err := os.Stat(filepath.Join(dir, id+".txt")); err != nil {
			t.Errorf("digest for %s missing: %v", id, err)
		}
	}
	for _, want := range []string{"session s1: 2 agents", "halluc:1", "review", "skip"} {
		if !strings.Contains(index, want) {
			t.Errorf("Write() index lacks %q:\n%s", want, index)
		}
	}
}

func TestSliceRedactsAndSkipsHarnessLines(t *testing.T) {
	out, err := scan.Slice(testSession(t), session.MainID, 2, 1)
	if err != nil {
		t.Fatalf("Slice() error = %v", err)
	}

	if strings.Contains(out, "ghp_") || !strings.Contains(out, "[REDACTED:github-token]") {
		t.Errorf("Slice() did not redact the token:\n%s", out)
	}
	if strings.Contains(out, "L2 ") {
		t.Errorf("Slice() includes the harness line L2:\n%s", out)
	}
	if !strings.Contains(out, "L3 assistant tool_use Read t1") {
		t.Errorf("Slice() lacks the tool call on L3:\n%s", out)
	}
}

func TestSliceUnknownAgent(t *testing.T) {
	if _, err := scan.Slice(testSession(t), "nope", 1, 1); err == nil {
		t.Error("Slice(unknown agent) error = nil, want an error")
	}
}

func TestWriteListsEachTouchedRepoOnce(t *testing.T) {
	root := t.TempDir()
	main := filepath.Join(root, "app")
	worktree := filepath.Join(root, "app-wt")
	plain := filepath.Join(root, "notes")
	for _, d := range []string{filepath.Join(main, ".git", "worktrees", "wt"), filepath.Join(main, "sub"), worktree, plain} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitFile := "gitdir: " + filepath.Join(main, ".git", "worktrees", "wt")
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte(gitFile), 0o644); err != nil {
		t.Fatal(err)
	}
	cwdLine := func(dir string) string {
		b, _ := json.Marshal(map[string]string{"type": "user", "cwd": dir})
		return string(b)
	}
	s := session.Session{ID: "s1", Agents: []*session.Agent{
		{ID: session.MainID, Type: session.MainID, Lines: lines(t, cwdLine(main), cwdLine(filepath.Join(main, "sub")), cwdLine(plain))},
		{ID: "sub1", Type: "claude", Parent: session.MainID, Depth: 1, Lines: lines(t, cwdLine(worktree))},
	}}

	index, err := scan.Write(s, t.TempDir(), 8)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if got := strings.Count(index, "repo "+main+"\n"); got != 1 {
		t.Errorf("Write() index has %d %q lines, want 1:\n%s", got, "repo "+main, index)
	}
	if strings.Contains(index, "repo "+plain) {
		t.Errorf("Write() index lists the non-repo %s:\n%s", plain, index)
	}
}
