package transcript

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

type Entry struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	Timestamp        string          `json:"timestamp"`
	Cwd              string          `json:"cwd"`
	IsMeta           bool            `json:"isMeta"`
	IsCompactSummary bool            `json:"isCompactSummary"`
	ToolUseResult    json.RawMessage `json:"toolUseResult"`
	CompactMetadata  struct {
		PreTokens  int `json:"preTokens"`
		PostTokens int `json:"postTokens"`
	} `json:"compactMetadata"`
	Message struct {
		ID      string          `json:"id"`
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
		Usage   Usage           `json:"usage"`
	} `json:"message"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

func (u Usage) In() int { return u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens }

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

type Line struct {
	N   int
	Raw []byte
	Entry
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

func (e Entry) Time() (time.Time, bool) {
	t, err := time.Parse(time.RFC3339Nano, e.Timestamp)
	return t, err == nil
}

func (e Entry) IsPrompt() bool {
	if e.Type != "user" || e.IsMeta || e.IsCompactSummary {
		return false
	}
	text, blocks := e.Parts()
	for _, b := range blocks {
		if b.Type == "tool_result" {
			return false
		}
	}
	text = strings.TrimSpace(text)
	return text != "" && !strings.HasPrefix(text, "<local-command-")
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

func Read(path string) ([]Line, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	rd := bufio.NewReaderSize(f, 1<<20)
	var lines []Line
	for n := 1; ; n++ {
		raw, err := rd.ReadBytes('\n')
		if len(raw) > 0 {
			l := Line{N: n, Raw: raw}
			_ = json.Unmarshal(raw, &l.Entry)
			lines = append(lines, l)
		}
		if errors.Is(err, io.EOF) {
			return lines, nil
		}
		if err != nil {
			return lines, err
		}
	}
}
