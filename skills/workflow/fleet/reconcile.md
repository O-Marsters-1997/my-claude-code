# Reconcile

Run after a wave's PRs have merged. It brings the board in line with GitHub, finds what is now
unblocked, and hands off. It never dispatches: the gap between waves is the user's review gate.

GitHub is the source of truth, so derive everything from issue and PR state, not from
`state.md`. That keeps it correct when PRs were merged by hand or the last dispatch was
abandoned partway.

## 1. Close what is done

List every open issue carrying the feature label and `status:in-progress` or `status:in-review`,
and find each one's PR:

```bash
gh issue list --state open --label "<label>" --label "status:in-review" --json number,title
gh pr list --state all --search "<N> in:body" --json number,state,mergedAt,headRefName,url
```

Match a PR to its issue by `Closes #<N>` or the branch name. Then, per ticket:

- **PR merged.** Move to `done` and close the issue (see Shared conventions in SKILL.md).
- **PR open.** Leave it.
- **PR closed unmerged.** Leave the ticket's label as it is, list it in the report, and don't
  promote its dependents. Whether to retry or rethink is the user's call.
- **No PR found.** Leave it and list it; the subagent may have blocked or the PR was never opened.

## 2. Promote the unblocked

For every open backlog ticket with the feature label (no `status:*` label, or `status:backlog`), read its
`## Blocked by` section. If every listed blocker is now closed, and none is a closed-unmerged
ticket from step 1, move it `backlog → ready`. A ticket with "None" that was already ready is
left alone. Promote per ticket: one stuck PR holds back only its own dependents.

## 3. Clean up

Remove the worktree and local branch of every ticket whose PR merged with
`tp remove <branch>` (branches are named `<type>-<N>/<short-title>`; find them with
`tp status --json`). Never use `git worktree remove` or `git branch -d`. Skip any worktree that
is dirty or whose PR is not merged. Do it after step 1 so nothing is removed on a guess.

## 4. Write the handoff

Write `<scratchpad>/fleet/state.md`. Open it with `Label: <label>`, then a section per ticket
promoted in step 2 and any other `ready` ticket with that label, each with:

- the issue number and title,
- `Base: main` (its blockers are merged, so nothing stacks across waves),
- the merged PR of its most relevant blocker, as the worked example.

Then report to the user in this order: tickets closed, tickets promoted, tickets needing a
decision (closed-unmerged, no PR), worktrees removed. End by telling them to `/clear` and run
`/fleet dispatch <label>`, with the label filled in; you cannot clear your own session.
