package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type entry struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	IsMeta    bool   `json:"isMeta"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type block struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   bool            `json:"is_error"`
	Content   json.RawMessage `json:"content"`
	Text      string          `json:"text"`
}

func (e entry) parts() (text string, blocks []block) {
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

func (b block) resultText() string {
	var s string
	if json.Unmarshal(b.Content, &s) == nil {
		return s
	}
	var parts []block
	_ = json.Unmarshal(b.Content, &parts)
	var texts []string
	for _, p := range parts {
		texts = append(texts, p.Text)
	}
	return strings.Join(texts, "\n")
}

func eachEntry(path string, from int64, fn func(entry)) (consumed int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return from, err
	}
	defer f.Close()
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
		var e entry
		if json.Unmarshal(line, &e) == nil {
			fn(e)
		}
	}
}

func subagentTranscript(transcript, agentID string) string {
	return filepath.Join(strings.TrimSuffix(transcript, ".jsonl"), "subagents", "agent-"+agentID+".jsonl")
}

func subagentTranscripts(transcript string) []string {
	files, _ := filepath.Glob(filepath.Join(strings.TrimSuffix(transcript, ".jsonl"), "subagents", "agent-*.jsonl"))
	return files
}

func agentIDOf(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	id, ok := strings.CutPrefix(base, "agent-")
	if !ok {
		return ""
	}
	return id
}

func (s store) sweep(path, sessionID string) {
	offsetFile := filepath.Join(s.dir, "offsets", filepath.Base(path)+".off")
	var from int64
	if b, err := os.ReadFile(offsetFile); err == nil {
		from, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	}
	agentID := agentIDOf(path)
	uses := map[string]block{}
	consumed, err := eachEntry(path, from, func(e entry) {
		_, blocks := e.parts()
		for _, b := range blocks {
			switch {
			case b.Type == "tool_use":
				uses[b.ID] = b
			case b.Type == "tool_result" && b.IsError:
				s.sweepResult(b, uses[b.ToolUseID], e, sessionID, agentID, path)
			}
		}
	})
	if err != nil || consumed == from {
		return
	}
	if os.MkdirAll(filepath.Dir(offsetFile), 0o755) == nil {
		_ = os.WriteFile(offsetFile, []byte(strconv.FormatInt(consumed, 10)), 0o644)
	}
}

func (s store) sweepResult(res, use block, e entry, sessionID, agentID, path string) {
	text := res.resultText()
	if !strings.Contains(text, "<tool_use_error>") {
		return
	}
	text = strings.NewReplacer("<tool_use_error>", "", "</tool_use_error>", "").Replace(text)
	class := classify(use.Name, text)
	if !isHallucination(class) && class != classReadFirst {
		return
	}
	ts, _ := time.Parse(time.RFC3339Nano, e.Timestamp)
	_ = s.append(Event{
		TS:         ts,
		Kind:       "tool_error",
		SessionID:  sessionID,
		AgentID:    agentID,
		ToolUseID:  res.ToolUseID,
		Tool:       use.Name,
		Class:      class,
		FP:         fingerprint(use.Name, class, use.Input, text),
		Input:      clean(string(use.Input), 500),
		Error:      clean(text, 1000),
		Transcript: path,
	})
}

func countTranscript(path string, prompts bool) (toolCalls, promptCount int) {
	_, _ = eachEntry(path, 0, func(e entry) {
		text, blocks := e.parts()
		for _, b := range blocks {
			if b.Type == "tool_use" {
				toolCalls++
			}
		}
		if prompts && e.Type == "user" && !e.IsMeta && includeMessage(text) {
			promptCount++
		}
	})
	return toolCalls, promptCount
}
