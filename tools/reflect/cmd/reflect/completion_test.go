package main

import (
	"os/exec"
	"strings"
	"testing"
)

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
