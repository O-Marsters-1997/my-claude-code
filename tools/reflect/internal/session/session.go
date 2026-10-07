package session

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

const (
	MainID        = "main"
	LinkToolUseID = "toolUseId"
	LinkAgentID   = "agentId"
	LinkPath      = "path"
	outcomeMax    = 300
)

var (
	ErrNotFound    = errors.New("session transcript not found")
	notification   = regexp.MustCompile(`<task-id>([\w-]+)</task-id>[\s\S]*?<status>(\w+)</status>`)
	reflectCommand = regexp.MustCompile(`<command-name>/(?:[\w-]+:)?reflect</command-name>`)
)

type Agent struct {
	ID          string
	Type        string
	Model       string
	Description string
	Parent      string
	ToolUseID   string
	Link        string
	Path        string
	Brief       string
	Outcome     string
	Depth       int
	Lines       []transcript.Line
}

type Session struct {
	ID       string
	Agents   []*Agent
	Warnings []string
}

func (s Session) Main() *Agent { return s.Agents[0] }

func (s Session) Agent(id string) *Agent {
	for _, a := range s.Agents {
		if a.ID == id {
			return a
		}
	}
	return nil
}

type LiveError struct{ Agents []string }

func (e *LiveError) Error() string {
	return "agents still running: " + strings.Join(e.Agents, ", ") + "; wait for them to finish"
}

type meta struct {
	AgentType     string `json:"agentType"`
	Description   string `json:"description"`
	ToolUseID     string `json:"toolUseId"`
	ParentAgentID string `json:"parentAgentId"`
	Model         string `json:"model"`
	WorktreePath  string `json:"worktreePath"`
}

type launch struct {
	AgentID string `json:"agentId"`
	IsAsync bool   `json:"isAsync"`
}

type site struct {
	agent string
	idx   int
}

type loader struct {
	projects string
	agents   map[string]*Agent
	metas    map[string]meta
	uses     map[string][]site
	results  map[string][]site
	launched map[string]string
	statuses map[string]string
	warnings []string
	keptMain map[int]bool
}

func Load(projectsDir, sid string) (Session, error) {
	matches, _ := filepath.Glob(filepath.Join(projectsDir, "*", sid+".jsonl"))
	if len(matches) == 0 {
		return Session{}, fmt.Errorf("%w: %s", ErrNotFound, sid)
	}
	mainPath := matches[0]
	l := &loader{
		projects: projectsDir,
		agents:   map[string]*Agent{},
		metas:    map[string]meta{},
		uses:     map[string][]site{},
		results:  map[string][]site{},
		launched: map[string]string{},
		statuses: map[string]string{},
		keptMain: map[int]bool{},
	}
	lines, err := transcript.Read(mainPath)
	if err != nil {
		return Session{}, err
	}
	main := &Agent{ID: MainID, Type: MainID, Path: mainPath, Lines: lines}
	l.agents[MainID] = main
	kept := cutReflectTurns(lines)
	for _, line := range kept {
		l.keptMain[line.N] = true
	}
	if err := l.readSubagents(filepath.Join(strings.TrimSuffix(mainPath, ".jsonl"), "subagents")); err != nil {
		return Session{}, err
	}
	l.index()
	l.link()
	s := Session{ID: sid, Agents: l.ordered(), Warnings: l.warnings}
	main.Lines = kept
	if live := l.live(s.Agents); len(live) > 0 {
		return s, &LiveError{Agents: live}
	}
	return s, nil
}

func cutReflectTurns(lines []transcript.Line) []transcript.Line {
	last := -1
	for i, l := range lines {
		if isReflect(l) {
			last = i
		}
	}
	if last >= 0 {
		lines = lines[:last]
	}
	var kept []transcript.Line
	inReflect := false
	for _, l := range lines {
		switch {
		case isReflect(l):
			inReflect = true
		case l.IsPrompt():
			inReflect = false
		}
		if !inReflect {
			kept = append(kept, l)
		}
	}
	return kept
}

func isReflect(l transcript.Line) bool {
	if l.Type != "user" {
		return false
	}
	text, _ := l.Parts()
	return reflectCommand.MatchString(text)
}

