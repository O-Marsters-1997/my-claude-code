package digest_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/digest"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

type script struct {
	t     *testing.T
	lines []transcript.Line
	ids   int
}

func (s *script) add(m map[string]any) {
	s.t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		s.t.Fatal(err)
	}
	l := transcript.Line{N: len(s.lines) + 1, Raw: raw}
	if err := json.Unmarshal(raw, &l.Entry); err != nil {
		s.t.Fatal(err)
	}
	s.lines = append(s.lines, l)
}

func (s *script) prompt(text string) {
	s.add(map[string]any{"type": "user", "message": map[string]any{"content": text}})
}

func (s *script) call(name string, input map[string]any, result string, isError bool) {
	s.ids++
	id := fmt.Sprintf("tu_%d", s.ids)
	s.add(map[string]any{"type": "assistant", "message": map[string]any{
		"id":      fmt.Sprintf("msg_%d", s.ids),
		"content": []any{map[string]any{"type": "tool_use", "id": id, "name": name, "input": input}},
		"usage":   map[string]any{"input_tokens": 1000, "output_tokens": 10},
	}})
	s.add(map[string]any{"type": "user", "message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "tool_use_id": id, "content": result, "is_error": isError},
	}}})
}

func (s *script) agent(id string) *session.Agent {
	return &session.Agent{ID: id, Type: "general-purpose", Model: "sonnet", Brief: "do the thing", Lines: s.lines}
}

func tagsAt(d digest.Digest) map[int][]string {
	out := map[int][]string{}
	for _, sig := range d.Signals {
		out[sig.Line] = append(out[sig.Line], sig.Tag)
	}
	return out
}

func TestBuildTagsSignals(t *testing.T) {
	s := &script{t: t}
	s.call("Read", map[string]any{"file_path": "/r/missing.go"}, "File does not exist.", true)
	s.call("Edit", map[string]any{"file_path": "/r/a.go", "old_string": "x", "new_string": "y"}, "<tool_use_error>String to replace not found in file.</tool_use_error>", true)
	for range 3 {
		s.call("Bash", map[string]any{"command": "go test ./..."}, "Exit code 1\nFAIL", true)
	}
	s.call("Read", map[string]any{"file_path": "/r/b.go"}, "package b", false)
	s.call("Read", map[string]any{"file_path": "/r/b.go"}, "package b", false)
	s.call("Edit", map[string]any{"file_path": "/r/b.go", "old_string": "p", "new_string": "q"}, "ok", false)
	s.call("Read", map[string]any{"file_path": "/r/b.go"}, "package b", false)
	s.call("Edit", map[string]any{"file_path": "/r/b.go", "old_string": "q", "new_string": "p"}, "ok", false)
	s.call("Edit", map[string]any{"file_path": "/r/b.go", "old_string": "m", "new_string": "n"}, "ok", false)
	s.call("Edit", map[string]any{"file_path": "/r/b.go", "old_string": "n", "new_string": "o"}, "ok", false)
	s.call("Bash", map[string]any{"command": "cat big.log"}, strings.Repeat("x", 25000), false)

	got := tagsAt(digest.Build(s.agent("a1")))

	want := map[int][]string{
		1:  {digest.Halluc},
		3:  {digest.Halluc},
		5:  {digest.Fail},
		7:  {digest.Fail},
		9:  {digest.Fail, digest.Repeat},
		13: {digest.Reread},
		19: {digest.Revert},
		23: {digest.Churn},
		25: {digest.Big},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Build() tags by line (-want +got):\n%s", diff)
	}
}

func TestBuildTagsCorrectionsOnlyForMain(t *testing.T) {
	s := &script{t: t}
	s.prompt("no, that's wrong, I told you to use the existing helper instead")

	if got := tagsAt(digest.Build(s.agent(session.MainID))); len(got[1]) == 0 || got[1][0] != digest.Correction {
		t.Errorf("main Build() tags = %v, want correction on line 1", got)
	}
	if got := tagsAt(digest.Build(s.agent("sub1"))); len(got) != 0 {
		t.Errorf("subagent Build() tags = %v, want none", got)
	}
}

