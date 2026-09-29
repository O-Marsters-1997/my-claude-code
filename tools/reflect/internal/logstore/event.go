package logstore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

const schemaVersion = 2

type Event struct {
	V           int               `json:"v"`
	TS          time.Time         `json:"ts"`
	Kind        string            `json:"kind"`
	SessionID   string            `json:"session_id"`
	AgentID     string            `json:"agent_id,omitempty"`
	AgentType   string            `json:"agent_type,omitempty"`
	ToolUseID   string            `json:"tool_use_id,omitempty"`
	Tool        string            `json:"tool,omitempty"`
	Class       string            `json:"class,omitempty"`
	FP          string            `json:"fp,omitempty"`
	Input       string            `json:"input,omitempty"`
	Error       string            `json:"error,omitempty"`
	File        string            `json:"file,omitempty"`
	OldHash     string            `json:"old,omitempty"`
	NewHash     string            `json:"new,omitempty"`
	Conf        float64           `json:"conf,omitempty"`
	Cwd         string            `json:"cwd,omitempty"`
	Transcript  string            `json:"transcript,omitempty"`
	Commit      string            `json:"commit,omitempty"`
	Branch      string            `json:"branch,omitempty"`
	Dirty       bool              `json:"dirty,omitempty"`
	InstrHash   string            `json:"instr_hash,omitempty"`
	Prev        string            `json:"prev,omitempty"`
	Added       map[string]string `json:"added,omitempty"`
	Changed     map[string]string `json:"changed,omitempty"`
	Removed     []string          `json:"removed,omitempty"`
	ToolCalls   int               `json:"tool_calls,omitempty"`
	Prompts     int               `json:"prompts,omitempty"`
	Errors      int               `json:"errors,omitempty"`
	Ended       bool              `json:"ended,omitempty"`
	Confusion   int               `json:"confusion,omitempty"`
	Corrections int               `json:"corrections,omitempty"`
	RepeatFail  int               `json:"repeat_fail,omitempty"`
	FPs         []string          `json:"fps,omitempty"`
}

type Store struct{ dir string }

func New(projectDir string) Store {
	return Store{filepath.Join(mainCheckout(projectDir), ".claude", "reflect")}
}

func At(dir string) Store { return Store{dir} }

func (s Store) Dir() string     { return s.dir }
func (s Store) LogPath() string { return filepath.Join(s.dir, "events.jsonl") }

func (s Store) ProposalsDir() (string, error) {
	dir := filepath.Join(s.dir, "proposals")
	if err := s.ensure(); err != nil {
		return "", err
	}
	return dir, os.MkdirAll(dir, 0o755)
}

func (s Store) ensure() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	ignore := filepath.Join(s.dir, ".gitignore")
	if _, err := os.Stat(ignore); err == nil {
		return nil
	}
	return os.WriteFile(ignore, []byte("*\n"), 0o644)
}

func Encode(e Event) ([]byte, error) {
	e.V = schemaVersion
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	return json.Marshal(e)
}

func (s Store) Append(e Event) error {
	if err := s.ensure(); err != nil {
		return err
	}
	line, err := Encode(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.LogPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return errors.Join(err, f.Close())
}

func (s Store) Rewrite(fn func(lines [][]byte) ([][]byte, error)) error {
	original, err := os.ReadFile(s.LogPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var lines [][]byte
	for _, line := range bytes.Split(original, []byte("\n")) {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	kept, err := fn(lines)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, "events-*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	w := bufio.NewWriter(tmp)
	for _, line := range kept {
		_, _ = w.Write(line)
		_ = w.WriteByte('\n')
	}
	if err := errors.Join(w.Flush(), appendGrowth(tmp, s.LogPath(), int64(len(original)))); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.LogPath())
}

func appendGrowth(dst *os.File, logPath string, from int64) error {
	src, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	if _, err := src.Seek(from, io.SeekStart); err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

func (s Store) Size() int64 {
	info, err := os.Stat(s.LogPath())
	if err != nil {
		return 0
	}
	return info.Size()
}

func (s Store) Read() ([]Event, error) {
	f, err := os.Open(s.LogPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var events []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			events = append(events, e)
		}
	}
	resolveTranscripts(events)
	return events, sc.Err()
}

func resolveTranscripts(events []Event) {
	main := map[string]string{}
	for _, e := range events {
		if e.Kind == "session" && e.Transcript != "" {
			main[e.SessionID] = e.Transcript
		}
	}
	for i := range events {
		e := &events[i]
		base, ok := main[e.SessionID]
		if e.Transcript != "" || !ok {
			continue
		}
		e.Transcript = base
		if e.AgentID != "" {
			e.Transcript = transcript.SubagentPath(base, e.AgentID)
		}
	}
}
