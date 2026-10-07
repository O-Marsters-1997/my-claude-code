package digest

import (
	"cmp"
	"slices"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
)

const (
	Review   = "review"
	Skip     = "skip"
	Overflow = "overflow"
)

func (d Digest) actionable() int {
	n := 0
	for _, s := range d.Signals {
		if s.Tag != Big && s.Tag != Reread && s.Cluster == "" {
			n++
		}
	}
	return n
}

func (d Digest) clustered() bool {
	return slices.ContainsFunc(d.Signals, func(s Signal) bool { return s.Cluster != "" })
}

func Triage(ds []Digest, limit int) map[string]string {
	heavy := heavyThreshold(ds)
	var wanted []Digest
	out := map[string]string{}
	for _, d := range ds {
		out[d.Agent.ID] = Skip
		if d.Agent.ID == session.MainID || d.actionable() > 0 || d.TokensIn > heavy && !d.clustered() {
			wanted = append(wanted, d)
		}
	}
	slices.SortStableFunc(wanted, func(a, b Digest) int {
		if aMain, bMain := a.Agent.ID == session.MainID, b.Agent.ID == session.MainID; aMain != bMain {
			if aMain {
				return -1
			}
			return 1
		}
		return cmp.Or(cmp.Compare(b.actionable(), a.actionable()), cmp.Compare(b.TokensIn, a.TokensIn))
	})
	for i, d := range wanted {
		if i < limit {
			out[d.Agent.ID] = Review
		} else {
			out[d.Agent.ID] = Overflow
		}
	}
	return out
}

func heavyThreshold(ds []Digest) int {
	tokens := make([]int, len(ds))
	for i, d := range ds {
		tokens[i] = d.TokensIn
	}
	slices.Sort(tokens)
	if len(tokens) == 0 {
		return 0
	}
	median := tokens[len(tokens)/2]
	quartile := tokens[len(tokens)*3/4]
	return max(median, quartile-1)
}
