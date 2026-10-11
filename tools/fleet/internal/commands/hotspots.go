package commands

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/hotspots"
)

func hotspotsCommand() *cli.Command {
	return &cli.Command{
		Name:      "hotspots",
		Usage:     "print each non-regen path named by 3 or more ready tickets, with their numbers",
		ArgsUsage: "<label>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
		},
		Action: runHotspots,
	}
}

func runHotspots(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return errors.New("usage: fleet hotspots <label>")
	}
	cfg, err := config.Load(cmd.String("repo"))
	if err != nil {
		return err
	}
	return hotspots.Run(ctx, commandDeps(cmd), hotspots.Input{
		Label:      cmd.Args().Get(0),
		RegenGlobs: cfg.RegenGlobs(),
	})
}
