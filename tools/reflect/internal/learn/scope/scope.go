package scope

import (
	"os"
	"path/filepath"
)

// Resolve maps the name in LEARN(name) to a scope and skill: a repo skill wins
// over a library one, "repo" is repo scope with no skill, and a bare or
// unknown name has an empty scope.
func Resolve(repoRoot, library, name string) (scopeName, skill string) {
	switch {
	case name == "":
		return "", ""
	case name == "." || name == "..":
		return "", name
	case name == "repo":
		return "repo", ""
	case isFile(filepath.Join(repoRoot, ".claude", "skills", name, "SKILL.md")):
		return "repo", name
	case installed(library, name):
		return "global", name
	}
	return "", name
}

func installed(library, name string) bool {
	if library != "" {
		if hits, _ := filepath.Glob(filepath.Join(library, "skills", "*", name)); len(hits) > 0 {
			return true
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return isDir(filepath.Join(home, ".agents", "skills", name)) ||
		isDir(filepath.Join(home, ".claude", "skills", name))
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}
