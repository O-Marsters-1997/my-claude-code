package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const seenSessionsMax = 64

type instrState struct {
	Hash  string            `json:"hash"`
	Files map[string]string `json:"files"`
	Seen  map[string]string `json:"seen"`
}

func seenMark(hash, branch string) string { return hash + "@" + branch }

func statePath(dir string) string { return filepath.Join(dir, "instr.json") }

func loadInstrState(dir string) instrState {
	var st instrState
	if b, err := os.ReadFile(statePath(dir)); err == nil {
		_ = json.Unmarshal(b, &st)
	}
	return st
}

func saveInstrState(dir string, st instrState) {
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	_ = os.WriteFile(statePath(dir), b, 0o644)
}

func diffFiles(prev, cur map[string]string) (added, changed map[string]string, removed []string) {
	added, changed = map[string]string{}, map[string]string{}
	for path, h := range cur {
		old, ok := prev[path]
		switch {
		case !ok:
			added[path] = h
		case old != h:
			changed[path] = h
		}
	}
	for path := range prev {
		if _, ok := cur[path]; !ok {
			removed = append(removed, path)
		}
	}
	sort.Strings(removed)
	return added, changed, removed
}

var instrGlobs = []string{
	"AGENTS.md",
	"CLAUDE.md",
	"settings.json",
	"rules/*.md",
	"agents/*.md",
	"skills/*/SKILL.md",
	"skills/*/references/*.md",
}

func hash12(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:6])
}

func instrFiles(projectDir string) map[string]string {
	roots := []string{projectDir, filepath.Join(projectDir, ".claude"), filepath.Join(projectDir, ".agents")}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".claude"))
	}
	files := map[string]string{}
	seen := map[string]bool{}
	for _, root := range roots {
		for _, g := range instrGlobs {
			matches, _ := filepath.Glob(filepath.Join(root, g))
			for _, m := range matches {
				real, err := filepath.EvalSymlinks(m)
				if err != nil {
					real = m
				}
				if seen[real] {
					continue
				}
				if b, err := os.ReadFile(m); err == nil {
					seen[real] = true
					files[m] = hash12(string(b))
				}
			}
		}
	}
	return files
}

func combinedHash(files map[string]string) string {
	lines := make([]string, 0, len(files))
	for path, h := range files {
		lines = append(lines, path+"\t"+h)
	}
	sort.Strings(lines)
	return hash12(lines...)
}

func git(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
