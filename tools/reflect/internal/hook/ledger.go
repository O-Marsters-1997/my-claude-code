package hook

import "github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"

const (
	inputMax  = 150
	errorMax  = 300
	bareAfter = 3
)

type ledger struct {
	store logstore.Store
	ids   map[string]bool
	fps   map[string]int
}

func loadLedger(store logstore.Store, sessionID string) *ledger {
	l := &ledger{store: store, ids: map[string]bool{}, fps: map[string]int{}}
	for _, e := range kindsOf(store, sessionID, "tool_error") {
		l.ids[e.ToolUseID] = true
		l.fps[e.FP]++
	}
	return l
}

func (l *ledger) append(e logstore.Event) {
	if e.ToolUseID != "" && l.ids[e.ToolUseID] {
		return
	}
	l.ids[e.ToolUseID] = true
	l.fps[e.FP]++
	if l.fps[e.FP] > bareAfter {
		e.Input, e.Error = "", ""
	}
	_ = l.store.Append(e)
}
