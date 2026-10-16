package land_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/land"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
)

func testMetrics(t *testing.T) {
	regen := map[string]string{"golden/**": regenAll}
	tests := []struct {
		name  string
		setup func(w *wave)
		in    land.Input
		want  metrics.Row
	}{
		{
			name:  "a clean run with resolver lines and a green suite",
			setup: func(w *wave) { w.addPR(10, false, "a.txt", "a\n") },
			in: land.Input{
				Suite: []string{"true"}, Started: time.Now().Add(-time.Hour),
				Resolved: []string{"#9 resolved append /r/9.md", "#8 resolved rewrite /r/8.md"},
			},
			want: metrics.Row{
				Label: "project:demo", Suite: "green", ResolverSpawns: 2,
				ConflictClasses: metrics.Classes{Append: 1, Rewrite: 1},
			},
		},
		{
			name:  "a failed resolver counts as a gate failure",
			setup: func(w *wave) { w.addPR(10, false, "a.txt", "a\n") },
			in:    land.Input{Resolved: []string{"#8 failed rewrite /r/8.md"}},
			want: metrics.Row{
				Label: "project:demo", Suite: "skipped", ResolverSpawns: 1, ResolverFailures: 1,
				GateFailures: 1, ConflictClasses: metrics.Classes{Rewrite: 1},
			},
		},
		{
			name: "a run stopped for subagents still records its conflicts",
			setup: func(w *wave) {
				w.addPR(10, false, "base.txt", "one\n")
				w.addPR(11, false, "base.txt", "two\n")
			},
			want: metrics.Row{Label: "project:demo", TrialConflicts: 1, Suite: "skipped"},
		},
		{
			name: "a regen fix is counted",
			setup: func(w *wave) {
				w.withGolden()
				w.addGoldenPR(10, "a")
				w.addGoldenPR(11, "b")
			},
			in:   land.Input{Regen: regen},
			want: metrics.Row{Label: "project:demo", TrialConflicts: 1, RegenFixes: 1, Suite: "skipped"},
		},
		{
			name: "a pre_push failure is a gate failure",
			setup: func(w *wave) {
				w.withGolden()
				w.addGoldenPR(10, "a")
				w.addGoldenPR(11, "b")
			},
			in:   land.Input{Regen: regen, PrePush: []string{"exit 1"}},
			want: metrics.Row{Label: "project:demo", TrialConflicts: 1, GateFailures: 1, Suite: "skipped"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWave(t)
			tt.setup(w)
			if _, err := w.run(tt.in); err != nil {
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
			if !tt.in.Started.IsZero() && got.WallClockS < 3600 {
				t.Errorf("wall_clock_s = %d, want at least 3600", got.WallClockS)
			}
			got.TS, got.WallClockS = time.Time{}, 0
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("row (-want +got):\n%s", diff)
			}
		})
	}
}
