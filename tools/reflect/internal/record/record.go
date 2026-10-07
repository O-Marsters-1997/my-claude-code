package record

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	SessionID string            `json:"sid"`
	Cwd       string            `json:"cwd"`
	TS        time.Time         `json:"ts"`
	InstrHash string            `json:"instr_hash"`
	Files     map[string]string `json:"files"`
	Commit    string            `json:"commit,omitempty"`
	Branch    string            `json:"branch,omitempty"`
	Dirty     bool              `json:"dirty,omitempty"`
}

func Path(dir string) string { return filepath.Join(dir, "sessions.jsonl") }

func Read(dir string) ([]Record, error) {
	f, err := os.Open(Path(dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<16), 1<<24)
	for sc.Scan() {
		var r Record
		if json.Unmarshal(sc.Bytes(), &r) == nil && r.SessionID != "" {
			out = append(out, r)
		}
	}
	return out, sc.Err()
}

func Append(dir string, r Record) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(Path(dir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, werr := f.Write(append(b, '\n'))
	return errors.Join(werr, f.Close())
}
