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
		Resolved: []string{"#9 resolved append /r/9.md", "#8 resolved rewrite /r/8.md"},
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
		ResolverSpawns: 2, ConflictClasses: metrics.Classes{Append: 1, Rewrite: 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("row (-want +got):\n%s", diff)
	}
}

func TestRunRecordsRowOnGateFailure(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")

	d := deps.DefaultDeps(&bytes.Buffer{}, &bytes.Buffer{})
	d.GH = w
	in := land.Input{Dir: w.repo, Label: "project:demo", Resolved: []string{"#8 failed rewrite /r/8.md"}}
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
	row := rows[0]
	row.TS = time.Time{}
	want := metrics.Row{
		Label: "project:demo", Suite: "skipped", ResolverSpawns: 1, ResolverFailures: 1,
		GateFailures: 1, ConflictClasses: metrics.Classes{Rewrite: 1},
	}
	if diff := cmp.Diff(want, row); diff != "" {
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
	if len(rows) != 1 {
		t.Fatalf("metrics rows = %d, want 1", len(rows))
	}
	row := rows[0]
	row.TS = time.Time{}
	want := metrics.Row{Label: "project:demo", PRsLanded: 1, TrialConflicts: 1, Suite: "skipped"}
	if diff := cmp.Diff(want, row); diff != "" {
		t.Errorf("row (-want +got):\n%s", diff)
	}
}

func TestRunRecordsRegenRows(t *testing.T) {
	regen := map[string]string{"golden/**": regenAll}
	tests := []struct {
		name  string
		in    land.Input
		fresh bool
		want  metrics.Row
	}{
		{
			name:  "regen fix counted per PR landed through the tier",
			in:    land.Input{Regen: regen},
			fresh: true,
			want:  metrics.Row{Label: "project:demo", PRsLanded: 2, RegenFixes: 1, Suite: "skipped"},
		},
		{
			name:  "pre_push gate failure counts as a gate failure",
			in:    land.Input{Regen: regen, PrePush: []string{"exit 1"}},
			fresh: true,
			want:  metrics.Row{Label: "project:demo", PRsLanded: 1, GateFailures: 1, Suite: "skipped"},
		},
		{
			name: "regen-stale still appends one row",
			in:   land.Input{Regen: regen},
			want: metrics.Row{Label: "project:demo", PRsLanded: 1, Suite: "skipped"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWave(t)
			w.withGolden()
			w.addGoldenPR(10, "a", tt.fresh)
			if tt.fresh {
				w.addGoldenPR(11, "b", true)
			}
			if _, err := w.runWith(tt.in); err != nil {
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
			got.TS = time.Time{}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("row (-want +got):\n%s", diff)
			}
		})
	}
}
