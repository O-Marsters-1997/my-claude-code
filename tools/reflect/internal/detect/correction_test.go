package detect_test

import (
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/detect"
)

func TestCorrectionDetects(t *testing.T) {
	for _, text := range []string{
		"no, use pnpm not npm",
		"No, that's wrong",
		"don't add comments unless I ask",
		"I told you to use tabs",
		"stop adding emojis",
		"remember: never edit generated files",
		"that is incorrect",
		"use tabs not spaces",
		"いや、そうじゃなくて",
		"不是，用另一个",
		"leave the tests alone",
	} {
		t.Run(text, func(t *testing.T) {
			if names, conf := detect.Correction(text); len(names) == 0 || conf < 0.5 {
				t.Errorf("Correction(%q) = %v, %.2f, want a correction", text, names, conf)
			}
		})
	}
}

func TestCorrectionIgnores(t *testing.T) {
	for _, text := range []string{
		"no problem",
		"No worries, thanks",
		"never mind",
		"no it works now",
		"can you fix this?",
		"please add a test",
		"/loop 5m /foo",
		"Perfect! Now let's add the endpoint",
		"ok",
		"the build failed with an error",
		"no dialog appeared",
		"I need a new endpoint",
	} {
		t.Run(text, func(t *testing.T) {
			if names, conf := detect.Correction(text); len(names) != 0 {
				t.Errorf("Correction(%q) = %v, %.2f, want none", text, names, conf)
			}
		})
	}
}

func TestCorrectionSkipsLongPromptsUnlessRemember(t *testing.T) {
	long := "no, " + strings.Repeat("x", 600)
	if names, _ := detect.Correction(long); len(names) != 0 {
		t.Errorf("long prompt matched %v", names)
	}
	if names, _ := detect.Correction("remember: " + long); len(names) == 0 {
		t.Error("long remember: prompt was dropped")
	}
}

func TestIncludeMessage(t *testing.T) {
	for text, want := range map[string]bool{
		"no, use pnpm":                         true,
		"<system-reminder>x</system-reminder>": false,
		"<command-name>/model</command-name>":  false,
		"   ":                                  false,
		`{"a":1}`:                              false,
	} {
		if got := detect.IncludeMessage(text); got != want {
			t.Errorf("IncludeMessage(%q) = %v, want %v", text, got, want)
		}
	}
}
