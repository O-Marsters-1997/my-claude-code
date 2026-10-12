# Reconcile

Run after a wave's PRs have merged. It brings the board in line with GitHub, finds what is now
unblocked, and hands off. It never dispatches: the gap between waves is the user's review gate.
Everything mechanical lives in the `fleet` CLI; this file only calls it and branches on its output.

## 1. Run the board

```bash
fleet reconcile "<label>" --state "<scratchpad>/fleet/state.md"
```

It closes merged tickets, promotes the unblocked ones, and writes `state.md`. Progress goes to
stderr, one line per ticket: `closed #N`, `promoted #N`, `closed-unmerged #N` and `no-pr #N`
(the last two need the user's decision; their dependents stay held). Read only the one line on
stdout: `finished` or `unfinished`, followed by the merged branches.

A non-zero exit is an error, not an outcome: show its stderr to the user and stop.

A `contract-change` issue is a ticket like any other: its stub PR carries `Closes #N`, and when
that merges into the feature branch the command above closes it. Tickets whose only open blocker
it was show up as `promoted #N`; a ticket with another open blocker stays in backlog.

## 2. Clean up

Run `tp remove <branch> --merged` for each branch on that line, and for no others. Skip any
worktree that is dirty. `tp remove` is the only teardown reconcile runs. Whatever a worktree
started beyond its checkout (services, caches, data) is the repo's to clean up in its treepad
remove hook, so fleet never names a technology.

## 3. Check completed slices

Run this before step 4 so a failing journey is filed before the feature PR goes up.

Read the bodies with `gh issue list --label "<label>" --state all --json number,body,state`. A slice is
complete when every ticket carrying the same `Slice: <n>` line is closed and merged (a
`closed-unmerged` or `no-pr` ticket keeps it open), and at least one of them merged this run (its
branch is on the stdout line). That keeps a later reconcile from rerunning a slice. A ticket without
a `Slice:` line belongs to no slice.

For each such slice, take its `Journeys:` numbers (union across its tickets) and read those journeys
from `docs/prds/<feature>.md`. Check out the feature branch `feat/<slug>` and run only those journeys
with the `run` skill. Save each output as evidence in
`<scratchpad>/fleet/journeys/slice-<n>-journey-<j>.md`.

Apply the Journey axis from `code-review` to each journey and its evidence. For each failure:

```bash
gh label create journey-fail --force
gh issue create --label "<label>" --label journey-fail --title "Journey <j> fails: <journey title>" --body "<body>"
```

The body names the slice and journey, quotes the `→` end state, and pastes the evidence. List the
journeys marked needs human eyes in the report; file no ticket for them.

## 4. Raise the feature PR

Skip this step unless the line starts with `finished` and no `journey-fail` ticket under the label
is open. A fix ticket's merge makes the next reconcile see `finished` again and raise the PR.

```bash
gh pr list --base main --head "feat/<slug>" --state open --json url     # reuse if present
```

If no PR from `feat/<slug>` into `main` is open, raise one. Nothing in it is new to
review, so the body only lists every ticket with the label, one `#<N> <title>` per line, from
`gh issue list --state closed --label "<label>"`:

```bash
gh pr create --base main --head "feat/<slug>" --title "<title>" --body "<body>"
```

It is not a draft: every ticket in it already passed review on its own PR.

## 5. Report

Report to the user in this order: tickets closed, tickets promoted, tickets needing a decision
(`closed-unmerged`, `no-pr`), slices checked with each journey's result, `journey-fail` tickets
filed, journeys needing human eyes, worktrees removed, and the feature PR's URL if step 4 raised or
found one.

If the line was `finished`, stop there. Otherwise end by telling them to `/clear` and run
`/fleet dispatch <label>`, with the label filled in; you cannot clear your own session.
