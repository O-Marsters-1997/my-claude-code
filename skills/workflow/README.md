# Product workflow suite

Eight skills that take work from a raw idea (or an already-scoped task) through to tickets on a
board. This file is the map for humans working on the suite.

## Installing

Install all eight — the skills call each other, so a partial install leaves broken handoffs:

```
for s in artifact-scan ideate chat-to-approach to-roadmap to-prd to-plan to-tickets ticket-tracker; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

**This suite must stay at `skills/workflow/<name>/`.** The `skills` CLI discovers
`skills/<category>/<name>/SKILL.md` and nothing deeper. When the suite briefly lived at
`skills/product/workflow/<name>/` it was invisible to `npx skills add` — six of the eight silently
went uninstalled. Nesting one directory further to tidy things up breaks installation, with no
error to tell you.

The routines the suite calls (`../product/grill-with-docs`, `../product/source-synthesis`,
`../product/design-an-interface`, `../product/triage-issue`) install the same way.

---

**Skills must not reference this file.** Installation is flat — each skill lands in its own directory
under the agent's skills root, so `../README.md` does not resolve once installed, and the map would
be unreachable at exactly the moment it's needed. Every skill therefore carries its own `## Where
this sits` block: the same diagram, an arrow marking its position, and the two rules below. When the
shape of the suite changes, update this file *and* all eight blocks.

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

A **roadmap card is a feature** — something that should be built — *not* a ticket. It is expected to
break down into several tickets, which happens at `to-tickets`, three stages later. `roadmap.html`
and the `ticket-tracker` board are different altitudes of the same work, not rival copies of it.

Crossing from portfolio to feature is a **fan-out, not a step**. A roadmap with six cards means six
independent spine runs, each ending in its own handful of tickets. When a spine skill is handed a
portfolio artifact as its source, its first job is to establish *which feature* — ask if it isn't
obvious.

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

From `../product/grill-with-docs/SKILL.md`, and it applies suite-wide:

> A **user-invoked skill** orchestrates a workflow and may call **model-invoked routines**. A routine
> is not triggered directly by the user — it is delegated to from within another skill's flow.

Every skill in this directory is an orchestrator except `artifact-scan`, which is both (a front door
when the user asks, a preflight when a skill calls it).

The routines these orchestrators call live **outside** this directory. They are not stages and never
appear in the map above:

| Routine | Called by | For |
|---|---|---|
| `../product/grill-with-docs` | `chat-to-approach`, `to-prd`, `to-roadmap` | the canonical grilling loop — never re-inline it |
| `../product/source-synthesis` | `chat-to-approach` | the optional `## Background Research` section of `approach.md` |
| `../product/design-an-interface` | `to-plan` | designing a module boundary twice before committing to one |
| `../product/triage-issue` | `ticket-tracker` | bug-driven tickets — a second door straight to the board, bypassing the spine |

## Not in scope

Implementation. The suite ends when tickets exist. The seam for later is `to-tickets` → `tdd`, with
the plan's technical design decisions as the constraint to check against.
