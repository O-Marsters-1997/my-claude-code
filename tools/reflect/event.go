package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const schemaVersion = 1

type Event struct {
	V          int               `json:"v"`
	TS         time.Time         `json:"ts"`
	Kind       string            `json:"kind"`
	SessionID  string            `json:"session_id"`
	AgentID    string            `json:"agent_id,omitempty"`
	AgentType  string            `json:"agent_type,omitempty"`
	ToolUseID  string            `json:"tool_use_id,omitempty"`
	Tool       string            `json:"tool,omitempty"`
	Class      string            `json:"class,omitempty"`
	FP         string            `json:"fp,omitempty"`
	Input      string            `json:"input,omitempty"`
	Error      string            `json:"error,omitempty"`
	File       string            `json:"file,omitempty"`
	OldHash    string            `json:"old,omitempty"`
	NewHash    string            `json:"new,omitempty"`
	Conf       float64           `json:"conf,omitempty"`
	Cwd        string            `json:"cwd,omitempty"`
	Transcript string            `json:"transcript,omitempty"`
	Commit     string            `json:"commit,omitempty"`
	Branch     string            `json:"branch,omitempty"`
	Dirty      bool              `json:"dirty,omitempty"`
	InstrHash  string            `json:"instr_hash,omitempty"`
	Files      map[string]string `json:"files,omitempty"`
	ToolCalls  int               `json:"tool_calls,omitempty"`
	Prompts    int               `json:"prompts,omitempty"`
}

type store struct{ dir string }

func newStore(projectDir string) store {
	return store{filepath.Join(mainCheckout(projectDir), ".claude", "reflect")}
}

func (s store) enabled() bool {
	_, err := os.Stat(filepath.Join(s.dir, "on"))
	return err == nil
}

func (s store) logPath() string { return filepath.Join(s.dir, "events.jsonl") }

func (s store) ensure() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	ignore := filepath.Join(s.dir, ".gitignore")
	if _, err := os.Stat(ignore); err == nil {
		return nil
	}
	return os.WriteFile(ignore, []byte("*\n"), 0o644)
}

func (s store) append(e Event) error {
	if err := s.ensure(); err != nil {
		return err
	}
	e.V = schemaVersion
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.logPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

func (s store) read() ([]Event, error) {
	f, err := os.Open(s.logPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var events []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			events = append(events, e)
		}
	}
	return events, sc.Err()
}
