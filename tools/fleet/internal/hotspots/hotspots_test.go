package hotspots_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/hotspots"
)

type fakeGH struct {
	deps.GH
	issues []gh.Issue
	err    error
}

func (f fakeGH) OpenIssues(context.Context, ...string) ([]gh.Issue, error) {
	return f.issues, f.err
}

func TestRunListError(t *testing.T) {
	boom := errors.New("boom")
	d := deps.Deps{GH: fakeGH{err: boom}, Out: &bytes.Buffer{}}
	err := hotspots.Run(context.Background(), d, hotspots.Input{Label: "l"})
	if !errors.Is(err, boom) {
		t.Errorf("Run() error = %v, want it to wrap %v", err, boom)
	}
}

func ticket(n int, files ...string) gh.Issue {
	body := "## What to build\n`ignored/outside.go`\n\n## Files to touch\n"
	for _, f := range files {
		body += fmt.Sprintf("- `%s`\n", f)
	}
	return gh.Issue{Number: n, Body: body + "\n## Blocked by\n\n- `not/a/path.go`\n"}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		issues []gh.Issue
		regen  []string
		want   string
	}{
		{
			name:   "three tickets make a hotspot",
			issues: []gh.Issue{ticket(3, "a/x.go"), ticket(1, "a/x.go", "b.go"), ticket(2, "./a/x.go")},
			want:   "a/x.go #1 #2 #3\n",
		},
		{
			name:   "two tickets are not enough",
			issues: []gh.Issue{ticket(1, "a/x.go"), ticket(2, "a/x.go")},
			want:   "none\n",
		},
		{
			name:   "regen globs are excluded",
			issues: []gh.Issue{ticket(1, "g/a.golden"), ticket(2, "g/a.golden"), ticket(3, "g/a.golden")},
			regen:  []string{"g/**"},
			want:   "none\n",
		},
		{
			name:   "tickets naming no paths",
			issues: []gh.Issue{{Number: 1, Body: "no sections"}, {Number: 2}, {Number: 3, Body: "## Files to touch\nTBD\n"}},
			want:   "none\n",
		},
		{
			name:   "identifiers are not paths",
			issues: []gh.Issue{ticket(1, "cfg.RegenGlobs"), ticket(2, "cfg.RegenGlobs"), ticket(3, "cfg.RegenGlobs")},
			want:   "none\n",
		},
		{
			name:   "directory path matches regen glob",
			issues: []gh.Issue{ticket(1, "t/golden/"), ticket(2, "t/golden/"), ticket(3, "t/golden/")},
			regen:  []string{"t/golden/**"},
			want:   "none\n",
		},
		{
			name:   "duplicate mention in one ticket counts once",
			issues: []gh.Issue{ticket(1, "a.go", "a.go"), ticket(2, "a.go")},
			want:   "none\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			d := deps.Deps{GH: fakeGH{issues: tt.issues}, Out: &out}
			if err := hotspots.Run(context.Background(), d, hotspots.Input{Label: "l", RegenGlobs: tt.regen}); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, out.String()); diff != "" {
				t.Errorf("output (-want +got):\n%s", diff)
			}
		})
	}
}
