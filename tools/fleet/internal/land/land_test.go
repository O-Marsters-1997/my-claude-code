package land_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
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
	issues []gh.Issue
	all    []gh.PR
	log    string

	origin     string
	worktrees  string
	comments   []string
	checks     map[int][]gh.Check
	pending    map[int]int
	checkCalls int
	plan       string
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
		origin: origin, worktrees: t.TempDir(), plan: filepath.Join(t.TempDir(), "plan.json"),
		checks: map[int][]gh.Check{}, pending: map[int]int{},
	}
}

func (w *wave) addPR(number int, draft bool, file, content string) {
	w.t.Helper()
	branch := fmt.Sprintf("issue-%d/pr", number)
	git(w.t, w.work, "fetch", "-q", "origin")
	git(w.t, w.work, "checkout", "-q", "-B", branch, "origin/feat/demo")
	write(w.t, w.work, file, content)
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "pr "+file)
	git(w.t, w.work, "push", "-q", "origin", branch)
	w.prs = append(w.prs, gh.PR{
		Number: number, HeadRefName: branch, IsDraft: draft,
		Body: fmt.Sprintf("Closes #%d\n", number),
		URL:  fmt.Sprintf("https://example.test/o/r/pull/%d", number),
	})
}

func (w *wave) subagentMerges(number, earlier int, file, content string) {
	w.t.Helper()
	branch := fmt.Sprintf("issue-%d/pr", number)
	git(w.t, w.work, "fetch", "-q", "origin")
	git(w.t, w.work, "checkout", "-q", "-B", branch, "origin/"+branch)
	_ = exec.Command("git", "-C", w.work, "merge", "--no-commit", "--no-ff", fmt.Sprintf("origin/issue-%d/pr", earlier)).Run()
	write(w.t, w.work, file, content)
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "resolve")
	git(w.t, w.work, "push", "-q", "origin", branch)
}

func (w *wave) OpenPRs(context.Context, string) ([]gh.PR, error) { return w.prs, nil }

func (w *wave) OpenPRsAll(context.Context) ([]gh.PR, error) { return w.all, nil }

func (w *wave) OpenIssues(context.Context, ...string) ([]gh.Issue, error) { return w.issues, nil }

