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
	res, err := harvest.Run(context.Background(), harvest.Options{Repo: dir, Ledger: ledgerPath})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

const marked = "package a\n\n// REFLECT(go-idiomatic): wrap errors, token=hunter2hunter2\nfunc f() {}\n"

func TestStagedMarkerIsRecordedOnceAndStripped(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", marked)
	gitIn(t, dir, "add", "a.go")

	if res := run(t, dir, lp); res.Recorded != 1 {
		t.Fatalf("Run = %+v, want one recorded", res)
	}
	if again := run(t, dir, lp); again.Recorded != 0 {
		t.Errorf("second Run = %+v, want nothing recorded", again)
	}
	if strings.Contains(gitOut(t, dir, "show", ":a.go"), "REFLECT") || strings.Contains(readFile(t, dir, "a.go"), "REFLECT") {
		t.Error("marker survived in the index or worktree")
	}
	got, err := ledger.Read(lp)
	if err != nil || len(got) != 1 {
		t.Fatalf("Read = %v, %v, want one learning", got, err)
	}
	l := got[0]
	if l.Status != "pending" || l.Skill != "go-idiomatic" || l.File != "a.go" || l.TargetText != "func f() {}" {
		t.Errorf("learning = %+v", l)
	}
	if strings.Contains(l.Text, "hunter2") || strings.Contains(l.Before, "hunter2") {
		t.Errorf("secret leaked: text %q before %q", l.Text, l.Before)
	}
	if !strings.Contains(l.Before, "func f() {}") {
		t.Errorf("before = %q, want the target line", l.Before)
	}
}

func TestUnstagedMarkerIsRecorded(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", marked)
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); res.Recorded != 1 {
		t.Errorf("Run = %+v, want one recorded", res)
	}
}

func TestCleanDiffRecordsNothing(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\nfunc f() { _ = 1 }\n")
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); res.Recorded != 0 {
		t.Errorf("Run = %+v, want nothing recorded", res)
	}
}

func TestMarkerOutsideDiffIgnoredAndLaterFormDoesNotParse(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\n// REFLECT: old\nfunc f() {}\n")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "has marker")
	body := "package a\n\n// REFLECT: old\nfunc f() {}\n// REFLECT later: tidy\nfunc g() {}\n"
	write(t, dir, "a.go", body)
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); res.Recorded != 0 {
		t.Errorf("Run = %+v, want nothing: old marker is not in the diff and later does not parse", res)
	}
	if got := readFile(t, dir, "a.go"); got != body {
		t.Errorf("worktree = %q, want untouched", got)
	}
}

func TestAddedPlusPlusLineDoesNotHideMarker(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\n++ x\n// REFLECT: after pluses\nfunc f() {}\n")
	if res := run(t, dir, filepath.Join(t.TempDir(), "l.jsonl")); res.Recorded != 1 {
		t.Errorf("Run = %+v, want one recorded", res)
	}
}

func TestSQLDashMarkerIsRecordedAndStripped(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "q.sql", "-- REFLECT: index this\nselect 1;\nselect 2; -- REFLECT: trailing\n")
	gitIn(t, dir, "add", "q.sql")
	if res := run(t, dir, lp); res.Recorded != 2 {
		t.Fatalf("Run = %+v, want two recorded", res)
	}
	if got := readFile(t, dir, "q.sql"); got != "select 1;\nselect 2;\n" {
		t.Errorf("worktree = %q", got)
	}
}

func TestRecordsResolvedScope(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := newRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "skills", "mine"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, ".claude", "skills", "mine"), "SKILL.md", "x")
	lib := t.TempDir()
	if err := os.MkdirAll(filepath.Join(lib, "skills", "code", "shared"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "a.go", "package a\n\n// REFLECT(mine): a\n// REFLECT(shared): b\n// REFLECT(repo): c\n// REFLECT(nope): d\n// REFLECT: e\nfunc f() {}\n")
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

func afterOf(t *testing.T, lp string) string {
	t.Helper()
	got, err := ledger.Read(lp)
	if err != nil || len(got) != 1 {
		t.Fatalf("Read = %v, %v, want one learning", got, err)
	}
	return got[0].After
}

func repoWithOldTarget(t *testing.T) string {
	t.Helper()
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\nfunc f() {} // old\n")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "old target")
	return dir
}

func TestAfterIsTheStagedHunkThatChangesTheTarget(t *testing.T) {
	dir := repoWithOldTarget(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", "package a\n\n// REFLECT: drop the comment\nfunc f() {}\n")
	gitIn(t, dir, "add", "a.go")
	run(t, dir, lp)
	if got := afterOf(t, lp); got != "-func f() {} // old\n+func f() {}" {
		t.Errorf("after = %q", got)
	}
}

func TestAfterStaysEmptyWhenCommitLeavesTargetAlone(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", marked)
	gitIn(t, dir, "add", "a.go")
	run(t, dir, lp)
	write(t, dir, "a.go", "package a\n\nfunc f() error { return nil }\n")
	gitIn(t, dir, "add", "a.go")
	run(t, dir, lp)
	if got := afterOf(t, lp); got != "" {
		t.Errorf("after = %q, want empty on a later commit", got)
	}
}

func TestAfterIgnoresHunksThatDoNotTouchTheTarget(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\nfunc f() {}\n\nfunc g() {}\n")
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-qm", "g")
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", "package a\n\nfunc f() {}\n\n// REFLECT: tidy\nfunc g() int { return 1 }\n")
	gitIn(t, dir, "add", "a.go")
	run(t, dir, lp)
	if got := afterOf(t, lp); got != "" {
		t.Errorf("after = %q, want empty: no staged hunk removes the target text", got)
	}
}
