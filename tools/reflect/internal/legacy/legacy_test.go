package legacy_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/legacy"
)

func TestUninstallStripsOnlyReflect(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, ".claude")
	if err := os.MkdirAll(filepath.Join(claude, "reflect", "edits"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := `{
  "permissions": {"allow": ["Bash(ls)"]},
  "hooks": {
    "SessionStart": [{"hooks": [{"type": "command", "command": "/u/.claude/bin/reflect hook 2>/dev/null || true", "async": true}]}],
    "Stop": [
      {"hooks": [{"type": "command", "command": "/u/.claude/bin/reflect hook 2>/dev/null || true", "async": true}]},
      {"hooks": [{"type": "command", "command": "say done"}]}
    ]
  }
}`
	for path, body := range map[string]string{
		filepath.Join(claude, "settings.local.json"):        settings,
		filepath.Join(claude, "reflect", "events.jsonl"):    "{}\n",
		filepath.Join(claude, "reflect", "reports", "r.md"): "keep",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := legacy.Uninstall(root); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	b, err := os.ReadFile(filepath.Join(claude, "settings.local.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(ls)"}},
		"hooks": map[string]any{"Stop": []any{
			map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "say done"}}},
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("settings after Uninstall (-want +got):\n%s", diff)
	}
	for _, gone := range []string{"events.jsonl", "edits"} {
		if _, err := os.Stat(filepath.Join(claude, "reflect", gone)); !os.IsNotExist(err) {
			t.Errorf("%s still exists after Uninstall", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(claude, "reflect", "reports", "r.md")); err != nil {
		t.Errorf("reports removed by Uninstall: %v", err)
	}
}
