package hook

import (
	"cmp"
	"encoding/json"
	"io"
	"slices"
	"time"

	"github.com/O-Marsters-1997/my-claude-code/tools/reflect/internal/record"
)

type payload struct {
	HookEventName string `json:"hook_event_name"`
	SessionID     string `json:"session_id"`
	Cwd           string `json:"cwd"`
}

func Run(in io.Reader, projectDir, recordDir string, now time.Time) error {
	var p payload
	if err := json.NewDecoder(in).Decode(&p); err != nil {
		return err
	}
	if p.HookEventName != "SessionStart" || p.SessionID == "" {
		return nil
	}
	existing, err := record.Read(recordDir)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(existing, func(r record.Record) bool { return r.SessionID == p.SessionID }) {
		return nil
	}
	dir := cmp.Or(projectDir, p.Cwd)
	files := instrFiles(dir)
	return record.Append(recordDir, record.Record{
		SessionID: p.SessionID,
		Cwd:       dir,
		TS:        now.UTC(),
		InstrHash: combinedHash(files),
		Files:     files,
		Commit:    git(dir, "rev-parse", "--short", "HEAD"),
		Branch:    git(dir, "branch", "--show-current"),
		Dirty:     git(dir, "status", "--porcelain") != "",
	})
}
