package scope

import (
	"os"
	"path/filepath"
)

// Resolve maps the name in LEARN(name) to a scope and skill. A skill in the
// repo wins over one in the library or the user's skill directories. "repo"
// is reserved for repo scope with no skill. A bare or unknown name has an
// empty scope; an unknown name is kept as the skill.
func Resolve(repoRoot, library, name string) (scope, skill string) {
	switch {
	case name == "":
		return "", ""
	case name == "repo":
		return "repo", ""
	case isDir(filepath.Join(repoRoot, ".claude", "skills", name)):
		return "repo", name
	case inLibrary(library, name):
		return "global", name
	}
	return "", name
}

func inLibrary(library, name string) bool {
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
