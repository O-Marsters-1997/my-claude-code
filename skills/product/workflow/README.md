# Product workflow suite

Eight skills that take work from a raw idea (or an already-scoped task) through to tickets on a
board. This file is the map. Read it before assuming how any two of these skills relate.

## The shape

```
DOORS — optional, pick whichever matches what you have
                                                            PORTFOLIO LEVEL
  ideate ──────────────┐                                    many features at once
  (no idea yet)        │
                       ├──→  approach.md  ──→  roadmap.html
  chat-to-approach ────┘     (alignment)       (Now/Next/Later)
  (handoff from a chat)                              │
                                                     │ pick ONE card
  a scoped task ───────────────────────────────┐     │
  (you already know what you want)             │     │
                                               ▼     ▼
                                    ┌───────────────────────────┐
                                    │  to-prd  →  to-plan  →    │  FEATURE LEVEL
                                    │  to-tickets               │  one feature per run
                                    └───────────────────────────┘
                                                     │
                                                     ▼
                                              ticket-tracker
                                              (live state on GitHub)
```

## Two rules that explain everything else

**1. The spine runs per feature. The portfolio artifacts describe many features.**

`approach.md` and `roadmap.html` are portfolio-level: they hold every feature you're considering.
`to-prd`, `to-plan` and `to-tickets` handle exactly one feature per run.

Crossing from portfolio to feature is a **fan-out, not a step**. A roadmap with six cards means six
independent spine runs. When a spine skill is handed a portfolio artifact as its source, its first
job is to establish *which feature* — ask if it isn't obvious.

**2. Nothing gates anything.**

Every skill runs standalone. Upstream artifacts make a stage faster and better-informed; they never
make it required. `to-prd` with an approach doc skips questions it can already answer; `to-prd` with
nothing runs the full interview. Both are correct uses.

This is deliberate. The suite is a set of loosely-coupled converters, not a pipeline with gates, so
that each stage can be iterated on alone as well as in sequence. Do **not** send a user backwards to
create a missing artifact — if they have what the stage needs, run the stage.

## The skills

| Skill | Level | Takes | Produces |
|---|---|---|---|
| `artifact-scan` | — | nothing | a report + one routing recommendation (also the preflight routine) |
| `ideate` | portfolio | the codebase | `./ideas/reports/YYYY-MM-DD-ideate.md`, `./ideas/CONTEXT.md` |
| `chat-to-approach` | portfolio | a pasted conversation | `./docs/approach.md` |
| `to-roadmap` | portfolio | `./docs/approach.md` | `./roadmap.html` |
| `to-prd` | feature | anything above, or an interview | `./docs/prd-<feature>.md` + a `[PRD]` issue |
| `to-plan` | feature | a PRD | `./plans/<feature>.md` |
| `to-tickets` | feature | a plan | GitHub issues |
| `ticket-tracker` | feature | GitHub issues | `status:*` label moves |

## Orchestrators vs routines

From `../grill-with-docs/SKILL.md`, and it applies suite-wide:

> A **user-invoked skill** orchestrates a workflow and may call **model-invoked routines**. A routine
> is not triggered directly by the user — it is delegated to from within another skill's flow.

Every skill in this directory is an orchestrator except `artifact-scan`, which is both (a front door
when the user asks, a preflight when a skill calls it).

The routines these orchestrators call live **outside** this directory. They are not stages and never
appear in the map above:

| Routine | Called by | For |
|---|---|---|
| `../grill-with-docs` | `chat-to-approach`, `to-prd`, `to-roadmap` | the canonical grilling loop — never re-inline it |
| `../source-synthesis` | `chat-to-approach` | the optional `## Background Research` section of `approach.md` |
| `../design-an-interface` | `to-plan` | designing a module boundary twice before committing to one |
| `../triage-issue` | `ticket-tracker` | bug-driven tickets — a second door straight to the board, bypassing the spine |

## Not in scope

Implementation. The suite ends when tickets exist. The seam for later is `to-tickets` → `tdd`, with
the plan's technical design decisions as the constraint to check against.
