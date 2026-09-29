package main

import "testing"

func errEvent(sid, tool, class, fp string) Event {
	return Event{Kind: "tool_error", SessionID: sid, Tool: tool, Class: class, FP: fp}
}

func kinds(sigs []Signal) map[string]int {
	out := map[string]int{}
	for _, s := range sigs {
		out[s.Kind]++
	}
	return out
}

func TestDeriveRepeatFailNeedsThree(t *testing.T) {
	two := []Event{errEvent("s", "Bash", classExit, "a"), errEvent("s", "Bash", classExit, "a")}
	if got := kinds(derive(two))[sigRepeatFail]; got != 0 {
		t.Errorf("2 failures gave %d repeat_fail signals, want 0", got)
	}
	three := append(two, errEvent("s", "Bash", classExit, "a"))
	if got := kinds(derive(three))[sigRepeatFail]; got != 1 {
		t.Errorf("3 failures gave %d repeat_fail signals, want 1", got)
	}
}

func TestDeriveChurnAndRevert(t *testing.T) {
	edit := func(old, new string) Event { return Event{Kind: "edit", File: "f.go", OldHash: old, NewHash: new} }
	sigs := derive([]Event{edit("a", "b"), edit("b", "c"), edit("c", "d"), edit("b", "a")})
	got := kinds(sigs)
	if got[sigChurn] != 1 || got[sigRevert] != 1 {
		t.Errorf("kinds = %v, want one churn and one revert", got)
	}
}

func TestQualifies(t *testing.T) {
	once := Signal{Kind: sigHallucination, Events: []Event{{}}}
	if once.qualifies() {
		t.Error("single unseen hallucination qualified")
	}
	once.Prior = 2
	if !once.qualifies() {
		t.Error("hallucination seen in 2 prior sessions did not qualify")
	}
	twice := Signal{Kind: sigHallucination, Events: []Event{{}, {}}}
	if !twice.qualifies() {
		t.Error("hallucination recurring in-session did not qualify")
	}
}

func TestPriorSessionsExcludesCurrent(t *testing.T) {
	events := []Event{
		errEvent("cur", "Read", classPathMissing, "x"),
		errEvent("old1", "Read", classPathMissing, "x"),
		errEvent("old2", "Read", classPathMissing, "x"),
	}
	if got := priorSessions(events, "cur")["x"]; got != 2 {
		t.Errorf("priorSessions[x] = %d, want 2", got)
	}
}
