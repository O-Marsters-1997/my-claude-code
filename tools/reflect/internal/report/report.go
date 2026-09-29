package report

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

const cleanupWarnDays = 90

func Status(s logstore.Store, enabled bool, library string) (string, error) {
	events, err := s.Read()
	if err != nil {
		return "", err
	}
	var w strings.Builder
	state := "off"
	if enabled {
		state = "on"
	}
	fmt.Fprintf(&w, "reflect: %s\nlog: %s (%d events, %d sessions)\nlibrary: %s\n",
		state, s.LogPath(), len(events), len(bySession(events)), library)
	if d := cleanupDays(); d < cleanupWarnDays {
		fmt.Fprintf(&w, "warn: cleanupPeriodDays=%d, transcripts the log points to are deleted after that\n", d)
	}
	return w.String(), nil
}

func cleanupDays() int {
	const claudeDefault = 30
	home, err := os.UserHomeDir()
	if err != nil {
		return claudeDefault
	}
	b, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		return claudeDefault
	}
	var cfg struct {
		CleanupPeriodDays int `json:"cleanupPeriodDays"`
	}
	if json.Unmarshal(b, &cfg) != nil {
		return claudeDefault
	}
	return cmp.Or(cfg.CleanupPeriodDays, claudeDefault)
}

func Show(s logstore.Store, sessionID string, all bool) (string, error) {
	events, err := s.Read()
	if err != nil {
		return "", err
	}
	var w strings.Builder
	session := bySession(events)[sessionID]
	prior := priorSessions(events, sessionID)
	shown := 0
	fmt.Fprintf(&w, "session %s instr=%s events=%d\n", sessionID, instrHashes(events)[sessionID], len(session))
	for _, sig := range derive(session) {
		sig.prior = prior[sig.fp]
		if !all && !sig.qualifies() {
			continue
		}
		shown++
		fmt.Fprintf(&w, "\n[%s] %s x%d fp=%s prior_sessions=%d\n", sig.kind, sig.label, len(sig.events), sig.fp, sig.prior)
		for _, e := range sig.events[:min(len(sig.events), 3)] {
			fmt.Fprintf(&w, "  %s agent=%s tool_use_id=%s transcript=%s\n",
				e.TS.Format("15:04:05"), cmp.Or(e.AgentID, "main"), cmp.Or(e.ToolUseID, "-"), e.Transcript)
			if detail := cmp.Or(e.Error, e.Input); detail != "" {
				fmt.Fprintf(&w, "    %.200s\n", strings.ReplaceAll(detail, "\n", " "))
			}
		}
	}
	if shown == 0 {
		fmt.Fprintln(&w, "no qualifying signals")
	}
	return w.String(), nil
}

type tally struct {
	sessions, toolCalls, prompts, confusion, corrections, repeatFail int
}

func (t *tally) add(o tally) {
	t.sessions += o.sessions
	t.toolCalls += o.toolCalls
	t.prompts += o.prompts
	t.confusion += o.confusion
	t.corrections += o.corrections
	t.repeatFail += o.repeatFail
}

func sessionTally(events []logstore.Event) (t tally, ended bool) {
	for _, e := range events {
		if e.Kind == "session_end" {
			ended = true
			t.toolCalls += e.ToolCalls
			t.prompts += e.Prompts
		}
	}
	t.sessions = 1
	for _, sig := range derive(events) {
		switch sig.kind {
		case hallucination:
			t.confusion += len(sig.events)
		case repeatFail:
			t.confusion++
			t.repeatFail++
		case churn, revert:
			t.confusion++
		case correction:
			t.corrections += len(sig.events)
		}
	}
	return t, ended
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func Metrics(s logstore.Store) (string, error) {
	events, err := s.Read()
	if err != nil {
		return "", err
	}
	hashes := instrHashes(events)
	byHash := map[string]*tally{}
	for sid, evs := range bySession(events) {
		t, ended := sessionTally(evs)
		if !ended {
			continue
		}
		if byHash[hashes[sid]] == nil {
			byHash[hashes[sid]] = &tally{}
		}
		byHash[hashes[sid]].add(t)
	}
	if len(byHash) == 0 {
		return "no completed sessions yet\n", nil
	}
	var w strings.Builder
	keys := sortedKeys(byHash)
	sort.SliceStable(keys, func(i, j int) bool { return byHash[keys[i]].sessions > byHash[keys[j]].sessions })
	fmt.Fprintf(&w, "%-14s %5s %11s %14s %10s %13s\n", "instr", "n", "tool_calls", "confusion/100", "corr_rate", "repeat/sess")
	for _, h := range keys {
		t := byHash[h]
		note := ""
		if t.sessions < 5 {
			note = "  (n<5: no verdict)"
		}
		fmt.Fprintf(&w, "%-14s %5d %11d %14.2f %10.2f %13.2f%s\n", cmp.Or(h, "-"), t.sessions, t.toolCalls,
			100*ratio(t.confusion, t.toolCalls), ratio(t.corrections, t.prompts), ratio(t.repeatFail, t.sessions), note)
	}
	return w.String(), nil
}
