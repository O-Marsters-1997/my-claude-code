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

A `contract-change` issue closes when its stub PR merges into the feature branch, like any
labelled ticket. Each ticket that listed it under `## Blocked by` then drops that line; if no
other open blocker remains, it moves from `status:backlog` back to `status:ready`. A ticket
with another open blocker stays in backlog.

## 2. Clean up

Run `tp remove <branch> --merged` for each branch on that line, and for no others. Skip any
worktree that is dirty. `tp remove` is the only teardown reconcile runs. Whatever a worktree
started beyond its checkout (services, caches, data) is the repo's to clean up in its treepad
remove hook, so fleet never names a technology.

## 3. Raise the feature PR

Skip this step unless the line starts with `finished`.

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

## 4. Report

Report to the user in this order: tickets closed, tickets promoted, tickets needing a decision
(`closed-unmerged`, `no-pr`), worktrees removed, and the feature PR's URL if step 3 raised or
found one.

If the line was `finished`, stop there. Otherwise end by telling them to `/clear` and run
`/fleet dispatch <label>`, with the label filled in; you cannot clear your own session.
