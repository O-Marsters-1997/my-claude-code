---
name: plan-to-issues
description: Break an implementation plan into independently-grabbable GitHub issues using tracer-bullet vertical slices. Use when the user wants to convert a plan to issues, create tickets from a plan file, or break a ./plans/ file into work items.
---

# Plan to Issues

Break an implementation plan (a local `./plans/*.md` file produced by prd-to-plan) into independently-grabbable GitHub issues using vertical slices (tracer bullets).

The plan is a technical design document. Its architectural decisions (data models, schema shapes, API contracts, module boundaries) are durable constraints; its user stories and acceptance criteria are source material.

## Process

### 1. Locate the plan

If a plan file is already in the conversation, use it. Otherwise ask the user for the path. Look first in `./plans/` (the default output location of prd-to-plan).

Read the plan in full. Extract:

- **Technical design decisions** (data models, schema, API contracts, module boundaries, routes)
- **Source PRD** link or identifier (from the plan header)
- **Per-phase user stories** and **acceptance criteria**

If the plan is thin on technical design (e.g. only describes product-level behaviour), note this to the user — the plan may need a prd-to-plan pass first.

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
## Source plan

`./plans/<file>.md`

<Source PRD link from the plan header, if present>

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

Reference by number from the source plan / PRD:

- User story 3
- User story 7

</issue-template>

Do NOT modify the source plan file.
