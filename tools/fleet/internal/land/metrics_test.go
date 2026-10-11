package land_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/land"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
)

func TestRunAppendsOneMetricsRowPerRun(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	w.addPR(11, true, "draft.txt", "d\n")

	d := deps.DefaultDeps(&bytes.Buffer{}, &bytes.Buffer{})
	d.GH = w
	in := land.Input{
		Dir: w.repo, Label: "project:demo", Suite: []string{"true"}, Started: time.Now().Add(-time.Hour),
		Resolved: []string{"9 resolved append /r/9.md", "8 failed rewrite /r/8.md"},
	}
	if err := land.Run(context.Background(), d, in); err != nil {
		t.Fatal(err)
	}

	rows, err := metrics.Read(w.repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("metrics rows = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.WallClockS < 3600 {
		t.Errorf("wall_clock_s = %d, want at least 3600", got.WallClockS)
	}
	got.TS, got.WallClockS = time.Time{}, 0
	want := metrics.Row{
		Label: "project:demo", PRsLanded: 1, DraftsSkipped: 1, Suite: "green",
		ResolverSpawns: 2, ResolverFailures: 1, ConflictClasses: metrics.Classes{Append: 1, Rewrite: 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("row (-want +got):\n%s", diff)
	}
}

func TestRunRecordsRowWhenStoppedEarly(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "base.txt", "one\n")
	w.addPR(11, false, "base.txt", "two\n")

	got, err := w.run()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "needs-resolver #11") {
		t.Fatalf("Run() = %q, want needs-resolver #11", got)
	}
	rows, err := metrics.Read(w.repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PRsLanded != 1 || rows[0].TrialConflicts != 1 || rows[0].Suite != "skipped" {
		t.Errorf("rows = %+v, want one row with 1 landed, 1 trial conflict, suite skipped", rows)
	}
}
