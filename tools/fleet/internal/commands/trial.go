package commands

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

func commandDeps(cmd *cli.Command) deps.Deps {
	root := cmd.Root()
	return deps.DefaultDeps(root.Writer, root.ErrWriter)
}

func trialCommand() *cli.Command {
	return &cli.Command{
		Name:      "trial",
		Usage:     "classify a branch against a tip: clean, regen <path>... or conflict <path>...",
		ArgsUsage: "<tip-ref> <branch-ref>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
		},
		Action: runTrial,
	}
}

func runTrial(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return errors.New("usage: fleet trial <tip-ref> <branch-ref>")
	}
	dir := cmd.String("repo")
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	return trial.Run(ctx, commandDeps(cmd), trial.Input{
		Dir:       dir,
		Tip:       cmd.Args().Get(0),
		Branch:    cmd.Args().Get(1),
		RegenGlob: cfg.RegenGlobs(),
	})
}
