package commands

import (
	"context"
	"errors"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/land"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
)

func landCommand() *cli.Command {
	return &cli.Command{
		Name:      "land",
		Usage:     "prepare the wave's PRs to merge in order without conflicts, never merging; prints needs-work, ready, gate-failed, regen-stale or suite-red (--check prints land or reconcile)",
		ArgsUsage: "<label>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
			&cli.StringSliceFlag{Name: "resolved", Usage: "a resolver's return line, repeatable"},
			&cli.BoolFlag{Name: "check", Usage: "only read CI and simulate merges, print land or reconcile; no pushes"},
			&cli.StringFlag{Name: "plan", Usage: "where to write the plan JSON shared with subagents (default .claude/fleet-land-plan.json)"},
			&cli.DurationFlag{Name: "ci-poll", Value: 30 * time.Second, Usage: "interval between CI status reads"},
			&cli.DurationFlag{Name: "ci-timeout", Value: 30 * time.Minute, Usage: "how long to wait for pending CI"},
			&cli.StringFlag{Name: "state", Usage: "path to state.md holding dispatch's Started timestamp"},
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
	label := cmd.Args().Get(0)
	return land.Run(ctx, commandDeps(cmd), land.Input{
		Dir:       dir,
		Label:     label,
		Check:     cmd.Bool("check"),
		Plan:      cmd.String("plan"),
		Regen:     cfg.Regen,
		PrePush:   cfg.Checks.PrePush,
		Suite:     cfg.Checks.Suite,
		Resolved:  cmd.StringSlice("resolved"),
		Started:   metrics.StartFromState(cmd.String("state"), label),
		CIPoll:    cmd.Duration("ci-poll"),
		CITimeout: cmd.Duration("ci-timeout"),
	})
}
