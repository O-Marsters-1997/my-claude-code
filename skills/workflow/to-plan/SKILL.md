---
name: to-plan
description: >
  Turns one feature's PRD into a skimmable plan in docs/plans/<feature>.md: a short overview, then
  per phase a glossary, the journeys it satisfies, and its new and modified definitions (types,
  enums, ports, schema), with the project:<slug> label. No code, branch or PR; the user's review
  of the file is the gate. Use for "plan this feature", "write the contract", "what are the
  types". One feature per run. Slash-only.
disable-model-invocation: true
---

# To Plan

```
FEATURE — one feature per run    to-prd → to-plan → to-tickets → fleet
                                          ^^^^^^^ you are here
```

Write the technical design as definitions, not code. The user skims the types, enums, ports and
schema to check that the product maps onto the domain; the implementer reads the same definitions
for the detail. Anything that follows a pattern the repo already has (how a type is fetched,
queried, mocked or routed) is left to the implementer, who knows the pattern; writing it out buries
the design.

Judge every boundary with the `codebase-design` vocabulary.

## Process

### 1. Locate the input

Read `docs/prds/<feature>.md`. If the user names a feature, use that file. If more than one PRD
exists and the target is not obvious, ask which one before starting; never plan several features in
one run. If no PRD exists, write the core inline (below) and carry on, rather than sending the user
to `/to-prd`.

`<feature>` is the PRD's filename without `.md`; `<slug>` below is the same string.

**No PRD: write the core.** Ask a short interview, one question at a time and skipping what the
conversation already answers: the problem in one line, the journeys (each with a surface, 2 to 5
steps and a `→ <observable state>` final line), then the delivery slices. Write it to
`docs/prds/<feature>.md` (kebab-case, creating `docs/prds/` if missing) in `to-prd`'s core format
(its `references/project-doc.md`): `### Problem`, numbered `### J1.` journeys, and numbered delivery
slices each naming what is usable after it and its `Journeys: J…`. Every journey sits in exactly one
slice. Show it to the user, confirm, then continue with step 2. Skip the optional layer.

### 2. Explore the codebase

Read the existing architecture, patterns, naming, the `CONTEXT.md` glossary if present, and the
types the feature will have to change. Write nothing outside `docs/` in this skill: no stubs, no
branch, no build.

**PRD drift.** When exploration shows a PRD decision won't hold (a cost model, an ownership rule, a
place the logic can't live), list each one and ask the user: edit the PRD now, or record it under
`PRD drift` in the plan's Notes. Never leave a changed decision only in the plan without asking,
or the PRD goes stale while still looking authoritative.

### 3. Design through usage, privately

For each journey, sketch the code a caller would write, in your own reasoning only. An awkward
sketch means the interface is wrong; change it now, while it costs nothing. The sketches never go
in the plan: the user reviews definitions, and call sites are pattern the implementer already knows.

### 4. Derive the definitions per phase

A phase is one PRD delivery slice, in the PRD's order, each building on the ones before. Place each
definition in the first phase that needs it, and cut any that no journey in its phase needs.

- **Definitions only.** Domain types, enums, DTO and wire shapes, port interfaces between modules,
  and the schema as table definitions. Never call sites, query or mutation wiring, fetch functions,
  hooks, mocks, fakes, route bindings or migrations: each follows a repo pattern, and the schema's
  end state is what the user reviews. Write them in the repo's own languages.
- **New and Modified.** `New` holds definitions that don't exist yet. `Modified` shows an existing
  definition's final shape (the whole type or table, or the part that matters), stated flatly. Never
  write "adds X" or "was Y": the reader needs the end state, and git holds the delta.
- **Nothing points forward.** Whatever a phase needs is defined in that phase or an earlier one.
  Tickets are built in parallel with no compiled stubs, so a forward reference is a phase that
  cannot start.
- **Glossary.** Open each phase with one line per new domain term: what it is and what it relates
  to. Name things in the domain's words, matching `CONTEXT.md` where it defines the term.

Judge each boundary with `codebase-design`: prefer deep modules, and put a port only where behaviour
must vary or be tested in isolation.

### 5. Quiz the user

Present the module boundaries and the phases as a numbered list, each named as the PRD names its
slice with its journey numbers; split one only if the user agrees, keeping the parts under its name.
For each phase show what is usable after it in product terms. Ask whether the granularity and
boundaries feel right. Iterate until approved. Phases are vertical: each is demoable on its own.

### 6. Settle the feature branch and label

The feature branch is the one the user names, else the branch checked out when the run started.
If that is the default branch, propose `feature/<slug>` and ask; tickets never build on the default
branch. The plan records the name only; this skill creates no branch.

```bash
gh label list --search "project:<slug>" --json name -q '.[].name' | grep -qx "project:<slug>" || \
  gh label create "project:<slug>" -d "<Feature> (docs/plans/<slug>.md)" -c "<any unused hex>"
```

### 7. Write the plan

Write `docs/plans/<feature>.md` in the format in [references/plan-format.md](references/plan-format.md),
matching the density of [references/example-plan.md](references/example-plan.md).
If the file already exists, show what would change and ask before overwriting: it may hold an
approved plan that tickets were already cut from.

Before writing, check:

- **Coverage.** Every definition appears in exactly one phase, and every phase heading matches a
  PRD delivery slice.
- **Skim test.** A reader gets each phase from its glossary and headings alone, and each phase fits
  about one screen. If it doesn't, cut detail (fields that follow a pattern, prose), not definitions.

Run `git check-ignore -q docs/plans/<feature>.md` and say in the closing message whether the plan is
local-only or will be committed.

Close by naming the plan path, the label and the feature branch, and ask the user to review the file:
their approval is the gate `/to-tickets` waits on.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
