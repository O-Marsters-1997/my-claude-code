# Dispatch

Per-ticket cost controls (scoped test commands, one batched code-simplifier pass,
/code-review medium, splitting a large ticket) live in /implement. Don't restate them in
dispatch prompts; each subagent runs /implement and gets them from there.

## 1. Select and triage the batch

The batch is every ready ticket for the feature label:

```bash
gh issue list --state open --label "<label>" --label "status:ready" --json number,title,body
```

If `<scratchpad>/fleet/state.md` exists and its label matches, reconcile already chose each
ticket's base and worked example. Take those as given and skip to step 2, keeping only the
inline check below. A `state.md` for a different label is stale; ignore it.

Otherwise, for each ticket, decide: inline, bundle or dispatch. A ticket is **tiny** when it
carries `size:xs`, or is a single-file change with no schema or API change and no new tests
beyond one case.

- **Inline.** The batch's only tiny ticket. Do it in this session. A `tp` worktree, a subagent
  spawn and a skill reload cost more than the ticket.
- **Bundle.** Two or more tiny tickets go to one subagent in one worktree, worked in sequence
  with one commit each, on the bundle branch (see Shared conventions in SKILL.md). One spawn
  and one review instead of one per ticket.
- **Everything else.** Dispatch every ready ticket in parallel off `feat/<label>`. Real
  dependencies live in `## Blocked by`, so a ready ticket is never waiting on another. Tickets
  that touch the same files are expected to conflict; step 6 resolves that after the work is done.

## 1b. Ensure the feature branch

Every PR in this wave targets `feat/<label>`. Create it from `main` if it doesn't exist yet:

```bash
git ls-remote --exit-code --heads origin "feat/<label>" >/dev/null \
  || { git branch "feat/<label>" origin/main && git push -u origin "feat/<label>"; }
```

Use `rtk proxy "git push -u origin feat/<label>"` where rtk is in play.

## 2. One shared exploration pass

Before the first wave, if two or more tickets touch the same ground or follow the same
pattern, explore it once:

- Run `codegraph_explore` once over the symbols and files named in the batch's tickets (the
  "Where to look" lists) when the repo has a `.codegraph/` index, and fold the result into the
  brief. Subagents start from it instead of re-reading the same source.
- Read the completed worked example, the governing ADR or design doc, and CONTEXT.md.
- Write a short brief (files, pattern, gotchas, test command) to
  `<scratchpad>/fleet/brief.md`.
- Pass the brief's path into every dispatch prompt. Don't let each subagent rediscover it.

When research is needed, spawn the `Explore` agent with `model: "haiku"`. Never let a
lookup default to general-purpose; it costs as much as the implementation.

## 3. Create the worktrees

All worktree work goes through `tp` (see the treepad skill), never `git worktree`. `tp new`
also syncs the local configs a bare worktree lacks.

**Name.** The ticket branch, `issue-<N>/<short-title>`, or the bundle branch for a bundle (see
Shared conventions in SKILL.md).

**Create.** One per dispatched ticket, capturing the path (`tp new` cannot cd for you). Fetch first
so the base is never stale and the subagent has no reason to reset its branch:

```bash
git fetch origin
WT=$(TREEPAD_CD_FD=3 tp new "issue-<N>/<short-title>" --base "origin/feat/<label>" 3>&1 1>&2)
```

Use `tp exec <branch> -- <cmd>`
or `tp status --json` to reach an existing worktree, not `cd` or `git -C` on a guessed path.

## 4. Write the dispatch prompt

Every prompt has the same shape, so the cached prefix stays identical across the batch.
Fixed text first, ticket-specific text last.

```
Run /implement for issue #<N> in worktree <path>, branch issue-<N>/<short-title>.

Brief: <scratchpad>/fleet/brief.md
Worked example: <merged PR of the previous wave, if any>
Files to touch: <exact paths, from the ticket, ADR table or CONTEXT.md>
Docker: COMPOSE_PROJECT_NAME=fleet-<N>
Base: feat/<label>
Size: <xs|s|unsized, from the ticket's size:* label>

Done means: the review passes /implement requires for this size run, findings applied, committed
on the branch, draft PR open with `gh pr create --draft --base feat/<label>` (never `main`), result written.
Skipping a required review pass is not allowed; return `blocked` instead.
Read with Read and Grep on absolute paths. Don't chain `cd … && cat; grep; …` across a sibling
worktree: the auto-mode classifier has denied such chains as destructive. If a command is denied,
return `blocked` with the denied command verbatim.
Write your full report to <scratchpad>/fleet/<N>.md.
Return exactly one line and nothing else:
<N> done|blocked <PR URL or blocker reason> reviewed=yes|no <scratchpad>/fleet/<N>.md
Example: 142 done https://github.com/o/r/pull/151 reviewed=yes /tmp/…/fleet/142.md
```

