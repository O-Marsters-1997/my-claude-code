package status

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
)

const (
	claudeDefaultCleanup = 30
	cleanupWarnDays      = 90
)

type settings struct {
	CleanupPeriodDays int `json:"cleanupPeriodDays"`
	Hooks             map[string][]struct {
		Hooks []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func Report(claudeDir, library string) (string, error) {
	var cfg settings
	if b, err := os.ReadFile(filepath.Join(claudeDir, "settings.json")); err == nil {
		_ = json.Unmarshal(b, &cfg)
	}
	records, err := record.Read(filepath.Join(claudeDir, "reflect"))
	if err != nil {
		return "", err
	}
	hook := "missing (run ./setup.sh --reflect)"
	if hasHook(cfg) {
		hook = "installed"
	}
	days := cmp.Or(cfg.CleanupPeriodDays, claudeDefaultCleanup)
	var w strings.Builder
	fmt.Fprintf(&w, "SessionStart hook: %s\nsession records: %d (%s)\ntranscripts kept: %d days\nlibrary: %s\n",
		hook, len(records), record.Path(filepath.Join(claudeDir, "reflect")), days, library)
	if days < cleanupWarnDays {
		fmt.Fprintf(&w, "warn: cleanupPeriodDays=%d; /reflect and metrics can only see sessions that recent\n", days)
	}
	return w.String(), nil
}

func hasHook(cfg settings) bool {
	for _, g := range cfg.Hooks["SessionStart"] {
		for _, h := range g.Hooks {
			if strings.Contains(h.Command, "reflect hook") {
				return true
			}
		}
	}
	return false
}
