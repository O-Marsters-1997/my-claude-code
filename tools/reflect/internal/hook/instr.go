package hook

import (
	"crypto/sha256"
	"encoding/hex"
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
