package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/commands"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRoot().Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "fleet:", err)
		stop()
		os.Exit(1)
	}
}

func newRoot() *cli.Command {
	return &cli.Command{
		Name:                  "fleet",
		Usage:                 "mechanics for /fleet: trial merges, landing and reconcile",
		Version:               fmt.Sprintf("%s (%s, %s)", Version, Commit, Date),
		EnableShellCompletion: true,
		Writer:                os.Stdout,
		ErrWriter:             os.Stderr,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Usage: "log progress to stderr"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			level := slog.LevelWarn
			if cmd.Bool("verbose") {
				level = slog.LevelDebug
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(cmd.Root().ErrWriter, &slog.HandlerOptions{Level: level})))
			return ctx, nil
		},
		Commands: commands.Router(),
	}
}