func (l *loader) readSubagents(dir string) error {
	metaPaths, _ := filepath.Glob(filepath.Join(dir, "agent-*.meta.json"))
	for _, p := range metaPaths {
		var m meta
		if b, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(b, &m)
		}
		l.metas[agentID(p, ".meta.json")] = m
	}
	paths, _ := filepath.Glob(filepath.Join(dir, "agent-*.jsonl"))
	for id := range l.metas {
		if p := filepath.Join(dir, "agent-"+id+".jsonl"); !slices.Contains(paths, p) {
			paths = append(paths, p)
		}
	}
	slices.Sort(paths)
	for _, p := range paths {
		id := agentID(p, ".jsonl")
		a := &Agent{ID: id, Path: p}
		lines, err := transcript.Read(p)
		switch {
		case errors.Is(err, os.ErrNotExist):
			a.Path = ""
		case err != nil:
			return err
		}
		a.Lines = lines
		l.agents[id] = a
	}
	return nil
}

func agentID(path, suffix string) string {
	return strings.TrimPrefix(strings.TrimSuffix(filepath.Base(path), suffix), "agent-")
}

func (l *loader) index() {
	for id, a := range l.agents {
		for i, line := range a.Lines {
			for _, m := range notification.FindAllSubmatch(line.Raw, -1) {
				if status := string(m[2]); status != "running" {
					l.statuses[string(m[1])] = status
				}
			}
			_, blocks := line.Parts()
			for _, b := range blocks {
				switch b.Type {
				case "tool_use":
					l.uses[b.ID] = append(l.uses[b.ID], site{id, i})
				case "tool_result":
					l.results[b.ToolUseID] = append(l.results[b.ToolUseID], site{id, i})
					var r launch
					if json.Unmarshal(line.ToolUseResult, &r) == nil && r.AgentID != "" {
						l.launched[r.AgentID] = b.ToolUseID
					}
				}
			}
		}
	}
}

func (l *loader) link() {
	for id, a := range l.agents {
		if id == MainID {
			continue
		}
		m := l.metas[id]
		a.Type = cmp.Or(m.AgentType, "unknown")
		a.Description = m.Description
		if m.Model != "inherit" {
			a.Model = m.Model
		}
		preferred := cmp.Or(m.ParentAgentID, MainID)
		switch {
		case m.ToolUseID != "" && l.spawnSite(m.ToolUseID, id, preferred) != nil:
			a.ToolUseID, a.Link = m.ToolUseID, LinkToolUseID
		case l.launched[id] != "" && l.spawnSite(l.launched[id], id, preferred) != nil:
			a.ToolUseID, a.Link = l.launched[id], LinkAgentID
		default:
			a.Parent, a.Link = preferred, LinkPath
			l.warnings = append(l.warnings, fmt.Sprintf("agent %s: no spawning call found, attached to %s by path", id, preferred))
		}
		if s := l.spawnSite(a.ToolUseID, id, preferred); s != nil {
			a.Parent = s.agent
			a.Brief = brief(l.agents[s.agent].Lines[s.idx], a.ToolUseID)
			a.Outcome = l.outcome(a)
		}
		if a.Path == "" && m.WorktreePath != "" {
			a.Path, a.Lines = l.worktreeTranscript(m.WorktreePath, a.Brief)
		}
		if a.Path == "" {
			l.warnings = append(l.warnings, fmt.Sprintf("agent %s: transcript not found", id))
		}
	}
	for _, a := range l.agents {
		if a.Model == "" {
			a.Model = firstModel(a.Lines)
		}
	}
}

func (l *loader) spawnSite(toolUseID, self, preferred string) *site {
	var found *site
	for _, s := range l.uses[toolUseID] {
		if s.agent == self {
			continue
		}
		if s.agent == preferred {
			return &s
		}
		if found == nil {
			found = &s
		}
	}
	return found
}

func brief(line transcript.Line, toolUseID string) string {
	_, blocks := line.Parts()
	for _, b := range blocks {
		if b.ID != toolUseID {
			continue
		}
		var in struct {
			Prompt string `json:"prompt"`
			Skill  string `json:"skill"`
			Args   string `json:"args"`
		}
		_ = json.Unmarshal(b.Input, &in)
		return cmp.Or(in.Prompt, strings.TrimSpace(in.Skill+" "+in.Args))
	}
	return ""
}

