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

A `contract-change` issue has no PR: the user closes it once the plan and the affected tickets'
`## Signatures` are updated. Tickets whose only open blocker
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

It is not a draft: every ticket in it already passed review on its own PR. Tell the user to merge
it with rebase-and-merge, so `main` reads as one commit per ticket (each ticket PR was squashed
into `feat/<slug>`).

## 5. Record metrics

Runs only when step 4 raised or found the feature PR; otherwise skip it. Counts come from GitHub, so
a re-run, or a feature finished before this step existed, gets its row too. `<label>` and `<slug>`
are the ones used in steps 1 and 4. Each count is a number: add `--limit 500 --json number --jq length`
to the `gh issue list` calls.

| Column | Derivation |
|---|---|
| contract changes | `gh issue list --label "<label>" --label contract-change --state all` |
| backlog-held | `gh issue list --label "<label>" --state all` with a `Blocked by` section naming at least one issue (not `None`): each sat in `status:backlog` while an earlier wave ran |
| journey failures | `gh issue list --label "<label>" --label journey-fail --state all` |
| read plan | ask the user "Did you read the plan?" (yes/no) and record the answer; `unanswered` if they cannot be asked now; when the row already exists with a yes/no answer, keep it and do not ask again |

The metrics file is `docs/plans/metrics.md`, gitignored with the rest of `docs/plans/`. Create it
with this header if absent:

```
| feature | date | contract changes | backlog-held | journey failures | read plan |
|---|---|---|---|---|---|
```

The first column is the feature label and the key; `date` is today as `YYYY-MM-DD`. If a row with
that label exists, rewrite that line with the new values; otherwise append one. Never leave two rows
for a label.

## 6. Report

Report to the user in this order: tickets closed, tickets promoted, tickets needing a decision
(`closed-unmerged`, `no-pr`), slices checked with each journey's result, `journey-fail` tickets
filed, journeys needing human eyes, worktrees removed, and the feature PR's URL and the metrics row
written, if step 4 raised or found a PR.

If the line was `finished`, stop there. Otherwise end by telling them to `/clear` and run
`/fleet dispatch <label>`, with the label filled in; you cannot clear your own session.