func (w *wave) Checks(_ context.Context, number int) ([]gh.Check, error) {
	w.checkCalls++
	if w.pending[number] > 0 {
		w.pending[number]--
		return []gh.Check{{Name: "ci", Bucket: "pending"}}, nil
	}
	return w.checks[number], nil
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

func (w *wave) run(in land.Input) (string, error) {
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	d.GH = w
	d.Sh = w
	in.Dir, in.Label, in.Plan = w.repo, "project:demo", w.plan
	err := land.Run(context.Background(), d, in)
	return strings.TrimSpace(out.String()), err
}

func (w *wave) runLast(in land.Input) string {
	w.t.Helper()
	out, err := w.run(in)
	if err != nil {
		w.t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	return lines[len(lines)-1]
}

func (w *wave) tipSHA() string {
	git(w.t, w.work, "fetch", "-q", "origin")
	return gitOut(w.t, w.work, "rev-parse", "origin/feat/demo")
}

func (w *wave) logged() []string {
	data, _ := os.ReadFile(w.log)
	return strings.Fields(string(data))
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

func (w *wave) addGoldenPR(number int, src string) {
	w.t.Helper()
	branch := fmt.Sprintf("issue-%d/pr", number)
	git(w.t, w.work, "fetch", "-q", "origin")
	git(w.t, w.work, "checkout", "-q", "-B", branch, "origin/feat/demo")
	write(w.t, w.work, "src/"+src+".txt", src+"\n")
	git(w.t, w.work, "add", "-A")
	cmd := exec.Command("sh", "-c", regenAll)
	cmd.Dir = w.work
	if out, err := cmd.CombinedOutput(); err != nil {
		w.t.Fatalf("regen: %v\n%s", err, out)
	}
	git(w.t, w.work, "add", "-A")
	git(w.t, w.work, "commit", "-q", "-m", "pr "+src)
	git(w.t, w.work, "push", "-q", "origin", branch)
	w.prs = append(w.prs, gh.PR{
		Number: number, HeadRefName: branch, Body: fmt.Sprintf("Closes #%d\n", number),
		URL: fmt.Sprintf("https://example.test/o/r/pull/%d", number),
	})
}

func TestRun(t *testing.T) {
	t.Run("--check", testCheck)
	t.Run("metrics", testMetrics)

	t.Run("CI still pending after the timeout is not ready", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		w.pending[10] = 1000
		got := w.runLast(land.Input{CIPoll: time.Millisecond, CITimeout: 20 * time.Millisecond})
		if want := "gate-failed #10 ci-pending"; got != want {
			t.Errorf("Run() = %q, want %q", got, want)
		}
	})

	t.Run("--check reads CI once instead of waiting", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		w.pending[10] = 1000
		got := w.runLast(land.Input{Check: true, CIPoll: time.Millisecond, CITimeout: time.Hour})
		if got != "reconcile" || w.checkCalls != 1 {
			t.Errorf("Run(--check) = %q after %d CI reads, want reconcile after 1", got, w.checkCalls)
		}
	})

	t.Run("errors when no PR targets the feature branch", func(t *testing.T) {
		w := newWave(t)
		if _, err := w.run(land.Input{}); err == nil {
			t.Error("Run() = nil, want error for a wave with no PRs")
		}
	})
	t.Run("a clean wave, drafts included, is ready and nothing is merged", func(t *testing.T) {
		w := newWave(t)
		w.addPR(12, false, "b.txt", "b\n")
		w.addPR(11, true, "draft.txt", "d\n")
		w.addPR(10, false, "a.txt", "a\n")
		before := w.tipSHA()

		out, err := w.run(land.Input{Suite: []string{"echo suite >> " + w.log}})
		if err != nil {
			t.Fatal(err)
		}
		var order []string
		for _, l := range strings.Split(out, "\n") {
			order = append(order, strings.Fields(l)[0])
		}
		if diff := cmp.Diff([]string{"#10", "#11", "#12", "ready"}, order); diff != "" {
			t.Errorf("Run() lines (-want +got):\n%s", diff)
		}
		if !strings.HasSuffix(out, "ready 3") {
			t.Errorf("Run() = %q, want it to end ready 3", out)
		}
		if got := w.tipSHA(); got != before {
			t.Errorf("feat/demo moved from %s to %s", before, got)
		}
		if diff := cmp.Diff([]string{"suite"}, w.logged()); diff != "" {
			t.Errorf("suite runs (-want +got):\n%s", diff)
		}
	})

	t.Run("a conflicting PR pulls in the earlier one and gets one subagent", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "shared.txt", "one\n")
		w.addPR(11, false, "shared.txt", "two\n")
		w.addPR(12, false, "c.txt", "c\n")

		out, err := w.run(land.Input{})
		if err != nil {
			t.Fatal(err)
		}
		if want := "#11 issue-11/pr ci=none conflicts=none pull-in=#10 needs-work"; !strings.Contains(out, want) {
			t.Errorf("Run() = %q, want line %q", out, want)
		}
		if !strings.HasSuffix(out, "\nneeds-work #11") {
			t.Errorf("Run() = %q, want terminal needs-work #11", out)
		}
		data, err := os.ReadFile(w.plan)
		if err != nil {
			t.Fatal(err)
		}
		var p struct {
			PRs []struct {
				Number int `json:"number"`
				PullIn []struct {
					PR    int      `json:"pr"`
					Paths []string `json:"paths"`
				} `json:"pull_in"`
			} `json:"prs"`
		}
		if err := json.Unmarshal(data, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.PRs) != 3 || len(p.PRs[1].PullIn) != 1 || p.PRs[1].PullIn[0].PR != 10 ||
			strings.Join(p.PRs[1].PullIn[0].Paths, " ") != "shared.txt" {
			t.Errorf("plan = %s, want #11 pulling in #10 over shared.txt", data)
		}
	})

	t.Run("after the fix pass the ordered simulation is clean and the PR links its own diff", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "shared.txt", "one\n")
		w.addPR(11, false, "shared.txt", "two\n")
		if got := w.runLast(land.Input{}); got != "needs-work #11" {
			t.Fatalf("first Run() = %q, want needs-work #11", got)
		}
		w.subagentMerges(11, 10, "shared.txt", "one\ntwo\n")

		got := w.runLast(land.Input{Resolved: []string{"#11 resolved append /r/11.md"}})
		if got != "ready 2" {
			t.Fatalf("second Run() = %q, want ready 2", got)
		}
		want := "compare/issue-10/pr...issue-11/pr"
		if len(w.comments) != 1 || !strings.HasPrefix(w.comments[0], "#11 ") || !strings.Contains(w.comments[0], want) {
			t.Errorf("comments = %q, want one on #11 containing %q", w.comments, want)
		}
	})

	t.Run("a failed resolver blocks the PRs that would pull it in", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "shared.txt", "one\n")
		w.addPR(11, false, "shared.txt", "two\n")
		w.addPR(12, false, "shared.txt", "three\n")

		out, err := w.run(land.Input{Resolved: []string{"#11 failed rewrite /r/11.md"}})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "pull-in=#10,#11 blocked-by #11") {
			t.Errorf("Run() = %q, want #12 blocked-by #11", out)
		}
		if !strings.HasSuffix(out, "\ngate-failed #11 resolver-failed") {
			t.Errorf("Run() = %q, want terminal gate-failed #11 resolver-failed", out)
		}
	})

	t.Run("a failed resolver for an unknown PR still stops", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		got := w.runLast(land.Input{Resolved: []string{"#21 failed rewrite /r/my report.md"}})
		if want := "gate-failed #21 resolver-failed"; got != want {
			t.Errorf("Run() = %q, want %q", got, want)
		}
	})

	t.Run("red CI gets one subagent and stays gate-failed if still red", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, true, "a.txt", "a\n")
		w.checks[10] = []gh.Check{{Name: "build", Bucket: "pass"}, {Name: "lint", Bucket: "fail"}}

		out, _ := w.run(land.Input{})
		if want := "#10 issue-10/pr ci=red conflicts=none pull-in=none needs-work"; !strings.Contains(out, want) {
			t.Errorf("Run() = %q, want line %q", out, want)
		}
		resolved := land.Input{Resolved: []string{"#10 resolved append /r/10.md"}}
		if got := w.runLast(resolved); got != "gate-failed #10 ci-red" {
			t.Errorf("Run() = %q, want gate-failed #10 ci-red", got)
		}
		w.checks[10] = []gh.Check{{Name: "lint", Bucket: "pass"}}
		if got := w.runLast(resolved); got != "ready 1" {
			t.Errorf("Run() = %q, want ready 1", got)
		}
	})

	t.Run("waits for pending checks before judging", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		w.pending[10] = 2
		got := w.runLast(land.Input{CIPoll: time.Millisecond, CITimeout: 5 * time.Second})
		if got != "ready 1" {
			t.Errorf("Run() = %q, want ready 1", got)
		}
		if w.checkCalls != 3 {
			t.Errorf("Checks called %d times, want 3", w.checkCalls)
		}
	})

	t.Run("a red suite names the culprit by bisecting the chain", func(t *testing.T) {
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
				got := w.runLast(land.Input{Suite: []string{"! grep -q bad f*.txt"}})
				if want := fmt.Sprintf("suite-red #%d", culprit); got != want {
					t.Errorf("Run() = %q, want %q", got, want)
				}
			})
		}
	})

	t.Run("a generated-file-only clash is regenerated with no subagent", func(t *testing.T) {
		w := newWave(t)
		w.withGolden()
		w.addGoldenPR(10, "a")
		w.addGoldenPR(11, "b")
		regen := map[string]string{"golden/**": regenAll}

		got := w.runLast(land.Input{Regen: regen, PrePush: []string{"echo gate >> " + w.log}})
		if got != "ready 2" {
			t.Fatalf("Run() = %q, want ready 2", got)
		}
		if diff := cmp.Diff([]string{"gate"}, w.logged()); diff != "" {
			t.Errorf("pre_push runs (-want +got):\n%s", diff)
		}
		if len(w.comments) != 1 || !strings.HasPrefix(w.comments[0], "#11 ") || !strings.Contains(w.comments[0], "+b") {
			t.Errorf("comments = %q, want one on #11 showing the regenerated diff", w.comments)
		}
		git(t, w.work, "fetch", "-q", "origin")
		merges := gitOut(t, w.work, "log", "--merges", "--format=%s", "origin/issue-11/pr")
		if !strings.Contains(merges, "origin/issue-10/pr") {
			t.Errorf("issue-11/pr merge commits = %q, want a merge from origin/issue-10/pr", merges)
		}
	})

	t.Run("a failing pre_push gate stops before pushing", func(t *testing.T) {
		w := newWave(t)
		w.withGolden()
		w.addGoldenPR(10, "a")
		w.addGoldenPR(11, "b")
		before := gitOut(t, w.work, "rev-parse", "origin/issue-11/pr")

		got := w.runLast(land.Input{Regen: map[string]string{"golden/**": regenAll}, PrePush: []string{"exit 1"}})
		if want := "gate-failed #11 exit 1"; got != want {
			t.Errorf("Run() = %q, want %q", got, want)
		}
		git(t, w.work, "fetch", "-q", "origin")
		if after := gitOut(t, w.work, "rev-parse", "origin/issue-11/pr"); after != before {
			t.Errorf("issue-11/pr moved from %s to %s", before, after)
		}
	})

	t.Run("stale generated files on the final commit are reported", func(t *testing.T) {
		w := newWave(t)
		w.withGolden()
		w.addPR(10, false, "src/a.txt", "a\n")
		got := w.runLast(land.Input{Regen: map[string]string{"golden/**": regenAll}})
		if want := "regen-stale golden/out.txt"; got != want {
			t.Errorf("Run() = %q, want %q", got, want)
		}
	})

	t.Run("rejects a malformed resolved line", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		if _, err := w.run(land.Input{Resolved: []string{"#10 maybe append r.md"}}); err == nil {
			t.Fatal("Run() = nil, want error for malformed --resolved line")
		}
	})

	t.Run("errors when the feature branch is missing", func(t *testing.T) {
		w := newWave(t)
		w.addPR(10, false, "a.txt", "a\n")
		git(t, w.work, "push", "-q", "origin", ":feat/demo")
		_, err := w.run(land.Input{})
		if err == nil || !strings.Contains(err.Error(), "feat/demo") {
			t.Errorf("Run() error = %v, want one naming feat/demo", err)
		}
	})

	t.Run("errors when open tickets have no PR on the feature branch", func(t *testing.T) {
		w := newWave(t)
		w.issues = []gh.Issue{{Number: 5}}
		w.all = []gh.PR{{Number: 50, BaseRefName: "feature/demo", Body: "Closes #5\n"}}
		_, err := w.run(land.Input{})
		if err == nil || !strings.Contains(err.Error(), "#50 (base feature/demo)") {
			t.Errorf("Run() error = %v, want one naming #50 aimed at feature/demo", err)
		}
	})
}

