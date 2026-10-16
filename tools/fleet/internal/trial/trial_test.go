package trial_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	r := &repo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "t@example.com")
	r.git("config", "user.name", "t")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r *repo) git(args ...string) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func (r *repo) commit(file, content string) {
	r.t.Helper()
	p := filepath.Join(r.dir, file)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "change "+file)
}

func (r *repo) run(globs ...string) string {
	r.t.Helper()
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	err := trial.Run(context.Background(), d, trial.Input{Dir: r.dir, Tip: "tip", Branch: "branch", RegenGlobs: globs})
	if err != nil {
		r.t.Fatal(err)
	}
	return out.String()
}

func base(t *testing.T) *repo {
	r := newRepo(t)
	r.commit("src.txt", "one\n")
	r.commit("dist/out.css", "a\n")
	return r
}

func TestRun(t *testing.T) {
	regen := []string{"dist/**"}
	tests := []struct {
		name  string
		setup func(r *repo)
		want  string
	}{
		{
			name: "clean",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "tip")
				r.commit("other.txt", "x\n")
				r.git("switch", "-q", "-c", "branch", "main")
				r.commit("src.txt", "two\n")
			},
			want: "clean\n",
		},
		{
			name: "regen only",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "tip")
				r.commit("dist/out.css", "tip\n")
				r.git("switch", "-q", "-c", "branch", "main")
				r.commit("dist/out.css", "branch\n")
			},
			want: "regen dist/out.css\n",
		},
		{
			name: "mixed",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "tip")
				r.commit("dist/out.css", "tip\n")
				r.commit("src.txt", "tip\n")
				r.git("switch", "-q", "-c", "branch", "main")
				r.commit("dist/out.css", "branch\n")
				r.commit("src.txt", "branch\n")
			},
			want: "conflict dist/out.css src.txt\n",
		},
		{
			name: "real conflict",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "tip")
				r.commit("src.txt", "tip\n")
				r.git("switch", "-q", "-c", "branch", "main")
				r.commit("src.txt", "branch\n")
			},
			want: "conflict src.txt\n",
		},
		{
			name: "sibling squash-merged into tip",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "a", "main")
				r.commit("a.txt", "a\n")
				r.git("switch", "-q", "-c", "branch", "main")
				r.commit("c.txt", "c\n")
				r.git("switch", "-q", "-c", "tip", "main")
				r.git("merge", "-q", "--squash", "a")
				r.git("commit", "-q", "-m", "squash a")
			},
			want: "clean\n",
		},
		{
			name: "child stacked on squash-merged parent",
			setup: func(r *repo) {
				r.git("switch", "-q", "-c", "a", "main")
				r.commit("a.txt", "a\n")
				r.git("switch", "-q", "-c", "branch", "a")
				r.commit("c.txt", "c\n")
				r.git("switch", "-q", "-c", "tip", "main")
				r.git("merge", "-q", "--squash", "a")
				r.git("commit", "-q", "-m", "squash a")
			},
			want: "clean\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := base(t)
			tt.setup(r)
			if got := r.run(regen...); got != tt.want {
				t.Errorf("Run() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	t.Run("clean merges chain", func(t *testing.T) {
		r := base(t)
		r.git("switch", "-q", "-c", "a", "main")
		r.commit("a.txt", "a\n")
		r.git("switch", "-q", "-c", "b", "main")
		r.commit("b.txt", "b\n")
		d := deps.DefaultDeps(&bytes.Buffer{}, &bytes.Buffer{})
		ctx := context.Background()

		sha, paths, err := trial.Merge(ctx, d, r.dir, "main", "a")
		if err != nil || len(paths) != 0 || sha == "" {
			t.Fatalf("Merge(main, a) = %q, %v, %v; want a commit", sha, paths, err)
		}
		sha, paths, err = trial.Merge(ctx, d, r.dir, sha, "b")
		if err != nil || len(paths) != 0 {
			t.Fatalf("Merge(<a>, b) = %q, %v, %v; want clean", sha, paths, err)
		}
		out, err := exec.Command("git", "-C", r.dir, "ls-tree", "--name-only", sha).Output()
		if err != nil {
			t.Fatal(err)
		}
		if got := string(out); !strings.Contains(got, "a.txt") || !strings.Contains(got, "b.txt") {
			t.Errorf("chained tree = %q, want a.txt and b.txt", got)
		}
	})

	t.Run("conflict returns paths and no commit", func(t *testing.T) {
		r := base(t)
		r.git("switch", "-q", "-c", "a", "main")
		r.commit("src.txt", "a\n")
		r.git("switch", "-q", "-c", "b", "main")
		r.commit("src.txt", "b\n")
		d := deps.DefaultDeps(&bytes.Buffer{}, &bytes.Buffer{})

		sha, paths, err := trial.Merge(context.Background(), d, r.dir, "a", "b")
		if err != nil || sha != "" || strings.Join(paths, " ") != "src.txt" {
			t.Errorf("Merge(a, b) = %q, %v, %v; want no commit and src.txt", sha, paths, err)
		}
	})
}

func TestRunBadRef(t *testing.T) {
	r := base(t)
	var out bytes.Buffer
	d := deps.DefaultDeps(&out, &bytes.Buffer{})
	err := trial.Run(context.Background(), d, trial.Input{Dir: r.dir, Tip: "nope", Branch: "main"})
	if err == nil {
		t.Error("Run() error = nil, want error")
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty on error", out.String())
	}
}
