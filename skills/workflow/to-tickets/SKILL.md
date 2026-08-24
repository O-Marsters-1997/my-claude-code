---
name: to-tickets
description: >
  Breaks a source — an implementation plan file (default), a PRD, or (only when neither exists)
  a raw conversation — into independently-grabbable tickets in Linear or GitHub, using
  tracer-bullet vertical slices. Cuts tickets for one feature per run: one roadmap card or one
  PRD becomes several tickets. The tracker is settled during the run. In a repo that runs
  command-centre, also emits `[[task]]` blocks into its config; in a plain treepad repo, emits a
  Batch Manifest instead so the tickets materialise as stacked worktrees. If the technical design
  isn't worked out yet, run /to-plan first.
disable-model-invocation: true
---

# Plan to Tickets

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./ideas/roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-plan → to-tickets → ticket-tracker
                                                    ^^^^^^^^^^ you are here
```

**Portfolio → feature is a fan-out, not a step.** A roadmap card is a *feature* — something to be
built — and this is the stage that turns exactly one of those features into its several tickets.
Never cut tickets for a whole roadmap or approach doc in one run.

**Nothing gates anything.** A missing plan is not a blocker — cut tickets from whatever exists, and
say so if the source is thin on technical design.

Break a source into independently-grabbable tickets using vertical slices (tracer bullets). The source is one of three, in priority order:

1. **A plan file** (DEFAULT) — a local `./plans/*.md` file produced by to-plan.
2. **A PRD** — a local `./docs/prd-<feature>.md` file, a Notion project doc, or a `[PRD]`-titled GitHub issue. Prefer the file when several exist; they hold the same content.
3. **A raw conversation** — a pasted or in-context discussion the user wants turned into tickets.

A plan is a technical design document: its architectural decisions (data models, schema shapes, API contracts, module boundaries) are durable constraints and its user stories are source material. A PRD or conversation carries product intent with less technical design — if a source is thin on technical design, note this to the user; it may need a to-plan pass first.

## Process

### 1. Locate the source

If the user named a source, use it. Otherwise detect what exists: run `artifact-scan` as a preflight if it's installed, else `ls ./plans/*.md ./docs/prd-*.md 2>/dev/null`. Apply the priority order above: prefer a plan file, then a PRD, then a conversation already in context.

Two cases where the priority order is the wrong answer:

- **A conversation in hand beats a stale artifact.** If the user is pasting a conversation *now* ("turn this into tickets"), that is the source — even if `./plans/*.md` exists. The priority order ranks source *quality*, not source *recency*; a plan file from another feature is not a better source than the thing the user just handed you.
- **More than one plan or PRD present.** These are per-feature artifacts, so several will accumulate. Confirm which feature you're cutting tickets for rather than defaulting to the newest.

Load the chosen source in full:

- **Plan file** — read the `./plans/*.md` file. Record its path and the **Source PRD** link from the plan header (if present).
- **PRD** — read `./docs/prd-<feature>.md`, fetch the Notion page, or fetch the issue with `gh issue view <number>` (with comments). Record whichever you used: the file path, the page URL, the issue number, or several.
- **Conversation** — use the in-context discussion, or ask the user to paste it.

Extract from whichever source: **technical design decisions** (data models, schema, API contracts, module boundaries, routes), **user stories**, and **acceptance criteria**. Note the source type — it selects the ticket-body source header in step 5.

Note the source's **provenance** too — a Linear project, a Notion page, a local file, a GitHub issue. It is what step 4 proposes a tracker from, and it costs nothing to record now.

### 2. Explore the codebase (optional)

If you have not already explored the codebase, do so to understand the current state of the code and how the plan's technical decisions map to existing structure.

### 3. Draft vertical slices

Break the plan into **tracer bullet** issues. Each issue is a thin vertical slice that cuts through ALL integration layers end-to-end, NOT a horizontal slice of one layer.

Re-derive slices freely. Do not map plan phases 1:1 to issues, and do not carry a plan's numbering
into a ticket title. A ticket called "Phase 3" or "Slice 2" is proof the slice was copied rather than
re-derived, so name it for what it delivers.

Slices may be 'HITL' or 'AFK'. HITL slices require human interaction, such as an architectural decision or a design review. AFK slices can be implemented and merged without human interaction. Prefer AFK over HITL where possible.

<vertical-slice-rules>
- Each slice delivers a narrow but COMPLETE path through every layer (schema, API, UI)
- A completed slice is demoable or verifiable on its own
- Prefer many thin slices over few thick ones
- Respect the plan's **deep modules** — substantial functionality behind a simple, stable, isolation-testable interface (vs shallow modules). A slice should build or extend a deep module behind its interface, not splay logic across shallow ones.
- A blocking edge is only real if you can **name the artifact** the blocked slice needs from its blocker: a type, table, route, file, or exported function. No nameable artifact means no edge. **Sequence in the source is not a dependency.** A plan lists phases in an order because prose has to be linear, not because each phase waits on the one before it. Given an ordered list and no criteria, the cheapest consistent answer is "each waits for the previous", and that answer is almost always wrong.
- Slice by feature, never by layer. "Extract helpers", then "core function", then "CLI wiring", then "docs" is a horizontal decomposition wearing four ticket titles, and it is serial by construction.
</vertical-slice-rules>

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each slice, show:

- **Title**: short descriptive name
- **Type**: HITL / AFK
- **Blocked by**: which other slices (if any) must complete first
- **Plan phases / user stories covered**: which phases and user stories from the plan this addresses

#### Measure the graph before proposing it

Print these three numbers alongside the breakdown, every run:

- **Day-one width** — how many slices have no blockers. This is how many agents can start at once.
- **Longest chain** — the deepest path through the blocking graph.
- **Acceptance criteria per slice** — a count each.

Target shape: day-one width of 3 or more, longest chain of 4 or less, roughly 3 acceptance criteria a
slice. A slice carrying 10 criteria is doing several things, and splitting it usually produces slices
that do not block each other.

**Stop and re-derive when day-one width is 1 and there are more than 3 slices.** Width 1 means the
fleet runs one agent and the parallelism is decorative. Never create a width-1 graph without the user
explicitly telling you to proceed.

Depth is a review-latency multiplier, so shallow and wide beats deep and narrow. It costs a staggered
start, since each layer waits for the layer below to *open* a pull request rather than to merge, and
then a serial merge tail, since a stack merges bottom-up. The last layer's agent can start early and
still land last.

Some convergence is genuine. Every phase of command-center really did touch `loop.go`, `server.go`
and `plan.go`, and forcing those apart would have produced tickets that conflict rather than tickets
that parallelise. So the gate asks you to justify each edge, not to invent independence: if the
slices truly stack, name the shared file or type that forces it and carry on.

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Which blocking edges are wrong? For each one kept, what artifact does it wait on?
- Should any slices be merged or split further?
- Are the correct slices marked as HITL and AFK?
- **Where do these get filed — Linear or GitHub?**

Settle the tracker here, in this approval — not from config, and not from a guess made silently at
write time. Propose an answer rather than asking cold, taking the first signal that applies:

1. **Where the repo sits.** Anything under `~/Documents/plain/` is Plain work — propose **Linear**.
   This outranks provenance because it still holds when the plan came from a local `./plans/*.md`
   file, which provenance on its own would misread as GitHub.
2. **The source's provenance** (step 1). A Linear project or a Notion page → **Linear**. Anything
   else → **GitHub**, which is also the fallback for personal repos.

Both are proposals, not rules. The user confirming one is a single word, whereas a wrong silent
guess is a whole batch of tickets in the wrong tracker — and a path check quietly stops matching if
the checkout ever moves, so never let it decide unconfirmed.

For Linear, settle the **project** in the same breath — take it from the source's Linear project if
it had one, else ask. The **team** is not yours to resolve: `create-adhoc-linear-issue` does that in
step 5.

Iterate until the user approves the breakdown.

### 5. Create the tickets

For each approved slice, create a ticket in the tracker settled in step 4, using the body template
below. Create in **dependency order (blockers first)** either way, so a blocker already exists — and
has a real identifier — by the time the ticket it blocks is created.

**GitHub** — `gh issue create --title "<title>" --body-file <file>`. Blocking is prose only: the
body's "Blocked by" section carries `#<number>`.

**Linear** — file each ticket through the **`create-adhoc-linear-issue`** skill rather than calling
`save_issue` directly. It is a Plain plugin skill, installed globally, and it owns the house
conventions for putting an issue into Linear: team resolution (`get_user` on `"me"`, dropping
`[Deprecated]` teams, never hardcoding one), the create call, and confirm-before-create. Keeping
those in one maintained place is the point — do not restate or second-guess them here.

It exists for drive-by tickets, though, so three of its behaviours are wrong for a planned batch.
Override them explicitly:

- **Body** — it drafts title and description from the branch's commits and diff. There is no branch
  yet, and the content is the vertical slice from the template below. Hand it the drafted title and
  description; do not let it derive them from git.
- **Project** — it files floating, deliberately, to keep ad-hoc work off project boards. These are
  planned feature tickets, so always pass the project settled in step 4.
- **Confirmation** — its per-issue confirm is already satisfied by the step 4 approval. Say so and
  let it create, rather than re-confirming every slice of a batch the user just signed off.

It sets neither blocking nor state, so those stay yours. Capture each returned `identifier`, then
record blocking with the **native relation, not prose** — one `save_issue` update per blocked
ticket, passing its `id` and `blockedBy: ["<identifier>", …]`, the identifiers of the slices it
waits on. Creating in dependency order is what makes those identifiers exist by then; `blockedBy` is
append-only, so the update adds relations without disturbing anything.

Then **omit the body's "Blocked by" section entirely** in Linear. The relation is the record — it is
what the board, the blocked-issue filters and the state machine act on, and Linear renders it on the
issue already. A prose copy alongside it is a second source of truth that drifts the moment a
dependency changes.

<ticket-template>
<Source-reference header — pick ONE to match the source type from step 1:>
<  Plan file:    "## Source plan" + `./plans/<file>.md` + Source PRD link from the plan header, if present>
<  PRD:          "## Parent PRD" + ./docs/prd-<feature>.md, the Notion URL, and/or #<prd-issue-number>>
<  Conversation: "## Source conversation" + one-line note on what the conversation was>

## What to build

A concise description of this vertical slice. Describe the end-to-end technical behaviour — what gets created, called, stored, and returned. Reference specific sections of the source plan rather than duplicating content.

## Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Blocked by
<GitHub only — in Linear this is the `blockedBy` relation, and the section is omitted>

- Blocked by #<issue-number> (if any)

Or "None - can start immediately" if no blockers.

## User stories addressed

Reference by number from the source (plan / PRD / conversation):

- User story 3
- User story 7

</ticket-template>

Do NOT modify the source plan file or the parent PRD.

### 6. Hand off to the board

Slices with no blockers are grabbable now; anything blocked by another slice is not, and lands
behind it.

**GitHub** — state rides on a `status:*` label, applied after creation:

```bash
gh issue edit <number> --add-label "status:ready"     # blocked slices: status:backlog
```

If those labels don't exist in the repo, create the two you need and move on — `status:ready`
(`-c "#0e8a16"`), `status:backlog` (`-c "#cccccc"`). `ticket-tracker` owns the full vocabulary;
don't invent names beyond these.

**Linear** — states are native, so there are no labels to bootstrap. `create-adhoc-linear-issue`
doesn't set state, so fold it into the same follow-up `save_issue` update that records `blockedBy`:
`Todo` for grabbable slices, `Backlog` for blocked ones. One update per ticket, not two. Confirm the
team's real state names with `list_issue_statuses` rather than assuming — teams rename them.

Report the created ticket identifiers with their states, and the blocking relations you set.

From this point `ticket-tracker` owns the tickets — it reads and moves them between states. This
skill does not track what it files.

### 7. Emit the dependency graph for whatever orchestrates this repo

Two things can read this graph, and they conflict if both get written: check for command-centre
first, since a command-centre repo already uses treepad internally to cut its worktrees, and a
Batch Manifest on top would hand the same worktrees to treepad's own `Reconcile` too.

**command-centre** — check for its config:

```bash
ls "$(git rev-parse --show-toplevel)/../.claude/command-centre.toml" 2>/dev/null
```

**Absent** — fall through to the treepad check below.

**Present** — for each ticket just created, write or update a `[[task]]` block in that file:

```toml
[[task]]
ticket_url = "<the ticket's URL>"
repo       = "<the checkout's [[repo]] name>"
branch     = "<repo-prefix>-<issue-number>-<short-slug>"
blocked_by = ["<ticket URL>", …]
seams      = ["<seam name>", …]
```

Edges only, no chain semantics: a fan-out (one ticket blocking four) is four `[[task]]` blocks,
each naming that one blocker — the app derives every base itself. Branch names follow whatever's
already in the file for that `[[repo]]` (e.g. `cc-33-fan-out`); check an existing `branch` value
there before inventing a new prefix.

**`seams`** names an agreed-but-unlanded interface in another repo that this ticket's agent needs
pasted into its prompt — written by `/to-seams`, never by this skill. Only a ticket whose slice is
a *consumer* of a seam carries the field; omit it entirely rather than writing `seams = []`, which
is the same "no seams" state a task with the field left out already gets. If a slice you're
cutting depends on an interface that hasn't landed anywhere yet, that's a sign to run `/to-seams`
first, then reference the name it gives you — this skill does not invent seam names.

**Upsert on `ticket_url`.** Read the file first. A `[[task]]` block already carrying this ticket's
`ticket_url` gets its fields replaced in place, so re-running over an edited feature updates rows
rather than minting duplicates. Otherwise append a new block.

Command-centre present means skip the treepad Batch Manifest below entirely — don't write both.

**treepad (only when command-centre is absent)** — check for a treepad batches directory:

```bash
ls "$(git rev-parse --git-common-dir)/treepad/batches/"
```

**Absent** — skip, and say nothing. Most repos use neither and the step is a no-op there.

**Present** — write `<feature>.toml` into that directory, per
[`references/treepad-manifest.md`](references/treepad-manifest.md). It declares the blocking graph
from step 4 as ordered Chains, which treepad materialises into stacked worktrees. Write it from the
graph you already hold — never read the relations back off the tracker.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
