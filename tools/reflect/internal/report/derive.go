package report

import (
	"slices"
	"sort"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

const (
	repeatFailMin = 3
	churnMin      = 4
	recurMin      = 2
	priorMin      = 2
	correctionMin = 0.6
)

const (
	hallucination = "hallucination"
	repeatFail    = "repeat_fail"
	churn         = "churn"
	revert        = "revert"
	correction    = "correction"
)

type signal struct {
	kind   string
	label  string
	fp     string
	events []logstore.Event
	prior  int
}

func (s signal) qualifies() bool {
	switch s.kind {
	case hallucination, correction:
		return len(s.events) >= recurMin || s.prior >= priorMin
	}
	return true
}

func groupBy(events []logstore.Event, key func(logstore.Event) string) map[string][]logstore.Event {
	groups := map[string][]logstore.Event{}
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

func ofKind(kind string, field func(logstore.Event) string) func(logstore.Event) string {
	return func(e logstore.Event) string {
		if e.Kind == kind {
			return field(e)
		}
		return ""
	}
}

func derive(events []logstore.Event) []signal {
	var out []signal

	errs := groupBy(events, ofKind("tool_error", func(e logstore.Event) string { return e.FP }))
	for _, fp := range sortedKeys(errs) {
		g := errs[fp]
		label := g[0].Class + " " + g[0].Tool
		switch class := g[0].Class; {
		case detect.IsHallucination(class) || class == detect.ReadFirst:
			out = append(out, signal{hallucination, label, fp, g, 0})
		case (class == detect.Exit || class == detect.Other) && len(g) >= repeatFailMin:
			out = append(out, signal{repeatFail, label, fp, g, 0})
		}
	}

	edits := groupBy(events, ofKind("edit", func(e logstore.Event) string { return e.File }))
	for _, file := range sortedKeys(edits) {
		g := edits[file]
		if len(g) >= churnMin {
			out = append(out, signal{churn, file, "churn|" + file, g, 0})
		}
		if pair := findRevert(g); pair != nil {
			out = append(out, signal{revert, file, "revert|" + file, pair, 0})
		}
	}

	var corrections []logstore.Event
	for _, e := range events {
		if e.Kind == "correction" && e.Conf >= correctionMin {
			corrections = append(corrections, e)
		}
	}
	if len(corrections) > 0 {
		out = append(out, signal{correction, "user corrections", "correction", corrections, 0})
	}
	return out
}

func findRevert(edits []logstore.Event) []logstore.Event {
	for j := range edits {
		for i := range j {
			if edits[j].NewHash == edits[i].OldHash && edits[j].OldHash == edits[i].NewHash {
				return []logstore.Event{edits[i], edits[j]}
			}
		}
	}
	return nil
}

func bySession(events []logstore.Event) map[string][]logstore.Event {
	return groupBy(events, func(e logstore.Event) string { return e.SessionID })
}

func priorSessions(events []logstore.Event, exclude string) map[string]int {
	prior := map[string]int{}
	for sid, evs := range bySession(events) {
		if sid == exclude {
			continue
		}
		seen := map[string]bool{}
		for _, sig := range derive(evs) {
			if !seen[sig.fp] {
				seen[sig.fp] = true
				prior[sig.fp]++
			}
		}
	}
	return prior
}

func instrHashes(events []logstore.Event) map[string]string {
	sorted := slices.Clone(events)
	slices.SortStableFunc(sorted, func(a, b logstore.Event) int { return a.TS.Compare(b.TS) })
	hashes := map[string]string{}
	for _, e := range sorted {
		if e.Kind == "session" {
			hashes[e.SessionID] = e.InstrHash
		}
	}
	return hashes
}