func testCheck(t *testing.T) {
	tests := []struct {
		name  string
		setup func(w *wave)
		want  string
	}{
		{"clean and green", func(w *wave) { w.addPR(10, false, "a.txt", "a\n") }, "reconcile"},
		{"conflict now", func(w *wave) {
			w.addPR(10, false, "base.txt", "ten\n")
			git(w.t, w.work, "checkout", "-q", "main")
			write(w.t, w.work, "base.txt", "tip\n")
			git(w.t, w.work, "commit", "-qam", "tip moves")
			git(w.t, w.work, "push", "-q", "origin", "main:feat/demo")
		}, "land"},
		{"conflict after an earlier PR merges", func(w *wave) {
			w.addPR(10, false, "shared.txt", "one\n")
			w.addPR(11, false, "shared.txt", "two\n")
		}, "land"},
		{"red CI", func(w *wave) {
			w.addPR(10, false, "a.txt", "a\n")
			w.checks[10] = []gh.Check{{Name: "ci", Bucket: "fail"}}
		}, "land"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWave(t)
			tt.setup(w)
			before := w.tipSHA()
			if got := w.runLast(land.Input{Check: true}); got != tt.want {
				t.Errorf("Run(--check) = %q, want %q", got, tt.want)
			}
			if got := w.tipSHA(); got != before || len(w.comments) != 0 {
				t.Errorf("--check moved feat/demo or commented: %q", w.comments)
			}
			if _, err := os.Stat(w.plan); err == nil {
				t.Error("--check wrote a plan, want none")
			}
		})
	}
}
