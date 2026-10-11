package detect

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/fs"
	"regexp"
	"sort"
	"strings"
)

type recipe struct {
	name string
	cmd  string
	body string
}

var recipeHeader = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_-]*)[^:=]*:([^=]|$)`)

func loadRecipes(fsys fs.FS) []recipe {
	var all []recipe
	for _, n := range []string{"justfile", "Justfile", ".justfile"} {
		if data, err := fs.ReadFile(fsys, n); err == nil {
			all = append(all, parseRecipes(data, "just")...)
			break
		}
	}
	for _, n := range []string{"Makefile", "makefile", "GNUmakefile"} {
		if data, err := fs.ReadFile(fsys, n); err == nil {
			all = append(all, parseRecipes(data, "make")...)
			break
		}
	}
	if data, err := fs.ReadFile(fsys, "package.json"); err == nil {
		all = append(all, parseScripts(data)...)
	}

	seen := map[string]bool{}
	var out []recipe
	for _, r := range all {
		if seen[r.name] {
			continue
		}
		seen[r.name] = true
		out = append(out, r)
	}
	return out
}

func parseRecipes(data []byte, tool string) []recipe {
	var out []recipe
	var cur *recipe
	flush := func() {
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if cur != nil {
				cur.body += line + "\n"
			}
			continue
		}
		flush()
		if m := recipeHeader.FindStringSubmatch(line); m != nil {
			cur = &recipe{name: m[1], cmd: tool + " " + m[1]}
		}
	}
	flush()
	return out
}

func parseScripts(data []byte) []recipe {
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}
	names := make([]string, 0, len(pkg.Scripts))
	for n := range pkg.Scripts {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]recipe, 0, len(names))
	for _, n := range names {
		out = append(out, recipe{name: n, cmd: "npm run " + n, body: pkg.Scripts[n]})
	}
	return out
}
