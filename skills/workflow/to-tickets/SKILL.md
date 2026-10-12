---
name: to-tickets
description: >
  Breaks a source into independently-grabbable GitHub tickets. For a code feature the source is the
  contract (`docs/contracts/<feature>.md`, the stub list from /to-scaffold) and every stub lands in
  exactly one ticket; a non-code feature is cut from the PRD's slices. Cuts tickets for one feature
  per run. If a code feature has no merged scaffold, offers /to-scaffold instead of cutting from
  prose. In a repo that runs command-centre, also emits `[[task]]` blocks into its config; in a
  plain treepad repo, emits a Batch Manifest instead so the tickets materialise as stacked worktrees.
disable-model-invocation: true
---

# Contract to Tickets

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./ideas/roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-scaffold → to-tickets → ticket-tracker
                                                        ^^^^^^^^^^ you are here
```

**Portfolio → feature is a fan-out, not a step.** A roadmap card is a *feature*, and this stage turns
exactly one of them into its several tickets. Never cut tickets for a whole roadmap or approach doc
in one run.

Break the source into independently-grabbable tickets using vertical slices (tracer bullets). Tickets
are filed on GitHub only.

## Process

### 1. Locate the source

If the user named a source, use it. Otherwise `ls ./docs/contracts/*.md ./docs/prds/*.md 2>/dev/null`.
Several present means several features: confirm which one rather than defaulting to the newest.

Decide whether the feature is code or not:

- **Code feature** — the source is the **contract**, `docs/contracts/<feature>.md`, in this format:

  ```markdown
  # Contract: <Feature>

  > Source: docs/prds/<feature>.md
  > Label: project:<slug>
  > Scaffold PR: #<n>

  The work is implementing the stubs. Signatures are frozen: changing one is a `contract-change` issue that blocks dependent work, never a quiet edit.

  ## Implementation order

  ### Slice 1: <what is usable after it, in product terms> (journeys 1, 2)
  - `path/to/file.go:Symbol`: one line on what it does

  ## Notes

  Only what code cannot say: rationale, ordering and invariants, gotchas, conventions, auth approach. About one screen.
  ```

  **No merged scaffold means no contract.** If there is no contract file, or `gh pr view <n>` shows
  its Scaffold PR is not merged, stop and offer `/to-scaffold`. Never cut a code feature from prose.
- **Non-code feature** (docs, skills, config) — the source is the PRD, `docs/prds/<feature>.md`, or a
  `[PRD]`-titled GitHub issue (`gh issue view <number>`, with comments). Cut from the PRD's slices.
  A conversation the user pastes stands in for the PRD here.

Record the source path or issue number. Settle the **feature label** every ticket gets: the contract's
`Label:` line, else ask the user for one before drafting. Never file unlabelled tickets. Feature
labels are `project:<slug>` (`gh label list`); offer the closest as a suggestion.

### 2. Explore the codebase (optional)

If you have not already explored the codebase, do so to see how the source maps to existing structure.

### 3. Draft vertical slices

Break the source into **tracer bullet** tickets. Each is a thin vertical slice that cuts through ALL
integration layers end-to-end, NOT a horizontal slice of one layer.

Re-derive slices freely. Do not carry the source's numbering into a ticket title. A ticket called
"Phase 3" or "Slice 2" is proof the slice was copied rather than re-derived, so name it for what it
delivers. The contract's slices are a starting point and its `Slice:` number is recorded on the
ticket; a ticket may still merge or split them.

For a code feature, **every stub belongs to exactly one ticket**. After drafting, list the contract's
stubs against the tickets and fix any stub that is uncovered or claimed twice.

Slices may be 'HITL' or 'AFK'. HITL slices require human interaction, such as an architectural decision or a design review. AFK slices can be implemented and merged without human interaction. Prefer AFK over HITL where possible.

<vertical-slice-rules>
- Each slice delivers a narrow but COMPLETE path through every layer (schema, API, UI)
- A completed slice is demoable or verifiable on its own
- Prefer many thin slices over few thick ones
- Respect **deep modules**: substantial functionality behind a simple, stable, isolation-testable interface. A slice should build or extend a deep module behind its interface, not splay logic across shallow ones.
- A blocking edge is only real if the blocked slice needs another stub's **behaviour**. Needing its types or signatures is never an edge: the scaffold already compiles them. **Sequence in the source is not a dependency.** Given an ordered list and no criteria, the cheapest consistent answer is "each waits for the previous", and that answer is almost always wrong.
- Slice by feature, never by layer. "Extract helpers", then "core function", then "CLI wiring", then "docs" is a horizontal decomposition wearing four ticket titles, and it is serial by construction.
</vertical-slice-rules>

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each slice, show:

- **Title**: short descriptive name
- **Type**: HITL / AFK
- **Blocked by**: which other slices (if any) must complete first
- **Stubs**: the `file:symbol` list (code feature), or the PRD slices covered (non-code)

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
- Which blocking edges are wrong? For each one kept, which stub's behaviour does it wait on?
- Should any slices be merged or split further?
- Are the correct slices marked as HITL and AFK?

Iterate until the user approves the breakdown.

### 5. Create the tickets

For each approved slice, run `gh issue create --title "<title>" --body-file <file>` with the body
template below. Create in **dependency order (blockers first)**, so a blocker already has a real
issue number by the time the ticket it blocks is created. Blocking is prose only: the body's
"Blocked by" section carries `#<number>`.

<ticket-template>
<Source-reference header, pick ONE:>
<  Contract:     "## Source contract" + `docs/contracts/<feature>.md` + Scaffold PR #<n>>
<  PRD:          "## Parent PRD" + `docs/prds/<feature>.md` and/or #<prd-issue-number>>

Slice: <n>
Journeys: <n, …>   (non-code feature: the PRD's journey numbers, or "none" if it lists none)

## Stubs

<Code feature: replaces prose "What to build".>

Signatures are frozen. Changing one is a `contract-change` issue, not an edit in this ticket.

- `path/to/file.go:Symbol`

<Non-code feature: use "## What to build" instead, a concise description of the slice, referencing the PRD rather than duplicating it.>

## Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Blocked by

- Blocked by #<issue-number> (if any)

Or "None - can start immediately" if no blockers.

</ticket-template>

Do NOT modify the contract or the PRD.

### 6. Hand off to the board

Slices with no blockers are grabbable now; anything blocked by another slice is not, and lands
behind it. State rides on a `status:*` label, and the feature label from step 1 goes on every
ticket. Apply both after creation:

```bash
gh issue edit <number> --add-label "status:ready" --add-label "<feature-label>"   # blocked slices: status:backlog
```

If the feature label doesn't exist yet, create it first with
`gh label create "<feature-label>" -d "<Feature Name> (docs/prds/<slug>.md)"`.

If those labels don't exist in the repo, create the two you need and move on — `status:ready`
(`-c "#0e8a16"`), `status:backlog` (`-c "#cccccc"`). `ticket-tracker` owns the full vocabulary;
don't invent names beyond these.

Report the created issue numbers with their states and the blocking edges. End with the command that
runs them: `/fleet dispatch <feature-label>`.

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
