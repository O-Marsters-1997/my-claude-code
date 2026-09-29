package report_test

import (
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/report"
)

func newStore(t *testing.T, events ...logstore.Event) logstore.Store {
	t.Helper()
	s := logstore.At(t.TempDir())
	for _, e := range events {
		if err := s.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func show(t *testing.T, s logstore.Store, sid string) string {
	t.Helper()
	out, err := report.Show(s, sid, false)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func fail(sid, class, fp string) logstore.Event {
	return logstore.Event{Kind: "tool_error", SessionID: sid, Tool: "Bash", Class: class, FP: fp}
}

func TestShowRepeatFailNeedsThree(t *testing.T) {
	two := newStore(t, fail("s", "exit_nonzero", "a"), fail("s", "exit_nonzero", "a"))
	if out := show(t, two, "s"); strings.Contains(out, "[repeat_fail]") {
		t.Errorf("2 failures reported as repeat_fail:\n%s", out)
	}
	three := newStore(t, fail("s", "exit_nonzero", "a"), fail("s", "exit_nonzero", "a"), fail("s", "exit_nonzero", "a"))
	if out := show(t, three, "s"); !strings.Contains(out, "[repeat_fail]") {
		t.Errorf("3 failures not reported:\n%s", out)
	}
}

func TestShowChurnAndRevert(t *testing.T) {
	edit := func(old, new string) logstore.Event {
		return logstore.Event{Kind: "edit", SessionID: "s", File: "f.go", OldHash: old, NewHash: new}
	}
	s := newStore(t, edit("a", "b"), edit("b", "c"), edit("c", "d"), edit("b", "a"))
	out := show(t, s, "s")
	for _, want := range []string{"[churn]", "[revert]"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %s:\n%s", want, out)
		}
	}
}

func TestShowHallucinationNeedsRecurrence(t *testing.T) {
	once := newStore(t, fail("s", "path_missing", "x"))
	if out := show(t, once, "s"); !strings.Contains(out, "no qualifying signals") {
		t.Errorf("one-off hallucination qualified:\n%s", out)
	}
	twice := newStore(t, fail("s", "path_missing", "x"), fail("s", "path_missing", "x"))
	if out := show(t, twice, "s"); !strings.Contains(out, "[hallucination]") {
		t.Errorf("recurring hallucination missing:\n%s", out)
	}
	seenBefore := newStore(t, fail("cur", "path_missing", "x"), fail("old1", "path_missing", "x"), fail("old2", "path_missing", "x"))
	if out := show(t, seenBefore, "cur"); !strings.Contains(out, "prior_sessions=2") {
		t.Errorf("fingerprint seen in 2 prior sessions did not qualify:\n%s", out)
	}
}

func TestShowIgnoresOtherSessions(t *testing.T) {
	s := newStore(t, fail("other", "path_missing", "x"), fail("other", "path_missing", "x"))
	if out := show(t, s, "mine"); !strings.Contains(out, "no qualifying signals") {
		t.Errorf("another session's events leaked in:\n%s", out)
	}
}

func TestMetricsGroupsByInstrHash(t *testing.T) {
	s := newStore(t,
		logstore.Event{Kind: "session", SessionID: "s", InstrHash: "abc"},
		fail("s", "path_missing", "x"), fail("s", "path_missing", "x"),
		logstore.Event{Kind: "session_end", SessionID: "s", ToolCalls: 10, Prompts: 2},
	)
	out, err := report.Metrics(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "abc") || !strings.Contains(out, "20.00") || !strings.Contains(out, "n<5") {
		t.Errorf("metrics output = %q, want hash abc, 20.00 confusion/100 and an n<5 note", out)
	}
}
