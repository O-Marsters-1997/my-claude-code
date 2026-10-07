package digest

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

const (
	Halluc     = "halluc"
	Fail       = "fail"
	Repeat     = "repeat"
	Churn      = "churn"
	Revert     = "revert"
	Reread     = "reread"
	Big        = "big"
	Correction = "correction"
)

const (
	repeatMin     = 3
	churnMin      = 4
	bigResult     = 20000
	correctionMin = 0.6
	briefMax      = 600
	targetMax     = 80
	targetPad     = 40
	sayMax        = 160
)

var leadingCd = regexp.MustCompile(`^cd \S+ && `)

type Signal struct {
	Tag   string
	Line  int
	Key   string
	Class string
}

func (s Signal) FP() string {
	sum := sha256.Sum256([]byte(s.Tag + "\x00" + s.Key))
	return hex.EncodeToString(sum[:4])
}

type Digest struct {
	Agent     *session.Agent
	Signals   []Signal
	Calls     int
	Fails     int
	TokensIn  int
	TokensOut int
	Prompts   int
	steps     []step
}

type step struct {
	line   int
	kind   string
	tool   string
	target string
	failed bool
	size   int
	text   string
}

type input struct {
	Command     string `json:"command"`
	FilePath    string `json:"file_path"`
	Path        string `json:"path"`
	Pattern     string `json:"pattern"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Skill       string `json:"skill"`
	URL         string `json:"url"`
	Query       string `json:"query"`
	OldString   string `json:"old_string"`
	NewString   string `json:"new_string"`
	Offset      int    `json:"offset"`
	Limit       int    `json:"limit"`
}

type edit struct{ oldS, newS string }

type builder struct {
	d         Digest
	uses      map[string]useSite
	failsByFP map[string]int
	edits     map[string][]edit
	reads     map[string]bool
	seenMsgs  map[string]bool
	stepAt    map[int]int
}

type useSite struct {
	line  int
	block transcript.Block
}

func Build(a *session.Agent) Digest {
	b := &builder{
		d:         Digest{Agent: a},
		uses:      map[string]useSite{},
		failsByFP: map[string]int{},
		edits:     map[string][]edit{},
		reads:     map[string]bool{},
		seenMsgs:  map[string]bool{},
		stepAt:    map[int]int{},
	}
	for _, l := range a.Lines {
		b.line(l)
	}
	return b.d
}

func (b *builder) line(l transcript.Line) {
	switch {
	case l.Type == "system" && l.Subtype == "compact_boundary":
		b.d.steps = append(b.d.steps, step{line: l.N, kind: "compact", text: fmt.Sprintf("compact_boundary (%s → %s)",
			Kilo(l.CompactMetadata.PreTokens), Kilo(l.CompactMetadata.PostTokens))})
	case l.Type == "assistant":
		b.assistant(l)
	case l.IsPrompt():
		b.prompt(l)
	case l.Type == "user":
		_, blocks := l.Parts()
		for _, blk := range blocks {
			if blk.Type == "tool_result" {
				b.result(blk)
			}
		}
	}
}

func (b *builder) assistant(l transcript.Line) {
	if id := l.Message.ID; id == "" || !b.seenMsgs[id] {
		b.seenMsgs[id] = true
		b.d.TokensIn += l.Message.Usage.In()
		b.d.TokensOut += l.Message.Usage.OutputTokens
	}
	text, blocks := l.Parts()
	if strings.TrimSpace(text) != "" {
		b.d.steps = append(b.d.steps, step{line: l.N, kind: "say", text: oneLine(redact.Clean(text, 4*sayMax), sayMax)})
	}
	for _, blk := range blocks {
		if blk.Type != "tool_use" {
			continue
		}
		b.d.Calls++
		b.uses[blk.ID] = useSite{l.N, blk}
		b.stepAt[l.N] = len(b.d.steps)
		b.d.steps = append(b.d.steps, step{line: l.N, kind: "call", tool: blk.Name, target: target(blk)})
	}
}

func (b *builder) prompt(l transcript.Line) {
	text, _ := l.Parts()
	if !detect.IncludeMessage(text) {
		return
	}
	b.d.Prompts++
	b.d.steps = append(b.d.steps, step{line: l.N, kind: "user", text: oneLine(redact.Clean(text, 4*sayMax), sayMax)})
	if b.d.Agent.ID != session.MainID {
		return
	}
	if names, conf := detect.Correction(text); conf >= correctionMin {
		b.tag(Correction, l.N, strings.Join(names, " "), "")
	}
}

func (b *builder) result(blk transcript.Block) {
	use, ok := b.uses[blk.ToolUseID]
	if !ok {
		return
	}
	text := blk.ResultText()
	st := &b.d.steps[b.stepAt[use.line]]
	st.size = len(text)
	var in input
	_ = json.Unmarshal(use.block.Input, &in)
	file := cmp.Or(in.FilePath, in.Path)
	if blk.IsError || strings.Contains(text, "<tool_use_error>") {
		st.failed = true
		b.failure(use, text)
		return
	}
	switch use.block.Name {
	case "Read":
		key := fmt.Sprintf("%s@%d+%d", file, in.Offset, in.Limit)
		if b.reads[key] {
			b.tag(Reread, use.line, file, "")
		}
		b.reads[key] = true
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
		b.edited(use.line, file, edit{in.OldString, in.NewString})
	}
	if st.size > bigResult {
		b.tag(Big, use.line, use.block.Name+"|"+st.target, "")
	}
}

func (b *builder) failure(use useSite, text string) {
	b.d.Fails++
	text = strings.NewReplacer("<tool_use_error>", "", "</tool_use_error>", "").Replace(text)
	class := detect.Classify(use.block.Name, text)
	key := detect.Fingerprint(use.block.Name, class, use.block.Input, text)
	if detect.IsHallucination(class) || class == detect.ReadFirst {
		b.tag(Halluc, use.line, key, class)
		return
	}
	b.tag(Fail, use.line, key, class)
	if b.failsByFP[key]++; b.failsByFP[key] >= repeatMin {
		b.tag(Repeat, use.line, key, class)
	}
}

func (b *builder) edited(line int, file string, e edit) {
	for k := range b.reads {
		if strings.HasPrefix(k, file+"@") {
			delete(b.reads, k)
		}
	}
	prior := b.edits[file]
	if e.oldS != "" && slices.ContainsFunc(prior, func(p edit) bool { return p.oldS == e.newS && p.newS == e.oldS }) {
		b.tag(Revert, line, file, "")
	}
	b.edits[file] = append(prior, e)
	if len(b.edits[file]) >= churnMin {
		b.tag(Churn, line, file, "")
	}
}

func (b *builder) tag(tag string, line int, key, class string) {
	b.d.Signals = append(b.d.Signals, Signal{Tag: tag, Line: line, Key: key, Class: class})
}

func (d Digest) Render() string {
	a := d.Agent
	var w strings.Builder
	fmt.Fprintf(&w, "agent %s %s %s depth=%d", a.ID, a.Type, cmp.Or(a.Model, "?"), a.Depth)
	if a.ID != session.MainID {
		fmt.Fprintf(&w, " link=%s parent=%s", a.Link, a.Parent)
	}
	fmt.Fprintf(&w, "\ntranscript: %s\n", a.Path)
	if a.Brief != "" {
		fmt.Fprintf(&w, "brief: %s\n", oneLine(a.Brief, briefMax))
	}
	if a.Outcome != "" {
		fmt.Fprintf(&w, "outcome: %s\n", a.Outcome)
	}
	tags := map[int][]string{}
	for _, s := range d.Signals {
		tags[s.Line] = append(tags[s.Line], fmt.Sprintf("%s fp=%s", s.Tag, s.FP()))
	}
	for _, st := range d.steps {
		switch st.kind {
		case "compact":
			fmt.Fprintf(&w, " ──── %s ────\n", st.text)
		case "say":
			fmt.Fprintf(&w, " L%-5d say   %s\n", st.line, st.text)
		case "user":
			fmt.Fprintf(&w, " L%-5d user  %s%s\n", st.line, st.text, bracket(tags[st.line]))
		case "call":
			status := "ok"
			if st.failed {
				status = "FAIL"
			}
			fmt.Fprintf(&w, " L%-5d %-6s %-*s %-4s %6s%s\n", st.line, st.tool, targetPad, st.target, status, Kilo(st.size), bracket(tags[st.line]))
		}
	}
	fmt.Fprintf(&w, "totals: %d calls, %d fail, %s tok in / %s out\n", d.Calls, d.Fails, Kilo(d.TokensIn), Kilo(d.TokensOut))
	return w.String()
}

func bracket(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "  [" + strings.Join(tags, "; ") + "]"
}

func target(b transcript.Block) string {
	var in input
	_ = json.Unmarshal(b.Input, &in)
	t := cmp.Or(leadingCd.ReplaceAllString(in.Command, ""), in.FilePath, in.Path, in.Pattern, in.URL, in.Query, in.Skill, in.Description, in.Prompt)
	return oneLine(redact.Clean(t, 4*targetMax), targetMax)
}

func oneLine(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "") + "…"
}

func Kilo(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprint(n)
	case n < 10000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%dk", (n+500)/1000)
}
