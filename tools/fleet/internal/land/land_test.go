package land_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/land"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func write(t *testing.T, dir, file, content string) {
	t.Helper()
	p := filepath.Join(dir, file)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func clone(t *testing.T, origin string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "clone")
	git(t, filepath.Dir(dir), "clone", "-q", origin, dir)
	git(t, dir, "config", "user.email", "t@example.com")
	git(t, dir, "config", "user.name", "t")
	git(t, dir, "config", "commit.gpgsign", "false")
	return dir
}

type wave struct {
	deps.GH
	t      *testing.T
	work   string
	repo   string
	prs    []gh.PR
	merged []int
	log    string
}

func newWave(t *testing.T) *wave {
	t.Helper()
	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, filepath.Dir(origin), "init", "-q", "--bare", "-b", "main", origin)
	work := clone(t, origin)
	write(t, work, "base.txt", "base\n")
	git(t, work, "add", "-A")
	git(t, work, "commit", "-q", "-m", "base")
	git(t, work, "push", "-q", "origin", "main")
	git(t, work, "push", "-q", "origin", "main:feat/demo")
	return &wave{t: t, work: work, repo: clone(t, origin), log: filepath.Join(t.TempDir(), "log")}
}

func (w *wave) addPR(number int, draft bool, file, content string) {
	w.t.Helper()
	branch := fmt.Sprintf("issue-%d/pr", number)
	git(w.t, w.work, "checkout", "-q", "-B", branch, "origin/feat/demo")
	write(w.t, w.work, file, content)
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "pr "+file)
	git(w.t, w.work, "push", "-q", "origin", branch)
	w.prs = append(w.prs, gh.PR{
		Number: number, HeadRefName: branch, IsDraft: draft,
		Body: fmt.Sprintf("Closes #%d\n", number),
	})
}

func (w *wave) OpenPRs(context.Context, string) ([]gh.PR, error) { return w.prs, nil }

func (w *wave) OpenIssues(context.Context, ...string) ([]gh.Issue, error) { return nil, nil }

func (w *wave) MergePR(_ context.Context, number int) error {
	for _, pr := range w.prs {
		if pr.Number != number {
			continue
		}
		git(w.t, w.work, "fetch", "-q", "origin")
		git(w.t, w.work, "checkout", "-q", "-B", "tip", "origin/feat/demo")
		git(w.t, w.work, "merge", "-q", "--squash", "origin/"+pr.HeadRefName)
		git(w.t, w.work, "commit", "-q", "-m", "squash "+pr.HeadRefName)
		git(w.t, w.work, "push", "-q", "origin", "tip:feat/demo")
	}
	w.merged = append(w.merged, number)
	f, err := os.OpenFile(w.log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "merge %d\n", number)
	return err
}

func (w *wave) run(suite ...string) (string, error) {
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	d.GH = w
	err := land.Run(context.Background(), d, land.Input{Dir: w.repo, Label: "project:demo", Suite: suite})
	return strings.TrimSpace(out.String()), err
}

func (w *wave) logged() []string {
	data, _ := os.ReadFile(w.log)
	return strings.Fields(strings.ReplaceAll(string(data), "\n", " "))
}

func TestRunLandsReadyAndReportsDraft(t *testing.T) {
	w := newWave(t)
	w.addPR(12, false, "b.txt", "b\n")
	w.addPR(11, true, "draft.txt", "d\n")
	w.addPR(10, false, "a.txt", "a\n")

	got, err := w.run("echo suite >> " + w.log)
	if err != nil {
		t.Fatal(err)
	}
	if got != "drafts #11" {
		t.Errorf("Run() = %q, want %q", got, "drafts #11")
	}
	want := []string{"merge", "10", "merge", "12", "suite"}
	if diff := cmp.Diff(want, w.logged()); diff != "" {
		t.Errorf("order of merges and suite (-want +got):\n%s", diff)
	}
}

