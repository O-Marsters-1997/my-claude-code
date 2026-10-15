package harvest

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
)

type fixHunk struct {
	oldStart int
	removed  []string
	added    []string
}

func (h fixHunk) text() string {
	lines := make([]string, 0, len(h.removed)+len(h.added))
	for _, l := range h.removed {
		lines = append(lines, "-"+l)
	}
	for _, l := range h.added {
		lines = append(lines, "+"+l)
	}
	return strings.Join(lines, "\n")
}

func recordFixes(ctx context.Context, top, ledgerPath string, recorded []ledger.Learning) error {
	for _, l := range recorded {
		diff, err := git(ctx, top, "-c", "core.quotepath=off", "diff", "--cached", "-U0", "--no-color", "--no-ext-diff", "--", l.File)
		if err != nil {
			return err
		}
		h, ok := pickHunk(parseHunks(diff), l)
		if !ok {
			continue
		}
		after := redact.Clean(h.text(), maxField)
		if err := ledger.Append(ledgerPath, ledger.Learning{ID: l.ID, After: after}); err != nil {
			return err
		}
	}
	return nil
}

func parseHunks(diff string) []fixHunk {
	var out []fixHunk
	for _, l := range strings.Split(diff, "\n") {
		if hunk.MatchString(l) {
			out = append(out, fixHunk{oldStart: oldStart(l)})
			continue
		}
		if len(out) == 0 {
			continue
		}
		h := &out[len(out)-1]
		switch {
		case strings.HasPrefix(l, "-"):
			h.removed = append(h.removed, l[1:])
		case strings.HasPrefix(l, "+"):
			h.added = append(h.added, l[1:])
		}
	}
	return out
}

func oldStart(header string) int {
	rest := strings.TrimPrefix(header, "@@ -")
	num, _, _ := strings.Cut(rest, " ")
	num, _, _ = strings.Cut(num, ",")
	n, _ := strconv.Atoi(num)
	return n
}

func pickHunk(hunks []fixHunk, l ledger.Learning) (fixHunk, bool) {
	target := strings.TrimSpace(l.TargetText)
	if target == "" {
		return fixHunk{}, false
	}
	for _, h := range hunks {
		if slices.ContainsFunc(h.removed, func(r string) bool { return strings.Contains(redact.Clean(r, maxField), target) }) {
			return h, true
		}
	}
	return fixHunk{}, false
}
