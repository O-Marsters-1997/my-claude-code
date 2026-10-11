package commands

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/land"
)

func landCommand() *cli.Command {
	return &cli.Command{
		Name:      "land",
		Usage:     "squash-merge the wave's ready PRs one at a time; prints done, drafts, needs-resolver or suite-red",
		ArgsUsage: "<label>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
		},
		Action: runLand,
	}
}

func runLand(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return errors.New("usage: fleet land <label>")
	}
	dir := cmd.String("repo")
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	return land.Run(ctx, commandDeps(cmd), land.Input{
		Dir:        dir,
		Label:      cmd.Args().Get(0),
		RegenGlobs: cfg.RegenGlobs(),
		Suite:      cfg.Checks.Suite,
	})
}
