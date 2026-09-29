package hook

import (
	"cmp"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

type payload struct {
	HookEventName       string          `json:"hook_event_name"`
	SessionID           string          `json:"session_id"`
	TranscriptPath      string          `json:"transcript_path"`
	AgentTranscriptPath string          `json:"agent_transcript_path"`
	Cwd                 string          `json:"cwd"`
	AgentID             string          `json:"agent_id"`
	AgentType           string          `json:"agent_type"`
	ToolName            string          `json:"tool_name"`
	ToolInput           json.RawMessage `json:"tool_input"`
	ToolUseID           string          `json:"tool_use_id"`
	Error               string          `json:"error"`
	Prompt              string          `json:"prompt"`
	Source              string          `json:"source"`
}

type handler struct {
	store      logstore.Store
	p          payload
	projectDir string
}

func Run(in io.Reader, projectDir string) {
	defer func() { _ = recover() }()
	var p payload
	if json.NewDecoder(in).Decode(&p) != nil {
		return
	}
	dir := cmp.Or(projectDir, p.Cwd)
	h := handler{store: logstore.New(dir), p: p, projectDir: dir}
	switch p.HookEventName {
	case "SessionStart":
		h.sessionStart()
	case "UserPromptSubmit":
		h.correction()
	case "PostToolUseFailure":
		h.toolFailure()
	case "PostToolUse":
		h.edit()
	case "Stop":
		sweep(h.store, p.TranscriptPath, p.SessionID)
	case "SubagentStop":
		sweep(h.store, p.AgentTranscriptPath, p.SessionID)
	case "SessionEnd":
		h.sessionEnd()
	}
}

func (h handler) event(kind string) logstore.Event {
	return logstore.Event{
		Kind:      kind,
		SessionID: h.p.SessionID,
		AgentID:   h.p.AgentID,
		AgentType: h.p.AgentType,
	}
}

func (h handler) sessionStart() {
	files := instrFiles(h.projectDir)
	hash := combinedHash(files)
	branch := git(h.projectDir, "branch", "--show-current")
	prev := loadInstrState(h.store.Dir())
	if mark := seenMark(hash, branch); h.p.Source == "compact" && prev.Seen[h.p.SessionID] == mark {
		return
	}
	e := h.event("session")
	e.Class = h.p.Source
	e.Cwd = h.p.Cwd
	e.Transcript = h.p.TranscriptPath
	e.Commit = git(h.projectDir, "rev-parse", "--short", "HEAD")
	e.Branch = branch
	e.Dirty = git(h.projectDir, "status", "--porcelain") != ""
	e.InstrHash = hash
	_ = h.store.Append(e)
	if added, changed, removed := diffFiles(prev.Files, files); prev.Hash != "" && len(added)+len(changed)+len(removed) > 0 {
		m := h.event("manifest")
		m.InstrHash = hash
		m.Prev = prev.Hash
		m.Added, m.Changed, m.Removed = added, changed, removed
		_ = h.store.Append(m)
	}
	seen := prev.Seen
	if seen == nil || len(seen) >= seenSessionsMax {
		seen = map[string]string{}
	}
	seen[h.p.SessionID] = seenMark(hash, branch)
	saveInstrState(h.store.Dir(), instrState{Hash: hash, Files: files, Seen: seen})
}

func (h handler) correction() {
	if !detect.IncludeMessage(h.p.Prompt) {
		return
	}
	names, conf := detect.Correction(h.p.Prompt)
	if len(names) == 0 {
		return
	}
	e := h.event("correction")
	e.Class = strings.Join(names, " ")
	e.Conf = conf
	e.Input = redact.Clean(h.p.Prompt, 500)
	if use, ok := transcript.LastToolUse(h.p.TranscriptPath); ok {
		e.ToolUseID, e.Tool = use.ID, use.Name
	}
	_ = h.store.Append(e)
}

func (h handler) toolFailure() {
	class := detect.Classify(h.p.ToolName, h.p.Error)
	e := h.event("tool_error")
	e.ToolUseID = h.p.ToolUseID
	e.Tool = h.p.ToolName
	e.Class = class
	e.FP = detect.Fingerprint(h.p.ToolName, class, h.p.ToolInput, h.p.Error)
	e.Input = redact.Clean(string(h.p.ToolInput), inputMax)
	e.Error = redact.Tail(h.p.Error, errorMax)
	loadLedger(h.store, h.p.SessionID).append(e)
}

func (h handler) edit() {
	if h.p.ToolName != "Edit" && h.p.ToolName != "Write" {
		return
	}
	var in struct {
		FilePath  string `json:"file_path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
		Content   string `json:"content"`
	}
	if json.Unmarshal(h.p.ToolInput, &in) != nil || in.FilePath == "" {
		return
	}
	e := h.event("edit")
	e.TS = time.Now().UTC()
	e.ToolUseID = h.p.ToolUseID
	e.Tool = h.p.ToolName
	e.File = in.FilePath
	e.OldHash = hash12(in.OldString)
	e.NewHash = hash12(cmp.Or(in.NewString, in.Content))
	h.logEdit(e)
}

func (h handler) sessionEnd() {
	subagents := transcript.Subagents(h.p.TranscriptPath)
	for _, path := range append([]string{h.p.TranscriptPath}, subagents...) {
		sweep(h.store, path, h.p.SessionID)
	}
	calls, prompts := transcript.Count(h.p.TranscriptPath, true)
	for _, path := range subagents {
		c, _ := transcript.Count(path, false)
		calls += c
	}
	e := h.event("session_end")
	e.ToolCalls = calls
	e.Prompts = prompts
	e.Errors = len(kindsOf(h.store, h.p.SessionID, "tool_error"))
	_ = h.store.Append(e)
	if dir := h.editsDir(); dir != "" {
		_ = os.RemoveAll(dir)
	}
}

func kindsOf(store logstore.Store, sessionID, kind string) []logstore.Event {
	events, _ := store.Read()
	var out []logstore.Event
	for _, e := range events {
		if e.Kind == kind && e.SessionID == sessionID {
			out = append(out, e)
		}
	}
	return out
}
