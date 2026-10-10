---
name: fleet
description: >
  Run a batch of tickets through parallel worktrees, wave by wave. Two subcommands:
  `/fleet dispatch <label>` fans /implement out across the ready tickets carrying that feature
  label, one subagent per ticket, stacking tickets that touch the same files so they merge
  cleanly in a fixed order;
  `/fleet reconcile <label>` runs after a wave has merged, closes the done tickets, promotes the
  newly unblocked ones from backlog to ready, writes the handoff for the next dispatch, and opens
  the feature's PR into main once no tickets are left. A /sweep's `sweep-YYYY-MM-DD` label
  works as a feature label.
  Use for "fan out these tickets", "implement these issues in parallel", "for each ready
  ticket create a worktree and delegate to a subagent", "reconcile the wave", "what's
  unblocked now", "close out the merged tickets", "set up the next wave". For a single
  ticket, use /implement directly.
---

Orchestration layer above /implement, split by when it runs:

- `dispatch <label>`: start of a wave. Read [dispatch.md](dispatch.md).
- `reconcile <label>`: after the wave's PRs have merged. Read [reconcile.md](reconcile.md).

Both subcommands require a **feature label** as the argument after the subcommand (for example
`/fleet dispatch cv-tailoring`). If it is missing, ask for it; never run across all issues.
Read the file for the subcommand in the first argument. With no subcommand, use `reconcile` if
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
- **Feature label.** The label `to-plan` creates and `to-tickets` applies to every ticket of a
  feature. It scopes every `gh issue list` in both subcommands with `--label "<label>"`.
- **Feature branch.** `feat/<label>`, the label as the feature name. Every ticket PR targets it,
  never `main`, so several features can run in parallel without touching each other. When the
  feature's tickets are all done, reconcile opens one PR merging `feat/<label>` into `main`.
- **Chain.** Tickets whose declared files overlap. Dispatch runs a chain in issue-number order,
  each ticket branched from its parent's branch, each PR targeting its parent. GitHub retargets
  the child to `feat/<label>` when its parent merges.
- **Fleet section.** A repo's CLAUDE.md may carry a `Fleet` section listing its generated,
  committed paths (minified CSS, goldens) and the command that rebuilds them. Those files conflict
  by construction, so fleet regenerates them mechanically instead of resolving them.
- **Ticket branch.** `issue-<N>/<short-title>`: `<N>` is the issue number, `<short-title>` a
  kebab-case slug of the issue title, a few words (`issue-142/stuck-scrape-run`). Dispatch
  creates it, /implement works on it, reconcile removes it.
- **Bundle branch.** `issue-<N>/tiny-bundle`, `<N>` the lowest issue number in the bundle. One
  PR whose body has a `Closes #<N>` line per bundled ticket, which is how reconcile matches it
  to each of them.
- **`state.md`.** Reconcile writes it, dispatch reads it. It opens with the feature label, then
  has one section per unblocked ticket with its number, base branch, and the merged PR to use as
  the worked example.
