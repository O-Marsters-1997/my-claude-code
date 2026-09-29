package main

import "testing"

func TestDetectCorrection(t *testing.T) {
	corrections := []string{
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
	}
	for _, text := range corrections {
		t.Run(text, func(t *testing.T) {
			names, conf := detectCorrection(text)
			if len(names) == 0 || conf < 0.5 {
				t.Errorf("detectCorrection(%q) = %v, %.2f, want a correction", text, names, conf)
			}
		})
	}
}

func TestDetectCorrectionIgnores(t *testing.T) {
	benign := []string{
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
	}
	for _, text := range benign {
		t.Run(text, func(t *testing.T) {
			if names, conf := detectCorrection(text); len(names) != 0 {
				t.Errorf("detectCorrection(%q) = %v, %.2f, want none", text, names, conf)
			}
		})
	}
}

func TestIncludeMessage(t *testing.T) {
	for text, want := range map[string]bool{
		"no, use pnpm":                         true,
		"<system-reminder>x</system-reminder>": false,
		"<command-name>/model</command-name>":  false,
		"   ":                                  false,
		"{\"a\":1}":                            false,
	} {
		if got := includeMessage(text); got != want {
			t.Errorf("includeMessage(%q) = %v, want %v", text, got, want)
		}
	}
}
