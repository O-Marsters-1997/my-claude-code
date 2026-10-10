package harvest_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/harvest"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q")
	write(t, dir, "a.go", "package a\n\nfunc f() {}\n")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "init")
	return dir
}

func run(t *testing.T, dir, ledgerPath string) harvest.Result {
	t.Helper()
	res, err := harvest.Run(context.Background(), harvest.Options{Repo: dir, Ledger: ledgerPath, Block: true})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

const marked = "package a\n\n// LEARN(go-idiomatic): wrap errors, token=hunter2hunter2\nfunc f() {}\n"

func TestStagedMarkerBlocksAndRecordsOnce(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", marked)
	gitIn(t, dir, "add", "a.go")

	res := run(t, dir, lp)
	if !res.Blocked || !strings.HasPrefix(res.Report(), "a.go:3: wrap errors") {
		t.Fatalf("Run = %+v, report %q, want blocked with a.go:3", res, res.Report())
	}
	if again := run(t, dir, lp); again.Recorded != 0 || !again.Blocked {
		t.Errorf("second Run = %+v, want blocked and nothing recorded", again)
	}
	got, err := ledger.Read(lp)
	if err != nil || len(got) != 1 {
		t.Fatalf("Read = %v, %v, want one learning", got, err)
	}
	l := got[0]
	if l.Kind != "fix" || l.Status != "pending" || l.Skill != "go-idiomatic" || l.File != "a.go" || l.TargetText != "func f() {}" {
		t.Errorf("learning = %+v", l)
	}
	if strings.Contains(l.Text, "hunter2") || strings.Contains(l.Before, "hunter2") {
		t.Errorf("secret leaked: text %q before %q", l.Text, l.Before)
	}
	if !strings.Contains(l.Before, "func f() {}") {
		t.Errorf("before = %q, want the target line", l.Before)
	}
}

func TestUnstagedMarkerBlocks(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", marked)
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); !res.Blocked {
		t.Errorf("Run = %+v, want blocked", res)
	}
}

func TestCleanDiffPasses(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\nfunc f() { _ = 1 }\n")
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	if res := run(t, dir, lp); res.Blocked || len(res.Found) != 0 {
		t.Errorf("Run = %+v, want pass", res)
	}
}

func TestMarkerOutsideDiffIgnoredAndLaterSkipped(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\n// LEARN: old\nfunc f() {}\n")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "has marker")
	write(t, dir, "a.go", "package a\n\n// LEARN: old\nfunc f() {}\n// LEARN later: tidy\nfunc g() {}\n")
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); res.Blocked {
		t.Errorf("Run = %+v, want pass: old marker is not in the diff and later does not block", res)
	}
}

func TestNonBlockingStillRecords(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", marked)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	res, err := harvest.Run(context.Background(), harvest.Options{Repo: dir, Ledger: lp})
	if err != nil || res.Blocked || res.Recorded != 1 {
		t.Errorf("Run = %+v, %v, want recorded without blocking", res, err)
	}
}

func TestAddedPlusPlusLineDoesNotHideMarker(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\n++ x\n// LEARN: after pluses\nfunc f() {}\n")
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); !res.Blocked {
		t.Errorf("Run = %+v, want blocked", res)
	}
}

func TestRecordsResolvedScope(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := newRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "skills", "mine"), 0o755); err != nil {
		t.Fatal(err)
	}
	lib := t.TempDir()
	if err := os.MkdirAll(filepath.Join(lib, "skills", "code", "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "a.go", "package a\n\n// LEARN(mine): a\n// LEARN(shared): b\n// LEARN(repo): c\n// LEARN(nope): d\n// LEARN: e\nfunc f() {}\n")
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	if _, err := harvest.Run(context.Background(), harvest.Options{Repo: dir, Ledger: lp, Library: lib}); err != nil {
		t.Fatal(err)
	}
	got, err := ledger.Read(lp)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][2]string{"a": {"repo", "mine"}, "b": {"global", "shared"}, "c": {"repo", ""}, "d": {"", "nope"}, "e": {"", ""}}
	for _, l := range got {
		if w := want[l.Text]; l.Scope != w[0] || l.Skill != w[1] {
			t.Errorf("%q: scope %q skill %q, want %v", l.Text, l.Scope, l.Skill, w)
		}
	}
	if len(got) != len(want) {
		t.Errorf("recorded %d, want %d", len(got), len(want))
	}
}
