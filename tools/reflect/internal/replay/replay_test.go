package replay_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/replay"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

func lines(t *testing.T, raws ...string) []transcript.Line {
	t.Helper()
	var out []transcript.Line
	for i, raw := range raws {
		l := transcript.Line{N: i + 1, Raw: []byte(raw + "\n")}
		if err := json.Unmarshal(l.Raw, &l.Entry); err != nil {
			t.Fatal(err)
		}
		out = append(out, l)
	}
	return out
}

func fixture(t *testing.T) (session.Session, string) {
	t.Helper()
	home := t.TempDir()
	hook := filepath.Join(home, "gate.sh")
	script := "#!/bin/bash\nin=$(cat)\ncase \"$in\" in *'\"cwd\":\"/blocked\"'*) echo nope >&2; exit 2;; esac\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	a := &session.Agent{ID: "a1", Lines: lines(t,
		`{"type":"assistant","cwd":"/blocked","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"grep x"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","is_error":true,"content":"PreToolUse:Bash hook error: [`+hook+`]: nope"}]}}`,
	)}
	return session.Session{ID: "s", Agents: []*session.Agent{a}}, home
}

func TestRunReportsHookExitForRecordedAndOverriddenCwd(t *testing.T) {
	s, home := fixture(t)

	recorded, err := replay.Run(s, "a1", 1, replay.Options{Home: home})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	moved, err := replay.Run(s, "a1", 1, replay.Options{Home: home, Cwd: "/elsewhere"})
	if err != nil {
		t.Fatalf("Run(--cwd) error = %v", err)
	}

	if !strings.Contains(recorded, "exit: 2") || !strings.Contains(recorded, "nope") {
		t.Errorf("Run() = %q, want exit 2 with the hook message", recorded)
	}
	if !strings.Contains(moved, "exit: 0") {
		t.Errorf("Run(--cwd) = %q, want exit 0", moved)
	}
}

func TestRunRefusesHookOutsideHome(t *testing.T) {
	s, _ := fixture(t)

	if _, err := replay.Run(s, "a1", 1, replay.Options{Home: t.TempDir()}); err == nil {
		t.Error("Run() error = nil, want a refusal for a hook outside home")
	}
}
