package report

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

type patternStats struct {
	pattern string
	n       int
	lo, hi  float64
	example string
}

func Corrections(s logstore.Store) (string, error) {
	events, err := s.Read()
	if err != nil {
		return "", err
	}
	byPattern := map[string]*patternStats{}
	for _, e := range events {
		if e.Kind != "correction" {
			continue
		}
		st, ok := byPattern[e.Class]
		if !ok {
			st = &patternStats{pattern: e.Class, lo: e.Conf, hi: e.Conf, example: oneLine(e.Input)}
			byPattern[e.Class] = st
		}
		st.n++
		st.lo, st.hi = min(st.lo, e.Conf), max(st.hi, e.Conf)
	}
	if len(byPattern) == 0 {
		return "no corrections logged\n", nil
	}
	stats := slices.Collect(maps.Values(byPattern))
	slices.SortFunc(stats, func(a, b *patternStats) int {
		return cmp.Or(cmp.Compare(b.n, a.n), strings.Compare(a.pattern, b.pattern))
	})
	var w strings.Builder
	fmt.Fprintf(&w, "%-24s %5s %-11s %s\n", "pattern", "n", "conf", "example")
	for _, st := range stats {
		conf := fmt.Sprintf("%.2f", st.hi)
		if st.lo != st.hi {
			conf = fmt.Sprintf("%.2f-%.2f", st.lo, st.hi)
		}
		fmt.Fprintf(&w, "%-24.24s %5d %-11s %s\n", st.pattern, st.n, conf, clip(st.example, logWidth))
	}
	return w.String(), nil
}
