package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/hook"
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
	s := logstore.New(cmp.Or(os.Getenv("CLAUDE_PROJECT_DIR"), wd))
	switch cmd {
	case "on":
		return "reflect on: logging to " + s.LogPath() + "\n", s.On()
	case "off":
		return "reflect off (log kept)\n", s.Off()
	case "status":
		return report.Status(s, library)
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
