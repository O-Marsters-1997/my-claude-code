package metrics

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	trialConflictShare = 0.30
	resolverFailShare  = 1.0 / 3
)

type Classes struct {
	Append  int `json:"append"`
	Rewrite int `json:"rewrite"`
}

type Row struct {
	TS               time.Time `json:"ts"`
	Label            string    `json:"label"`
	PRsLanded        int       `json:"prs_landed"`
	DraftsSkipped    int       `json:"drafts_skipped"`
	TrialConflicts   int       `json:"trial_conflicts"`
	RegenFixes       int       `json:"regen_fixes"`
	ResolverSpawns   int       `json:"resolver_spawns"`
	ResolverFailures int       `json:"resolver_failures"`
	ResolverTokens   int       `json:"resolver_tokens"`
	ConflictClasses  Classes   `json:"conflict_classes"`
	GateFailures     int       `json:"gate_failures"`
	Suite            string    `json:"suite"`
	Bisects          int       `json:"bisects"`
	WallClockS       int       `json:"wall_clock_s"`
}

func Path(repo string) string {
	return filepath.Join(repo, ".claude", "fleet-metrics.jsonl")
}

func (r *Row) AddResolution(failed bool, class string) {
	r.ResolverSpawns++
	if failed {
		r.ResolverFailures++
	}
	switch class {
	case "append":
		r.ConflictClasses.Append++
	case "rewrite":
		r.ConflictClasses.Rewrite++
	}
}

func Append(repo string, row Row) error {
	p := Path(repo)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(row)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func Read(repo string) ([]Row, error) {
	f, err := os.Open(Path(repo))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows []Row
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}
		var r Row
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", Path(repo), n, err)
		}
		rows = append(rows, r)
	}
	return rows, sc.Err()
}

// Select keeps rows for label (all when empty) and then the last n (all when n <= 0).
func Select(rows []Row, label string, n int) []Row {
	var out []Row
	for _, r := range rows {
		if label == "" || r.Label == label {
			out = append(out, r)
		}
	}
	if n > 0 && len(out) > n {
		out = out[len(out)-n:]
	}
	return out
}

func (r Row) Flags() []string {
	var flags []string
	prs := r.PRsLanded + r.DraftsSkipped + r.TrialConflicts
	if prs > 0 && float64(r.TrialConflicts)/float64(prs) > trialConflictShare {
		flags = append(flags, fmt.Sprintf("trial conflicts %d of %d PRs (above 30%%)", r.TrialConflicts, prs))
	}
	if r.ResolverSpawns > 0 && float64(r.ResolverFailures)/float64(r.ResolverSpawns) > resolverFailShare {
		flags = append(flags, fmt.Sprintf("resolver failures %d of %d (above one in three)", r.ResolverFailures, r.ResolverSpawns))
	}
	if r.ConflictClasses.Append*2 > r.ConflictClasses.Append+r.ConflictClasses.Rewrite {
		flags = append(flags, "append is the majority conflict class: consider mergiraf or a hotspot refactor")
	}
	return flags
}

func (r Row) String() string {
	return fmt.Sprintf("%s %s landed=%d drafts=%d trial_conflicts=%d regen=%d resolvers=%d failed=%d append=%d rewrite=%d gate_failures=%d suite=%s bisects=%d wall=%ds",
		r.TS.UTC().Format("2006-01-02T15:04Z"), r.Label, r.PRsLanded, r.DraftsSkipped, r.TrialConflicts,
		r.RegenFixes, r.ResolverSpawns, r.ResolverFailures, r.ConflictClasses.Append, r.ConflictClasses.Rewrite,
		r.GateFailures, r.Suite, r.Bisects, r.WallClockS)
}

// Trend compares the first and last row's trial conflicts and wall clock.
func Trend(rows []Row) string {
	if len(rows) < 2 {
		return "trend: need at least two waves"
	}
	first, last := rows[0], rows[len(rows)-1]
	return fmt.Sprintf("trend over %d waves: trial_conflicts %d -> %d, resolver_spawns %d -> %d, wall_clock_s %d -> %d",
		len(rows), first.TrialConflicts, last.TrialConflicts, first.ResolverSpawns, last.ResolverSpawns,
		first.WallClockS, last.WallClockS)
}

// StartFromState reads the "Started: <RFC3339>" line dispatch writes to state.md.
// A missing file, line or mismatched label yields the zero time.
func StartFromState(path, label string) time.Time {
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}
	}
	var started time.Time
	matches := false
	for _, l := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(l, "Label: "):
			matches = strings.TrimPrefix(l, "Label: ") == label
		case strings.HasPrefix(l, "Started: "):
			started, _ = time.Parse(time.RFC3339, strings.TrimPrefix(l, "Started: "))
		}
	}
	if !matches {
		return time.Time{}
	}
	return started
}
