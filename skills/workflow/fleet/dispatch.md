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
- **Everything else.** Dispatch every ready ticket in parallel off `feat/<slug>`. Shared files
  never chain. Real dependencies live in `## Blocked by`, so a ready ticket is never waiting on
  another: one that needs another ready ticket's code is held and reported as a ticket bug.

Then run `fleet hotspots "<label>"` and keep its output for the report. Each line is a path
named by three or more ready tickets and their numbers (`none` when there are no hotspots). It
changes no scheduling: every ticket still dispatches in parallel.

## 1b. Ensure the feature branch

Every PR in this wave targets `feat/<slug>`. Create it from `main` if it doesn't exist yet:

```bash
git ls-remote --exit-code --heads origin "feat/<slug>" >/dev/null \
  || { git branch "feat/<slug>" origin/main && git push -u origin "feat/<slug>"; }
```

Use `rtk proxy "git push -u origin feat/<slug>"` where rtk is in play.

## 1c. Record the start time

`fleet land` measures the wave's wall clock from here. Add one line under `Label:` in
`<scratchpad>/fleet/state.md`, creating the file with just the `Label:` line if it doesn't exist,
and replacing any earlier `Started:` line:

```
Started: <date -u +%Y-%m-%dT%H:%M:%SZ>
```

Without it, `fleet land --state <scratchpad>/fleet/state.md` falls back to the earliest wave PR's
creation time.

## 2. One shared exploration pass

Before the first wave, if two or more tickets touch the same ground or follow the same
pattern, explore it once:

- Run `codegraph_explore` once over the symbols and files named in the batch's tickets (the
  "Where to look" lists) when the repo has a `.codegraph/` index, and fold the result into the
  brief. Subagents start from it instead of re-reading the same source.
- Read the completed worked example, the governing ADR or design doc, and CONTEXT.md.
- Write a short brief (files, pattern, gotchas, test command) to
  `<scratchpad>/fleet/brief.md`.
- Write the worked example's `git show --stat <merge sha>` output and its route registration
  into the brief once. Prompts only point at the brief.
- Resolve every doc, prototype and example path to an absolute path when writing the brief.
- Pass the brief's path into every dispatch prompt. Don't let each subagent rediscover it.

When research is needed, spawn the `Explore` agent with `model: "haiku"`. Never let a
lookup default to general-purpose; it costs as much as the implementation.

## 3. Create the worktrees

All worktree work goes through `tp` (see the treepad skill), never `git worktree`. `tp new`
also syncs the local configs a bare worktree lacks.

**Name.** The ticket branch, `issue-<N>/<short-title>`, or the bundle branch for a bundle (see
Shared conventions in SKILL.md).

**Create.** One per dispatched ticket, capturing the path (`tp new` cannot cd for you). Fetch first
so the base is never stale and the subagent has no reason to reset its branch. Create every
worktree up front from `origin/feat/<slug>`:

```bash
git fetch origin
BRANCH="issue-<N>/<short-title>"
tp new "$BRANCH" --base "origin/feat/<slug>"
WT=$(tp status --json | jq -r --arg b "$BRANCH" '.[] | select(.branch == $b) | .path')
[ -n "$WT" ] || { echo "no worktree path for $BRANCH" >&2; exit 1; }
```

An empty `$WT` stops dispatch: spawn no subagent for that ticket.

Use `tp exec <branch> -- <cmd>`
or `tp status --json` to reach an existing worktree, not `cd` or `git -C` on a guessed path.

**Scratch dir.** Create one per dispatched ticket (`mkdir -p <scratchpad>/fleet/<N>`). The
prompt passes its absolute path.

## 4. Write the dispatch prompt

Every prompt has the same shape, so the cached prefix stays identical across the batch.
Fixed text first, ticket-specific text last.

```
Run /implement for issue #<N> in worktree <path>, branch issue-<N>/<short-title>.

Brief: <scratchpad>/fleet/brief.md (use only the absolute paths in it and here)
Worked example: <merged PR of the previous wave, if any>
Files to touch: <exact absolute paths, from the ticket, ADR table or CONTEXT.md>
Base: feat/<slug>
Regen: <the repo's `Fleet` section in CLAUDE.md: generated paths and command, or none>
Size: <xs|s|unsized, from the ticket's size:* label>
Scratch dir: <scratchpad>/fleet/<N>/ (write every flag, log and temp file here, never /tmp)

Done means: the review passes /implement requires for this size run, findings applied, committed
on the branch, draft PR open with `gh pr create --draft --base <Base>` (never `main`), result written.
On a conflict in a Regen path, take either side, run the Regen command and commit; never hand-merge it.
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

## 5. After the batch

Spot-check two subagent transcripts: compare `cache_read_input_tokens` with
`cache_creation_input_tokens` in their `usage` fields. A falling read-to-creation ratio
across the batch means the prompts' shared prefix drifted (reordered fields,
per-ticket text too early). Fix the template before the next batch.

Report to the user:

- the PRs as a numbered list in issue order, one PR per line
  (`1. #151 issue-142/stuck-scrape-run`);
- the blocked tickets with their reason, and any held ticket as a ticket bug;
- each hotspot from `fleet hotspots`, with its ticket numbers and a suggested refactor ticket
  that would split the shared file (omit the section when the output was `none`).

Once every PR is open, run `fleet land "<label>" --check`. It reads CI and simulates merging the
PRs in order, drafts included, without spawning anything or pushing, and prints one word. End the
report by telling the user to review each PR, then:

- `land`: a PR conflicts now or after an earlier PR merges, or CI is red. Recommend
  `/fleet land <label>` before they merge anything, so every PR can merge in order without
  conflicts. Land never merges; the user squash-merges each PR into `feat/<slug>` themselves.
- `reconcile`: nothing needs preparing. Recommend merging the PRs in order, then
  `/fleet reconcile <label>`.

A non-zero exit is an error (for example `feat/<slug>` missing, or tickets whose PRs target
another base): show its stderr and stop.

## 6. CI follow-up

Read the top-level `ci` key from `.claude/fleet.toml` (`pre_push` lives under `[checks]`; `fleet init --dry-run` previews a detected file). With no file or `ci = "off"`, do nothing: never read, watch
or poll CI.

With `ci = "async"`, after a ticket returns `done` with `reviewed=yes`, watch its PR's checks in the background
(`gh pr checks <PR> --watch` with `run_in_background`) without blocking the wave. Green costs
nothing. On the first red run, resume that ticket's subagent once with SendMessage, passing the
`gh run view <run-id> --log-failed` excerpt, and ask for one fix pushed. Don't wait for the
re-run, but keep the background watch on that PR. If the subagent can't be resumed, or a later run on that PR is red again, report the PR and the failing check to the user
and stop.
