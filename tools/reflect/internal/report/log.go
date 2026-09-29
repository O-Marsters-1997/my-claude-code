package report

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

const logWidth = 120

type LogFilter struct {
	Session string
	Kind    string
	Last    int
}

func Log(s logstore.Store, f LogFilter) (string, error) {
	events, err := s.Read()
	if err != nil {
		return "", err
	}
	var picked []logstore.Event
	for _, e := range events {
		if f.Kind != "" && e.Kind != f.Kind {
			continue
		}
		if f.Session != "" && !strings.HasPrefix(e.SessionID, f.Session) {
			continue
		}
		picked = append(picked, e)
	}
	if f.Last > 0 && len(picked) > f.Last {
		picked = picked[len(picked)-f.Last:]
	}
	var w strings.Builder
	for _, e := range picked {
		fmt.Fprintln(&w, clip(logLine(e), logWidth))
	}
	return w.String(), nil
}

func logLine(e logstore.Event) string {
	return fmt.Sprintf("%s %-11s %-8.8s %-5.5s %s",
		e.TS.Local().Format("01-02 15:04:05"), e.Kind, e.SessionID, cmp.Or(e.AgentID, "main"), logDetail(e))
}

func logDetail(e logstore.Event) string {
	switch e.Kind {
	case "session":
		dirty := ""
		if e.Dirty {
			dirty = " dirty"
		}
		return fmt.Sprintf("%s %s@%s instr=%s%s", e.Class, e.Branch, e.Commit, e.InstrHash, dirty)
	case "manifest":
		return fmt.Sprintf("instr=%s prev=%s +%d ~%d -%d", e.InstrHash, cmp.Or(e.Prev, "-"), len(e.Added), len(e.Changed), len(e.Removed))
	case "session_end":
		return fmt.Sprintf("calls=%d prompts=%d errors=%d", e.ToolCalls, e.Prompts, e.Errors)
	case "tool_error":
		return strings.TrimSpace(fmt.Sprintf("%s %s %s", e.Class, e.Tool, oneLine(cmp.Or(e.Error, e.Input))))
	case "edit":
		return e.Tool + " " + e.File
	case "correction":
		return fmt.Sprintf("%s conf=%.2f %s", e.Class, e.Conf, oneLine(e.Input))
	case "summary":
		return fmt.Sprintf("calls=%d confusion=%d corrections=%d instr=%s", e.ToolCalls, e.Confusion, e.Corrections, e.InstrHash)
	}
	return e.Class
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func clip(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}
