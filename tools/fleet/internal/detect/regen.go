package detect

import (
	"bufio"
	"bytes"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
)

const headerBytes = 2048

var (
	generateRecipe = regexp.MustCompile(`^(generate|gen|codegen|generated)$`)
	updateFlag     = regexp.MustCompile(`flag\.Bool\(\s*"(update[a-z-]*|golden)"`)
	outputArg      = regexp.MustCompile(`(?:\s-o|--output|--outfile|--out|>)[ =]*(\S+)`)
	generatorConfs = []struct{ file, cmd string }{
		{"sqlc.yaml", "sqlc generate"},
		{"sqlc.yml", "sqlc generate"},
		{"sqlc.json", "sqlc generate"},
		{"buf.gen.yaml", "buf generate"},
		{"gqlgen.yml", "gqlgen generate"},
	}
)

type generator struct {
	recipeCmd  string
	configCmds []string
	goGenerate bool
}

// detectRegen returns glob -> command for generated files, and the paths it
// found to be generated but could not map to a command.
func detectRegen(fsys fs.FS, tracked []string, recipes []recipe) (map[string]string, []string) {
	gen := findGenerator(fsys, tracked, recipes)
	regen := map[string]string{}
	var unmapped []string

	add := func(glob, cmd string) {
		if config.MatchAny(sortedKeys(regen), glob) {
			return
		}
		regen[glob] = cmd
	}
	covered := func(file string) bool {
		return config.MatchAny(sortedKeys(regen), file)
	}

	for glob, cmd := range outputsNamedByRecipes(tracked, recipes) {
		add(glob, cmd)
	}
	for glob, cmd := range goldens(fsys, tracked) {
		add(glob, cmd)
	}

	globs := append(generatedGlobs(fsys, tracked), attributeGlobs(fsys, tracked)...)
	for _, glob := range globs {
		if covered(glob) {
			continue
		}
		if cmd := gen.commandFor(); cmd != "" {
			add(glob, cmd)
		} else {
			unmapped = append(unmapped, glob)
		}
	}
	sort.Strings(unmapped)
	if len(regen) == 0 {
		regen = nil
	}
	return regen, unmapped
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func findGenerator(fsys fs.FS, tracked []string, recipes []recipe) generator {
	var g generator
	for _, r := range recipes {
		if generateRecipe.MatchString(r.name) {
			g.recipeCmd = r.cmd
			break
		}
	}
	for _, c := range generatorConfs {
		if _, err := fs.Stat(fsys, c.file); err == nil {
			g.configCmds = append(g.configCmds, c.cmd)
		}
	}
	for _, f := range tracked {
		if !strings.HasSuffix(f, ".go") {
			continue
		}
		if data, err := fs.ReadFile(fsys, f); err == nil && bytes.Contains(data, []byte("\n//go:generate ")) {
			g.goGenerate = true
			break
		}
	}
	return g
}

func (g generator) commandFor() string {
	switch {
	case g.recipeCmd != "":
		return g.recipeCmd
	case len(g.configCmds) > 0:
		return g.configCmds[0]
	case g.goGenerate:
		return "go generate ./..."
	}
	return ""
}

func outputsNamedByRecipes(tracked []string, recipes []recipe) map[string]string {
	set := make(map[string]bool, len(tracked))
	for _, f := range tracked {
		set[f] = true
	}
	out := map[string]string{}
	for _, r := range recipes {
		for _, m := range outputArg.FindAllStringSubmatch(r.body, -1) {
			p := path.Clean(strings.Trim(m[1], `"'`))
			if set[p] {
				out[p] = r.cmd
			}
		}
	}
	return out
}

func goldens(fsys fs.FS, tracked []string) map[string]string {
	out := map[string]string{}
	for _, f := range tracked {
		if !strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := fs.ReadFile(fsys, f)
		if err != nil {
			continue
		}
		m := updateFlag.FindSubmatch(data)
		if m == nil {
			continue
		}
		dir := path.Dir(f)
		prefix := path.Join(dir, "testdata") + "/"
		hasData := false
		for _, t := range tracked {
			hasData = hasData || strings.HasPrefix(t, prefix)
		}
		if !hasData {
			continue
		}
		pkg := "./" + dir
		if dir == "." {
			pkg = "."
		}
		out[prefix+"**"] = "go test " + pkg + " -" + string(m[1])
	}
	return out
}

func generatedGlobs(fsys fs.FS, tracked []string) []string {
	byDir := map[string][]string{}
	generated := map[string]bool{}
	for _, f := range tracked {
		d := path.Dir(f)
		byDir[d] = append(byDir[d], f)
		if hasGeneratedHeader(fsys, f) {
			generated[f] = true
		}
	}

	var globs []string
	seenDir := map[string]bool{}
	for _, f := range tracked {
		if !generated[f] {
			continue
		}
		d := path.Dir(f)
		if d == "." {
			globs = append(globs, f)
			continue
		}
		if seenDir[d] {
			continue
		}
		all := true
		for _, sibling := range byDir[d] {
			all = all && generated[sibling]
		}
		if all {
			seenDir[d] = true
			globs = append(globs, d+"/**")
		} else {
			globs = append(globs, f)
		}
	}
	return globs
}

func hasGeneratedHeader(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, headerBytes)
	n, _ := f.Read(buf)
	buf = buf[:n]
	if bytes.IndexByte(buf, 0) >= 0 {
		return false
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for i := 0; i < 5 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") || strings.Contains(line, "DO NOT EDIT") || strings.Contains(line, "@generated") {
			return true
		}
	}
	return false
}

func attributeGlobs(fsys fs.FS, tracked []string) []string {
	data, err := fs.ReadFile(fsys, ".gitattributes")
	if err != nil {
		return nil
	}
	var globs []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		generated := false
		for _, attr := range fields[1:] {
			switch attr {
			case "linguist-generated", "linguist-generated=true":
				generated = true
			case "-linguist-generated", "linguist-generated=false":
				generated = false
			}
		}
		if !generated {
			continue
		}
		glob := attributeToGlob(fields[0])
		for _, f := range tracked {
			if config.MatchAny([]string{glob}, f) {
				globs = append(globs, glob)
				break
			}
		}
	}
	return globs
}

func attributeToGlob(pattern string) string {
	switch {
	case strings.HasSuffix(pattern, "/"):
		pattern += "**"
	case !strings.Contains(strings.TrimPrefix(pattern, "/"), "/"):
		if !strings.HasPrefix(pattern, "/") {
			pattern = "**/" + pattern
		}
	}
	return strings.TrimPrefix(pattern, "/")
}
