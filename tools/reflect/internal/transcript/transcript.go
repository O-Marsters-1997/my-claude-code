package transcript

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
)

type Entry struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	IsMeta    bool   `json:"isMeta"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type Block struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error"`
	Content   json.RawMessage `json:"content"`
	Text      string          `json:"text"`
}

func (e Entry) Parts() (text string, blocks []Block) {
	raw := e.Message.Content
	if len(raw) > 0 && raw[0] == '"' {
		_ = json.Unmarshal(raw, &text)
		return text, nil
	}
	_ = json.Unmarshal(raw, &blocks)
	var texts []string
	for _, b := range blocks {
		if b.Type == "text" {
			texts = append(texts, b.Text)
		}
	}
	return strings.Join(texts, "\n"), blocks
}

func (b Block) ResultText() string {
	var s string
	if json.Unmarshal(b.Content, &s) == nil {
		return s
	}
	var parts []Block
	_ = json.Unmarshal(b.Content, &parts)
	texts := make([]string, len(parts))
	for i, p := range parts {
		texts[i] = p.Text
	}
	return strings.Join(texts, "\n")
}

func Each(path string, from int64, fn func(Entry)) (consumed int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return from, err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Seek(from, io.SeekStart); err != nil {
		return from, err
	}
	rd := bufio.NewReaderSize(f, 1<<20)
	consumed = from
	for {
		line, err := rd.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			return consumed, nil
		}
		if err != nil {
			return consumed, err
		}
		consumed += int64(len(line))
		var e Entry
		if json.Unmarshal(line, &e) == nil {
			fn(e)
		}
	}
}

func SubagentPath(transcript, agentID string) string {
	return filepath.Join(strings.TrimSuffix(transcript, ".jsonl"), "subagents", "agent-"+agentID+".jsonl")
}

func Subagents(transcript string) []string {
	files, _ := filepath.Glob(filepath.Join(strings.TrimSuffix(transcript, ".jsonl"), "subagents", "agent-*.jsonl"))
	return files
}

func AgentID(path string) string {
	id, _ := strings.CutPrefix(strings.TrimSuffix(filepath.Base(path), ".jsonl"), "agent-")
	if id == filepath.Base(path) {
		return ""
	}
	return id
}

func Count(path string, countPrompts bool) (toolCalls, prompts int) {
	_, _ = Each(path, 0, func(e Entry) {
		text, blocks := e.Parts()
		for _, b := range blocks {
			if b.Type == "tool_use" {
				toolCalls++
			}
		}
		if countPrompts && e.Type == "user" && !e.IsMeta && detect.IncludeMessage(text) {
			prompts++
		}
	})
	return toolCalls, prompts
}

func LastToolUse(path string) (last Block, ok bool) {
	_, _ = Each(path, 0, func(e Entry) {
		_, blocks := e.Parts()
		for _, b := range blocks {
			if b.Type == "tool_use" {
				last, ok = b, true
			}
		}
	})
	return last, ok
}
