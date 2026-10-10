package marker

import (
	"regexp"
	"strings"
)

// Marker is one LEARN marker found in a file. Line, End and Target are
// 1-based; Target is 0 when an own-line marker has no code line after it.
// Col is the byte offset of the LEARN keyword within line Line.
type Marker struct {
	Skill      string
	Later      bool
	Text       string
	Line       int
	Col        int
	End        int
	Target     int
	TargetText string
}

var (
	marked   = regexp.MustCompile(`\bLEARN(?:\(([\w.-]+)\))?(\s+later)?:\s*(.*)`)
	opener   = regexp.MustCompile(`(?:/{2,}|#+|\{?/\*+|<!--|\*)\s*$`)
	leading  = regexp.MustCompile(`^\s*(?:/{2,}|#+(?:\s|$)|\{?/\*+|<!--|\*(?:[\s/]|$))`)
	closer   = regexp.MustCompile(`\s*(?:\*/\}?|-->)\s*$`)
	opening  = regexp.MustCompile(`^\s*(?:/{2,}|#+|\{?/\*+|<!--|\*)\s?`)
	hasClose = regexp.MustCompile(`(?:\*/\}?|-->)\s*$`)
)

// ParseComment reads a marker from the text of one comment, opener already
// removed. ok is false when the body holds no marker.
func ParseComment(body string) (m Marker, ok bool) {
	loc := marked.FindStringSubmatch(body)
	if loc == nil {
		return Marker{}, false
	}
	return Marker{
		Skill: loc[1],
		Later: loc[2] != "",
		Text:  strings.TrimSpace(closer.ReplaceAllString(loc[3], "")),
	}, true
}

// Parse finds every marker in lines. An own-line marker targets the next
// line that is neither a comment nor blank, and comment lines directly after
// it continue its text. A trailing marker targets its own line.
func Parse(lines []string) []Marker {
	var out []Marker
	for i := 0; i < len(lines); i++ {
		loc := marked.FindStringIndex(lines[i])
		if loc == nil || !inComment(lines[i][:loc[0]]) {
			continue
		}
		m, _ := ParseComment(lines[i][loc[0]:])
		m.Line, m.End, m.Col = i+1, i+1, loc[0]
		own := leading.MatchString(lines[i])
		closed := hasClose.MatchString(lines[i])
		for own && !closed && m.End < len(lines) {
			next := lines[m.End]
			if !leading.MatchString(next) || marked.MatchString(next) {
				break
			}
			body := strings.TrimSpace(opening.ReplaceAllString(closer.ReplaceAllString(next, ""), ""))
			if body == "" {
				break
			}
			m.Text += " " + body
			m.End++
			closed = hasClose.MatchString(next)
		}
		if own {
			m.Target = nextCode(lines, m.End)
		} else {
			m.Target = m.Line
		}
		if m.Target > 0 {
			m.TargetText = strings.TrimSpace(lines[m.Target-1])
		}
		out = append(out, m)
		i = m.End - 1
	}
	return out
}

func inComment(prefix string) bool {
	om := opener.FindStringIndex(prefix)
	if om == nil {
		return false
	}
	before := prefix[:om[0]]
	if before != "" && !strings.HasSuffix(before, " ") && !strings.HasSuffix(before, "\t") {
		return false
	}
	return strings.Count(before, `"`)%2 == 0 && strings.Count(before, "`")%2 == 0
}

func nextCode(lines []string, from int) int {
	for j := from; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "" || leading.MatchString(lines[j]) {
			continue
		}
		return j + 1
	}
	return 0
}
