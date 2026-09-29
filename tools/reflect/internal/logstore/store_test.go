package logstore_test

import (
	"os"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
)

func TestRewriteKeepsLinesAppendedWhileRewriting(t *testing.T) {
	s := logstore.At(t.TempDir())
	for _, sid := range []string{"a", "b"} {
		if err := s.Append(logstore.Event{Kind: "edit", SessionID: sid}); err != nil {
			t.Fatal(err)
		}
	}
	err := s.Rewrite(func(lines [][]byte) ([][]byte, error) {
		if err := s.Append(logstore.Event{Kind: "edit", SessionID: "late"}); err != nil {
			t.Fatal(err)
		}
		return lines[:1], nil
	})
	if err != nil {
		t.Fatal(err)
	}
	events, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].SessionID != "a" || events[1].SessionID != "late" {
		t.Errorf("events after rewrite = %+v, want a then late", events)
	}
}

func TestRewriteLeavesLogUntouchedOnError(t *testing.T) {
	s := logstore.At(t.TempDir())
	if err := s.Append(logstore.Event{Kind: "edit", SessionID: "a"}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(s.LogPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Rewrite(func([][]byte) ([][]byte, error) { return nil, os.ErrInvalid }); err == nil {
		t.Fatal("Rewrite returned nil, want the callback's error")
	}
	after, err := os.ReadFile(s.LogPath())
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("log changed: %q -> %q", before, after)
	}
	leftovers, _ := os.ReadDir(s.Dir())
	for _, f := range leftovers {
		if strings.HasSuffix(f.Name(), ".tmp") {
			t.Errorf("temp file left behind: %s", f.Name())
		}
	}
}

func TestRewriteOnMissingLogIsANoop(t *testing.T) {
	s := logstore.At(t.TempDir())
	if err := s.Rewrite(func(l [][]byte) ([][]byte, error) { return l, nil }); err != nil {
		t.Errorf("Rewrite on a missing log = %v, want nil", err)
	}
}
