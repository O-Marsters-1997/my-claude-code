package detect

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
)

var formatterRecipe = regexp.MustCompile(`^(fmt|format)$`)

type CIHistory struct {
	Runs       int
	FailedRuns int
	FailedJobs []string
}

type Advice struct {
	Config      config.Config
	CINote      string
	PrePushNote string
	SuiteNote   string
	RegenNotes  map[string]string
}

func Advise(res Result, hist CIHistory) Advice {
	cfg := res.Config
	cfg.Checks.PrePush = formattersFirst(cfg.Checks.PrePush)
	a := Advice{Config: cfg, RegenNotes: map[string]string{}}

	a.PrePushNote = "Run in order before every push; formatters come first so linters see formatted code."
	if len(cfg.Checks.Suite) > 0 && !slices.Contains(cfg.Checks.PrePush, cfg.Checks.Suite[0]) {
		a.PrePushNote += fmt.Sprintf(" '%s' is left out; pre_push replaces the agent's own end-of-work suite, so agents will not run it.", cfg.Checks.Suite[0])
	}
	if len(res.Excluded) > 0 {
		a.PrePushNote += " Not quick enough or not a check: " + strings.Join(res.Excluded, ", ") + "."
	}

	if len(cfg.Checks.Suite) > 0 {
		a.SuiteNote = fmt.Sprintf("The full suite agents run once at the end; '%s' is the broadest recipe found.", cfg.Checks.Suite[0])
	}
	for glob, cmd := range cfg.Regen {
		a.RegenNotes[glob] = fmt.Sprintf("Files matching %s are generated; a branch conflict there is fixed by running '%s', not by hand-merging.", glob, cmd)
	}

	a.Config.CI, a.CINote = chooseCI(cfg.Checks.PrePush, hist)
	return a
}

func chooseCI(prePush []string, hist CIHistory) (string, string) {
	if hist.Runs == 0 {
		return "off", "No CI history was available (gh missing, unauthenticated or no runs), so CI is not watched."
	}
	evidence := fmt.Sprintf("%d of the last %d CI runs failed", hist.FailedRuns, hist.Runs)
	failed := slices.Clone(hist.FailedJobs)
	slices.Sort(failed)
	uncovered := jobsWithoutMatch(slices.Compact(failed), prePush)
	if len(uncovered) > 0 {
		return "async", fmt.Sprintf("%s, including jobs pre_push cannot cover (%s); async watches each PR's checks without blocking the wave, at the cost of background polling.", evidence, strings.Join(uncovered, ", "))
	}
	return "off", fmt.Sprintf("%s, but only in jobs pre_push already covers; off avoids polling CI for failures agents catch locally.", evidence)
}

func formattersFirst(cmds []string) []string {
	isFmt := func(c string) bool {
		f := strings.Fields(c)
		return len(f) > 0 && formatterRecipe.MatchString(f[len(f)-1])
	}
	out := slices.Clone(cmds)
	slices.SortStableFunc(out, func(a, b string) int {
		switch {
		case isFmt(a) == isFmt(b):
			return 0
		case isFmt(a):
			return -1
		}
		return 1
	})
	return out
}
