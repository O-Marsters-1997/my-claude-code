package main

import (
	"slices"
	"sort"
)

var thresholds = struct {
	RepeatFail int
	Churn      int
	Recur      int
	Prior      int
	Correction float64
}{RepeatFail: 3, Churn: 4, Recur: 2, Prior: 2, Correction: 0.6}

const (
	sigHallucination = "hallucination"
	sigRepeatFail    = "repeat_fail"
	sigChurn         = "churn"
	sigRevert        = "revert"
	sigCorrection    = "correction"
)

type Signal struct {
	Kind   string
	Label  string
	FP     string
	Events []Event
	Prior  int
}

func (s Signal) qualifies() bool {
	switch s.Kind {
	case sigHallucination, sigCorrection:
		return len(s.Events) >= thresholds.Recur || s.Prior >= thresholds.Prior
	}
	return true
}

func groupBy(events []Event, key func(Event) string) map[string][]Event {
	groups := map[string][]Event{}
	for _, e := range events {
		if k := key(e); k != "" {
			groups[k] = append(groups[k], e)
		}
	}
	return groups
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func derive(events []Event) []Signal {
	var out []Signal

	errs := groupBy(events, func(e Event) string {
		if e.Kind == "tool_error" {
			return e.FP
		}
		return ""
	})
	for _, fp := range sortedKeys(errs) {
		g := errs[fp]
		label := g[0].Class + " " + g[0].Tool
		switch {
		case isHallucination(g[0].Class) || g[0].Class == classReadFirst:
			out = append(out, Signal{Kind: sigHallucination, Label: label, FP: fp, Events: g})
		case (g[0].Class == classExit || g[0].Class == classOther) && len(g) >= thresholds.RepeatFail:
			out = append(out, Signal{Kind: sigRepeatFail, Label: label, FP: fp, Events: g})
		}
	}

	edits := groupBy(events, func(e Event) string {
		if e.Kind == "edit" {
			return e.File
		}
		return ""
	})
	for _, file := range sortedKeys(edits) {
		g := edits[file]
		if len(g) >= thresholds.Churn {
			out = append(out, Signal{Kind: sigChurn, Label: file, FP: hash12("churn", file), Events: g})
		}
		if pair := findRevert(g); pair != nil {
			out = append(out, Signal{Kind: sigRevert, Label: file, FP: hash12("revert", file), Events: pair})
		}
	}

	var corrections []Event
	for _, e := range events {
		if e.Kind == "correction" && e.Conf >= thresholds.Correction {
			corrections = append(corrections, e)
		}
	}
	if len(corrections) > 0 {
		out = append(out, Signal{Kind: sigCorrection, Label: "user corrections", FP: hash12("correction"), Events: corrections})
	}
	return out
}

func findRevert(edits []Event) []Event {
	for j := range edits {
		for i := range j {
			if edits[j].NewHash == edits[i].OldHash && edits[j].OldHash == edits[i].NewHash {
				return []Event{edits[i], edits[j]}
			}
		}
	}
	return nil
}

func bySession(events []Event) map[string][]Event {
	return groupBy(events, func(e Event) string { return e.SessionID })
}

func priorSessions(events []Event, exclude string) map[string]int {
	prior := map[string]int{}
	for sid, evs := range bySession(events) {
		if sid == exclude {
			continue
		}
		seen := map[string]bool{}
		for _, sig := range derive(evs) {
			if !seen[sig.FP] {
				seen[sig.FP] = true
				prior[sig.FP]++
			}
		}
	}
	return prior
}

func instrHashes(events []Event) map[string]string {
	sorted := slices.Clone(events)
	slices.SortStableFunc(sorted, func(a, b Event) int { return a.TS.Compare(b.TS) })
	hashes := map[string]string{}
	for _, e := range sorted {
		if e.Kind == "session" {
			hashes[e.SessionID] = e.InstrHash
		}
	}
	return hashes
}
