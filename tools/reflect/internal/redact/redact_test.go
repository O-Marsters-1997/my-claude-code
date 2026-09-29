package redact_test

import (
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
)

func TestClean(t *testing.T) {
	tests := map[string]string{
		"AKIAABCDEFGHIJKLMNOP":                                                "[REDACTED:aws-key]",
		"ghp_" + strings.Repeat("a", 36):                                      "[REDACTED:github-token]",
		"sk-ant-" + strings.Repeat("x", 30):                                   "[REDACTED:api-key]",
		"Authorization: Bearer abcdefghijklmnop12345":                         "Authorization: Bearer [REDACTED:bearer]",
		"export API_KEY=hunter2hunter2":                                       "export API_KEY=[REDACTED:secret]",
		`{"password": "s3cretvalue"}`:                                         `{"password": "[REDACTED:secret]"}`,
		"eyJhbGciOiJI.eyJzdWIiOiIx.SflKxwRJSMeKKF2QT4fwpMeJf36POk":            "[REDACTED:jwt]",
		"-----BEGIN RSA PRIVATE KEY-----\nabc\n-----END RSA PRIVATE KEY-----": "[REDACTED:private-key]",
		"go test ./... -run TestFoo":                                          "go test ./... -run TestFoo",
	}
	for in, want := range tests {
		if got := redact.Clean(in, 1000); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCleanRedactsBeforeTruncating(t *testing.T) {
	in := strings.Repeat("a", 490) + " token=abcdefghijklmnop"
	if got := redact.Clean(in, 500); strings.Contains(got, "abcdefghij") {
		t.Errorf("Clean leaked a secret at the truncation boundary: %q", got[480:])
	}
}
