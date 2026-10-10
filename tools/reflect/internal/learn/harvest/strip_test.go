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

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const laterMarked = "package a\n\n// LEARN later: tidy this\n// up soon\nfunc f() {}\n"

func TestLaterMarkerIsStrippedFromIndexAndWorktreeWithoutBlocking(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", laterMarked)
	gitIn(t, dir, "add", "a.go")

	res := run(t, dir, lp)
	if res.Blocked || res.Recorded != 1 {
		t.Fatalf("Run = %+v, want recorded and not blocked", res)
	}
	want := "package a\n\nfunc f() {}\n"
	if got := gitOut(t, dir, "show", ":a.go"); got != want {
		t.Errorf("staged blob = %q, want %q", got, want)
	}
	if got := readFile(t, dir, "a.go"); got != want {
		t.Errorf("worktree = %q, want %q", got, want)
	}
	got, err := ledger.Read(lp)
	if err != nil || len(got) != 1 || got[0].Kind != "later" || got[0].Text != "tidy this up soon" {
		t.Errorf("ledger = %+v, %v, want one later learning", got, err)
	}
}

func TestStripKeepsUnstagedHunksUnstaged(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", laterMarked)
	gitIn(t, dir, "add", "a.go")
	write(t, dir, "a.go", laterMarked+"\nfunc g() {}\n")

	run(t, dir, filepath.Join(t.TempDir(), "l.jsonl"))

	if got := gitOut(t, dir, "show", ":a.go"); got != "package a\n\nfunc f() {}\n" {
		t.Errorf("staged blob = %q, want marker gone and g unstaged", got)
	}
	if got := readFile(t, dir, "a.go"); got != "package a\n\nfunc f() {}\n\nfunc g() {}\n" {
		t.Errorf("worktree = %q", got)
	}
	if diff := gitOut(t, dir, "diff"); !strings.Contains(diff, "+func g() {}") || strings.Contains(diff, "LEARN") {
		t.Errorf("unstaged diff = %q, want only g", diff)
	}
}

func TestStripDeletesOnlyTheMarkerLines(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\n// keep me\n// LEARN later: one\n// two\n\n// other\nfunc f() {}\n")
	run(t, dir, filepath.Join(t.TempDir(), "l.jsonl"))
	want := "package a\n\n// keep me\n\n// other\nfunc f() {}\n"
	if got := readFile(t, dir, "a.go"); got != want {
		t.Errorf("worktree = %q, want %q", got, want)
	}
}

func TestStripKeepsCodeBesideTrailingMarker(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", "package a\n\nfunc f() {} // LEARN later: trailing\n")
	run(t, dir, filepath.Join(t.TempDir(), "l.jsonl"))
	if got := readFile(t, dir, "a.go"); got != "package a\n\nfunc f() {}\n" {
		t.Errorf("worktree = %q, want code kept", got)
	}
}

func TestStripFlagRemovesFixMarkersWithoutBlocking(t *testing.T) {
	dir := newRepo(t)
	lp := filepath.Join(t.TempDir(), "l.jsonl")
	write(t, dir, "a.go", marked)
	gitIn(t, dir, "add", "a.go")

	res, err := harvest.Run(context.Background(), harvest.Options{Repo: dir, Ledger: lp, Block: true, Strip: true})
	if err != nil || res.Blocked || res.Recorded != 1 {
		t.Fatalf("Run = %+v, %v, want recorded and not blocked", res, err)
	}
	if strings.Contains(gitOut(t, dir, "show", ":a.go"), "LEARN") || strings.Contains(readFile(t, dir, "a.go"), "LEARN") {
		t.Error("fix marker survived --strip")
	}
	if got, _ := ledger.Read(lp); len(got) != 1 || got[0].Kind != "fix" {
		t.Errorf("ledger = %+v, want one fix learning", got)
	}
}

func TestFixMarkerWithoutStripStaysPut(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "a.go", marked)
	run(t, dir, filepath.Join(t.TempDir(), "l.jsonl"))
	if got := readFile(t, dir, "a.go"); got != marked {
		t.Errorf("worktree = %q, want untouched", got)
	}
}

func TestCommitFormsProduceMarkerFreeTrees(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "reflect")
	build := exec.Command("go", "build", "-o", bin, "../../../cmd/reflect")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	for _, tt := range []struct {
		name   string
		commit []string
	}{
		{"commit -a", []string{"commit", "-qam", "x"}},
		{"commit path", []string{"commit", "-qm", "x", "a.go"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := newRepo(t)
			hook := "#!/bin/sh\nexec " + bin + " learn harvest --block\n"
			if err := os.WriteFile(filepath.Join(dir, ".git", "hooks", "pre-commit"), []byte(hook), 0o755); err != nil {
				t.Fatal(err)
			}
			write(t, dir, "a.go", laterMarked)
			gitIn(t, dir, tt.commit...)
			if tree := gitOut(t, dir, "show", "HEAD:a.go"); strings.Contains(tree, "LEARN") {
				t.Errorf("committed tree = %q, want no marker", tree)
			}
			if got := readFile(t, dir, "a.go"); strings.Contains(got, "LEARN") {
				t.Errorf("worktree = %q, want no marker", got)
			}
		})
	}
}
