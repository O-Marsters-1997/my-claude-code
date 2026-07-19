---
name: to-plan
description: Turn a PRD — or an approach doc or raw conversation — into a multi-phase implementation plan for one feature, using tracer-bullet vertical slices, saved as a local Markdown file in ./plans/. Use when the user wants to go from PRD to plan, break down a PRD, create an implementation plan, plan phases from a PRD, plan straight from an approach doc (./docs/approach.md) or a pasted conversation, "skip straight to a plan", or mentions "tracer bullets".
---

# PRD to Plan

This skill produces the **technical design document** — it owns the HOW that the PRD deliberately leaves out. Where the PRD specifies WHAT and WHY in product language, this plan specifies HOW in technical language: data models, schema shapes, API contracts, module boundaries, and integration points. The plan must be concrete enough that to-tickets can derive independently-grabbable tickets from it without re-deriving the design.

Break a PRD into a phased implementation plan using vertical slices (tracer bullets). Output is a Markdown file in `./plans/`.

## Process

### 1. Locate the input (PRD by default)

A PRD is the default source and takes priority. It exists in two forms, either of which is fine: a local `./docs/prd-<feature>.md` file (the artifact `to-prd` writes) or a `[PRD]`-titled GitHub issue (a copy of the same content). Prefer the file when both are present — it needs no network and is what `to-prd` treats as canonical.

If a PRD is already in the conversation — or the user points you to one — use it and skip the rest of this step; behave exactly as the PRD path always has.

Otherwise the goal is to **skip straight to a plan** from whatever artifact exists. To locate inputs, run the `artifact-scan` skill as a preflight — it reports which artifacts are present. Do not re-implement its detection here. Build the plan from the first available of:

- an approach doc at `./docs/approach.md`
- a raw conversation summary or transcript the user has pasted

If none of these and no PRD exist, ask the user to paste one or point you to a file or GitHub issue. Whatever the source, treat it as the WHAT/WHY input and produce the same technical design plan described below.

**A plan covers one feature.** An approach doc or a pasted conversation is portfolio-level — it may hold several. If the target feature isn't obvious from the invocation, ask which one before drafting; don't silently plan all of them or pick the first. (See `../README.md`.)

### 2. Explore the codebase

If you have not already explored the codebase, do so to understand the current architecture, existing patterns, and integration layers.

### 3. Identify durable technical design decisions

Before slicing, identify technical decisions that are unlikely to change throughout implementation. This is the core of the plan as a technical design document. Capture:

- **Route structures / URL patterns**
- **Database schema shape** — table names, key columns, relationships
- **Key data models** — names, fields, and their responsibilities
- **Module boundaries** — what each module owns and exposes. Prefer **deep modules**: substantial functionality behind a simple, stable interface that can be tested in isolation (as opposed to a shallow module, whose interface is complex relative to what it does). Note which modules warrant isolated tests.

  When a module boundary is central to the design and the right shape isn't obvious, run the `design-an-interface` skill on it before committing — it generates several radically different shapes in parallel and compares them. A boundary is the most expensive thing in this document to get wrong: slices are re-derived freely by `to-tickets`, but every slice is built against the interface. Use it sparingly, on the one or two boundaries the feature actually hinges on.
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
- **User stories covered**: which user stories or features from the source this addresses

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Should any phases be merged or split further?

Iterate until the user approves the breakdown.

### 6. Write the plan file

Create `./plans/` if it doesn't exist. Write the plan as a Markdown file named after the feature (e.g. `./plans/user-onboarding.md`). Use the template below.

<plan-template>
# Plan: <Feature Name>

> Source: <./docs/prd-<feature>.md, PRD issue URL, ./docs/approach.md, or brief identifier>

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

**User stories**: <list from source>

### What to build

A concise description of this vertical slice. Describe the end-to-end technical behaviour — what gets created, called, stored, and returned. Reference the technical design decisions above rather than repeating them.

### Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

---

## Phase 2: <Title>

**User stories**: <list from source>

### What to build

...

### Acceptance criteria

- [ ] ...

<!-- Repeat for each phase -->
</plan-template>

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
