package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/metrics"
)

func metricsCommand() *cli.Command {
	return &cli.Command{
		Name:  "metrics",
		Usage: "print one row per wave from .claude/fleet-metrics.jsonl, a trend line and threshold flags",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
			&cli.StringFlag{Name: "label", Usage: "only waves for this feature label"},
			&cli.IntFlag{Name: "last", Usage: "only the last N waves"},
		},
		Action: runMetrics,
	}
}

func runMetrics(_ context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 0 {
		return errors.New("usage: fleet metrics [--label X] [--last N]")
	}
	all, err := metrics.Read(cmd.String("repo"))
	if err != nil {
		return err
	}
	rows := metrics.Select(all, cmd.String("label"), cmd.Int("last"))
	deps := commandDeps(cmd)
	for _, r := range rows {
		fmt.Fprintln(deps.Out, r)
		for _, f := range r.Flags() {
			fmt.Fprintf(deps.Out, "  flag: %s\n", f)
		}
	}
	fmt.Fprintln(deps.Out, metrics.Trend(rows))
	return nil
}
