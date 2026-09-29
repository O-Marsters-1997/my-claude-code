package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
)

const (
	classEditMiss    = "edit_miss"
	classReadFirst   = "read_first"
	classPathMissing = "path_missing"
	classNotFound    = "not_found"
	classSymbol      = "symbol_missing"
	classExit        = "exit_nonzero"
	classOther       = "other"
)

var (
	symbolMissing = regexp.MustCompile(`(?i)cannot find (?:module|package|name|symbol)|no module named|modulenotfounderror|\bundefined: \w|is not defined|has no attribute|no such (?:function|method|export)|not exported|unresolved import|could not resolve|is not a function`)
	exitCode      = regexp.MustCompile(`^Exit code \d+`)
	evalPrefix    = regexp.MustCompile(`^\(eval\):\d+: `)
	digits        = regexp.MustCompile(`\d+`)
)

func isHallucination(class string) bool {
	switch class {
	case classEditMiss, classPathMissing, classNotFound, classSymbol:
		return true
	}
	return false
}

func classify(tool, errText string) string {
	lower := strings.ToLower(errText)
	switch {
	case strings.Contains(lower, "string to replace not found"):
		return classEditMiss
	case strings.Contains(lower, "has not been read yet"):
		return classReadFirst
	case strings.Contains(lower, "file does not exist"), strings.Contains(lower, "no such file or directory"), strings.Contains(lower, "enoent"):
		return classPathMissing
	case strings.Contains(lower, "command not found"):
		return classNotFound
	case symbolMissing.MatchString(errText):
		return classSymbol
	case tool == "Bash" && exitCode.MatchString(errText):
		return classExit
	}
	return classOther
}

func hash12(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:6])
}

func normalizeLine(s string) string {
	s = evalPrefix.ReplaceAllString(strings.TrimSpace(s), "")
	return digits.ReplaceAllString(truncate(s, 200), "#")
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}

type toolArgs struct {
	Command  string `json:"command"`
	FilePath string `json:"file_path"`
	Path     string `json:"path"`
	Pattern  string `json:"pattern"`
}

func (a toolArgs) target() string {
	switch {
	case a.FilePath != "":
		return a.FilePath
	case a.Path != "":
		return a.Path
	}
	return a.Pattern
}

func fingerprint(tool, class string, input json.RawMessage, errText string) string {
	var a toolArgs
	_ = json.Unmarshal(input, &a)
	switch class {
	case classEditMiss, classReadFirst, classPathMissing:
		if t := a.target(); t != "" {
			return hash12(tool, class, t)
		}
	case classNotFound, classSymbol:
		return hash12(class, normalizeLine(lastLine(errText)))
	}
	if a.Command != "" {
		return hash12(tool, class, strings.Join(strings.Fields(a.Command), " "))
	}
	return hash12(tool, class, normalizeLine(errText))
}
