package hotspots

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
)

const threshold = 3

type Input struct {
	Label      string
	RegenGlobs []string
}

var (
	sectionHead = regexp.MustCompile(`(?i)^\s*#+\s*.*(where to look|files to touch)`)
	anyHead     = regexp.MustCompile(`^\s*#+\s`)
	fileExt     = regexp.MustCompile(`\.[a-z0-9]+$`)
	codeSpan    = regexp.MustCompile("`([^`\\s]+)`")
)

func Run(ctx context.Context, d deps.Deps, in Input) error {
	issues, err := d.GH.OpenIssues(ctx, in.Label, "status:ready")
	if err != nil {
		return fmt.Errorf("list ready issues for %q: %w", in.Label, err)
	}
	named := map[string][]int{}
	for _, issue := range issues {
		for _, p := range paths(issue.Body) {
			if !config.MatchAny(in.RegenGlobs, p) {
				named[p] = append(named[p], issue.Number)
			}
		}
	}

	var hot []string
	for p, nums := range named {
		if len(nums) >= threshold {
			hot = append(hot, p)
		}
	}
	if len(hot) == 0 {
		_, err := fmt.Fprintln(d.Out, "none")
		return err
	}
	sort.Strings(hot)
	for _, p := range hot {
		nums := named[p]
		sort.Ints(nums)
		refs := make([]string, len(nums))
		for i, n := range nums {
			refs[i] = "#" + strconv.Itoa(n)
		}
		if _, err := fmt.Fprintf(d.Out, "%s %s\n", p, strings.Join(refs, " ")); err != nil {
			return err
		}
	}
	return nil
}

func paths(body string) []string {
	seen := map[string]bool{}
	var out []string
	inSection := false
	for _, line := range strings.Split(body, "\n") {
		switch {
		case sectionHead.MatchString(line):
			inSection = true
			continue
		case anyHead.MatchString(line):
			inSection = false
		}
		if !inSection {
			continue
		}
		for _, m := range codeSpan.FindAllStringSubmatch(line, -1) {
			p := strings.TrimPrefix(m[1], "./")
			if looksLikePath(p) && !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
}

func looksLikePath(s string) bool {
	return strings.Contains(s, "/") || fileExt.MatchString(s)
}
