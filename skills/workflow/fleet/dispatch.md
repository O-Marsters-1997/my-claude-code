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
inline and stack-depth checks below. A `state.md` for a different label is stale; ignore it.

Otherwise, for each ticket, decide: inline, dispatch now, or dispatch later.

- **Inline.** Single-file change, no schema or API change, no new tests beyond one case.
  Do it in this session. A `tp` worktree, a subagent spawn and a skill reload cost more than
  the ticket.
- **Dependency order.** If tickets form a chain (a staged migration, a series of
  refactors), stack them: ticket N branches off N-1's branch and its PR targets that
  branch, so N starts once N-1 is committed, not merged. Its prompt points at N-1's
  diff as the worked example. Independent tickets branch off `main` and share a wave.
- **Stack depth.** Cap a stack at 3. A rejected approach low in the stack wastes
  everything above it, so if the chain is longer, leave the rest for the next wave.

## 2. One shared exploration pass

Before the first wave, if two or more tickets touch the same ground or follow the same
pattern, explore it once:

- Read the completed worked example, the governing ADR or design doc, and CONTEXT.md.
- Write a short brief (files, pattern, gotchas, test command) to
  `<scratchpad>/fleet/brief.md`.
- Pass the brief's path into every dispatch prompt. Don't let each subagent rediscover it.

When research is needed, spawn the `Explore` agent with `model: "haiku"`. Never let a
lookup default to general-purpose; it costs as much as the implementation.

## 3. Create the worktrees

All worktree work goes through `tp` (see the treepad skill), never `git worktree`. `tp new`
also syncs the local configs a bare worktree lacks.

**Name.** `<type>-<N>/<short-title>`: `<N>` is the issue number, `<short-title>` a kebab-case
slug of the ticket title, a few words. `<type>` is `bug` for a bug ticket (labelled `bug`, or
filed by `triage-issue`), otherwise `feat`. For other kinds of ticket use the matching
conventional prefix (`chore-`, `docs-`, `refactor-`), and fall back to `feat`.

```
bug-142/stuck-scrape-run
feat-143/schedule-every
```

**Create.** One per dispatched ticket, capturing the path (`tp new` cannot cd for you):

```bash
WT=$(TREEPAD_CD_FD=3 tp new "<type>-<N>/<short-title>" --base "<Base>" 3>&1 1>&2)
```

`<Base>` is `main`, or the parent ticket's branch for a stack. Use `tp exec <branch> -- <cmd>`
or `tp status --json` to reach an existing worktree, not `cd` or `git -C` on a guessed path.

## 4. Write the dispatch prompt

Every prompt has the same shape, so the cached prefix stays identical across the batch.
Fixed text first, ticket-specific text last.

```
Run /implement for issue #<N> in worktree <path>, branch <type>-<N>/<short-title>.

Brief: <scratchpad>/fleet/brief.md
Worked example: <commit or PR of the previous wave, if any>
Files to touch: <exact paths, from the ticket, ADR table or CONTEXT.md>
Docker: COMPOSE_PROJECT_NAME=fleet-<N>
Base: <parent ticket's branch, or main>

Done means: committed on the branch, draft PR open with `gh pr create --draft --base <Base>`,
result written.
Write your full report to <scratchpad>/fleet/<N>.md.
Return exactly one line and nothing else:
<N> done|blocked <PR URL or blocker reason> <scratchpad>/fleet/<N>.md
Example: 142 done https://github.com/o/r/pull/151 /tmp/…/fleet/142.md
```

- Name exact files when the mapping is written down anywhere. "Explore the area" is for
  the case where it isn't.
- The one-line return is the contract. Read a ticket's report file only when its status
  is `blocked` or its PR needs a look.
- Move the ticket `ready → in-progress` when you spawn its subagent, and `→ in-review` when
  it returns `done`. Reconcile relies on these labels to find the wave.

## 5. Docker-backed tests

Parallel testcontainers fight over ports and networks, and each clash burns a full test
run on a flaky failure.

- Give each worktree its own `COMPOSE_PROJECT_NAME` (as in the prompt above), so
  networks and volumes don't collide.
- Cap concurrent Docker-backed test runs at 2 across the batch. If the tickets need more,
  split the wave.

## 6. Stacked chains within the wave

Only for chains dispatched together. Cross-wave sequencing is `reconcile`.

1. Link each chain's PRs into a GitHub stack, bottom to top: `gh stack link <b1> <b2> …`.
   It works from branch names alone, with no local stack state, so it doesn't care
   which worktree built each branch. If `gh stack` is missing, skip this; the `--base`
   chaining in the dispatch prompt already gives reviewers the ordered diffs. Install with
   `gh extension install github/gh-stack`.
2. If review changed a lower PR, restack before dispatching anything above it. Remove the
   finished worktrees first with `tp remove <branch>` (git won't rebase a branch checked out
   in another worktree),
   then from the main checkout: `gh stack init <b1> <b2> …` to adopt the branches, and
   `gh stack sync`. On a conflict, sync restores every branch; resolve it with
   `gh stack rebase` yourself, not in a subagent.

## 7. After the batch

Spot-check two subagent transcripts: compare `cache_read_input_tokens` with
`cache_creation_input_tokens` in their `usage` fields. A falling read-to-creation ratio
across the batch means the prompts' shared prefix drifted (reordered fields,
per-ticket text too early). Fix the template before the next batch.

Report to the user: one status line per ticket, and the blocked ones with their reason.
Once the PRs are reviewed and merged, tell the user to run `/fleet reconcile`.
