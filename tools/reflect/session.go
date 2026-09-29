package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var instrGlobs = []string{
	"AGENTS.md",
	"CLAUDE.md",
	"settings.json",
	"rules/*.md",
	"agents/*.md",
	"skills/*/SKILL.md",
	"skills/*/references/*.md",
}

func instrFiles(projectDir string) map[string]string {
	roots := []string{projectDir, filepath.Join(projectDir, ".claude"), filepath.Join(projectDir, ".agents")}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".claude"))
	}
	files := map[string]string{}
	for _, root := range roots {
		for _, g := range instrGlobs {
			matches, _ := filepath.Glob(filepath.Join(root, g))
			for _, m := range matches {
				if b, err := os.ReadFile(m); err == nil {
					files[m] = hash12(string(b))
				}
			}
		}
	}
	return files
}

func combinedHash(files map[string]string) string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p+"\t"+files[p])
	}
	sort.Strings(paths)
	return hash12(paths...)
}

func git(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
