package detect_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/detect"
)

func advise(t *testing.T, justfile string, hist detect.CIHistory) detect.Advice {
	t.Helper()
	res := detect.Scan(fstest.MapFS{"justfile": file(justfile)}, nil, detect.Options{})
	return detect.Advise(res, hist)
}

func TestAdvise(t *testing.T) {
	t.Run("formatters run before linters", func(t *testing.T) {
		a := advise(t, "lint:\n    x\nfmt:\n    y\ntest:\n    z\n", detect.CIHistory{})

		want := []string{"just fmt", "just lint", "just test"}
		if diff := cmp.Diff(want, a.Config.Checks.PrePush); diff != "" {
			t.Errorf("pre_push (-want +got):\n%s", diff)
		}
	})

	t.Run("suite kept in pre_push raises no warning", func(t *testing.T) {
		a := advise(t, "lint:\n    x\ntest:\n    z\n", detect.CIHistory{})
		if strings.Contains(a.PrePushNote, "end-of-work suite") {
			t.Errorf("PrePushNote = %q, want no warning", a.PrePushNote)
		}
	})

	t.Run("suite left out of pre_push is warned about", func(t *testing.T) {
		a := advise(t, "lint:\n    x\nci:\n    slow\n", detect.CIHistory{})
		if !strings.Contains(a.PrePushNote, "end-of-work suite") || !strings.Contains(a.PrePushNote, "just ci") {
			t.Errorf("PrePushNote = %q, want warning naming just ci", a.PrePushNote)
		}
	})

	ciTests := []struct {
		name string
		hist detect.CIHistory
		want string
		note string
	}{
		{"no history", detect.CIHistory{}, "off", "No CI history"},
		{"failures covered", detect.CIHistory{Runs: 53, FailedRuns: 41, FailedJobs: []string{"lint", "lint", "test"}}, "off", "41 of the last 53"},
		{"failures uncovered", detect.CIHistory{Runs: 20, FailedRuns: 5, FailedJobs: []string{"lint", "e2e", "e2e"}}, "async", "e2e"},
	}
	for _, tt := range ciTests {
		t.Run("ci: "+tt.name, func(t *testing.T) {
			a := advise(t, "lint:\n    x\ntest:\n    z\n", tt.hist)
			if a.Config.CI != tt.want {
				t.Errorf("ci = %q, want %q", a.Config.CI, tt.want)
			}
			if !strings.Contains(a.CINote, tt.note) {
				t.Errorf("CINote = %q, want it to contain %q", a.CINote, tt.note)
			}
		})
	}
}
