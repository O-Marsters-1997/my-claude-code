package main

import (
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strings"
)

const maxAreaMatches = 50

func areaMatches(root, fragment string) ([]string, error) {
	out, err := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	seen := map[string]bool{}
	for file := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		for p := file; p != "." && p != "" && p != "/"; p = path.Dir(p) {
			seen[p] = true
		}
	}
	needle := strings.ToLower(fragment)
	var matches []string
	for p := range seen {
		if strings.Contains(strings.ToLower(p), needle) {
			matches = append(matches, p)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if len(matches[i]) != len(matches[j]) {
			return len(matches[i]) < len(matches[j])
		}
		return matches[i] < matches[j]
	})
	return matches[:min(len(matches), maxAreaMatches)], nil
}