func TestRunStopsAtFirstConflict(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	w.addPR(11, false, "shared.txt", "one\n")
	w.addPR(12, false, "shared.txt", "two\n")
	w.addPR(13, false, "c.txt", "c\n")

	got, err := w.run("echo suite >> " + w.log)
	if err != nil {
		t.Fatal(err)
	}
	if want := "needs-resolver #12 issue-12/pr shared.txt"; got != want {
		t.Errorf("Run() = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]int{10, 11}, w.merged); diff != "" {
		t.Errorf("merged (-want +got):\n%s", diff)
	}
}

func TestRunNeverMergesDraftOrRunsSuiteWithoutLanding(t *testing.T) {
	w := newWave(t)
	w.addPR(11, true, "draft.txt", "d\n")

	got, err := w.run("echo suite >> " + w.log)
	if err != nil {
		t.Fatal(err)
	}
	if got != "drafts #11" {
		t.Errorf("Run() = %q, want %q", got, "drafts #11")
	}
	if len(w.merged) != 0 || len(w.logged()) != 0 {
		t.Errorf("merged %v, log %v; want neither a merge nor a suite run", w.merged, w.logged())
	}
}

func TestRunDoneWithoutDrafts(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	got, err := w.run()
	if err != nil {
		t.Fatal(err)
	}
	if got != "done 1" {
		t.Errorf("Run() = %q, want %q", got, "done 1")
	}
}

func TestRunSuiteRedSingleLandedPR(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	got, err := w.run("exit 1")
	if err != nil {
		t.Fatal(err)
	}
	if want := "suite-red #10"; got != want {
		t.Errorf("Run() = %q, want %q", got, want)
	}
}

func TestRunSuiteRedNamesCulpritPR(t *testing.T) {
	for _, culprit := range []int{10, 12, 14} {
		t.Run(fmt.Sprintf("culprit %d", culprit), func(t *testing.T) {
			w := newWave(t)
			for n := 10; n <= 14; n++ {
				content := "ok\n"
				if n == culprit {
					content = "bad\n"
				}
				w.addPR(n, false, fmt.Sprintf("f%d.txt", n), content)
			}

			got, err := w.run("! grep -q bad f*.txt")
			if err != nil {
				t.Fatal(err)
			}
			if want := fmt.Sprintf("suite-red #%d", culprit); got != want {
				t.Errorf("Run() = %q, want %q", got, want)
			}
		})
	}
}

func TestRunRejectsMalformedResolvedLine(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	d := deps.DefaultDeps(&bytes.Buffer{}, &bytes.Buffer{})
	d.GH = w
	in := land.Input{Dir: w.repo, Label: "project:demo", Resolved: []string{"#10 maybe append r.md"}}
	if err := land.Run(context.Background(), d, in); err == nil {
		t.Fatal("Run() = nil, want error for malformed --resolved line")
	}
	if len(w.merged) != 0 {
		t.Errorf("merged %v, want none", w.merged)
	}
}

func TestRunAcceptsResolvedLineAndContinues(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	var out, log bytes.Buffer
	d := deps.DefaultDeps(&out, &log)
	d.GH = w
	in := land.Input{Dir: w.repo, Label: "project:demo", Resolved: []string{"#10 resolved append /r/10.md"}}
	if err := land.Run(context.Background(), d, in); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "done 1" {
		t.Errorf("Run() = %q, want %q", got, "done 1")
	}
	if !strings.Contains(log.String(), "resolver: #10 resolved append /r/10.md") {
		t.Errorf("log %q missing the resolver line", log.String())
	}
}

func TestRunFailedResolutionStopsBeforeLanding(t *testing.T) {
	w := newWave(t)
	w.addPR(10, false, "a.txt", "a\n")
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	d.GH = w
	in := land.Input{Dir: w.repo, Label: "project:demo", Resolved: []string{"#21 failed rewrite /r/my report.md"}}
	if err := land.Run(context.Background(), d, in); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "gate-failed #21" {
		t.Errorf("Run() = %q, want %q", got, "gate-failed #21")
	}
	if len(w.merged) != 0 {
		t.Errorf("merged %v, want none", w.merged)
	}
}
