package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/harvest"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/ledger"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/learn/pull"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/legacy"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/metrics"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/replay"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/repo"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/scan"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/session"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/status"
)

const usage = "usage: reflect scan|slice|replay|metrics|status|learn|reports|uninstall-legacy|hook"

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
	case "replay":
		return runReplay(e, args)
	case "metrics":
		return runMetrics(e, args)
	case "learn":
		return runLearn(e, args)
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

func runReplay(e env, args []string) (string, error) {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	cwd := fs.String("cwd", "", "directory to run the hook from, default the recorded one")
	command := fs.String("command", "", "Bash command to feed the hook instead of the recorded one")
	const replayUsage = "usage: reflect replay <session_id> <agent|main> <line> [--cwd dir] [--command cmd]"
	if err := fs.Parse(reorder(args)); err != nil || fs.NArg() != 3 {
		return "", errors.New(replayUsage)
	}
	line, err := strconv.Atoi(fs.Arg(2))
	if err != nil {
		return "", errors.New(replayUsage)
	}
	s, err := session.Load(e.projects(), fs.Arg(0))
	var live *session.LiveError
	if err != nil && !errors.As(err, &live) {
		return "", err
	}
	home, _ := os.UserHomeDir()
	return replay.Run(s, fs.Arg(1), line, replay.Options{Home: home, Cwd: *cwd, Command: *command})
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

func runLearn(e env, args []string) (string, error) {
	const learnUsage = "usage: reflect learn harvest [--block] [--strip] | pull | ls [--json] [--status s] | mark <id> <status> [--issue url] [--proposal url] [--scope global|repo] [--skill name] | status [--remind]"
	if len(args) == 0 {
		return "", errors.New(learnUsage)
	}
	fs := flag.NewFlagSet("learn "+args[0], flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	switch args[0] {
	case "harvest":
		block := fs.Bool("block", false, "exit 1 while fix markers remain")
		strip := fs.Bool("strip", false, "delete every recorded marker, fix markers included")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			return "", errors.New(learnUsage)
		}
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		res, err := harvest.Run(context.Background(), harvest.Options{Repo: wd, Ledger: ledger.DefaultPath(), Library: library, Block: *block, Strip: *strip})
		if err != nil {
			return "", err
		}
		if res.Blocked {
			return res.Report(), fmt.Errorf("%d fix marker(s) remain; resolve or remove them before committing", len(res.Found))
		}
		return "", nil
	case "pull":
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			return "", errors.New(learnUsage)
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		res, err := pull.Run(context.Background(), pull.Options{
			Ledger: ledger.DefaultPath(),
			Roots:  []string{filepath.Join(home, "Documents", "coding")},
		})
		for _, w := range res.Warnings {
			fmt.Fprintln(os.Stderr, "reflect: warning:", w)
		}
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("recorded %d learning(s) from PR comments\n", res.Recorded), nil
	case "ls":
		asJSON := fs.Bool("json", false, "print the folded ledger as JSON")
		status := fs.String("status", "", "only learnings with this status")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			return "", errors.New(learnUsage)
		}
		learnings, err := ledger.Read(ledger.DefaultPath())
		if err != nil {
			return "", err
		}
		if *status != "" && !ledger.ValidStatus(*status) {
			return "", fmt.Errorf("unknown status %q", *status)
		}
		if *status != "" {
			learnings = slices.DeleteFunc(learnings, func(l ledger.Learning) bool { return l.Status != *status })
		}
		if *asJSON {
			if learnings == nil {
				learnings = []ledger.Learning{}
			}
			b, err := json.MarshalIndent(learnings, "", "  ")
			return string(b) + "\n", err
		}
		var sb strings.Builder
		for _, l := range learnings {
			fmt.Fprintf(&sb, "%s\t%s\t%s\t%s:%d\t%s\n", l.ID, l.Status, l.Kind, l.File, l.Line, l.Text)
		}
		return sb.String(), nil
	case "mark":
		issue := fs.String("issue", "", "issue URL, required for promoted")
		proposal := fs.String("proposal", "", "proposal issue URL grouping this learning")
		scope := fs.String("scope", "", "global or repo")
		skill := fs.String("skill", "", "owning skill")
		if err := fs.Parse(reorder(args[1:])); err != nil || fs.NArg() != 2 {
			return "", errors.New(learnUsage)
		}
		fields := ledger.MarkFields{Issue: *issue, Proposal: *proposal, Scope: *scope, Skill: *skill}
		return "", ledger.Mark(ledger.DefaultPath(), fs.Arg(0), fs.Arg(1), fields)
	case "status":
		remind := fs.Bool("remind", false, "print the triage reminder instead of the count, only at or above the threshold")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			return "", errors.New(learnUsage)
		}
		return learnStatus(ledger.DefaultPath(), *remind)
	}
	return "", errors.New(learnUsage)
}

const remindThreshold = 10

func learnStatus(path string, remind bool) (string, error) {
	n, err := pendingCount(path)
	if err != nil {
		return "", err
	}
	if !remind {
		return fmt.Sprintf("%d\n", n), nil
	}
	return reminder(n), nil
}

func pendingCount(path string) (int, error) {
	learnings, err := ledger.Read(path)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, l := range learnings {
		if l.Status == "pending" {
			n++
		}
	}
	return n, nil
}

func reminder(pending int) string {
	if pending < remindThreshold {
		return ""
	}
	return fmt.Sprintf("%d pending learnings: run /triage-learnings\n", pending)
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
