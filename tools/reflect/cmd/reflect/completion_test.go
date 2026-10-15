package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestAreaMatches(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	for _, f := range []string{"skills/code/tdd/SKILL.md", "tools/reflect/main.go"} {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := areaMatches(dir, "TDD")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"skills/code/tdd", "skills/code/tdd/SKILL.md"}
	if !slices.Equal(got, want) {
		t.Errorf("areaMatches(TDD) = %v, want %v", got, want)
	}
}

func TestCompletionScripts(t *testing.T) {
	for _, shell := range []string{"zsh", "bash"} {
		t.Run(shell, func(t *testing.T) {
			got, err := completionScript([]string{shell})
			if err != nil {
				t.Fatal(err)
			}
			words := append([]string{}, commands...)
			for _, m := range markerCommands {
				words = append(words, m.name)
				words = append(words, m.flags...)
			}
			for _, w := range words {
				if !strings.Contains(got, w) {
					t.Errorf("completionScript(%q) missing %q", shell, w)
				}
			}
			if path, err := exec.LookPath(shell); err == nil {
				if out, err := exec.Command(path, "-n", "-c", got).CombinedOutput(); err != nil {
					t.Errorf("%s -n: %v\n%s", shell, err, out)
				}
			}
		})
	}
}

func TestCompletionUnknownShell(t *testing.T) {
	if _, err := completionScript([]string{"fish"}); err == nil {
		t.Error("completionScript(fish) accepted")
	}
	if !strings.Contains(usage, "completion") {
		t.Errorf("usage = %q, want completion listed", usage)
	}
}
