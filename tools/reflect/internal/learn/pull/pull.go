package pull

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/redact"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/repo"
)

type Gh func(ctx context.Context, args ...string) ([]byte, error)

// Options configures Run. Roots are directories whose children are checked
// for GitHub remotes; Gh defaults to the gh binary.
type Options struct {
	Ledger string
	Roots  []string
	Gh     Gh
	Now    func() time.Time
}

// Result reports what a Run recorded and the repos it skipped.
type Result struct {
	Recorded int
	Warnings []string
}

const (
	maxField  = 4000
	stampFile = "last-pull"
)

var (
	learnComment = regexp.MustCompile(`(?is)^\s*learn(?:\(([\w.-]+)\))?(\s+later)?:\s*(.*)`)
	githubRemote = regexp.MustCompile(`github\.com[:/]([\w.-]+)/([\w.-]+?)(?:\.git)?/?$`)
)

type comment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	Path      string `json:"path"`
	Line      int    `json:"line"`
	OrigLine  int    `json:"original_line"`
	DiffHunk  string `json:"diff_hunk"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// Run records the authenticated user's learn: review comments on every
// GitHub remote found since the previous pull as pending learnings. A repo
// gh cannot read becomes a warning and keeps the last-pull time where it was.
func Run(ctx context.Context, o Options) (Result, error) {
	if o.Gh == nil {
		o.Gh = runGh
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	started := o.Now().UTC()

	out, err := o.Gh(ctx, "api", "user")
	if err != nil {
		return Result{}, fmt.Errorf("gh user: %w", err)
	}
	var me struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(out, &me); err != nil || me.Login == "" {
		return Result{}, errors.New("gh api user returned no login")
	}

	known, err := ledger.Read(o.Ledger)
	if err != nil {
		return Result{}, err
	}
	have := map[string]bool{}
	for _, l := range known {
		have[l.ID] = true
	}
	since := readStamp(o.Ledger)
	repos := discover(ctx, known, o.Roots)

	var res Result
	for _, slug := range slices.Sorted(maps.Keys(repos)) {
		comments, err := fetch(ctx, o.Gh, slug, since)
		if err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("skipping %s: %v", slug, err))
			continue
		}
		for _, c := range comments {
			m := learnComment.FindStringSubmatch(c.Body)
			if m == nil || c.User.Login != me.Login {
				continue
			}
			id := learningID(slug, c.ID)
			if have[id] {
				continue
			}
			have[id] = true
			kind := "fix"
			if m[2] != "" {
				kind = "later"
			}
			line := c.Line
			if line == 0 {
				line = c.OrigLine
			}
			l := ledger.Learning{
				ID: id, TS: cmp.Or(c.CreatedAt, started.Format(time.RFC3339)), Source: "pr", Kind: kind, Skill: m[1],
				Text: redact.Clean(strings.TrimSpace(m[3]), maxField), Repo: repos[slug],
				Origin: "https://github.com/" + slug, File: c.Path, Line: line,
				Before: redact.Clean(c.DiffHunk, maxField), Status: "pending",
			}
			if err := ledger.Append(o.Ledger, l); err != nil {
				return res, err
			}
			res.Recorded++
		}
	}
	if len(res.Warnings) == 0 {
		if err := writeStamp(o.Ledger, started); err != nil {
			return res, err
		}
	}
	return res, nil
}

func fetch(ctx context.Context, gh Gh, slug, since string) ([]comment, error) {
	endpoint := "repos/" + slug + "/pulls/comments?per_page=100"
	if since != "" {
		endpoint += "&since=" + since
	}
	out, err := gh(ctx, "api", "--paginate", endpoint)
	if err != nil {
		return nil, err
	}
	// --paginate prints one JSON array per page, back to back.
	dec := json.NewDecoder(bytes.NewReader(out))
	var all []comment
	for {
		var page []comment
		if err := dec.Decode(&page); errors.Is(err, io.EOF) {
			return all, nil
		} else if err != nil {
			return nil, err
		}
		all = append(all, page...)
	}
}

func discover(ctx context.Context, known []ledger.Learning, roots []string) map[string]string {
	repos := map[string]string{}
	add := func(remote, dir string) {
		if m := githubRemote.FindStringSubmatch(strings.TrimSpace(remote)); m != nil {
			slug := m[1] + "/" + m[2]
			if repos[slug] == "" {
				repos[slug] = dir
			}
		}
	}
	for _, l := range known {
		add(l.Origin, l.Repo)
	}
	dirs := map[string]bool{}
	for _, l := range known {
		if l.Repo != "" {
			dirs[l.Repo] = true
		}
	}
	for _, root := range roots {
		entries, _ := os.ReadDir(root)
		for _, e := range entries {
			if e.IsDir() {
				dirs[filepath.Join(root, e.Name())] = true
			}
		}
	}
	for dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
		cmd.Dir = dir
		if b, err := cmd.Output(); err == nil {
			add(string(b), repo.MainCheckout(dir))
		}
	}
	return repos
}

func learningID(slug string, commentID int64) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "pr %s %d", slug, commentID))
	return hex.EncodeToString(sum[:])[:12]
}

func readStamp(ledgerPath string) string {
	b, err := os.ReadFile(filepath.Join(filepath.Dir(ledgerPath), stampFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func writeStamp(ledgerPath string, t time.Time) error {
	dir := filepath.Dir(ledgerPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, stampFile), []byte(t.Format(time.RFC3339)+"\n"), 0o644)
}

func runGh(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
