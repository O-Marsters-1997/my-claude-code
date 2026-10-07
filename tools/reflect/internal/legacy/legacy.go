package legacy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookSuffix = " hook 2>/dev/null || true"

var leftovers = []string{"events.jsonl", "edits", "instr.json", "offsets", "ledger"}

func Uninstall(root string) (string, error) {
	var w strings.Builder
	settings := filepath.Join(root, ".claude", "settings.local.json")
	removed, err := stripHooks(settings)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(&w, "%s: removed %d reflect hook groups\n", settings, removed)
	dir := filepath.Join(root, ".claude", "reflect")
	for _, name := range leftovers {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if err := os.RemoveAll(p); err != nil {
			return w.String(), err
		}
		fmt.Fprintf(&w, "deleted %s\n", p)
	}
	return w.String(), nil
}

func stripHooks(path string) (int, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		return 0, err
	}
	hooks, _ := cfg["hooks"].(map[string]any)
	removed := 0
	for event, groups := range hooks {
		list, _ := groups.([]any)
		var kept []any
		for _, g := range list {
			if isReflectGroup(g) {
				removed++
				continue
			}
			kept = append(kept, g)
		}
		if len(kept) == 0 {
			delete(hooks, event)
			continue
		}
		hooks[event] = kept
	}
	if removed == 0 {
		return 0, nil
	}
	if len(hooks) == 0 {
		delete(cfg, "hooks")
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return 0, err
	}
	return removed, os.WriteFile(path, append(out, '\n'), 0o644)
}

func isReflectGroup(group any) bool {
	g, _ := group.(map[string]any)
	hooks, _ := g["hooks"].([]any)
	for _, h := range hooks {
		m, _ := h.(map[string]any)
		cmd, _ := m["command"].(string)
		if strings.HasSuffix(cmd, hookSuffix) && strings.Contains(cmd, "reflect") {
			return true
		}
	}
	return false
}
