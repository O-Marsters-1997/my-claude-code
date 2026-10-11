package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFullFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".claude/fleet.toml", `
ci = "async"

[checks]
pre_push = ["just lint", "just test"]
suite = ["just ci"]

[regen]
"web/dist/**" = "npm run build"
"*.golden" = "go test -update"
`)
	got, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{
		CI:     "async",
		Checks: config.Checks{PrePush: []string{"just lint", "just test"}, Suite: []string{"just ci"}},
		Regen:  map[string]string{"web/dist/**": "npm run build", "*.golden": "go test -update"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Load() (-want +got):\n%s", diff)
	}
}

func TestLoadMissingFile(t *testing.T) {
	got, err := config.Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(config.Config{}, got); diff != "" {
		t.Errorf("Load() (-want +got):\n%s", diff)
	}
}

func TestLoadClaudeMDFallback(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "CLAUDE.md", "# Project\n\n## Fleet\n\n- `web/dist/**`: `npm run build`\n- `golden/*.txt`: `go test -update`\n\n## Other\n\n- `x`: `y`\n")
	got, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"web/dist/**": "npm run build", "golden/*.txt": "go test -update"}
	if diff := cmp.Diff(want, got.Regen); diff != "" {
		t.Errorf("Load().Regen (-want +got):\n%s", diff)
	}
}

func TestLoadTomlRegenWinsOverClaudeMD(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".claude/fleet.toml", "[regen]\n\"a\" = \"b\"\n")
	write(t, dir, "CLAUDE.md", "## Fleet\n- `c`: `d`\n")
	got, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(map[string]string{"a": "b"}, got.Regen); diff != "" {
		t.Errorf("Load().Regen (-want +got):\n%s", diff)
	}
}

func TestLoadInvalidToml(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".claude/fleet.toml", "ci = [")
	if _, err := config.Load(dir); err == nil {
		t.Error("Load() error = nil, want error")
	}
}

func TestMatchAny(t *testing.T) {
	globs := []string{"web/dist/**", "*.golden"}
	tests := []struct {
		name string
		want bool
	}{
		{"web/dist/app.css", true},
		{"web/dist/a/b/app.css", true},
		{"x.golden", true},
		{"sub/x.golden", false},
		{"web/src/app.css", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.MatchAny(globs, tt.name); got != tt.want {
				t.Errorf("MatchAny(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
