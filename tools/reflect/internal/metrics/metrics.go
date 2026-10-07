package metrics

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/digest"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/repo"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
)

const verdictMin = 5

type Options struct {
	Projects string
	Records  []record.Record
	Root     string
	Exclude  string
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

func Report(o Options) string {
	records := inRepo(o.Records, o.Root)
	byHash := map[string]*tally{}
	var order []string
	for _, r := range records {
		if r.SessionID == o.Exclude {
			continue
		}
		s, err := session.Load(o.Projects, r.SessionID)
		if err != nil {
			continue
		}
		if byHash[r.InstrHash] == nil {
			byHash[r.InstrHash] = &tally{}
			order = append(order, r.InstrHash)
		}
		byHash[r.InstrHash].add(sessionTally(s))
	}
	if len(byHash) == 0 {
		return "no recorded sessions with transcripts for " + o.Root + "\n"
	}
	slices.SortStableFunc(order, func(a, b string) int { return cmp.Compare(byHash[b].sessions, byHash[a].sessions) })
	var w strings.Builder
	fmt.Fprintf(&w, "%-14s %5s %11s %14s %10s %13s\n", "instr", "n", "tool_calls", "confusion/100", "corr_rate", "repeat/sess")
	for _, h := range order {
		t := byHash[h]
		note := ""
		if t.sessions < verdictMin {
			note = fmt.Sprintf("  (n<%d: no verdict)", verdictMin)
		}
		fmt.Fprintf(&w, "%-14s %5d %11d %14.2f %10.2f %13.2f%s\n", cmp.Or(h, "-"), t.sessions, t.toolCalls,
			100*ratio(t.confusion, t.toolCalls), ratio(t.corrections, t.prompts), ratio(t.repeatFail, t.sessions), note)
		if prev, cur, ok := firstChange(records, h); ok {
			fmt.Fprintf(&w, "  changed vs %s: %s\n", prev.InstrHash, changeSummary(prev.Files, cur.Files))
		}
	}
	return w.String()
}

func inRepo(records []record.Record, root string) []record.Record {
	seen := map[string]bool{}
	var out []record.Record
	for _, r := range records {
		if seen[r.SessionID] || repo.MainCheckout(r.Cwd) != root {
			continue
		}
		seen[r.SessionID] = true
		out = append(out, r)
	}
	slices.SortStableFunc(out, func(a, b record.Record) int { return a.TS.Compare(b.TS) })
	return out
}

func sessionTally(s session.Session) tally {
	t := tally{sessions: 1}
	for _, a := range s.Agents {
		d := digest.Build(a)
		t.toolCalls += d.Calls
		t.prompts += d.Prompts
		distinct := map[string]bool{}
		for _, sig := range d.Signals {
			switch sig.Tag {
			case digest.Halluc:
				t.confusion++
			case digest.Correction:
				t.corrections++
			case digest.Repeat, digest.Churn, digest.Revert:
				if k := sig.Tag + sig.Key; !distinct[k] {
					distinct[k] = true
					t.confusion++
					if sig.Tag == digest.Repeat {
						t.repeatFail++
					}
				}
			}
		}
	}
	return t
}

func firstChange(records []record.Record, hash string) (prev, cur record.Record, ok bool) {
	for i, r := range records {
		if r.InstrHash != hash {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			if records[j].InstrHash != hash {
				return records[j], r, true
			}
		}
		return record.Record{}, record.Record{}, false
	}
	return record.Record{}, record.Record{}, false
}

func changeSummary(prev, cur map[string]string) string {
	const shown = 4
	var names []string
	for _, p := range sortedKeys(cur) {
		switch old, had := prev[p]; {
		case !had:
			names = append(names, "+"+p)
		case old != cur[p]:
			names = append(names, "~"+p)
		}
	}
	for _, p := range sortedKeys(prev) {
		if _, has := cur[p]; !has {
			names = append(names, "-"+p)
		}
	}
	if len(names) > shown {
		names = append(names[:shown], fmt.Sprintf("… %d more", len(names)-shown))
	}
	return strings.Join(names, " ")
}

func sortedKeys(m map[string]string) []string {
	return slices.Sorted(maps.Keys(m))
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}
