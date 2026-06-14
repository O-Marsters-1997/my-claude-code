---
name: prd-to-plan
description: Turn a PRD into a multi-phase implementation plan using tracer-bullet vertical slices, saved as a local Markdown file in ./plans/. Use when user wants to break down a PRD, create an implementation plan, plan phases from a PRD, or mentions "tracer bullets".
---

# PRD to Plan

This skill produces the **technical design document** — it owns the HOW that the PRD deliberately leaves out. Where the PRD specifies WHAT and WHY in product language, this plan specifies HOW in technical language: data models, schema shapes, API contracts, module boundaries, and integration points. The plan must be concrete enough that plan-to-issues can derive independently-grabbable tickets from it without re-deriving the design.

Break a PRD into a phased implementation plan using vertical slices (tracer bullets). Output is a Markdown file in `./plans/`.

## Process

### 1. Confirm the PRD is in context

The PRD should already be in the conversation. If it isn't, ask the user to paste it or point you to the file or GitHub issue.

### 2. Explore the codebase

If you have not already explored the codebase, do so to understand the current architecture, existing patterns, and integration layers.

### 3. Identify durable technical design decisions

Before slicing, identify technical decisions that are unlikely to change throughout implementation. This is the core of the plan as a technical design document. Capture:

- **Route structures / URL patterns**
- **Database schema shape** — table names, key columns, relationships
- **Key data models** — names, fields, and their responsibilities
- **Module boundaries** — what each module owns and exposes. Prefer **deep modules**: substantial functionality behind a simple, stable interface that can be tested in isolation (as opposed to a shallow module, whose interface is complex relative to what it does). Note which modules warrant isolated tests.
- **API / interface contracts** — endpoint shapes, function signatures (at the module boundary level, not internal implementation)
- **Integration points** — third-party services, async boundaries, storage adapters
- **Key flows / algorithms** — non-trivial logic described at the level of "what calls what"
- **Authentication / authorization approach**

These go in the plan header so every phase can reference them. Be concrete.

### 4. Draft vertical slices

Break the PRD into **tracer bullet** phases. Each phase is a thin vertical slice that cuts through ALL integration layers end-to-end, NOT a horizontal slice of one layer.

<vertical-slice-rules>
- Each slice delivers a narrow but COMPLETE path through every layer (schema, API, UI, tests)
- A completed slice is demoable or verifiable on its own
- Prefer many thin slices over few thick ones
- INCLUDE durable technical design: schema shapes, route paths, model and module names, contract shapes — the details a developer needs to implement the slice
- AVOID volatile line-level detail: exact function names or internal file paths that are likely to change as later phases are built
</vertical-slice-rules>

### 5. Quiz the user

Present the proposed breakdown as a numbered list. For each phase show:

- **Title**: short descriptive name
- **User stories covered**: which user stories from the PRD this addresses

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Should any phases be merged or split further?

Iterate until the user approves the breakdown.

### 6. Write the plan file

Create `./plans/` if it doesn't exist. Write the plan as a Markdown file named after the feature (e.g. `./plans/user-onboarding.md`). Use the template below.

<plan-template>
# Plan: <Feature Name>

> Source PRD: <issue URL or brief identifier>

## Technical design decisions

Durable decisions that apply across all phases. Be concrete — this section is the technical design document:

- **Routes**: ...
- **Schema**: ...
- **Key models**: (names, fields, responsibilities)
- **Module boundaries**: (what each module owns and exposes)
- **API / interface contracts**: (endpoint or function shapes at the boundary level)
- **Integration points**: ...
- (add/remove sections as appropriate)

---

## Phase 1: <Title>

**User stories**: <list from PRD>

### What to build

A concise description of this vertical slice. Describe the end-to-end technical behaviour — what gets created, called, stored, and returned. Reference the technical design decisions above rather than repeating them.

### Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

---

## Phase 2: <Title>

**User stories**: <list from PRD>

### What to build

...

### Acceptance criteria

- [ ] ...

<!-- Repeat for each phase -->
</plan-template>
