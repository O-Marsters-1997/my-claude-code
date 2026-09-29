package redact

import (
	"regexp"
	"strings"
)

var secrets = []struct {
	re   *regexp.Regexp
	repl string
}{
	{regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?(?:-----END [A-Z ]*PRIVATE KEY-----|$)`), "[REDACTED:private-key]"},
	{regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`), "[REDACTED:aws-key]"},
	{regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36,}\b|\bgithub_pat_[A-Za-z0-9_]{22,}`), "[REDACTED:github-token]"},
	{regexp.MustCompile(`\bsk-(?:ant-)?[A-Za-z0-9_-]{20,}`), "[REDACTED:api-key]"},
	{regexp.MustCompile(`\bxox[abprs]-[A-Za-z0-9-]{10,}`), "[REDACTED:slack-token]"},
	{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{5,}\.[A-Za-z0-9_-]{5,}\.[A-Za-z0-9_-]{5,}`), "[REDACTED:jwt]"},
	{regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{16,}`), "Bearer [REDACTED:bearer]"},
	{regexp.MustCompile(`(?i)\b(password|passwd|secret|token|api[_-]?key)(["']?\s*[=:]\s*["']?)[^\s"',;&]{4,}`), "${1}${2}[REDACTED:secret]"},
}

func scrub(s string) string {
	for _, p := range secrets {
		s = p.re.ReplaceAllString(s, p.repl)
	}
	return s
}

func Clean(s string, maxLen int) string {
	s = scrub(s)
	if len(s) <= maxLen {
		return s
	}
	return strings.ToValidUTF8(s[:maxLen], "") + "…"
}

func Tail(s string, maxLen int) string {
	s = scrub(s)
	if len(s) <= maxLen {
		return s
	}
	return "…" + strings.ToValidUTF8(s[len(s)-maxLen:], "")
}
