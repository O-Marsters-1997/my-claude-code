package metrics_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
)

func TestAppendReadRoundTrip(t *testing.T) {
	repo := t.TempDir()
	a := metrics.Row{TS: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Label: "project:a", PRsLanded: 2, Suite: "green"}
	b := metrics.Row{TS: a.TS.Add(time.Hour), Label: "project:b", Suite: "skipped"}
	for _, r := range []metrics.Row{a, b} {
		if err := metrics.Append(repo, r); err != nil {
			t.Fatal(err)
		}
	}
	got, err := metrics.Read(repo)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]metrics.Row{a, b}, got); diff != "" {
		t.Errorf("Read() (-want +got):\n%s", diff)
	}
	data, _ := os.ReadFile(metrics.Path(repo))
	if n := strings.Count(string(data), "\n"); n != 2 {
		t.Errorf("file has %d lines, want 2", n)
	}
}

func TestReadMissingFile(t *testing.T) {
	got, err := metrics.Read(t.TempDir())
	if err != nil || got != nil {
		t.Errorf("Read(empty) = %v, %v; want nil, nil", got, err)
	}
}

func TestApplyResolved(t *testing.T) {
	var r metrics.Row
	err := r.ApplyResolved([]string{"11 resolved append /r/1.md", "12 failed rewrite /r/2.md", "13 resolved rewrite /r/3.md"})
	if err != nil {
		t.Fatal(err)
	}
	want := metrics.Row{ResolverSpawns: 3, ResolverFailures: 1, ConflictClasses: metrics.Classes{Append: 1, Rewrite: 2}}
	if diff := cmp.Diff(want, r); diff != "" {
		t.Errorf("ApplyResolved (-want +got):\n%s", diff)
	}
}

func TestApplyResolvedRejectsMalformed(t *testing.T) {
	for _, l := range []string{"11 resolved", "11 maybe append x", "11 resolved merge x"} {
		var r metrics.Row
		if err := r.ApplyResolved([]string{l}); err == nil {
			t.Errorf("ApplyResolved(%q) succeeded, want error", l)
		}
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		name string
		row  metrics.Row
		want int
	}{
		{"quiet", metrics.Row{PRsLanded: 5, TrialConflicts: 1}, 0},
		{"trial conflicts at 30% do not flag", metrics.Row{PRsLanded: 7, TrialConflicts: 3}, 0},
		{"trial conflicts above 30%", metrics.Row{PRsLanded: 2, TrialConflicts: 1}, 1},
		{"resolver failures at one in three do not flag", metrics.Row{ResolverSpawns: 3, ResolverFailures: 1}, 0},
		{"resolver failures above one in three", metrics.Row{ResolverSpawns: 2, ResolverFailures: 1}, 1},
		{"append majority", metrics.Row{ConflictClasses: metrics.Classes{Append: 2, Rewrite: 1}}, 1},
		{"append tie does not flag", metrics.Row{ConflictClasses: metrics.Classes{Append: 1, Rewrite: 1}}, 0},
		{"all three", metrics.Row{PRsLanded: 1, TrialConflicts: 1, ResolverSpawns: 1, ResolverFailures: 1, ConflictClasses: metrics.Classes{Append: 1}}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.row.Flags(); len(got) != tt.want {
				t.Errorf("Flags() = %v, want %d flags", got, tt.want)
			}
		})
	}
}

func TestSelect(t *testing.T) {
	rows := []metrics.Row{{Label: "a", PRsLanded: 1}, {Label: "b", PRsLanded: 2}, {Label: "a", PRsLanded: 3}, {Label: "a", PRsLanded: 4}}
	got := metrics.Select(rows, "a", 2)
	want := []metrics.Row{{Label: "a", PRsLanded: 3}, {Label: "a", PRsLanded: 4}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Select (-want +got):\n%s", diff)
	}
	if got := metrics.Select(rows, "", 0); len(got) != 4 {
		t.Errorf("Select(all) kept %d rows, want 4", len(got))
	}
}

func TestTrend(t *testing.T) {
	if got := metrics.Trend([]metrics.Row{{}}); !strings.Contains(got, "at least two") {
		t.Errorf("Trend(one) = %q", got)
	}
	got := metrics.Trend([]metrics.Row{{TrialConflicts: 3, WallClockS: 600}, {}, {TrialConflicts: 1, WallClockS: 300}})
	for _, want := range []string{"3 waves", "trial_conflicts 3 -> 1", "wall_clock_s 600 -> 300"} {
		if !strings.Contains(got, want) {
			t.Errorf("Trend() = %q, missing %q", got, want)
		}
	}
}

func TestStartFromState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.md")
	if err := os.WriteFile(path, []byte("Label: project:demo\nStarted: 2026-01-02T03:04:05Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := metrics.StartFromState(path, "project:demo"); !got.Equal(want) {
		t.Errorf("StartFromState = %v, want %v", got, want)
	}
	if got := metrics.StartFromState(path, "project:other"); !got.IsZero() {
		t.Errorf("StartFromState(other label) = %v, want zero", got)
	}
	if got := metrics.StartFromState(filepath.Join(t.TempDir(), "none"), "project:demo"); !got.IsZero() {
		t.Errorf("StartFromState(missing) = %v, want zero", got)
	}
}
