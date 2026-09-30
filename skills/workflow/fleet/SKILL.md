---
name: fleet
description: >
  Run a batch of tickets through parallel worktrees, wave by wave. Two subcommands:
  `/fleet dispatch` fans /implement out across ready tickets, one subagent per ticket;
  `/fleet reconcile` runs after a wave has merged, closes the done tickets, promotes the
  newly unblocked ones from backlog to ready, and writes the handoff for the next dispatch.
  Use for "fan out these tickets", "implement these issues in parallel", "for each ready
  ticket create a worktree and delegate to a subagent", "reconcile the wave", "what's
  unblocked now", "close out the merged tickets", "set up the next wave". For a single
  ticket, use /implement directly.
---

Orchestration layer above /implement, split by when it runs:

- `dispatch`: start of a wave. Read [dispatch.md](dispatch.md).
- `reconcile`: after the wave's PRs have merged. Read [reconcile.md](reconcile.md).

Read the file for the subcommand in the first argument. With no argument, use `reconcile` if
the user says the wave has merged, otherwise `dispatch`. If it is still unclear, ask.

## Shared conventions

Both halves rely on these, so they are defined once here.

- **Scratchpad.** `<scratchpad>/fleet/` holds `brief.md`, one `<N>.md` report per ticket, and
  `state.md`. It is scratch, not truth: the board and PR state on GitHub win over it.
- **Status labels.** ticket-tracker's `status:backlog → ready → in-progress → in-review → done`.
  Move a ticket with `gh issue edit <N> --remove-label "status:<from>" --add-label "status:<to>"`.
  Moving to `done` also needs `gh issue close <N>`. A ticket with no `status:*` label is backlog.
- **Blockers.** A ticket's `## Blocked by` section lists `- Blocked by #<N>` lines, or "None".
  This is the only record of dependencies; both subcommands read it.
- **`state.md`.** Reconcile writes it, dispatch reads it: one section per unblocked ticket with
  its number, base branch, and the merged PR to use as the worked example.
