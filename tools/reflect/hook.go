package main

import (
	"cmp"
	"encoding/json"
	"io"
	"os"
	"strings"
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

type editInput struct {
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
	Content   string `json:"content"`
}

type hook struct {
	s          store
	p          payload
	projectDir string
}

func runHook(in io.Reader, projectDir string) {
	defer func() { _ = recover() }()
	var p payload
	if json.NewDecoder(in).Decode(&p) != nil {
		return
	}
	h := hook{s: newStore(cmp.Or(projectDir, p.Cwd)), p: p, projectDir: cmp.Or(projectDir, p.Cwd)}
	if !h.s.enabled() {
		return
	}
	switch p.HookEventName {
	case "SessionStart":
		h.sessionStart()
	case "UserPromptSubmit":
		h.prompt()
	case "PostToolUseFailure":
		h.toolFailure()
	case "PostToolUse":
		h.edit()
	case "Stop":
		h.s.sweep(p.TranscriptPath, p.SessionID)
	case "SubagentStop":
		h.s.sweep(p.AgentTranscriptPath, p.SessionID)
	case "SessionEnd":
		h.sessionEnd()
	}
}

func (h hook) base(kind string) Event {
	transcript := h.p.TranscriptPath
	if h.p.AgentID != "" {
		transcript = subagentTranscript(transcript, h.p.AgentID)
	}
	return Event{
		Kind:       kind,
		SessionID:  h.p.SessionID,
		AgentID:    h.p.AgentID,
		AgentType:  h.p.AgentType,
		Cwd:        h.p.Cwd,
		Transcript: transcript,
	}
}

func (h hook) sessionStart() {
	files := instrFiles(h.projectDir)
	e := h.base("session")
	e.Class = h.p.Source
	e.Commit = git(h.projectDir, "rev-parse", "--short", "HEAD")
	e.Branch = git(h.projectDir, "branch", "--show-current")
	e.Dirty = git(h.projectDir, "status", "--porcelain") != ""
	e.Files = files
	e.InstrHash = combinedHash(files)
	_ = h.s.append(e)
}

func (h hook) prompt() {
	text := h.p.Prompt
	if !includeMessage(text) || (len(text) > maxCapturePromptLen && !explicit.MatchString(text)) {
		return
	}
	names, conf := detectCorrection(text)
	if len(names) == 0 {
		return
	}
	e := h.base("correction")
	e.Class = strings.Join(names, " ")
	e.Conf = conf
	e.Input = clean(text, 500)
	_ = h.s.append(e)
}

func (h hook) toolFailure() {
	class := classify(h.p.ToolName, h.p.Error)
	e := h.base("tool_error")
	e.ToolUseID = h.p.ToolUseID
	e.Tool = h.p.ToolName
	e.Class = class
	e.FP = fingerprint(h.p.ToolName, class, h.p.ToolInput, h.p.Error)
	e.Input = clean(string(h.p.ToolInput), 500)
	e.Error = clean(h.p.Error, 1000)
	_ = h.s.append(e)
}

func (h hook) edit() {
	if h.p.ToolName != "Edit" && h.p.ToolName != "Write" {
		return
	}
	var in editInput
	if json.Unmarshal(h.p.ToolInput, &in) != nil || in.FilePath == "" {
		return
	}
	e := h.base("edit")
	e.ToolUseID = h.p.ToolUseID
	e.Tool = h.p.ToolName
	e.File = in.FilePath
	e.OldHash = hash12(in.OldString)
	e.NewHash = hash12(cmp.Or(in.NewString, in.Content))
	_ = h.s.append(e)
}

func (h hook) sessionEnd() {
	h.s.sweep(h.p.TranscriptPath, h.p.SessionID)
	for _, f := range subagentTranscripts(h.p.TranscriptPath) {
		h.s.sweep(f, h.p.SessionID)
	}
	calls, prompts := countTranscript(h.p.TranscriptPath, true)
	for _, f := range subagentTranscripts(h.p.TranscriptPath) {
		c, _ := countTranscript(f, false)
		calls += c
	}
	e := h.base("session_end")
	e.ToolCalls = calls
	e.Prompts = prompts
	_ = h.s.append(e)
}

func hookMain() {
	runHook(os.Stdin, os.Getenv("CLAUDE_PROJECT_DIR"))
}
