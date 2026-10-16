package land

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/config"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/gh"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/queue"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/trial"
)

type pullIn struct {
	PR     int      `json:"pr"`
	Branch string   `json:"branch"`
	Paths  []string `json:"paths"`
}

type node struct {
	Number        int      `json:"number"`
	Branch        string   `json:"branch"`
	URL           string   `json:"url"`
	Draft         bool     `json:"draft"`
	CI            string   `json:"ci"`
	FailingChecks []string `json:"failing_checks"`
	Conflicts     []string `json:"conflicts"`
	PullIn        []pullIn `json:"pull_in"`
	NeedsWork     bool     `json:"needs_work"`
	Status        string   `json:"status"`

	regenFrom []regenFix
}

type regenFix struct {
	from  string
	paths []string
}

type plan struct {
	Label string `json:"label"`
	Tip   string `json:"tip"`
	PRs   []node `json:"prs"`
}

func PlanPath(dir string) string {
	return filepath.Join(dir, ".claude", "fleet-land-plan.json")
}

func writePlan(path string, p plan) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func readPlan(path, label string) plan {
	data, err := os.ReadFile(path)
	if err != nil {
		return plan{}
	}
	var p plan
	if json.Unmarshal(data, &p) != nil || p.Label != label {
		return plan{}
	}
	return p
}

func requireWave(ctx context.Context, d deps.Deps, in Input, tip string, prs []gh.PR, issues []gh.Issue) error {
	if _, err := d.Git.Run(ctx, in.Dir, nil, "rev-parse", "--verify", "--quiet", "origin/"+tip); err != nil {
		return fmt.Errorf("feature branch %s is not on origin: push it, or dispatch the label first", tip)
	}
	if len(prs) > 0 {
		return nil
	}
	open, err := d.GH.OpenPRsAll(ctx)
	if err != nil {
		return err
	}
	wanted := map[int]bool{}
	for _, i := range issues {
		wanted[i.Number] = true
	}
	var elsewhere []string
	for _, e := range queue.Build(open, nil) {
		if slices.ContainsFunc(e.Tickets, func(n int) bool { return wanted[n] }) {
			elsewhere = append(elsewhere, fmt.Sprintf("#%d (base %s)", e.PR.Number, e.PR.BaseRefName))
		}
	}
	msg := fmt.Sprintf("no open PR targets %s (label %s has %d open tickets)", tip, in.Label, len(issues))
	if len(elsewhere) > 0 {
		msg += "; PRs aimed at another base: " + strings.Join(elsewhere, ", ")
	}
	return errors.New(msg)
}

func analyse(ctx context.Context, d deps.Deps, in Input, tip string, entries []queue.Entry) ([]node, error) {
	globs := config.Config{Regen: in.Regen}.RegenGlobs()
	nodes := make([]node, 0, len(entries))
	needsWork := map[int]bool{}
	for _, e := range entries {
		n := node{Number: e.PR.Number, Branch: e.PR.HeadRefName, URL: e.PR.URL, Draft: e.PR.IsDraft}
		var err error
		if n.CI, n.FailingChecks, err = waitCI(ctx, d, in, e.PR.Number); err != nil {
			return nil, err
		}
		classify := func(base, label string) (kind string, paths []string, err error) {
			v, err := trial.Classify(ctx, d, trial.Input{Dir: in.Dir, Tip: base, Branch: "origin/" + e.PR.HeadRefName, RegenGlobs: globs})
			if err != nil {
				return "", nil, fmt.Errorf("trial #%d against %s: %w", e.PR.Number, label, err)
			}
			kind, rest, _ := strings.Cut(v, " ")
			return kind, strings.Fields(rest), nil
		}
		tipKind, tipPaths, err := classify("origin/"+tip, tip)
		if err != nil {
			return nil, err
		}
		source := tipKind == "conflict"
		if tipKind != "clean" {
			n.Conflicts = tipPaths
		}
		var pairs []pullIn
		pairKinds := map[int]string{}
		for _, prev := range nodes {
			kind, paths, err := classify("origin/"+prev.Branch, fmt.Sprintf("#%d", prev.Number))
			if err != nil {
				return nil, err
			}
			if kind == "clean" {
				continue
			}
			pairs = append(pairs, pullIn{PR: prev.Number, Branch: prev.Branch, Paths: paths})
			pairKinds[prev.Number] = kind
			source = source || kind == "conflict" || needsWork[prev.Number]
		}
		if source {
			n.PullIn = pairs
		} else {
			if tipKind == "regen" {
				n.regenFrom = append(n.regenFrom, regenFix{from: tip, paths: tipPaths})
			}
			for _, p := range pairs {
				n.regenFrom = append(n.regenFrom, regenFix{from: p.Branch, paths: p.Paths})
			}
		}
		n.NeedsWork = source || n.CI == "red"
		needsWork[n.Number] = n.NeedsWork
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func waitCI(ctx context.Context, d deps.Deps, in Input, number int) (string, []string, error) {
	deadline := time.Now().Add(in.CITimeout)
	for {
		checks, err := d.GH.Checks(ctx, number)
		if err != nil {
			return "", nil, err
		}
		var failing []string
		pending := false
		for _, c := range checks {
			switch c.Bucket {
			case "fail", "cancel":
				failing = append(failing, c.Name)
			case "pending":
				pending = true
			}
		}
		switch {
		case len(checks) == 0:
			return "none", nil, nil
		case len(failing) > 0:
			return "red", failing, nil
		case !pending:
			return "green", nil, nil
		case !time.Now().Before(deadline):
			return "pending", nil, nil
		}
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(in.CIPoll):
		}
	}
}

func recommend(nodes []node) string {
	for _, n := range nodes {
		if n.CI == "red" || len(n.Conflicts) > 0 || len(n.PullIn) > 0 || len(n.regenFrom) > 0 {
			return "land"
		}
	}
	return "reconcile"
}
