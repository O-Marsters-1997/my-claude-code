package ledger

import (
	"errors"
	"fmt"
	"slices"
)

var statuses = []string{"pending", "promoted", "deferred", "rejected"}

// ValidStatus reports whether s is a ledger status.
func ValidStatus(s string) bool { return slices.Contains(statuses, s) }

// MarkFields are the optional ledger fields a mark writes when non-empty.
type MarkFields struct {
	Issue, Proposal, Scope, Skill string
}

// Mark appends a status line for the learning id, writing the non-empty
// fields. It fails for an unknown id or status, and for "promoted" when neither
// f nor the ledger already has an issue.
func Mark(path, id, status string, f MarkFields) error {
	if !ValidStatus(status) {
		return fmt.Errorf("status %q must be one of %v", status, statuses)
	}
	if f.Scope != "" && f.Scope != "global" && f.Scope != "repo" {
		return fmt.Errorf("scope %q must be global or repo", f.Scope)
	}
	all, err := Read(path)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(all, func(l Learning) bool { return l.ID == id })
	if i < 0 {
		return fmt.Errorf("no learning with id %q", id)
	}
	if status == "promoted" && f.Issue == "" && all[i].Issue == "" {
		return errors.New("promoted needs an issue URL")
	}
	return Append(path, Learning{ID: id, Status: status, Issue: f.Issue, Proposal: f.Proposal, Scope: f.Scope, Skill: f.Skill})
}
