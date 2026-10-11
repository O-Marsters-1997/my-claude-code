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

	origin    string
	worktrees string
	comments  []string
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
	return &wave{
		t: t, work: work, repo: clone(t, origin), log: filepath.Join(t.TempDir(), "log"),
		origin: origin, worktrees: t.TempDir(),
	}
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

func (w *wave) CommentPR(_ context.Context, number int, body string) error {
	w.comments = append(w.comments, fmt.Sprintf("#%d %s", number, body))
	return nil
}

func (w *wave) Run(ctx context.Context, dir, command string) ([]byte, error) {
	tp := `tp() { shift; b=$1; shift; d="$WT/$(echo "$b" | tr / _)"
if [ ! -d "$d" ]; then git clone -q "$ORIGIN" "$d" && git -C "$d" checkout -q "$b" &&
git -C "$d" config user.email t@example.com && git -C "$d" config user.name t &&
git -C "$d" config commit.gpgsign false; fi
cd "$d" && "$@"; }
`
	cmd := exec.CommandContext(ctx, "sh", "-c", tp+command)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "WT="+w.worktrees, "ORIGIN="+w.origin)
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("sh %q: %w", command, err)
	}
	return out, nil
}

func (w *wave) runWith(in land.Input) (string, error) {
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	d.GH = w
	d.Sh = w
	in.Dir, in.Label = w.repo, "project:demo"
	err := land.Run(context.Background(), d, in)
	return strings.TrimSpace(out.String()), err
}

const regenAll = "cat src/*.txt > golden/out.txt"

func (w *wave) withGolden() {
	w.t.Helper()
	write(w.t, w.work, "src/base.txt", "base\n")
	write(w.t, w.work, "golden/out.txt", "base\n")
	git(w.t, w.work, "checkout", "-q", "main")
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "golden")
	git(w.t, w.work, "push", "-q", "origin", "main", "main:feat/demo")
}

func (w *wave) addGoldenPR(number int, src string, regenerated bool) {
	w.t.Helper()
	branch := fmt.Sprintf("issue-%d/pr", number)
	git(w.t, w.work, "fetch", "-q", "origin")
	git(w.t, w.work, "checkout", "-q", "-B", branch, "origin/feat/demo")
	write(w.t, w.work, "src/"+src+".txt", src+"\n")
	if regenerated {
		git(w.t, w.work, "add", "-A")
		cmd := exec.Command("sh", "-c", regenAll)
		cmd.Dir = w.work
		if out, err := cmd.CombinedOutput(); err != nil {
			w.t.Fatalf("regen: %v\n%s", err, out)
		}
	}
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "pr "+src)
	git(w.t, w.work, "push", "-q", "origin", branch)
	w.prs = append(w.prs, gh.PR{Number: number, HeadRefName: branch, Body: fmt.Sprintf("Closes #%d\n", number)})
}

func TestRunRegeneratesGeneratedOnlyConflict(t *testing.T) {
	w := newWave(t)
	w.withGolden()
	w.addGoldenPR(10, "a", true)
	w.addGoldenPR(11, "b", true)
	regen := map[string]string{"golden/**": regenAll}

	got, err := w.runWith(land.Input{Regen: regen, PrePush: []string{"echo gate >> " + w.log}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "done 2" {
		t.Fatalf("Run() = %q, want %q", got, "done 2")
	}
	if diff := cmp.Diff([]int{10, 11}, w.merged); diff != "" {
		t.Errorf("merged (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"merge", "10", "gate", "merge", "11"}, w.logged()); diff != "" {
		t.Errorf("order of merges and pre_push (-want +got):\n%s", diff)
	}
	if len(w.comments) != 1 || !strings.HasPrefix(w.comments[0], "#11 ") || !strings.Contains(w.comments[0], "+b") {
		t.Errorf("comments = %q, want one on #11 showing the regenerated diff", w.comments)
	}

	git(t, w.work, "fetch", "-q", "origin")
	tipOut, err := exec.Command("git", "-C", w.work, "show", "origin/feat/demo:golden/out.txt").Output()
	if err != nil {
		t.Fatal(err)
	}
	if want := "a\nb\nbase\n"; string(tipOut) != want {
		t.Errorf("tip golden/out.txt = %q, want %q", tipOut, want)
	}
	merges, err := exec.Command("git", "-C", w.work, "log", "--merges", "--format=%s", "origin/issue-11/pr").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merges), "origin/feat/demo") {
		t.Errorf("issue-11/pr merge commits = %q, want a merge from origin/feat/demo", merges)
	}
}

func TestRunGateFailedStopsBeforePush(t *testing.T) {
	w := newWave(t)
	w.withGolden()
	w.addGoldenPR(10, "a", true)
	w.addGoldenPR(11, "b", true)

	got, err := w.runWith(land.Input{Regen: map[string]string{"golden/**": regenAll}, PrePush: []string{"exit 1"}})
	if err != nil {
		t.Fatal(err)
	}
	if want := "gate-failed #11 exit 1"; got != want {
		t.Errorf("Run() = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]int{10}, w.merged); diff != "" {
		t.Errorf("merged (-want +got):\n%s", diff)
	}
	if len(w.comments) != 0 {
		t.Errorf("comments = %q, want none", w.comments)
	}
}

func TestRunReportsStaleGeneratedFiles(t *testing.T) {
	w := newWave(t)
	w.withGolden()
	w.addGoldenPR(10, "a", false)

	got, err := w.runWith(land.Input{Regen: map[string]string{"golden/**": regenAll}})
	if err != nil {
		t.Fatal(err)
	}
	if want := "regen-stale golden/out.txt"; got != want {
		t.Errorf("Run() = %q, want %q", got, want)
	}
}
