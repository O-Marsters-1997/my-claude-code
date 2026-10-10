package scope_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/scope"
)

func mkdir(t *testing.T, parts ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(parts...), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestResolve(t *testing.T) {
	home, repo, lib := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	mkdir(t, repo, ".claude", "skills", "both")
	mkdir(t, repo, ".claude", "skills", "mine")
	mkdir(t, lib, "skills", "code", "both")
	mkdir(t, lib, "skills", "code", "libonly")
	mkdir(t, home, ".agents", "skills", "agents")
	mkdir(t, home, ".claude", "skills", "claude")

	tests := []struct {
		name, wantScope, wantSkill string
	}{
		{"mine", "repo", "mine"},
		{"both", "repo", "both"},
		{"libonly", "global", "libonly"},
		{"agents", "global", "agents"},
		{"claude", "global", "claude"},
		{"repo", "repo", ""},
		{"", "", ""},
		{"nope", "", "nope"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, k := scope.Resolve(repo, lib, tt.name)
			if s != tt.wantScope || k != tt.wantSkill {
				t.Errorf("Resolve(%q) = %q, %q, want %q, %q", tt.name, s, k, tt.wantScope, tt.wantSkill)
			}
		})
	}
}
