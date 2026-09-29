package main

import (
	"encoding/json"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		tool, err, want string
	}{
		{"Edit", "String to replace not found in file.\nString: x", classEditMiss},
		{"Edit", "File has not been read yet. Read it first", classReadFirst},
		{"Read", "File does not exist. Note: your cwd", classPathMissing},
		{"Bash", "Exit code 1\ncat: x: No such file or directory", classPathMissing},
		{"Bash", "Exit code 127\n(eval):1: command not found: foo", classNotFound},
		{"Bash", "Exit code 2\n./a.go:3: undefined: Foo", classSymbol},
		{"Bash", "Exit code 1\nModuleNotFoundError: No module named 'x'", classSymbol},
		{"Bash", "Exit code 3", classExit},
		{"Read", "boom", classOther},
	}
	for _, tt := range tests {
		if got := classify(tt.tool, tt.err); got != tt.want {
			t.Errorf("classify(%q, %q) = %q, want %q", tt.tool, tt.err, got, tt.want)
		}
	}
}

func TestFingerprintGroupsSameCommand(t *testing.T) {
	a := fingerprint("Bash", classExit, json.RawMessage(`{"command":"go  test ./..."}`), "Exit code 1")
	b := fingerprint("Bash", classExit, json.RawMessage(`{"command":"go test ./...","description":"x"}`), "Exit code 1\nFAIL")
	c := fingerprint("Bash", classExit, json.RawMessage(`{"command":"go build"}`), "Exit code 1")
	if a != b {
		t.Errorf("same command gave different fingerprints: %s vs %s", a, b)
	}
	if a == c {
		t.Errorf("different commands shared fingerprint %s", a)
	}
}

func TestFingerprintNotFoundIgnoresCaller(t *testing.T) {
	a := fingerprint("Bash", classNotFound, json.RawMessage(`{"command":"foo 1"}`), "Exit code 127\n(eval):1: command not found: foo")
	b := fingerprint("Bash", classNotFound, json.RawMessage(`{"command":"foo 2"}`), "Exit code 127\n(eval):9: command not found: foo")
	if a != b {
		t.Errorf("same missing command gave different fingerprints: %s vs %s", a, b)
	}
}
