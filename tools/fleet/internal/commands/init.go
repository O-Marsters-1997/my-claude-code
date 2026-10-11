package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v3"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/detect"
)

const fleetTOML = ".claude/fleet.toml"

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "detect checks, suite and regen paths and write .claude/fleet.toml",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "repo", Value: ".", Usage: "repository to run in"},
			&cli.BoolFlag{Name: "dry-run", Usage: "print the TOML instead of writing it"},
			&cli.BoolFlag{Name: "force", Usage: "overwrite an existing " + fleetTOML},
			&cli.BoolFlag{Name: "measure", Usage: "run each candidate check and drop failing or slow ones"},
		},
		Action: runInit,
	}
}

func runInit(ctx context.Context, cmd *cli.Command) error {
	d := commandDeps(cmd)
	dryRun, force := cmd.Bool("dry-run"), cmd.Bool("force")

	top, err := d.Git.Run(ctx, cmd.String("repo"), nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	root := strings.TrimSpace(string(top))
	out := filepath.Join(root, fleetTOML)
	if _, err := os.Stat(out); err == nil && !force && !dryRun {
		return fmt.Errorf("%s exists; pass --force to overwrite or --dry-run to preview", fleetTOML)
	}

	listing, err := d.Git.Run(ctx, root, nil, "ls-files", "-z")
	if err != nil {
		return err
	}
	var tracked []string
	for _, f := range strings.Split(string(listing), "\x00") {
		if f != "" {
			tracked = append(tracked, f)
		}
	}

	var opts detect.Options
	if cmd.Bool("measure") && !dryRun {
		opts.Measure = func(command string) (time.Duration, error) { return measure(ctx, root, command) }
	}
	res := detect.Scan(os.DirFS(root), tracked, opts)

	for _, n := range res.Notes {
		fmt.Fprintln(d.Log, n)
	}
	if len(res.Unmapped) > 0 {
		fmt.Fprintf(d.Log, "Generated paths with no regen command: %s\n", strings.Join(res.Unmapped, ", "))
	}

	body, err := renderTOML(res.Config)
	if err != nil {
		return err
	}
	return emit(d, out, body, dryRun)
}

func emit(d deps.Deps, out string, body []byte, dryRun bool) error {
	if dryRun {
		_, err := d.Out.Write(body)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, body, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(d.Out, "wrote %s\n", fleetTOML)
	return nil
}

func renderTOML(cfg config.Config) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("# Written by fleet init. Review before committing.\n")
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func measure(ctx context.Context, dir, command string) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	start := time.Now()
	err := cmd.Run()
	return time.Since(start), err
}
