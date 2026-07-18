---
name: to-tickets
description: Break a source — an implementation plan file (default), a PRD GitHub issue, or a raw conversation — into independently-grabbable GitHub issues using tracer-bullet vertical slices. Use when the user wants to convert a plan to issues, create tickets from a plan file, go from PRD to issues, cut tickets from a PRD, create tickets from a PRD, turn a conversation into issues, or break work into grabbable GitHub items.
---

# Plan to Issues

Break a source into independently-grabbable GitHub issues using vertical slices (tracer bullets). The source is one of three, in priority order:

1. **A plan file** (DEFAULT) — a local `./plans/*.md` file produced by to-plan.
2. **A PRD GitHub issue** — a `[PRD]`-titled issue.
3. **A raw conversation** — a pasted or in-context discussion the user wants turned into tickets.

A plan is a technical design document: its architectural decisions (data models, schema shapes, API contracts, module boundaries) are durable constraints and its user stories are source material. A PRD or conversation carries product intent with less technical design — if a source is thin on technical design, note this to the user; it may need a to-plan pass first.

## Process

### 1. Locate the source

If the user named a source, use it. Otherwise detect what exists by running the `artifact-scan` skill as a preflight — it reports whether a plan file or a `[PRD]` issue is present. Apply the priority order above: prefer a plan file, then a PRD issue, then a conversation already in context. If more than one is present, confirm which to use.

Load the chosen source in full:

- **Plan file** — read the `./plans/*.md` file. Record its path and the **Source PRD** link from the plan header (if present).
- **PRD issue** — fetch with `gh issue view <number>` (with comments). Record the issue number.
- **Conversation** — use the in-context discussion, or ask the user to paste it.

Extract from whichever source: **technical design decisions** (data models, schema, API contracts, module boundaries, routes), **user stories**, and **acceptance criteria**. Note the source type — it selects the issue-body source header in step 5.

### 2. Explore the codebase (optional)

If you have not already explored the codebase, do so to understand the current state of the code and how the plan's technical decisions map to existing structure.

### 3. Draft vertical slices

Break the plan into **tracer bullet** issues. Each issue is a thin vertical slice that cuts through ALL integration layers end-to-end, NOT a horizontal slice of one layer.

Re-derive slices freely — do not map plan phases 1:1 to issues.

Slices may be 'HITL' or 'AFK'. HITL slices require human interaction, such as an architectural decision or a design review. AFK slices can be implemented and merged without human interaction. Prefer AFK over HITL where possible.

<vertical-slice-rules>
- Each slice delivers a narrow but COMPLETE path through every layer (schema, API, UI)
- A completed slice is demoable or verifiable on its own
- Prefer many thin slices over few thick ones
- Respect the plan's **deep modules** — substantial functionality behind a simple, stable, isolation-testable interface (vs shallow modules). A slice should build or extend a deep module behind its interface, not splay logic across shallow ones.
</vertical-slice-rules>

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each slice, show:

- **Title**: short descriptive name
- **Type**: HITL / AFK
- **Blocked by**: which other slices (if any) must complete first
- **Plan phases / user stories covered**: which phases and user stories from the plan this addresses

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Are the dependency relationships correct?
- Should any slices be merged or split further?
- Are the correct slices marked as HITL and AFK?

Iterate until the user approves the breakdown.

### 5. Create the GitHub issues

For each approved slice, create a GitHub issue using `gh issue create`. Use the issue body template below.

Create issues in dependency order (blockers first) so you can reference real issue numbers in the "Blocked by" field.

<issue-template>
<Source-reference header — pick ONE to match the source type from step 1:>
<  Plan file:    "## Source plan" + `./plans/<file>.md` + Source PRD link from the plan header, if present>
<  PRD issue:    "## Parent PRD" + #<prd-issue-number>>
<  Conversation: "## Source conversation" + one-line note on what the conversation was>

## What to build

A concise description of this vertical slice. Describe the end-to-end technical behaviour — what gets created, called, stored, and returned. Reference specific sections of the source plan rather than duplicating content.

## Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Blocked by

- Blocked by #<issue-number> (if any)

Or "None - can start immediately" if no blockers.

## User stories addressed

Reference by number from the source (plan / PRD / conversation):

- User story 3
- User story 7

</issue-template>

Do NOT modify the source plan file or the parent PRD issue.
