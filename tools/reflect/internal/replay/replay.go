package replay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
)

const hookTimeout = 30 * time.Second

var hookSource = regexp.MustCompile(`hook (?:blocking )?error: \[(.+?)\]: `)

type Options struct {
	Home    string
	Cwd     string
	Command string
}

type hookInput struct {
	ToolName            string          `json:"tool_name"`
	ToolInput           json.RawMessage `json:"tool_input"`
	Cwd                 string          `json:"cwd"`
	TranscriptPath      string          `json:"transcript_path"`
	AgentTranscriptPath string          `json:"agent_transcript_path"`
	AgentID             string          `json:"agent_id,omitempty"`
}

// Run feeds the tool call recorded at line through the hook that blocked it,
// against the transcript as it stood before that line, and reports the hook's
// exit code and stderr.
func Run(s session.Session, agentID string, line int, o Options) (string, error) {
	a := s.Agent(agentID)
	if a == nil {
		return "", fmt.Errorf("no agent %q in session %s", agentID, s.ID)
	}
	idx := -1
	for i, l := range a.Lines {
		if l.N == line {
			idx = i
		}
	}
	if idx < 0 {
		return "", fmt.Errorf("agent %s has no line L%d", agentID, line)
	}
	_, blocks := a.Lines[idx].Parts()
	for _, b := range blocks {
		if b.Type != "tool_use" {
			continue
		}
		hook, err := blockingHook(a, idx, b.ID, o.Home)
		if err != nil {
			return "", err
		}
		return execute(a, idx, b.Name, b.Input, hook, o)
	}
	return "", fmt.Errorf("L%d of agent %s has no tool call", line, agentID)
}

func blockingHook(a *session.Agent, from int, useID, home string) (string, error) {
	for _, l := range a.Lines[from:] {
		_, blocks := l.Parts()
		for _, b := range blocks {
			if b.Type != "tool_result" || b.ToolUseID != useID {
				continue
			}
			m := hookSource.FindStringSubmatch(b.ResultText())
			if m == nil {
				return "", errors.New("the call was not blocked by a hook")
			}
			return resolve(m[1], home)
		}
	}
	return "", errors.New("the call has no recorded result")
}

func resolve(path, home string) (string, error) {
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		path = filepath.Join(home, rest)
	}
	path = filepath.Clean(path)
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("hook %s is outside %s", path, home)
	}
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("hook %s is not a file", path)
	}
	return path, nil
}

func execute(a *session.Agent, idx int, tool string, input json.RawMessage, hook string, o Options) (string, error) {
	cwd := o.Cwd
	if cwd == "" {
		cwd = a.Lines[idx].Cwd
	}
	if o.Command != "" {
		var m map[string]any
		if err := json.Unmarshal(input, &m); err != nil {
			return "", err
		}
		m["command"] = o.Command
		input, _ = json.Marshal(m)
	}
	dir, err := os.MkdirTemp("", "reflect-replay-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	var before bytes.Buffer
	for _, l := range a.Lines[:idx] {
		before.Write(l.Raw)
	}
	transcript := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(transcript, before.Bytes(), 0o600); err != nil {
		return "", err
	}
	payload, _ := json.Marshal(hookInput{
		ToolName: tool, ToolInput: input, Cwd: cwd,
		TranscriptPath: transcript, AgentTranscriptPath: transcript,
	})
	ctx, cancel := context.WithTimeout(context.Background(), hookTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", hook)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Dir = cwd
	if info, err := os.Stat(cwd); err != nil || !info.IsDir() {
		cmd.Dir = ""
	}
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+cwd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return "", err
		}
		code = exit.ExitCode()
	}
	return fmt.Sprintf("hook: %s\ncwd: %s\nexit: %d\n%s", hook, cwd, code, stderr.String()), nil
}
