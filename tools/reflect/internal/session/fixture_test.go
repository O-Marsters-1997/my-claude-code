package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	t        *testing.T
	projects string
	slug     string
	sid      string
	n        int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	return &fixture{t: t, projects: t.TempDir(), slug: "-repo", sid: "11111111-2222-3333-4444-555555555555"}
}

func (f *fixture) mainPath() string {
	return filepath.Join(f.projects, f.slug, f.sid+".jsonl")
}

func (f *fixture) agentPath(aid string) string {
	return filepath.Join(f.projects, f.slug, f.sid, "subagents", "agent-"+aid+".jsonl")
}

func (f *fixture) write(path string, lines ...map[string]any) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	for _, l := range lines {
		if err := enc.Encode(l); err != nil {
			f.t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) meta(aid string, m map[string]any) {
	f.t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		f.t.Fatal(err)
	}
	path := strings.TrimSuffix(f.agentPath(aid), ".jsonl") + ".meta.json"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) ts() string {
	f.n++
	return fmt.Sprintf("2026-10-07T10:%02d:%02d.000Z", f.n/60, f.n%60)
}

func (f *fixture) prompt(text string) map[string]any {
	return map[string]any{"type": "user", "timestamp": f.ts(), "message": map[string]any{"role": "user", "content": text}}
}

func (f *fixture) say(text string) map[string]any {
	return map[string]any{"type": "assistant", "timestamp": f.ts(), "message": map[string]any{
		"model": "claude-opus-5-5", "content": []any{map[string]any{"type": "text", "text": text}},
		"usage": map[string]any{"input_tokens": 100, "output_tokens": 10},
	}}
}

func (f *fixture) use(id, name string, input map[string]any) map[string]any {
	return map[string]any{"type": "assistant", "timestamp": f.ts(), "message": map[string]any{
		"model":   "claude-opus-5-5",
		"content": []any{map[string]any{"type": "tool_use", "id": id, "name": name, "input": input}},
		"usage":   map[string]any{"input_tokens": 1000, "output_tokens": 50, "cache_read_input_tokens": 500},
	}}
}

func (f *fixture) result(id, text string, isError bool, toolUseResult any) map[string]any {
	l := map[string]any{"type": "user", "timestamp": f.ts(), "message": map[string]any{"role": "user", "content": []any{
		map[string]any{"type": "tool_result", "tool_use_id": id, "content": text, "is_error": isError},
	}}}
	if toolUseResult != nil {
		l["toolUseResult"] = toolUseResult
	}
	return l
}

func (f *fixture) notify(aid, toolUseID, status string) map[string]any {
	return map[string]any{"type": "queue-operation", "operation": "enqueue", "timestamp": f.ts(), "content": fmt.Sprintf(
		"<task-notification>\n<task-id>%s</task-id>\n<tool-use-id>%s</tool-use-id>\n<status>%s</status>\n</task-notification>", aid, toolUseID, status)}
}

func (f *fixture) command(name string) map[string]any {
	return f.prompt("<command-message>" + name + "</command-message>\n<command-name>/" + name + "</command-name>")
}

func (f *fixture) compact() map[string]any {
	return map[string]any{"type": "system", "subtype": "compact_boundary", "timestamp": f.ts(),
		"compactMetadata": map[string]any{"trigger": "auto", "preTokens": 180000, "postTokens": 22000}}
}

func spawn(prompt string) map[string]any {
	return map[string]any{"description": "d", "prompt": prompt, "subagent_type": "general-purpose"}
}
