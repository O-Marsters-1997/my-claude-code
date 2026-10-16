package queue

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
)

type Entry struct {
	PR      gh.PR
	Tickets []int
}

var (
	closesRef  = regexp.MustCompile(`(?i)\bcloses?\s+#(\d+)`)
	issueRef   = regexp.MustCompile(`#(\d+)`)
	blockedSec = regexp.MustCompile(`(?i)^\s*#+\s*blocked by`)
	anyHeading = regexp.MustCompile(`^\s*#+\s`)
)

func Build(prs []gh.PR, issues []gh.Issue) []Entry {
	entries := make([]Entry, len(prs))
	for i, pr := range prs {
		entries[i] = Entry{PR: pr, Tickets: tickets(pr)}
	}
	return order(entries, Blockers(issues))
}

func tickets(pr gh.PR) []int {
	var nums []int
	for _, m := range closesRef.FindAllStringSubmatch(pr.Body, -1) {
		n, _ := strconv.Atoi(m[1])
		nums = append(nums, n)
	}
	if len(nums) == 0 {
		nums = []int{pr.Number}
	}
	sort.Ints(nums)
	return nums
}

func Blockers(issues []gh.Issue) map[int][]int {
	out := map[int][]int{}
	for _, issue := range issues {
		inSection := false
		for _, line := range strings.Split(issue.Body, "\n") {
			switch {
			case blockedSec.MatchString(line):
				inSection = true
				continue
			case anyHeading.MatchString(line):
				inSection = false
			}
			if !inSection {
				continue
			}
			for _, m := range issueRef.FindAllStringSubmatch(line, -1) {
				n, _ := strconv.Atoi(m[1])
				out[issue.Number] = append(out[issue.Number], n)
			}
		}
	}
	return out
}

func order(entries []Entry, blockedBy map[int][]int) []Entry {
	sort.Slice(entries, func(i, j int) bool { return entries[i].Tickets[0] < entries[j].Tickets[0] })
	owner := map[int]int{}
	for i, e := range entries {
		for _, t := range e.Tickets {
			owner[t] = i
		}
	}
	waitsOn := make([]map[int]bool, len(entries))
	for i, e := range entries {
		waitsOn[i] = map[int]bool{}
		for _, t := range e.Tickets {
			for _, b := range blockedBy[t] {
				if j, ok := owner[b]; ok && j != i {
					waitsOn[i][j] = true
				}
			}
		}
	}

	var out []Entry
	placed := make([]bool, len(entries))
	for len(out) < len(entries) {
		next := -1
		for i := range entries {
			if !placed[i] && allPlaced(waitsOn[i], placed) {
				next = i
				break
			}
		}
		if next < 0 {
			next = firstUnplaced(placed)
		}
		placed[next] = true
		out = append(out, entries[next])
	}
	return out
}

func allPlaced(deps map[int]bool, placed []bool) bool {
	for j := range deps {
		if !placed[j] {
			return false
		}
	}
	return true
}

func firstUnplaced(placed []bool) int {
	for i, p := range placed {
		if !p {
			return i
		}
	}
	return -1
}
