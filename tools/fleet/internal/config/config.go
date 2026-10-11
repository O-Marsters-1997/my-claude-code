package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CI     string            `toml:"ci"`
	Checks Checks            `toml:"checks"`
	Regen  map[string]string `toml:"regen"`
}

type Checks struct {
	PrePush []string `toml:"pre_push"`
	Suite   []string `toml:"suite"`
}

// Load reads .claude/fleet.toml under dir. A missing file yields the zero
// Config. When the file has no [regen] table, regen comes from the Fleet
// section of CLAUDE.md.
func Load(dir string) (Config, error) {
	var cfg Config
	tomlPath := filepath.Join(dir, ".claude", "fleet.toml")
	meta, err := toml.DecodeFile(tomlPath, &cfg)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return Config{}, fmt.Errorf("load %s: %w", tomlPath, err)
	}
	if meta.IsDefined("regen") {
		return cfg, nil
	}
	regen, err := claudeMDRegen(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		return Config{}, err
	}
	cfg.Regen = regen
	return cfg, nil
}

func (c Config) RegenGlobs() []string {
	globs := make([]string, 0, len(c.Regen))
	for g := range c.Regen {
		globs = append(globs, g)
	}
	sort.Strings(globs)
	return globs
}

// MatchAny reports whether name matches any glob. A "**" segment matches
// zero or more directories; other segments use path.Match.
func MatchAny(globs []string, name string) bool {
	for _, g := range globs {
		if matchSegments(strings.Split(g, "/"), strings.Split(name, "/")) {
			return true
		}
	}
	return false
}

func matchSegments(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(name); i++ {
			if matchSegments(pat[1:], name[i:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], name[0])
	return err == nil && ok && matchSegments(pat[1:], name[1:])
}

// claudeMDRegen reads list items under a "Fleet" heading, each naming a glob
// and a command in backticks: - `web/dist/**`: `npm run build`.
func claudeMDRegen(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	regen := map[string]string{}
	inFleet := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			inFleet = strings.EqualFold(strings.TrimSpace(strings.TrimLeft(line, "#")), "fleet")
			continue
		}
		if !inFleet || !strings.HasPrefix(line, "-") {
			continue
		}
		spans := strings.Split(line, "`")
		if len(spans) >= 5 {
			regen[spans[1]] = spans[3]
		}
	}
	return regen, scanner.Err()
}
