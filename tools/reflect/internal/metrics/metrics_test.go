package metrics_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/metrics"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
)

const (
	failedRead = `{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/x.go"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"File does not exist.","is_error":true}]}}
`
	goodRead = `{"type":"user","message":{"content":"please read x"}}
{"type":"assistant","message":{"id":"m1","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/x.go"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"package x"}]}}
`
)

const abandonedAgent = `{"type":"assistant","message":{"id":"m2","content":[{"type":"tool_use","id":"t2","name":"Agent","input":{"prompt":"bg"}}]}}
{"type":"user","toolUseResult":{"isAsync":true,"agentId":"bg1"},"message":{"content":[{"type":"tool_result","tool_use_id":"t2","content":"launched"}]}}
`

func writeSession(t *testing.T, projects, sid, body string) {
	t.Helper()
	dir := filepath.Join(projects, "-repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, sid+".jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestReportGroupsByInstructionHash(t *testing.T) {
	projects := t.TempDir()
	root := gitRepo(t)
	other := gitRepo(t)
	writeSession(t, projects, "s1", failedRead)
	writeSession(t, projects, "s2", goodRead+abandonedAgent)
	subagents := filepath.Join(projects, "-repo", "s2", "subagents")
	if err := os.MkdirAll(subagents, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"agent-bg1.meta.json": `{"agentType":"general-purpose","toolUseId":"t2"}`,
		"agent-bg1.jsonl":     `{"type":"user","message":{"content":"bg"}}` + "\n",
	} {
		if err := os.WriteFile(filepath.Join(subagents, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeSession(t, projects, "s3", failedRead)
	writeSession(t, projects, "now", failedRead)
	at := func(min int) time.Time { return time.Date(2026, 10, 7, 10, min, 0, 0, time.UTC) }
	records := []record.Record{
		{SessionID: "s1", Cwd: root, TS: at(1), InstrHash: "h1", Files: map[string]string{"AGENTS.md": "a"}},
		{SessionID: "s2", Cwd: root, TS: at(2), InstrHash: "h2", Files: map[string]string{"AGENTS.md": "b", "rules/x.md": "c"}},
		{SessionID: "s3", Cwd: other, TS: at(3), InstrHash: "h1"},
		{SessionID: "now", Cwd: root, TS: at(4), InstrHash: "h2"},
	}

	out := metrics.Report(metrics.Options{Projects: projects, Records: records, Root: root, Exclude: "now"})

	rows := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) > 0 {
			rows[f[0]] = line
		}
	}
	for hash, want := range map[string][]string{
		"h1": {"h1", "1", "1", "100.00", "0.00", "0.00", "(n<5:"},
		"h2": {"h2", "1", "2", "0.00", "0.00", "0.00", "(n<5:"},
	} {
		if got := strings.Fields(rows[hash]); len(got) < len(want) || strings.Join(got[:len(want)], " ") != strings.Join(want, " ") {
			t.Errorf("Report() row %s = %q, want prefix %q\n%s", hash, rows[hash], want, out)
		}
	}
	if !strings.Contains(out, "changed vs h1: ~AGENTS.md +rules/x.md") {
		t.Errorf("Report() lacks change summary:\n%s", out)
	}
}
