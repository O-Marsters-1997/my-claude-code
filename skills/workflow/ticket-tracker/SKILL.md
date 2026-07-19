---
name: ticket-tracker
description: >
  Live triage board over GitHub issues — reads the current tickets via the gh CLI and moves each
  between states (backlog → ready → in-progress → in-review → done) using status labels. This is
  the *ticket-level* board, complementing the *feature-level* ./roadmap.html rather than replacing
  it: a roadmap card is a feature that becomes several tickets here. Use when the user asks "where
  are my tickets", "show the board", "what's in progress", "move issue N to in-review", "what's
  ready to pick up", "update the ticket status", "mark N done", or otherwise wants to see or
  change where work sits. Trigger whenever the request is about the *current state* of GitHub
  issues rather than generating a new roadmap or filing a new issue.
---

# Ticket Tracker

The live state-machine layer over GitHub issues. Where `triage-issue`/`to-tickets` *create* tickets,
this skill is the ongoing surface that reads and moves them. State lives on GitHub, not in a file,
so it is never stale.

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-plan → to-tickets → ticket-tracker
                                                                 ^^^^^^^^^^^^^^ you are here
```

`./roadmap.html` and this board are **different altitudes, not rival copies**. A roadmap card is a
*feature* — something to be built — and one card fans out into several tickets, which is what this
board tracks. Do not tell the user the roadmap is superseded, and do not render a feature-level
roadmap from issues; a board full of tickets is not a roadmap.

## State machine

A ticket carries exactly one `status:*` label. States, in order:

`status:backlog` → `status:ready` → `status:in-progress` → `status:in-review` → `status:done`

- **backlog** — captured, not yet scoped to pick up. Open issues with no `status:*` label are treated as backlog.
- **ready** — scoped and grabbable now.
- **in-progress** — actively being worked.
- **in-review** — PR open / awaiting review.
- **done** — merged/closed. Closing the issue is the terminal action.

Forward moves are the norm; backward moves (e.g. review kicked back to in-progress) are allowed.
Reject a move that skips more than one state unless the user is explicit — a jump usually means a
missing label, not real progress.

## Setup (once per repo)

Before relying on the labels, check they exist:

```bash
gh label list --search status: | grep -q status: || {
  gh label create "status:backlog"     -c "#cccccc" -f
  gh label create "status:ready"        -c "#0e8a16" -f
  gh label create "status:in-progress"  -c "#1d76db" -f
  gh label create "status:in-review"    -c "#fbca04" -f
  gh label create "status:done"         -c "#5319e7" -f
}
```

## Show the board

Read live state and group by column. Never render from a cached file.

```bash
gh issue list --state open --limit 200 \
  --json number,title,labels,assignees \
  --jq 'sort_by(.number)'
```

Derive each ticket's state from its `status:*` label (none → backlog). Print one column per
state in order, each ticket as `#<number> <title> [@assignee]`. Put `done`/closed items last or
omit unless asked. Keep it to a scannable text board.

## Move a ticket

Swap the label so exactly one `status:*` remains, then confirm:

```bash
gh issue edit <number> \
  --remove-label "status:<from>" \
  --add-label "status:<to>"
```

- If the current state is unknown, read the issue's labels first (`gh issue view <number> --json labels`).
- Moving *to* `done`: also close the issue — `gh issue close <number>`.
- Moving *out of* `done`: reopen it — `gh issue reopen <number>`.
- After editing, re-read and state the ticket's new position: `#<n> is now <state>`.

## Getting tickets onto the board

This skill tracks tickets; it does not investigate or file them. To add new work:

- Bugs / reported problems → defer to **triage-issue** (investigate + file), then the new issue
  lands in **backlog**.
- Planned work → defer to **to-tickets**, then label the batch
  **ready**.

Do not duplicate those flows. When a ticket-tracker request implies creating an issue, hand off,
then apply the initial `status:*` label and show the board.
