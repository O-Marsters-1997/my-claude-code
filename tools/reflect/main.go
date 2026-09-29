package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: reflect on|off|status|show|metrics|proposals|hook")
		os.Exit(2)
	}
	if os.Args[1] == "hook" {
		hookMain()
		return
	}
	if err := run(os.Stdout, os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "reflect:", err)
		os.Exit(1)
	}
}

func run(w io.Writer, cmd string, args []string) error {
	switch cmd {
	case "on":
		return cmdOn(w)
	case "off":
		return cmdOff(w)
	case "status":
		return cmdStatus(w)
	case "show":
		return cmdShow(w, args)
	case "metrics":
		return cmdMetrics(w)
	case "proposals":
		return cmdProposals(w, args)
	}
	return fmt.Errorf("unknown command %q", cmd)
}
