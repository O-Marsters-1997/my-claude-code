package commands

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/board"
)

func reconcileCommand() *cli.Command {
	return &cli.Command{
		Name:      "reconcile",
		Usage:     "close merged tickets, promote unblocked ones, write state.md; prints finished|unfinished then the merged branches for tp remove --merged",
		ArgsUsage: "<label>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "state", Required: true, Usage: "path to write state.md"},
		},
		Action: runReconcile,
	}
}

func runReconcile(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return errors.New("usage: fleet reconcile <label> --state <path>")
	}
	return board.Run(ctx, commandDeps(cmd), board.Input{
		Label:     cmd.Args().Get(0),
		StatePath: cmd.String("state"),
	})
}