func TestBuildTotals(t *testing.T) {
	s := &script{t: t}
	s.call("Read", map[string]any{"file_path": "/r/a.go"}, "a", false)
	s.call("Read", map[string]any{"file_path": "/r/missing.go"}, "File does not exist.", true)
	s.add(map[string]any{"type": "assistant", "message": map[string]any{
		"id": "msg_2", "content": []any{map[string]any{"type": "text", "text": "same message, second block"}},
		"usage": map[string]any{"input_tokens": 1000, "output_tokens": 10},
	}})

	d := digest.Build(s.agent("a1"))

	if d.Calls != 2 || d.Fails != 1 || d.TokensIn != 2000 || d.TokensOut != 20 {
		t.Errorf("Build() totals = calls %d fails %d in %d out %d, want 2 1 2000 20", d.Calls, d.Fails, d.TokensIn, d.TokensOut)
	}
}

func TestRenderShowsContext(t *testing.T) {
	s := &script{t: t}
	s.call("Read", map[string]any{"file_path": "/r/missing.go"}, "File does not exist.", true)
	s.add(map[string]any{"type": "system", "subtype": "compact_boundary", "compactMetadata": map[string]any{"preTokens": 180000, "postTokens": 22000}})
	a := s.agent("a1")
	a.Outcome = "used the result"

	out := digest.Build(a).Render()

	for _, want := range []string{"agent a1 general-purpose sonnet", "brief: do the thing", "outcome: used the result", "L1", "Read", "/r/missing.go", "FAIL", "[halluc fp=", "compact_boundary (180k → 22k)", "totals: 1 calls, 1 fail"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() lacks %q:\n%s", want, out)
		}
	}
}

func TestTriage(t *testing.T) {
	quiet := func(id string, tokens int) digest.Digest {
		return digest.Digest{Agent: &session.Agent{ID: id}, TokensIn: tokens}
	}
	noisy := func(id string, signals int) digest.Digest {
		d := quiet(id, 10)
		for range signals {
			d.Signals = append(d.Signals, digest.Signal{Tag: digest.Halluc})
		}
		return d
	}
	ds := []digest.Digest{
		quiet(session.MainID, 10),
		noisy("loud", 3),
		noisy("louder", 5),
		quiet("heavy", 900),
		quiet("idle1", 10),
		quiet("idle2", 10),
		quiet("idle3", 10),
		{Agent: &session.Agent{ID: "reread"}, Signals: []digest.Signal{{Tag: digest.Reread}}},
	}

	got := digest.Triage(ds, 3)

	want := map[string]string{
		session.MainID: digest.Review,
		"louder":       digest.Review,
		"loud":         digest.Review,
		"heavy":        digest.Overflow,
		"idle1":        digest.Skip,
		"idle2":        digest.Skip,
		"idle3":        digest.Skip,
		"reread":       digest.Skip,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Triage() (-want +got):\n%s", diff)
	}
}

func TestRenderMarksTheFailedCallOfAParallelPair(t *testing.T) {
	s := &script{t: t}
	s.add(map[string]any{"type": "assistant", "message": map[string]any{"id": "m1", "content": []any{
		map[string]any{"type": "tool_use", "id": "ta", "name": "Read", "input": map[string]any{"file_path": "/r/a.go"}},
		map[string]any{"type": "tool_use", "id": "tb", "name": "Read", "input": map[string]any{"file_path": "/r/b.go"}},
	}}})
	s.add(map[string]any{"type": "user", "message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "tool_use_id": "ta", "content": "File does not exist.", "is_error": true},
		map[string]any{"type": "tool_result", "tool_use_id": "tb", "content": "package b"},
	}}})

	out := digest.Build(s.agent("a1")).Render()

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "/r/b.go") && strings.Contains(line, "FAIL") {
			t.Errorf("Render() marks b.go as failed:\n%s", out)
		}
		if strings.Contains(line, "/r/a.go") && !strings.Contains(line, "FAIL") {
			t.Errorf("Render() does not mark a.go as failed:\n%s", out)
		}
	}
}
