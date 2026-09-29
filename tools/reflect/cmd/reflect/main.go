package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/install"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/logstore"
	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/report"
)

var library string

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: reflect on|off|status|show|metrics|proposals|hook")
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	if cmd == "hook" {
		hook.Run(os.Stdin, os.Getenv("CLAUDE_PROJECT_DIR"))
		return
	}
	out, err := run(cmd, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reflect:", err)
		os.Exit(1)
	}
	fmt.Print(out)
}

func run(cmd string, args []string) (string, error) {
	wd, _ := os.Getwd()
	projectDir := cmp.Or(os.Getenv("CLAUDE_PROJECT_DIR"), wd)
	root := logstore.CheckoutRoot(projectDir)
	s := logstore.New(projectDir)
	switch cmd {
	case "on":
		return enable(root, s)
	case "off":
		return "reflect off for " + root + " (log kept)\n", install.Disable(root)
	case "status":
		return report.Status(s, install.Enabled(root), library)
	case "show":
		if len(args) == 0 {
			return "", errors.New("usage: reflect show <session_id> [--all]")
		}
		return report.Show(s, args[0], len(args) > 1 && args[1] == "--all")
	case "metrics":
		return report.Metrics(s)
	case "proposals":
		if len(args) > 0 && args[0] == "library" {
			s = logstore.At(filepath.Join(library, ".claude", "reflect"))
		}
		dir, err := s.ProposalsDir()
		return dir + "\n", err
	}
	return "", fmt.Errorf("unknown command %q", cmd)
}

func enable(root string, s logstore.Store) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if err := install.Enable(root, exe); err != nil {
		return "", err
	}
	out := fmt.Sprintf("reflect on for %s\nhooks: %s\nlog: %s\nRestart the session for the hooks to take effect.\n",
		root, install.SettingsPath(root), s.LogPath())
	if !install.GitIgnored(root) {
		out += "warn: .claude/settings.local.json is not git-ignored here; it holds a machine-specific path\n"
	}
	return out, nil
}
