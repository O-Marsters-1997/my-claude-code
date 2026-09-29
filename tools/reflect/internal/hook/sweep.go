package hook

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/transcript"
)

func sweep(store logstore.Store, path, sessionID string) {
	offsetFile := filepath.Join(store.Dir(), "offsets", filepath.Base(path)+".off")
	var from int64
	if b, err := os.ReadFile(offsetFile); err == nil {
		from, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	}
	agentID := transcript.AgentID(path)
	errs := loadLedger(store, sessionID)
	uses := map[string]transcript.Block{}
	consumed, err := transcript.Each(path, from, func(e transcript.Entry) {
		_, blocks := e.Parts()
		for _, b := range blocks {
			switch {
			case b.Type == "tool_use":
				uses[b.ID] = b
			case b.Type == "tool_result" && b.IsError:
				recordToolUseError(errs, b, uses[b.ToolUseID], e.Timestamp, logstore.Event{SessionID: sessionID, AgentID: agentID})
			}
		}
	})
	if err != nil || consumed == from || os.MkdirAll(filepath.Dir(offsetFile), 0o755) != nil {
		return
	}
	_ = os.WriteFile(offsetFile, []byte(strconv.FormatInt(consumed, 10)), 0o644)
}

func recordToolUseError(errs *ledger, result, use transcript.Block, timestamp string, e logstore.Event) {
	text := result.ResultText()
	if !strings.Contains(text, "<tool_use_error>") {
		return
	}
	text = strings.NewReplacer("<tool_use_error>", "", "</tool_use_error>", "").Replace(text)
	class := detect.Classify(use.Name, text)
	if !detect.IsHallucination(class) && class != detect.ReadFirst {
		return
	}
	e.TS, _ = time.Parse(time.RFC3339Nano, timestamp)
	e.Kind = "tool_error"
	e.ToolUseID = result.ToolUseID
	e.Tool = use.Name
	e.Class = class
	e.FP = detect.Fingerprint(use.Name, class, use.Input, text)
	e.Input = redact.Clean(string(use.Input), inputMax)
	e.Error = redact.Tail(text, errorMax)
	errs.append(e)
}
