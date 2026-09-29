package detect

import (
	"cmp"
	"encoding/json"
	"regexp"
	"strings"
)

const (
	EditMiss    = "edit_miss"
	ReadFirst   = "read_first"
	PathMissing = "path_missing"
	NotFound    = "not_found"
	Symbol      = "symbol_missing"
	Exit        = "exit_nonzero"
	Other       = "other"
)

var (
	symbolMissing = regexp.MustCompile(`(?i)cannot find (?:module|package|name|symbol)|no module named|modulenotfounderror|\bundefined: \w|is not defined|has no attribute|no such (?:function|method|export)|not exported|unresolved import|could not resolve|is not a function`)
	exitCode      = regexp.MustCompile(`^Exit code \d+`)
	evalPrefix    = regexp.MustCompile(`^\(eval\):\d+: `)
	digits        = regexp.MustCompile(`\d+`)
)

func IsHallucination(class string) bool {
	switch class {
	case EditMiss, PathMissing, NotFound, Symbol:
		return true
	}
	return false
}

func Classify(tool, errText string) string {
	lower := strings.ToLower(errText)
	switch {
	case strings.Contains(lower, "string to replace not found"):
		return EditMiss
	case strings.Contains(lower, "has not been read yet"):
		return ReadFirst
	case strings.Contains(lower, "file does not exist"), strings.Contains(lower, "no such file or directory"), strings.Contains(lower, "enoent"):
		return PathMissing
	case strings.Contains(lower, "command not found"):
		return NotFound
	case symbolMissing.MatchString(errText):
		return Symbol
	case tool == "Bash" && exitCode.MatchString(errText):
		return Exit
	}
	return Other
}

func Fingerprint(tool, class string, input json.RawMessage, errText string) string {
	var a struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Pattern  string `json:"pattern"`
	}
	_ = json.Unmarshal(input, &a)
	switch class {
	case EditMiss, ReadFirst, PathMissing:
		if target := cmp.Or(a.FilePath, a.Path, a.Pattern); target != "" {
			return join(tool, class, target)
		}
	case NotFound, Symbol:
		return join(class, normalize(lastLine(errText)))
	}
	if a.Command != "" {
		return join(tool, class, strings.Join(strings.Fields(a.Command), " "))
	}
	return join(tool, class, normalize(errText))
}

func join(parts ...string) string { return strings.Join(parts, "|") }

func normalize(s string) string {
	s = evalPrefix.ReplaceAllString(strings.TrimSpace(s), "")
	if len(s) > 200 {
		s = s[:200]
	}
	return digits.ReplaceAllString(s, "#")
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}
