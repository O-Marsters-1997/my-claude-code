package report

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

const summaryFPs = 20

func Prune(s logstore.Store, days int, now time.Time) (string, error) {
	if days < 0 {
		return "", fmt.Errorf("older-than must be positive, got %d", days)
	}
	days = cmp.Or(days, cleanupDays())
	cutoff := now.AddDate(0, 0, -days)
	before := s.Size()
	pruned := 0
	err := s.Rewrite(func(lines [][]byte) ([][]byte, error) {
		events := make([]*logstore.Event, len(lines))
		bySession := map[string][]logstore.Event{}
		latest := map[string]time.Time{}
		for i, line := range lines {
			var e logstore.Event
			if json.Unmarshal(line, &e) != nil {
				continue
			}
			events[i] = &e
			bySession[e.SessionID] = append(bySession[e.SessionID], e)
			if e.TS.After(latest[e.SessionID]) {
				latest[e.SessionID] = e.TS
			}
		}
		old := map[string]bool{}
		for sid, evs := range bySession {
			if sid != "" && latest[sid].Before(cutoff) && slices.ContainsFunc(evs, hasRawActivity) {
				old[sid] = true
			}
		}
		pruned = len(old)
		var kept [][]byte
		summarised := map[string]bool{}
		for i, line := range lines {
			e := events[i]
			if e == nil || !old[e.SessionID] || e.Kind == "manifest" {
				kept = append(kept, line)
				continue
			}
			if summarised[e.SessionID] {
				continue
			}
			summarised[e.SessionID] = true
			summary, err := logstore.Encode(summarize(e.SessionID, bySession[e.SessionID], latest[e.SessionID]))
			if err != nil {
				return nil, err
			}
			kept = append(kept, summary)
		}
		return kept, nil
	})
	if err != nil {
		return "", err
	}
	removeStaleEditState(s, cutoff)
	return fmt.Sprintf("pruned %d sessions older than %d days: %s -> %s\n", pruned, days, humanSize(before), humanSize(s.Size())), nil
}

func hasRawActivity(e logstore.Event) bool { return e.Kind != "manifest" && e.Kind != "summary" }

func summarize(sessionID string, events []logstore.Event, last time.Time) logstore.Event {
	t, ended := sessionTally(events)
	fps := slices.Clone(sessionFPs(events))
	slices.Sort(fps)
	return logstore.Event{
		Kind: "summary", SessionID: sessionID, TS: last, InstrHash: instrHashes(events)[sessionID],
		ToolCalls: t.toolCalls, Prompts: t.prompts, Confusion: t.confusion, Corrections: t.corrections,
		RepeatFail: t.repeatFail, Ended: ended, FPs: fps[:min(len(fps), summaryFPs)],
	}
}

func removeStaleEditState(s logstore.Store, cutoff time.Time) {
	dir := filepath.Join(s.Dir(), "edits")
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if info, err := entry.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(dir, entry.Name()))
		}
	}
}
