package install

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const hookArgs = " hook 2>/dev/null || true"

var events = []struct{ name, matcher string }{
	{"SessionStart", ""},
	{"UserPromptSubmit", ""},
	{"PostToolUseFailure", ""},
	{"PostToolUse", "Edit|Write"},
	{"Stop", ""},
	{"SubagentStop", ""},
	{"SessionEnd", ""},
}

func SettingsPath(root string) string {
	return filepath.Join(root, ".claude", "settings.local.json")
}

func Enable(root, exe string) error {
	cfg, err := load(root)
	if err != nil {
		return err
	}
	hooks := strip(asMap(cfg["hooks"]))
	for _, ev := range events {
		group := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": exe + hookArgs, "async": true}}}
		if ev.matcher != "" {
			group["matcher"] = ev.matcher
		}
		hooks[ev.name] = append(asList(hooks[ev.name]), group)
	}
	cfg["hooks"] = hooks
	return save(root, cfg)
}

func Disable(root string) error {
	cfg, err := load(root)
	if err != nil {
		return err
	}
	hooks := strip(asMap(cfg["hooks"]))
	if len(hooks) == 0 {
		delete(cfg, "hooks")
	} else {
		cfg["hooks"] = hooks
	}
	if _, err := os.Stat(SettingsPath(root)); os.IsNotExist(err) {
		return nil
	}
	return save(root, cfg)
}

func Enabled(root string) bool {
	cfg, err := load(root)
	if err != nil {
		return false
	}
	for _, groups := range asMap(cfg["hooks"]) {
		for _, g := range asList(groups) {
			if isOurs(g) {
				return true
			}
		}
	}
	return false
}

func GitIgnored(root string) bool {
	return exec.Command("git", "-C", root, "check-ignore", "-q", SettingsPath(root)).Run() == nil
}

func isOurs(group any) bool {
	for _, h := range asList(asMap(group)["hooks"]) {
		cmd, _ := asMap(h)["command"].(string)
		if strings.HasSuffix(cmd, hookArgs) && strings.Contains(cmd, "reflect") {
			return true
		}
	}
	return false
}

func strip(hooks map[string]any) map[string]any {
	for name, groups := range hooks {
		var kept []any
		for _, g := range asList(groups) {
			if !isOurs(g) {
				kept = append(kept, g)
			}
		}
		if len(kept) == 0 {
			delete(hooks, name)
			continue
		}
		hooks[name] = kept
	}
	return hooks
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		m = map[string]any{}
	}
	return m
}

func asList(v any) []any {
	l, _ := v.([]any)
	return l
}

func load(root string) (map[string]any, error) {
	b, err := os.ReadFile(SettingsPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return asMap(cfg), nil
}

func save(root string, cfg map[string]any) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(SettingsPath(root)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(SettingsPath(root), append(b, '\n'), 0o644)
}
