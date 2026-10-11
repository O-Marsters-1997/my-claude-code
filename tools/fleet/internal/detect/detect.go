// Package detect infers a repository's .claude/fleet.toml from its recipes,
// CI workflows and generated files.
package detect

import (
	"bufio"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
)

const maxQuickSeconds = 60

var (
	quickRecipe = regexp.MustCompile(`^(lint|fmt|format|fmtcheck|check|typecheck|tc|vet|test|unit|assets|css)$`)
	workflowJob = regexp.MustCompile(`^  ([A-Za-z0-9_-]+):`)
)

// Options configures Scan. A nil Measure skips timing the candidate checks.
type Options struct {
	Measure func(cmd string) (time.Duration, error)
}

// Result is the inferred config plus what Scan could not place.
type Result struct {
	Config config.Config
	// Notes are human-readable lines about checks left out or CI jobs with no local equivalent.
	Notes []string
	// Unmapped are generated paths for which no regenerating command was found.
	Unmapped []string
}

// Scan reads fsys, whose git-tracked files are listed in tracked.
func Scan(fsys fs.FS, tracked []string, opts Options) Result {
	recipes := loadRecipes(fsys)
	cfg := config.Config{CI: "off"}
	var res Result

	var excluded []string
	for _, r := range recipes {
		if !quickRecipe.MatchString(r.name) {
			excluded = append(excluded, r.cmd)
			continue
		}
		if opts.Measure != nil {
			d, err := opts.Measure(r.cmd)
			if err != nil {
				res.Notes = append(res.Notes, fmt.Sprintf("'%s' failed while measuring, left out", r.cmd))
				excluded = append(excluded, r.cmd)
				continue
			}
			if d > maxQuickSeconds*time.Second {
				res.Notes = append(res.Notes, fmt.Sprintf("'%s' took %ds, left out", r.cmd, int(d.Seconds())))
				excluded = append(excluded, r.cmd)
				continue
			}
		}
		cfg.Checks.PrePush = append(cfg.Checks.PrePush, r.cmd)
	}
	if len(excluded) > 0 {
		res.Notes = append(res.Notes, "Left out of pre_push (slow or not a quick check): "+strings.Join(excluded, ","))
	}
	if jobs := unmatchedJobs(fsys, cfg.Checks.PrePush); len(jobs) > 0 {
		res.Notes = append(res.Notes, "CI jobs with no local equivalent: "+strings.Join(jobs, ","))
	}

	cfg.Checks.Suite = suite(recipes)
	cfg.Regen, res.Unmapped = detectRegen(fsys, tracked, recipes)
	res.Config = cfg
	return res
}

func suite(recipes []recipe) []string {
	for _, want := range []string{"ci", "test"} {
		for _, r := range recipes {
			if r.name == want {
				return []string{r.cmd}
			}
		}
	}
	return nil
}

func unmatchedJobs(fsys fs.FS, picked []string) []string {
	files, _ := fs.Glob(fsys, ".github/workflows/*.y*ml")
	seen := map[string]bool{}
	var jobs []string
	for _, f := range files {
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			continue
		}
		inJobs := false
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "jobs:" {
				inJobs = true
				continue
			}
			if inJobs && line != "" && line[0] != ' ' && line[0] != '#' {
				inJobs = false
			}
			if m := workflowJob.FindStringSubmatch(line); inJobs && m != nil && !seen[m[1]] {
				seen[m[1]] = true
				jobs = append(jobs, m[1])
			}
		}
	}
	sort.Strings(jobs)

	var unmatched []string
	for _, j := range jobs {
		stem, _, _ := strings.Cut(strings.ReplaceAll(j, "_", "-"), "-")
		hit := false
		for _, p := range picked {
			hit = hit || strings.Contains(p, stem)
		}
		if !hit {
			unmatched = append(unmatched, j)
		}
	}
	return unmatched
}