- For a bundle, the first line names every issue in order (`Run /implement for issues #142,
  #145, #150 in worktree …`), `<N>` elsewhere is the lowest issue number, and `Size: xs`. The
  return line lists every number comma-separated (`142,145,150 done <PR URL> …`), and the
  ticket labels move together.
- Name exact files when the mapping is written down anywhere. "Explore the area" is for
  the case where it isn't.
- The one-line return is the contract. Read a ticket's report file only when its status
  is `blocked` or its PR needs a look.
- Treat `done` with `reviewed=no` as not done: don't move the ticket to `in-review`; send the
  subagent back to run the missing passes.
- Move the ticket `ready → in-progress` when you spawn its subagent, and `→ in-review` when
  it returns `done`. Reconcile relies on these labels to find the wave.

## 5. Docker-backed tests

Parallel testcontainers fight over ports and networks, and each clash burns a full test
run on a flaky failure.

- Give each worktree its own `COMPOSE_PROJECT_NAME` (as in the prompt above), so
  networks and volumes don't collide.
- Cap concurrent Docker-backed test runs at 2 across the batch. If the tickets need more,
  split the wave.

## 6. Integrate

Once every subagent has returned, make the wave merge into `feat/<label>` with no conflict in a
fixed order. The tickets were built in parallel, so conflicts are expected and are fixed here,
before review, so the next wave starts from merged work.

Simulate with `git merge-tree`, which needs no worktree and touches no branch:

```bash
git fetch origin
TIP=$(git rev-parse "origin/feat/<label>")
# for each PR branch B, in order:
TREE=$(git merge-tree --write-tree "$TIP" "origin/$B") \
  && TIP=$(git commit-tree "$TREE" -p "$TIP" -p "origin/$B" -m "sim $B")
```

It exits non-zero on a conflict and prints the conflicting paths.

1. **Order.** Greedily pick the next PR that merges cleanly onto the current `TIP`, lowest issue
   number on a tie. A PR that conflicts whatever comes before it goes last.
2. **Fix.** For each PR `B` that conflicts, find the earlier PR `A` whose files it collides
   with. In B's worktree, rebase B onto A's branch (`tp exec <B> -- git rebase origin/<A>`) and
   give any conflict to a resolver subagent, one per PR. It gets both diffs and both tickets' intent,
   resolves, runs the scoped tests, and continues the rebase. Then
   `git push --force-with-lease` and `gh pr edit <B-PR> --base <A>`, so B's review diff shows only
   its own ticket. When A merges, GitHub retargets B to `feat/<label>`, and B merges cleanly
   because it already carries A's change.
3. **Verify.** Re-run the simulation over the final order. Every step must be clean. If a PR
   still conflicts, repeat step 2 once. If it still conflicts, stop and name that PR and its
   conflicting paths for the user to resolve by prompt; leave the rest as they are.

Comment on each rebased PR with the files the resolver touched, so review checks them first.
If review changes a PR that others are stacked on, restack them with
`git rebase --onto origin/feat/<label> <old-parent> <child>` before merging.

## 7. After the batch

Spot-check two subagent transcripts: compare `cache_read_input_tokens` with
`cache_creation_input_tokens` in their `usage` fields. A falling read-to-creation ratio
across the batch means the prompts' shared prefix drifted (reordered fields,
per-ticket text too early). Fix the template before the next batch.

Report to the user:

- the merge order as a numbered list, one PR per line (`1. #151 issue-142/stuck-scrape-run`),
  marking each PR the resolver rebased and why;
- the blocked tickets with their reason, and any PR left for a manual conflict prompt.

Tell them to merge in that order, and once the wave has merged, to run
`/fleet reconcile <label>` with the label filled in.