func (l *loader) outcome(a *Agent) string {
	parent := l.agents[a.Parent]
	anchor := -1
	for _, s := range l.results[a.ToolUseID] {
		if s.agent == a.Parent {
			anchor = s.idx
		}
	}
	if anchor < 0 {
		return ""
	}
	var r launch
	if json.Unmarshal(parent.Lines[anchor].ToolUseResult, &r) == nil && r.IsAsync {
		anchor = notificationLine(parent.Lines, a.ID, anchor)
	}
	for _, line := range parent.Lines[anchor+1:] {
		if line.Type != "assistant" {
			continue
		}
		if text, _ := line.Parts(); strings.TrimSpace(text) != "" {
			return clip(text, outcomeMax)
		}
	}
	return ""
}

func notificationLine(lines []transcript.Line, aid string, from int) int {
	for i := from; i < len(lines); i++ {
		for _, m := range notification.FindAllSubmatch(lines[i].Raw, -1) {
			if string(m[1]) == aid {
				return i
			}
		}
	}
	return from
}

func (l *loader) worktreeTranscript(worktreePath, brief string) (string, []transcript.Line) {
	slug := strings.NewReplacer("/", "-", ".", "-").Replace(worktreePath)
	paths, _ := filepath.Glob(filepath.Join(l.projects, slug, "*.jsonl"))
	for _, p := range paths {
		lines, err := transcript.Read(p)
		if err != nil {
			continue
		}
		if len(paths) == 1 || firstPrompt(lines) == strings.TrimSpace(brief) {
			return p, lines
		}
	}
	return "", nil
}

func firstPrompt(lines []transcript.Line) string {
	for _, l := range lines {
		if l.IsPrompt() {
			text, _ := l.Parts()
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func firstModel(lines []transcript.Line) string {
	for _, l := range lines {
		if l.Message.Model != "" {
			return l.Message.Model
		}
	}
	return ""
}

func (l *loader) ordered() []*Agent {
	children := map[string][]*Agent{}
	spawnedAt := map[string]int{}
	for _, a := range l.agents {
		if a.ID == MainID {
			continue
		}
		children[a.Parent] = append(children[a.Parent], a)
		spawnedAt[a.ID] = len(l.agents[a.Parent].Lines) + 1
		if s := l.spawnSite(a.ToolUseID, a.ID, a.Parent); s != nil {
			spawnedAt[a.ID] = s.idx
		}
	}
	var out []*Agent
	var walk func(a *Agent, depth int)
	walk = func(a *Agent, depth int) {
		a.Depth = depth
		out = append(out, a)
		kids := children[a.ID]
		slices.SortFunc(kids, func(x, y *Agent) int {
			return cmp.Or(cmp.Compare(spawnedAt[x.ID], spawnedAt[y.ID]), cmp.Compare(x.ID, y.ID))
		})
		for _, k := range kids {
			if l.kept(k) {
				walk(k, depth+1)
			}
		}
	}
	walk(l.agents[MainID], 0)
	return out
}

func (l *loader) kept(a *Agent) bool {
	if a.Link == LinkPath {
		return true
	}
	s := l.spawnSite(a.ToolUseID, a.ID, a.Parent)
	if s == nil || s.agent != MainID {
		return true
	}
	return l.keptMain[l.agents[MainID].Lines[s.idx].N]
}

func (l *loader) live(agents []*Agent) []string {
	var live []string
	for _, a := range agents {
		if a.ID == MainID || a.Link == LinkPath {
			continue
		}
		results := l.results[a.ToolUseID]
		if len(results) == 0 {
			live = append(live, a.ID)
			continue
		}
		var r launch
		res := results[0]
		if json.Unmarshal(l.agents[res.agent].Lines[res.idx].ToolUseResult, &r) == nil && r.IsAsync && l.statuses[a.ID] == "" {
			live = append(live, a.ID)
		}
	}
	return live
}

func clip(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "") + "…"
}
