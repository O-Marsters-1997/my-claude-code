package scan

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/digest"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

var fileHeading = regexp.MustCompile("(?m)^\\*\\*`([^`]+)`\\*\\*")

const (
	maxHeadings   = 20
	sliceTextMax  = 2000
	sliceInputMax = 1000
)

func Write(s session.Session, dir string, limit int) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ds := make([]digest.Digest, len(s.Agents))
	for i, a := range s.Agents {
		ds[i] = digest.Build(a)
	}
	clusters := digest.Clusters(ds)
	verdicts := digest.Triage(ds, limit)
	var w strings.Builder
	fmt.Fprintf(&w, "session %s: %d agents, review cap %d\ndigests: %s/<agent>.txt\n", s.ID, len(s.Agents), limit, dir)
	fmt.Fprintf(&w, "%-18s %-16s %-16s %5s %-9s %-6s %5s %5s %7s %-8s %s\n",
		"agent", "type", "model", "depth", "link", "parent", "calls", "fail", "tok_in", "verdict", "signals")
	for _, d := range ds {
		a := d.Agent
		path := filepath.Join(dir, a.ID+".txt")
		if err := os.WriteFile(path, []byte(d.Render()), 0o644); err != nil {
			return "", err
		}
		fmt.Fprintf(&w, "%-18s %-16s %-16s %5d %-9s %-6s %5d %5d %7s %-8s %s\n",
			a.ID, a.Type, cmp.Or(a.Model, "?"), a.Depth, cmp.Or(a.Link, "-"), short(cmp.Or(a.Parent, "-")),
			d.Calls, d.Fails, digest.Kilo(d.TokensIn), verdicts[a.ID], signalCounts(d))
	}
	for _, c := range clusters {
		fmt.Fprintf(&w, "cluster %s: %d instances, %d agents:", c.Mechanism, len(c.Instances), len(c.Agents()))
		for _, i := range c.Instances {
			fmt.Fprintf(&w, " %s L%d", i.Agent, i.Line)
		}
		w.WriteString("\n")
	}
	for _, warn := range s.Warnings {
		fmt.Fprintf(&w, "warn: %s\n", warn)
	}
	return w.String(), nil
}

func short(id string) string {
	if len(id) > 6 {
		return id[:6]
	}
	return id
}

func signalCounts(d digest.Digest) string {
	counts := map[string]int{}
	var order []string
	for _, s := range d.Signals {
		if counts[s.Tag] == 0 {
			order = append(order, s.Tag)
		}
		counts[s.Tag]++
	}
	if len(order) == 0 {
		return "-"
	}
	parts := make([]string, len(order))
	for i, tag := range order {
		parts[i] = fmt.Sprintf("%s:%d", tag, counts[tag])
	}
	return strings.Join(parts, ",")
}

func Slice(s session.Session, agentID string, line, context int) (string, error) {
	a := s.Agent(agentID)
	if a == nil {
		return "", fmt.Errorf("no agent %q in session %s", agentID, s.ID)
	}
	var w strings.Builder
	for _, l := range a.Lines {
		if l.N < line-context || l.N > line+context || l.Message.Content == nil && l.Subtype == "" {
			continue
		}
		fmt.Fprintf(&w, "L%d %s%s\n", l.N, l.Type, render(l))
	}
	if w.Len() == 0 {
		return "", fmt.Errorf("agent %s has no lines near L%d", agentID, line)
	}
	return w.String(), nil
}

func render(l transcript.Line) string {
	text, blocks := l.Parts()
	var w strings.Builder
	if strings.TrimSpace(text) != "" {
		fmt.Fprintf(&w, " text: %s", redact.Clean(text, sliceTextMax))
	}
	for _, b := range blocks {
		switch b.Type {
		case "tool_use":
			fmt.Fprintf(&w, " tool_use %s %s: %s", b.Name, b.ID, redact.Clean(compact(b.Input), sliceInputMax))
		case "tool_result":
			status := "ok"
			if b.IsError {
				status = "error"
			}
			text := b.ResultText()
			fmt.Fprintf(&w, " tool_result %s %s: %s%s", b.ToolUseID, status, redact.Clean(text, sliceTextMax), headings(text))
		}
	}
	if l.Subtype != "" {
		fmt.Fprintf(&w, " %s", l.Subtype)
	}
	return w.String()
}

func headings(text string) string {
	if len(text) <= sliceTextMax {
		return ""
	}
	var files []string
	for _, m := range fileHeading.FindAllStringSubmatch(text, -1) {
		files = append(files, m[1])
	}
	if len(files) == 0 {
		return ""
	}
	return " [files: " + strings.Join(files[:min(len(files), maxHeadings)], ", ") + "]"
}

func compact(raw json.RawMessage) string {
	return strings.Join(strings.Fields(string(raw)), " ")
}
