package harvest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/marker"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/scope"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/repo"
)

// Options configures Run. Repo is any directory inside the repository,
// Ledger the learnings.jsonl path and Library the skill library root.
type Options struct {
	Repo    string
	Ledger  string
	Library string
	Block   bool
	Strip   bool
}

// Found is one fix marker still present in a staged or unstaged diff.
type Found struct {
	File string
	Line int
	Text string
}

// Result reports what a Run saw. Blocked is true when Block was set and fix
// markers remain.
type Result struct {
	Found    []Found
	Recorded int
	Blocked  bool
}

// Report lists the found markers as file:line: text, one per line.
func (r Result) Report() string {
	var b strings.Builder
	for _, f := range r.Found {
		fmt.Fprintf(&b, "%s:%d: %s\n", f.File, f.Line, f.Text)
	}
	return b.String()
}

const maxField = 4000

var hunk = regexp.MustCompile(`^@@ -\S+ \+(\d+)(?:,(\d+))? @@`)

// Run records the markers in the staged and unstaged diffs as pending
// learnings, skipping ones already in the ledger. Later markers are removed
// from the staged blobs and the working tree without blocking; Strip removes
// fix markers the same way.
func Run(ctx context.Context, o Options) (Result, error) {
	top, err := git(ctx, o.Repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return Result{}, err
	}
	top = strings.TrimSpace(top)
	known, err := ledger.Read(o.Ledger)
	if err != nil {
		return Result{}, err
	}
	have := map[string]bool{}
	for _, l := range known {
		have[l.ID] = true
	}
	root := repo.MainCheckout(top)
	origin, _ := git(ctx, top, "config", "--get", "remote.origin.url")
	origin = strings.TrimSpace(origin)

	var res Result
	toStrip := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, staged := range []bool{true, false} {
		diffArgs := []string{"-c", "core.quotepath=off", "diff", "-U0", "--src-prefix=a/", "--dst-prefix=b/", "--no-color", "--no-ext-diff"}
		if staged {
			diffArgs = append(diffArgs, "--cached")
		}
		diff, err := git(ctx, top, diffArgs...)
		if err != nil {
			return Result{}, err
		}
		changed := addedLines(diff)
		for _, file := range slices.Sorted(maps.Keys(changed)) {
			added := changed[file]
			content, err := read(ctx, top, file, staged)
			if err != nil {
				return Result{}, err
			}
			lines := strings.Split(strings.ReplaceAll(strings.TrimSuffix(content, "\n"), "\r\n", "\n"), "\n")
			for _, m := range marker.Parse(lines) {
				if !slices.Contains(added, m.Line) {
					continue
				}
				text := redact.Clean(m.Text, maxField)
				id := learningID(root, file, text)
				stripped := m.Later || o.Strip
				if stripped {
					if toStrip[file] == nil {
						toStrip[file] = map[string]bool{}
					}
					toStrip[file][id] = true
				}
				if seen[id] {
					continue
				}
				seen[id] = true
				kind := "fix"
				if m.Later {
					kind = "later"
				}
				if !stripped {
					res.Found = append(res.Found, Found{File: file, Line: m.Line, Text: text})
				}
				if have[id] {
					continue
				}
				scopeName, skill := scope.Resolve(root, o.Library, m.Skill)
				l := ledger.Learning{
					ID: id, TS: time.Now().UTC().Format(time.RFC3339), Source: "editor", Kind: kind,
					Scope: scopeName, Skill: skill, Text: text, Repo: root, Origin: origin, File: file, Line: m.Line,
					TargetText: redact.Clean(m.TargetText, maxField),
					Before:     redact.Clean(snippet(lines, m), maxField), Status: "pending",
				}
				if err := ledger.Append(o.Ledger, l); err != nil {
					return Result{}, err
				}
				res.Recorded++
			}
		}
	}
	if err := stripMarkers(ctx, top, root, toStrip); err != nil {
		return Result{}, err
	}
	slices.SortFunc(res.Found, func(a, b Found) int {
		if c := strings.Compare(a.File, b.File); c != 0 {
			return c
		}
		return a.Line - b.Line
	})
	if err := recordFixes(ctx, top, root, o.Ledger); err != nil {
		return Result{}, err
	}
	res.Blocked = o.Block && len(res.Found) > 0
	return res, nil
}

func learningID(root, file, text string) string {
	sum := sha256.Sum256([]byte(root + file + text))
	return hex.EncodeToString(sum[:])[:12]
}

func snippet(lines []string, m marker.Marker) string {
	end := max(m.End, m.Target)
	return strings.Join(lines[m.Line-1:end], "\n")
}

func read(ctx context.Context, top, file string, staged bool) (string, error) {
	if staged {
		return git(ctx, top, "show", ":"+file)
	}
	b, err := os.ReadFile(filepath.Join(top, file))
	return string(b), err
}

func addedLines(diff string) map[string][]int {
	out := map[string][]int{}
	var file string
	inHunk := false
	for _, l := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(l, "diff --git "):
			inHunk = false
		case !inHunk && strings.HasPrefix(l, "+++ "):
			file = ""
			if p, ok := strings.CutPrefix(l, "+++ b/"); ok {
				file = p
			}
		case file != "":
			m := hunk.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			inHunk = true
			start, _ := strconv.Atoi(m[1])
			n := 1
			if m[2] != "" {
				n, _ = strconv.Atoi(m[2])
			}
			for i := range n {
				out[file] = append(out[file], start+i)
			}
		}
	}
	return out
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
