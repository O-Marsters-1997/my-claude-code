package harvest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/marker"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
)

var commentOpener = regexp.MustCompile(`\s*(?:/{2,}|#+|\{?/\*+|<!--)\s*$`)

// stripMarkers removes the markers whose ids are listed per file from both
// the staged blob and the working tree copy, each independently, so unstaged
// hunks stay unstaged. It never runs git add.
func stripMarkers(ctx context.Context, top, root string, ids map[string]map[string]bool) error {
	for _, file := range slices.Sorted(maps.Keys(ids)) {
		drop := func(m marker.Marker) bool {
			return ids[file][learningID(root, file, redact.Clean(m.Text, maxField))]
		}
		if err := stripIndex(ctx, top, file, drop); err != nil {
			return err
		}
		if err := stripWorktree(top, file, drop); err != nil {
			return err
		}
	}
	return nil
}

func stripIndex(ctx context.Context, top, file string, drop func(marker.Marker) bool) error {
	entry, err := git(ctx, top, "ls-files", "-s", "--", file)
	if err != nil {
		return err
	}
	fields := strings.Fields(entry)
	if len(fields) < 3 || fields[2] != "0" {
		return nil
	}
	content, err := git(ctx, top, "show", ":"+file)
	if err != nil {
		return err
	}
	stripped, changed := removeMarkers(content, drop)
	if !changed {
		return nil
	}
	sha, err := gitStdin(ctx, top, stripped, "hash-object", "-w", "--stdin")
	if err != nil {
		return err
	}
	_, err = git(ctx, top, "update-index", "--cacheinfo", fields[0]+","+strings.TrimSpace(sha)+","+file)
	return err
}

func stripWorktree(top, file string, drop func(marker.Marker) bool) error {
	path := filepath.Join(top, file)
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	stripped, changed := removeMarkers(string(b), drop)
	if !changed {
		return nil
	}
	return os.WriteFile(path, []byte(stripped), info.Mode().Perm())
}

// removeMarkers deletes the comment lines of every marker drop accepts. A
// marker trailing code loses only its comment, so the code stays.
func removeMarkers(content string, drop func(marker.Marker) bool) (string, bool) {
	raw := strings.Split(content, "\n")
	lines := make([]string, len(raw))
	for i, l := range raw {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	gone := map[int]bool{}
	changed := false
	for _, m := range marker.Parse(lines) {
		if !drop(m) {
			continue
		}
		changed = true
		if m.Target != m.Line {
			for n := m.Line; n <= m.End; n++ {
				gone[n-1] = true
			}
			continue
		}
		code := strings.TrimRight(commentOpener.ReplaceAllString(lines[m.Line-1][:m.Col], ""), " \t")
		if strings.TrimSpace(code) == "" {
			gone[m.Line-1] = true
			continue
		}
		raw[m.Line-1] = code + strings.TrimPrefix(raw[m.Line-1], lines[m.Line-1])
	}
	if !changed {
		return content, false
	}
	kept := raw[:0:0]
	for i, l := range raw {
		if !gone[i] {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n"), true
}

func gitStdin(ctx context.Context, dir, stdin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
