package main

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/legacy"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/metrics"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/repo"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/scan"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/status"
)

const usage = "usage: reflect scan|slice|metrics|status|reports|uninstall-legacy|hook"

var library string

type env struct {
	claudeDir  string
	projectDir string
}

func (e env) projects() string  { return filepath.Join(e.claudeDir, "projects") }
func (e env) recordDir() string { return filepath.Join(e.claudeDir, "reflect") }
func (e env) root() string      { return repo.MainCheckout(e.projectDir) }

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	home, _ := os.UserHomeDir()
	wd, _ := os.Getwd()
	e := env{
		claudeDir:  cmp.Or(os.Getenv("CLAUDE_CONFIG_DIR"), filepath.Join(home, ".claude")),
		projectDir: cmp.Or(os.Getenv("CLAUDE_PROJECT_DIR"), wd),
	}
	cmd, args := os.Args[1], os.Args[2:]
	if cmd == "hook" {
		_ = hook.Run(os.Stdin, os.Getenv("CLAUDE_PROJECT_DIR"), e.recordDir(), time.Now())
		return
	}
	out, err := run(e, cmd, args)
	fmt.Print(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reflect:", err)
		os.Exit(1)
	}
}

func run(e env, cmd string, args []string) (string, error) {
	switch cmd {
	case "scan":
		return runScan(e, args)
	case "slice":
		return runSlice(e, args)
	case "metrics":
		return runMetrics(e, args)
	case "status":
		return status.Report(e.claudeDir, library)
	case "reports":
		dir, err := reportsDir(e.root())
		return dir + "\n", err
	case "uninstall-legacy":
		return legacy.Uninstall(e.root())
	}
	return "", fmt.Errorf("unknown command %q; %s", cmd, usage)
}

func runScan(e env, args []string) (string, error) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	limit := fs.Int("cap", 8, "most agents to review")
	if err := fs.Parse(reorder(args)); err != nil || fs.NArg() != 1 {
		return "", errors.New("usage: reflect scan <session_id> [--cap n]")
	}
	sid := fs.Arg(0)
	s, err := session.Load(e.projects(), sid)
	if err != nil {
		return "", err
	}
	return scan.Write(s, filepath.Join(os.TempDir(), "reflect", sid), *limit)
}

func runSlice(e env, args []string) (string, error) {
	fs := flag.NewFlagSet("slice", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	context := fs.Int("C", 20, "lines either side")
	const sliceUsage = "usage: reflect slice <session_id> <agent|main> <line> [-C n]"
	if err := fs.Parse(reorder(args)); err != nil || fs.NArg() != 3 {
		return "", errors.New(sliceUsage)
	}
	line, err := strconv.Atoi(fs.Arg(2))
	if err != nil {
		return "", errors.New(sliceUsage)
	}
	s, err := session.Load(e.projects(), fs.Arg(0))
	var live *session.LiveError
	if err != nil && !errors.As(err, &live) {
		return "", err
	}
	return scan.Slice(s, fs.Arg(1), line, *context)
}

func runMetrics(e env, args []string) (string, error) {
	fs := flag.NewFlagSet("metrics", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	exclude := fs.String("exclude", "", "session id to leave out, usually the current one")
	if err := fs.Parse(args); err != nil {
		return "", errors.New("usage: reflect metrics [--exclude session_id]")
	}
	records, err := record.Read(e.recordDir())
	if err != nil {
		return "", err
	}
	return metrics.Report(metrics.Options{Projects: e.projects(), Records: records, Root: e.root(), Exclude: *exclude}), nil
}

func reportsDir(root string) (string, error) {
	dir := filepath.Join(root, ".claude", "reflect", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ignore := filepath.Join(root, ".claude", "reflect", ".gitignore")
	if _, err := os.Stat(ignore); errors.Is(err, os.ErrNotExist) {
		return dir, os.WriteFile(ignore, []byte("*\n"), 0o644)
	}
	return dir, nil
}

func reorder(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 1 && a[0] == '-' {
			flags = append(flags, a)
			if i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		positional = append(positional, a)
	}
	return append(flags, positional...)
}
