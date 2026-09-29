package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var library string

func cliStore() store {
	wd, _ := os.Getwd()
	return newStore(cmp.Or(os.Getenv("CLAUDE_PROJECT_DIR"), wd))
}

func cmdOn(w io.Writer) error {
	s := cliStore()
	if err := s.ensure(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.dir, "on"), nil, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(w, "reflect on: logging to %s\n", s.logPath())
	return nil
}

func cmdOff(w io.Writer) error {
	err := os.Remove(filepath.Join(cliStore().dir, "on"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Fprintln(w, "reflect off (log kept)")
	return nil
}

func cleanupDays() int {
	home, err := os.UserHomeDir()
	if err != nil {
		return 30
	}
	b, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		return 30
	}
	var cfg struct {
		CleanupPeriodDays int `json:"cleanupPeriodDays"`
	}
	if json.Unmarshal(b, &cfg) != nil || cfg.CleanupPeriodDays == 0 {
		return 30
	}
	return cfg.CleanupPeriodDays
}

func cmdStatus(w io.Writer) error {
	s := cliStore()
	events, err := s.read()
	if err != nil {
		return err
	}
	state := "off"
	if s.enabled() {
		state = "on"
	}
	fmt.Fprintf(w, "reflect: %s\nlog: %s (%d events, %d sessions)\nlibrary: %s\n",
		state, s.logPath(), len(events), len(bySession(events)), library)
	if d := cleanupDays(); d < 90 {
		fmt.Fprintf(w, "warn: cleanupPeriodDays=%d, transcripts the log points to are deleted after that\n", d)
	}
	return nil
}

func cmdProposals(w io.Writer, args []string) error {
	s := cliStore()
	if len(args) > 0 && args[0] == "library" {
		s = store{filepath.Join(library, ".claude", "reflect")}
	}
	dir := filepath.Join(s.dir, "proposals")
	if err := s.ensure(); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fmt.Fprintln(w, dir)
	return nil
}

func cmdShow(w io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: reflect show <session_id> [--all]")
	}
	sid, all := args[0], len(args) > 1 && args[1] == "--all"
	events, err := cliStore().read()
	if err != nil {
		return err
	}
	session := bySession(events)[sid]
	prior := priorSessions(events, sid)
	shown := 0
	fmt.Fprintf(w, "session %s instr=%s events=%d\n", sid, instrHashes(events)[sid], len(session))
	for _, sig := range derive(session) {
		sig.Prior = prior[sig.FP]
		if !all && !sig.qualifies() {
			continue
		}
		shown++
		fmt.Fprintf(w, "\n[%s] %s x%d fp=%s prior_sessions=%d\n", sig.Kind, sig.Label, len(sig.Events), sig.FP, sig.Prior)
		for _, e := range sig.Events[:min(len(sig.Events), 3)] {
			agent := cmp.Or(e.AgentID, "main")
			fmt.Fprintf(w, "  %s agent=%s tool_use_id=%s transcript=%s\n", e.TS.Format("15:04:05"), agent, cmp.Or(e.ToolUseID, "-"), e.Transcript)
			if detail := cmp.Or(e.Error, e.Input); detail != "" {
				fmt.Fprintf(w, "    %s\n", truncate(strings.ReplaceAll(detail, "\n", " "), 200))
			}
		}
	}
	if shown == 0 {
		fmt.Fprintln(w, "no qualifying signals")
	}
	return nil
}

type tally struct {
	sessions, toolCalls, prompts, confusion, corrections, repeatFail int
}

func cmdMetrics(w io.Writer) error {
	events, err := cliStore().read()
	if err != nil {
		return err
	}
	hashes := instrHashes(events)
	tallies := map[string]*tally{}
	for sid, evs := range bySession(events) {
		ended := false
		t := &tally{}
		for _, e := range evs {
			if e.Kind == "session_end" {
				ended = true
				t.toolCalls += e.ToolCalls
				t.prompts += e.Prompts
			}
		}
		if !ended {
			continue
		}
		t.sessions = 1
		for _, sig := range derive(evs) {
			switch sig.Kind {
			case sigHallucination:
				t.confusion += len(sig.Events)
			case sigRepeatFail:
				t.confusion++
				t.repeatFail++
			case sigChurn, sigRevert:
				t.confusion++
			case sigCorrection:
				t.corrections += len(sig.Events)
			}
		}
		h := hashes[sid]
		if tallies[h] == nil {
			tallies[h] = &tally{}
		}
		tallies[h].add(*t)
	}
	if len(tallies) == 0 {
		fmt.Fprintln(w, "no completed sessions yet")
		return nil
	}
	keys := sortedKeys(tallies)
	sort.SliceStable(keys, func(i, j int) bool { return tallies[keys[i]].sessions > tallies[keys[j]].sessions })
	fmt.Fprintf(w, "%-14s %5s %11s %14s %10s %13s\n", "instr", "n", "tool_calls", "confusion/100", "corr_rate", "repeat/sess")
	for _, h := range keys {
		t := tallies[h]
		note := ""
		if t.sessions < 5 {
			note = "  (n<5: no verdict)"
		}
		fmt.Fprintf(w, "%-14s %5d %11d %14.2f %10.2f %13.2f%s\n", cmp.Or(h, "-"), t.sessions, t.toolCalls,
			per(t.confusion*100, t.toolCalls), per(t.corrections, t.prompts), per(t.repeatFail, t.sessions), note)
	}
	return nil
}

func (t *tally) add(o tally) {
	t.sessions += o.sessions
	t.toolCalls += o.toolCalls
	t.prompts += o.prompts
	t.confusion += o.confusion
	t.corrections += o.corrections
	t.repeatFail += o.repeatFail
}

func per(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}
