package pull_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/pull"
)

type fakeGh struct {
	pages map[string]string
	calls []string
}

func (f *fakeGh) run(_ context.Context, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	f.calls = append(f.calls, joined)
	if joined == "api user" {
		return []byte(`{"login":"me"}`), nil
	}
	for slug, out := range f.pages {
		if strings.Contains(joined, "repos/"+slug+"/") {
			if out == "" {
				return nil, errors.New("HTTP 404")
			}
			return []byte(out), nil
		}
	}
	return nil, errors.New("unexpected " + joined)
}

func checkout(t *testing.T, root, name, remote string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", remote}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

const pageOne = `[
 {"id":1,"body":"learn(go-idiomatic): use errors.Is\nbecause wrapping","path":"a.go","line":7,"diff_hunk":"@@ -1 +1 @@\n+x","user":{"login":"me"}},
 {"id":2,"body":"learn: not mine","path":"a.go","line":8,"user":{"login":"other"}},
 {"id":3,"body":"looks fine","path":"a.go","line":9,"user":{"login":"me"}}
]`

const pageTwo = `[{"id":4,"body":"Learn later: rename this","path":"b.go","original_line":3,"user":{"login":"me"}}]`

func setup(t *testing.T) (pull.Options, *fakeGh) {
	t.Helper()
	root := t.TempDir()
	checkout(t, root, "good", "git@github.com:o/good.git")
	checkout(t, root, "denied", "https://github.com/o/denied")
	checkout(t, root, "elsewhere", "https://gitlab.com/o/elsewhere")
	gh := &fakeGh{pages: map[string]string{"o/good": pageOne + pageTwo, "o/denied": ""}}
	return pull.Options{
		Ledger: filepath.Join(t.TempDir(), "learn", "learnings.jsonl"),
		Roots:  []string{root},
		Gh:     gh.run,
		Now:    func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	}, gh
}

func TestRunRecordsOwnLearnComments(t *testing.T) {
	o, _ := setup(t)
	res, err := pull.Run(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Recorded != 2 {
		t.Fatalf("Recorded = %d, want 2", res.Recorded)
	}
	got, err := ledger.Read(o.Ledger)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ledger has %d learnings, want 2", len(got))
	}
	fix, later := got[0], got[1]
	want := ledger.Learning{
		Source: "pr", Kind: "fix", Skill: "go-idiomatic", Text: "use errors.Is\nbecause wrapping", File: "a.go", Line: 7,
		Before: "@@ -1 +1 @@\n+x", Status: "pending", Origin: "https://github.com/o/good",
	}
	fix.ID, fix.TS, fix.Repo = "", "", ""
	if fix != want {
		t.Errorf("Run() fix learning = %+v, want %+v", fix, want)
	}
	if later.Kind != "later" || later.Line != 3 || later.Text != "rename this" {
		t.Errorf("later learning = %+v", later)
	}
}

func TestRunFollowsPagination(t *testing.T) {
	o, gh := setup(t)
	if _, err := pull.Run(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	for _, c := range gh.calls {
		if strings.Contains(c, "repos/") && !strings.Contains(c, "--paginate") {
			t.Errorf("gh call %q lacks --paginate", c)
		}
	}
}

func TestRunTwiceRecordsNothingNew(t *testing.T) {
	o, _ := setup(t)
	if _, err := pull.Run(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	res, err := pull.Run(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if res.Recorded != 0 {
		t.Errorf("second Run recorded %d, want 0", res.Recorded)
	}
	got, err := ledger.Read(o.Ledger)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("ledger has %d learnings, want 2", len(got))
	}
}

func TestRunSkipsUnreadableRepoWithWarning(t *testing.T) {
	o, _ := setup(t)
	res, err := pull.Run(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "o/denied") {
		t.Errorf("Warnings = %q, want one naming o/denied", res.Warnings)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(o.Ledger), "last-pull")); err == nil {
		t.Error("last-pull written despite a skipped repo")
	}
}

func TestRunUsesLastPullAsSince(t *testing.T) {
	o, gh := setup(t)
	gh.pages["o/denied"] = "[]"
	if _, err := pull.Run(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	gh.calls = nil
	if _, err := pull.Run(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	var since int
	for _, c := range gh.calls {
		if strings.HasSuffix(c, "&since=2026-01-02T03:04:05Z") {
			since++
		}
	}
	if since != 2 {
		t.Errorf("calls with since = %d, want 2: %q", since, gh.calls)
	}
}
