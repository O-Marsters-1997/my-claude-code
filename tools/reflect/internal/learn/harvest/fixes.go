package harvest

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/marker"
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

func recordFixes(ctx context.Context, top, root, ledgerPath string) error {
	known, err := ledger.Read(ledgerPath)
	if err != nil {
		return err
	}
	for _, l := range known {
		if l.Kind != "fix" || l.Status != "pending" || l.After != "" || l.Repo != root || l.File == "" {
			continue
		}
		content, err := git(ctx, top, "show", ":"+l.File)
		if err != nil || markerPresent(content, root, l) {
			continue
		}
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

func markerPresent(content, root string, l ledger.Learning) bool {
	lines := strings.Split(strings.ReplaceAll(strings.TrimSuffix(content, "\n"), "\r\n", "\n"), "\n")
	for _, m := range marker.Parse(lines) {
		if !m.Later && learningID(root, l.File, redact.Clean(m.Text, maxField)) == l.ID {
			return true
		}
	}
	return false
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
	if len(hunks) == 0 {
		return fixHunk{}, false
	}
	if target := strings.TrimSpace(l.TargetText); target != "" {
		for _, h := range hunks {
			if slices.ContainsFunc(h.removed, func(r string) bool { return strings.Contains(r, target) }) {
				return h, true
			}
		}
	}
	best := hunks[0]
	for _, h := range hunks[1:] {
		if abs(h.oldStart-l.Line) < abs(best.oldStart-l.Line) {
			best = h
		}
	}
	return best, true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
