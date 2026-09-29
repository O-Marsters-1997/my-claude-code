package report_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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

func TestShowReadsV1LogLines(t *testing.T) {
	s := logstore.At(t.TempDir())
	lines := []string{
		`{"v":1,"ts":"2026-09-01T10:00:00Z","kind":"session","session_id":"s","class":"startup","cwd":"/r","transcript":"/t/s.jsonl","instr_hash":"old111","files":{"/r/CLAUDE.md":"aaaaaaaaaaaa"}}`,
		`{"v":1,"ts":"2026-09-01T10:01:00Z","kind":"tool_error","session_id":"s","tool_use_id":"t1","tool":"Read","class":"path_missing","fp":"Read|path_missing|/x","transcript":"/t/s.jsonl"}`,
		`{"v":1,"ts":"2026-09-01T10:02:00Z","kind":"tool_error","session_id":"s","agent_id":"a1","tool_use_id":"t2","tool":"Read","class":"path_missing","fp":"Read|path_missing|/x","transcript":"/t/s/subagents/agent-a1.jsonl"}`,
	}
	if err := os.WriteFile(s.LogPath(), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := show(t, s, "s")
	for _, want := range []string{"instr=old111", "[hallucination]", "transcript=/t/s.jsonl", "transcript=/t/s/subagents/agent-a1.jsonl"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestShowResolvesTranscriptsFromTheSessionEvent(t *testing.T) {
	s := newStore(t,
		logstore.Event{Kind: "session", SessionID: "s", Transcript: "/t/s.jsonl"},
		logstore.Event{Kind: "tool_error", SessionID: "s", AgentID: "a1", ToolUseID: "t1", Tool: "Read", Class: "path_missing", FP: "x"},
		logstore.Event{Kind: "tool_error", SessionID: "s", ToolUseID: "t2", Tool: "Read", Class: "path_missing", FP: "x"},
	)
	out := show(t, s, "s")
	for _, want := range []string{"transcript=/t/s/subagents/agent-a1.jsonl", "transcript=/t/s.jsonl"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func manifestStore(t *testing.T) logstore.Store {
	t.Helper()
	return newStore(t,
		logstore.Event{Kind: "session", SessionID: "s1", InstrHash: "h1"},
		logstore.Event{Kind: "manifest", SessionID: "s1", InstrHash: "h1", Added: map[string]string{"/r/CLAUDE.md": "a", "/r/rules/x.md": "b"}},
		logstore.Event{Kind: "session_end", SessionID: "s1", ToolCalls: 5},
		logstore.Event{Kind: "session", SessionID: "s2", InstrHash: "h2"},
		logstore.Event{Kind: "manifest", SessionID: "s2", InstrHash: "h2", Prev: "h1", Changed: map[string]string{"/r/CLAUDE.md": "c"}, Removed: []string{"/r/rules/x.md"}},
		logstore.Event{Kind: "session_end", SessionID: "s2", ToolCalls: 5},
	)
}

func TestShowNamesFilesChangedSincePreviousHash(t *testing.T) {
	s := manifestStore(t)
	out := show(t, s, "s2")
	for _, want := range []string{"instr changes vs h1", "~/r/CLAUDE.md", "-/r/rules/x.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if out := show(t, s, "s1"); strings.Contains(out, "instr changes") {
		t.Errorf("first hash reported changes:\n%s", out)
	}
}

func TestMetricsNamesFilesChangedSincePreviousHash(t *testing.T) {
	out, err := report.Metrics(manifestStore(t))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "changed vs h1") || !strings.Contains(out, "~/r/CLAUDE.md") {
		t.Errorf("metrics output missing the change line:\n%s", out)
	}
}

func TestStatusReportsLogSize(t *testing.T) {
	s := newStore(t, fail("s", "path_missing", "x"))
	out, err := report.Status(s, true, "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`1 events, 1 sessions, \d+(\.\d)? (B|KB|MB)\)`).MatchString(out) {
		t.Errorf("status output = %q, want the log size", out)
	}
}

func logOut(t *testing.T, s logstore.Store, f report.LogFilter) []string {
	t.Helper()
	out, err := report.Log(s, f)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
}

func TestLogPrintsOneLinePerEventWithinWidth(t *testing.T) {
	s := newStore(t,
		logstore.Event{Kind: "session", SessionID: "abcdef123456", Class: "startup", Branch: "main", Commit: "c91e20e", InstrHash: "h1"},
		logstore.Event{Kind: "tool_error", SessionID: "abcdef123456", Tool: "Read", Class: "path_missing", Error: strings.Repeat("e", 400)},
		logstore.Event{Kind: "edit", SessionID: "abcdef123456", Tool: "Edit", File: "/r/a.go"},
		logstore.Event{Kind: "session_end", SessionID: "abcdef123456", ToolCalls: 4, Prompts: 2, Errors: 1},
	)
	lines := logOut(t, s, report.LogFilter{})
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	for _, line := range lines {
		if n := len([]rune(line)); n > 120 {
			t.Errorf("line is %d columns: %q", n, line)
		}
		if !strings.Contains(line, "abcdef12") {
			t.Errorf("line missing the short session id: %q", line)
		}
	}
	for i, want := range []string{"session", "tool_error", "edit", "session_end"} {
		if !strings.Contains(lines[i], want) {
			t.Errorf("line %d = %q, want kind %s", i, lines[i], want)
		}
	}
	if !strings.Contains(lines[1], "path_missing") || !strings.Contains(lines[2], "/r/a.go") || !strings.Contains(lines[3], "calls=4") {
		t.Errorf("lines lack their detail:\n%s", strings.Join(lines, "\n"))
	}
}

func TestLogFilters(t *testing.T) {
	s := newStore(t,
		logstore.Event{Kind: "edit", SessionID: "aaaa1111", File: "/1"},
		logstore.Event{Kind: "tool_error", SessionID: "aaaa1111", Class: "path_missing"},
		logstore.Event{Kind: "edit", SessionID: "bbbb2222", File: "/2"},
		logstore.Event{Kind: "edit", SessionID: "bbbb2222", File: "/3"},
	)
	if got := logOut(t, s, report.LogFilter{Kind: "edit"}); len(got) != 3 {
		t.Errorf("kind filter returned %d lines, want 3", len(got))
	}
	if got := logOut(t, s, report.LogFilter{Session: "bbbb"}); len(got) != 2 {
		t.Errorf("session prefix filter returned %d lines, want 2", len(got))
	}
	got := logOut(t, s, report.LogFilter{Kind: "edit", Last: 2})
	if len(got) != 2 || !strings.Contains(got[0], "/2") || !strings.Contains(got[1], "/3") {
		t.Errorf("--last 2 = %q, want the final two edits", got)
	}
}

func TestLogSkipsMalformedLines(t *testing.T) {
	s := newStore(t, logstore.Event{Kind: "edit", SessionID: "s", File: "/a"})
	f, err := os.OpenFile(s.LogPath(), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("not json\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if got := logOut(t, s, report.LogFilter{}); len(got) != 1 {
		t.Errorf("got %d lines, want the one valid event: %q", len(got), got)
	}
}

var now = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func at(days int, e logstore.Event) logstore.Event {
	e.TS = now.AddDate(0, 0, -days)
	return e
}

func rawLog(t *testing.T, s logstore.Store) string {
	t.Helper()
	b, err := os.ReadFile(s.LogPath())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func pruneFixture(t *testing.T) logstore.Store {
	t.Helper()
	old := func(sid string, hash string) []logstore.Event {
		return []logstore.Event{
			at(60, logstore.Event{Kind: "session", SessionID: sid, InstrHash: hash}),
			at(60, logstore.Event{Kind: "tool_error", SessionID: sid, ToolUseID: sid + "1", Tool: "Read", Class: "path_missing", FP: "Read|path_missing|/x", Input: "in", Error: "err"}),
			at(60, logstore.Event{Kind: "tool_error", SessionID: sid, ToolUseID: sid + "2", Tool: "Bash", Class: "exit_nonzero", FP: "Bash|exit_nonzero|make"}),
			at(60, logstore.Event{Kind: "tool_error", SessionID: sid, ToolUseID: sid + "3", Tool: "Bash", Class: "exit_nonzero", FP: "Bash|exit_nonzero|make"}),
			at(60, logstore.Event{Kind: "tool_error", SessionID: sid, ToolUseID: sid + "4", Tool: "Bash", Class: "exit_nonzero", FP: "Bash|exit_nonzero|make"}),
			at(60, logstore.Event{Kind: "correction", SessionID: sid, Conf: 0.8, Class: "no,"}),
			at(60, logstore.Event{Kind: "session_end", SessionID: sid, ToolCalls: 20, Prompts: 4}),
		}
	}
	events := append(old("old1", "h1"), old("old2", "h1")...)
	events = append(events,
		at(60, logstore.Event{Kind: "manifest", SessionID: "old1", InstrHash: "h1", Added: map[string]string{"/r/CLAUDE.md": "a"}}),
		at(1, logstore.Event{Kind: "session", SessionID: "cur", InstrHash: "h2"}),
		at(1, logstore.Event{Kind: "tool_error", SessionID: "cur", ToolUseID: "c1", Tool: "Read", Class: "path_missing", FP: "Read|path_missing|/x"}),
		at(1, logstore.Event{Kind: "session_end", SessionID: "cur", ToolCalls: 10}),
	)
	return newStore(t, events...)
}

func pruned(t *testing.T, s logstore.Store) string {
	t.Helper()
	out, err := report.Prune(s, 30, now)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestPruneKeepsMetricsAndPriorSessionCounts(t *testing.T) {
	s := pruneFixture(t)
	metricsBefore, err := report.Metrics(s)
	if err != nil {
		t.Fatal(err)
	}
	showBefore := show(t, s, "cur")
	before := s.Size()

	pruned(t, s)

	metricsAfter, err := report.Metrics(s)
	if err != nil {
		t.Fatal(err)
	}
	if metricsAfter != metricsBefore {
		t.Errorf("metrics changed by pruning:\nbefore:\n%s\nafter:\n%s", metricsBefore, metricsAfter)
	}
	if showAfter := show(t, s, "cur"); showAfter != showBefore {
		t.Errorf("show changed by pruning:\nbefore:\n%s\nafter:\n%s", showBefore, showAfter)
	}
	if !strings.Contains(show(t, s, "cur"), "prior_sessions=2") {
		t.Error("prior session counts were lost")
	}
	if s.Size() >= before {
		t.Errorf("log grew from %d to %d bytes", before, s.Size())
	}
}

func TestPruneLeavesRecentSessionsAndManifestsByteIdentical(t *testing.T) {
	s := pruneFixture(t)
	var keep []string
	for _, line := range strings.Split(strings.TrimSpace(rawLog(t, s)), "\n") {
		if strings.Contains(line, `"session_id":"cur"`) || strings.Contains(line, `"kind":"manifest"`) {
			keep = append(keep, line)
		}
	}
	pruned(t, s)
	after := rawLog(t, s)
	for _, line := range keep {
		if !strings.Contains(after, line+"\n") {
			t.Errorf("line changed or dropped: %s", line)
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(after), "\n") {
		var e logstore.Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(e.SessionID, "old") && e.Kind != "manifest" && e.Kind != "summary" {
			t.Errorf("raw %s event of an old session survived", e.Kind)
		}
	}
	if n := strings.Count(after, `"kind":"summary"`); n != 2 {
		t.Errorf("got %d summary events, want one per old session", n)
	}
}

func TestPruneIsIdempotent(t *testing.T) {
	s := pruneFixture(t)
	pruned(t, s)
	once := rawLog(t, s)
	pruned(t, s)
	if twice := rawLog(t, s); twice != once {
		t.Errorf("second prune changed the log:\n%s\nvs\n%s", once, twice)
	}
}

func TestPruneRemovesStaleEditState(t *testing.T) {
	s := pruneFixture(t)
	stale := filepath.Join(s.Dir(), "edits", "old")
	fresh := filepath.Join(s.Dir(), "edits", "new")
	for _, dir := range []string{stale, fresh} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(stale, now.AddDate(0, 0, -90), now.AddDate(0, 0, -90)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(fresh, now, now); err != nil {
		t.Fatal(err)
	}
	pruned(t, s)
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale edit state kept: %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("fresh edit state removed: %v", err)
	}
}

func TestPruneKeepsUnparseableLinesAndHandlesMissingLog(t *testing.T) {
	empty := logstore.At(t.TempDir())
	if _, err := report.Prune(empty, 30, now); err != nil {
		t.Errorf("Prune on a missing log = %v, want nil", err)
	}
	s := pruneFixture(t)
	f, err := os.OpenFile(s.LogPath(), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("garbage line\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	pruned(t, s)
	if !strings.Contains(rawLog(t, s), "garbage line\n") {
		t.Error("unparseable line was dropped")
	}
}

func TestPruneFoldsActivityLoggedAfterASessionWasSummarised(t *testing.T) {
	s := pruneFixture(t)
	pruned(t, s)
	if err := s.Append(at(0, logstore.Event{Kind: "session_end", SessionID: "old1", ToolCalls: 5})); err != nil {
		t.Fatal(err)
	}
	out, err := report.Metrics(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, " 45 ") {
		t.Errorf("resumed activity missing from metrics (want 20+20+5 tool calls):\n%s", out)
	}
	if _, err := report.Prune(s, 30, now.AddDate(0, 0, 90)); err != nil {
		t.Fatal(err)
	}
	after, err := report.Metrics(s)
	if err != nil {
		t.Fatal(err)
	}
	if after != out {
		t.Errorf("re-pruning changed metrics:\n%s\nvs\n%s", out, after)
	}
	if n := strings.Count(rawLog(t, s), `"kind":"summary"`); n != 3 {
		t.Errorf("got %d summaries, want one per session (old1, old2, cur)", n)
	}
}

func TestPruneRejectsNegativeDays(t *testing.T) {
	s := pruneFixture(t)
	before := rawLog(t, s)
	if _, err := report.Prune(s, -1, now); err == nil {
		t.Error("Prune(-1) = nil, want an error")
	}
	if rawLog(t, s) != before {
		t.Error("a rejected prune changed the log")
	}
}
