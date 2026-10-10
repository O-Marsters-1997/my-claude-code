package ledger

import (
	"errors"
	"fmt"
	"slices"
)

var statuses = []string{"pending", "promoted", "deferred", "rejected"}

// ValidStatus reports whether s is a ledger status.
func ValidStatus(s string) bool { return slices.Contains(statuses, s) }

// Mark appends a status line for the learning id. Issue, scope and skill are
// written only when non-empty. It fails for an unknown id or status, and for
// "promoted" without an issue.
func Mark(path, id, status, issue, scope, skill string) error {
	if !ValidStatus(status) {
		return fmt.Errorf("status %q must be one of %v", status, statuses)
	}
	if status == "promoted" && issue == "" {
		return errors.New("promoted needs an issue URL")
	}
	if scope != "" && scope != "global" && scope != "repo" {
		return fmt.Errorf("scope %q must be global or repo", scope)
	}
	all, err := Read(path)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(all, func(l Learning) bool { return l.ID == id }) {
		return fmt.Errorf("no learning with id %q", id)
	}
	return Append(path, Learning{ID: id, Status: status, Issue: issue, Scope: scope, Skill: skill})
}
