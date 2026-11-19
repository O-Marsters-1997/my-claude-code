package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/detect"
)

const (
	historyRuns    = 50
	historyInspect = 10
	commentWidth   = 120
)

func ciHistory(ctx context.Context, sh deps.Shell, dir string) detect.CIHistory {
	out, err := sh.Run(ctx, dir, fmt.Sprintf("gh run list --limit %d --json databaseId,conclusion", historyRuns))
	if err != nil {
		return detect.CIHistory{}
	}
	var runs []struct {
		DatabaseID int64  `json:"databaseId"`
		Conclusion string `json:"conclusion"`
	}
	if err := json.Unmarshal(out, &runs); err != nil {
		return detect.CIHistory{}
	}
	hist := detect.CIHistory{Runs: len(runs)}
	for _, r := range runs {
		if r.Conclusion != "failure" {
			continue
		}
		hist.FailedRuns++
		if hist.FailedRuns > historyInspect {
			continue
		}
		hist.FailedJobs = append(hist.FailedJobs, failedJobs(ctx, sh, dir, r.DatabaseID)...)
	}
	return hist
}

func failedJobs(ctx context.Context, sh deps.Shell, dir string, runID int64) []string {
	out, err := sh.Run(ctx, dir, fmt.Sprintf("gh run view %d --json jobs", runID))
	if err != nil {
		return nil
	}
	var view struct {
		Jobs []struct {
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
		} `json:"jobs"`
	}
	if json.Unmarshal(out, &view) != nil {
		return nil
	}
	var names []string
	for _, j := range view.Jobs {
		if j.Conclusion == "failure" {
			names = append(names, j.Name)
		}
	}
	return names
}

func renderAdvised(a detect.Advice) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("# Written by fleet init --advise. Review before committing.\n\n")
	if err := writeKey(&buf, a.CINote, "ci", a.Config.CI); err != nil {
		return nil, err
	}
	buf.WriteString("\n[checks]\n")
	if err := writeKey(&buf, a.PrePushNote, "pre_push", a.Config.Checks.PrePush); err != nil {
		return nil, err
	}
	if len(a.Config.Checks.Suite) > 0 {
		if err := writeKey(&buf, a.SuiteNote, "suite", a.Config.Checks.Suite); err != nil {
			return nil, err
		}
	}
	if len(a.Config.Regen) > 0 {
		buf.WriteString("\n[regen]\n")
		for _, g := range slices.Sorted(maps.Keys(a.Config.Regen)) {
			if err := writeKey(&buf, a.RegenNotes[g], g, a.Config.Regen[g]); err != nil {
				return nil, err
			}
		}
	}
	return buf.Bytes(), nil
}

func writeKey(buf *bytes.Buffer, note, key string, value any) error {
	for _, line := range wrap(note, commentWidth) {
		buf.WriteString("# " + line + "\n")
	}
	return toml.NewEncoder(buf).Encode(map[string]any{key: value})
}

func wrap(s string, width int) []string {
	var lines []string
	var cur string
	for _, w := range strings.Fields(s) {
		if cur != "" && len(cur)+1+len(w) > width {
			lines = append(lines, cur)
			cur = ""
		}
		if cur != "" {
			cur += " "
		}
		cur += w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
