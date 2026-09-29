package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/install"
)

const exe = "/home/u/.claude/bin/reflect"

func write(t *testing.T, root, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(install.SettingsPath(root), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(install.SettingsPath(root))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func groups(cfg map[string]any, event string) []any {
	hooks, _ := cfg["hooks"].(map[string]any)
	l, _ := hooks[event].([]any)
	return l
}

func TestEnableKeepsExistingSettingsAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	write(t, root, `{"permissions":{"allow":["Bash(ls)"]},"hooks":{"Stop":[{"hooks":[{"type":"command","command":"other.sh"}]}]}}`)
	for range 2 {
		if err := install.Enable(root, exe); err != nil {
			t.Fatal(err)
		}
	}
	cfg := read(t, root)
	if _, ok := cfg["permissions"]; !ok {
		t.Error("Enable dropped unrelated settings")
	}
	if n := len(groups(cfg, "Stop")); n != 2 {
		t.Errorf("Stop has %d groups after enabling twice, want other.sh plus one reflect", n)
	}
	if n := len(groups(cfg, "SessionEnd")); n != 1 {
		t.Errorf("SessionEnd has %d groups, want 1", n)
	}
	if !install.Enabled(root) {
		t.Error("Enabled = false after Enable")
	}
}

func TestEnableCommandUsesGivenBinary(t *testing.T) {
	root := t.TempDir()
	if err := install.Enable(root, exe); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(install.SettingsPath(root))
	if !strings.Contains(string(b), exe+" hook") {
		t.Errorf("settings do not call %s:\n%s", exe, b)
	}
}

func TestDisableRemovesOnlyReflectHooks(t *testing.T) {
	root := t.TempDir()
	write(t, root, `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"other.sh"}]}]}}`)
	if err := install.Enable(root, exe); err != nil {
		t.Fatal(err)
	}
	if err := install.Disable(root); err != nil {
		t.Fatal(err)
	}
	cfg := read(t, root)
	if install.Enabled(root) {
		t.Error("Enabled = true after Disable")
	}
	if n := len(groups(cfg, "Stop")); n != 1 {
		t.Errorf("Stop has %d groups after Disable, want the original 1", n)
	}
	if n := len(groups(cfg, "SessionEnd")); n != 0 {
		t.Errorf("SessionEnd still has %d groups", n)
	}
}

func TestDisableWithoutSettingsCreatesNothing(t *testing.T) {
	root := t.TempDir()
	if err := install.Disable(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(install.SettingsPath(root)); !os.IsNotExist(err) {
		t.Error("Disable created a settings file")
	}
}
