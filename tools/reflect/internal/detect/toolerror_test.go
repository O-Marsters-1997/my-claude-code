package detect_test

import (
	"encoding/json"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		tool, err, want string
	}{
		{"Edit", "String to replace not found in file.\nString: x", detect.EditMiss},
		{"Edit", "File has not been read yet. Read it first", detect.ReadFirst},
		{"Read", "File does not exist. Note: your cwd", detect.PathMissing},
		{"Bash", "Exit code 1\ncat: x: No such file or directory", detect.PathMissing},
		{"Bash", "Exit code 127\n(eval):1: command not found: foo", detect.NotFound},
		{"Bash", "Exit code 2\n./a.go:3: undefined: Foo", detect.Symbol},
		{"Bash", "Exit code 1\nModuleNotFoundError: No module named 'x'", detect.Symbol},
		{"Bash", "Exit code 3", detect.Exit},
		{"Read", "boom", detect.Other},
	}
	for _, tt := range tests {
		if got := detect.Classify(tt.tool, tt.err); got != tt.want {
			t.Errorf("Classify(%q, %q) = %q, want %q", tt.tool, tt.err, got, tt.want)
		}
	}
}

func TestFingerprintGroupsSameCommand(t *testing.T) {
	a := detect.Fingerprint("Bash", detect.Exit, json.RawMessage(`{"command":"go  test ./..."}`), "Exit code 1")
	b := detect.Fingerprint("Bash", detect.Exit, json.RawMessage(`{"command":"go test ./...","description":"x"}`), "Exit code 1\nFAIL")
	c := detect.Fingerprint("Bash", detect.Exit, json.RawMessage(`{"command":"go build"}`), "Exit code 1")
	if a != b {
		t.Errorf("same command gave different fingerprints: %q vs %q", a, b)
	}
	if a == c {
		t.Errorf("different commands shared fingerprint %q", a)
	}
}

func TestFingerprintNotFoundIgnoresCaller(t *testing.T) {
	a := detect.Fingerprint("Bash", detect.NotFound, json.RawMessage(`{"command":"foo 1"}`), "Exit code 127\n(eval):1: command not found: foo")
	b := detect.Fingerprint("Bash", detect.NotFound, json.RawMessage(`{"command":"foo 2"}`), "Exit code 127\n(eval):9: command not found: foo")
	if a != b {
		t.Errorf("same missing command gave different fingerprints: %q vs %q", a, b)
	}
}
