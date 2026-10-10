package ledger

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Learning is one ledger record. A ledger line may carry only some fields;
// Read folds lines sharing an id, later fields overriding earlier ones.
type Learning struct {
	ID         string `json:"id"`
	TS         string `json:"ts,omitempty"`
	Source     string `json:"source,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Scope      string `json:"scope,omitempty"`
	Skill      string `json:"skill,omitempty"`
	Text       string `json:"text,omitempty"`
	Repo       string `json:"repo,omitempty"`
	Origin     string `json:"origin,omitempty"`
	File       string `json:"file,omitempty"`
	Line       int    `json:"line,omitempty"`
	TargetText string `json:"target_line_text,omitempty"`
	Commit     string `json:"commit,omitempty"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Status     string `json:"status,omitempty"`
	Issue      string `json:"issue,omitempty"`
}

// DefaultPath is $XDG_DATA_HOME/learn/learnings.jsonl, falling back to
// ~/.local/share/learn/learnings.jsonl.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(cmp.Or(os.Getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share")), "learn", "learnings.jsonl")
}

// Append adds one line to the ledger at path, creating it if needed.
func Append(path string, l Learning) error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Read returns the folded learnings in order of first appearance. A missing
// ledger reads as empty and lines that are not ledger records are skipped.
func Read(path string) ([]Learning, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Learning
	index := map[string]int{}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 16<<20)
	for n := 1; sc.Scan(); n++ {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var probe struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(line, &probe); err != nil || probe.ID == "" {
			continue
		}
		i, seen := index[probe.ID]
		if !seen {
			i = len(out)
			index[probe.ID] = i
			out = append(out, Learning{})
		}
		if err := json.Unmarshal(line, &out[i]); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
	}
	return out, sc.Err()
}
