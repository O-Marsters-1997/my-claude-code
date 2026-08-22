---
name: artifact-scan
description: >
  Use whenever the user asks "which skill should I use?", "where are we in the workflow?",
  "what should I run next?", "what product docs already exist?", "do we have a roadmap/PRD/plan
  yet?", or "what's left before I start coding?". Front door to the product workflow suite:
  detects the ideate report, approach doc, roadmap, PRDs, plan files and tickets, then names the
  one skill to run next based on what exists rather than a fixed order. Routes between *skills* —
  for "what feature should I build next?" use ideate. Also usable as a preflight routine from
  other product skills.
---

# Artifact Scan

A fast, read-only routine. Scan the repo for the product artifacts, report what exists, and point
the user at exactly one next skill. No interview, no grilling, no writing. Just look, report, route.

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-plan → to-tickets → ticket-tracker
```

- **The doors** (`ideate`, `chat-to-approach`, or an already-scoped task) are optional ways in.
- **Portfolio → feature is a fan-out, not a step.** A roadmap card is a *feature* — something to be
  built, which then breaks down into several tickets. Six cards means six runs down the spine, not
  one. When a skill is handed a portfolio artifact, it must establish *which feature* first.
- **Nothing gates anything.** Upstream artifacts make a stage faster, they never make it required.

That last point drives this skill's whole job: **report what exists, don't compute a first gap.**
An earlier artifact being absent is not a problem to solve — it is usually a deliberate skip.

## The artifacts

| Artifact         | Level     | How to detect it                                       | Skill that produces it |
|------------------|-----------|--------------------------------------------------------|------------------------|
| ideate report    | portfolio | `ls ./ideas/reports/*-ideate.md`                       | `ideate`               |
| approach doc     | portfolio | `test -f ./docs/approach.md`                           | `chat-to-approach`     |
| roadmap board    | portfolio | `test -f ./roadmap.html`                               | `to-roadmap`           |
| PRD              | feature   | `ls ./docs/prd-*.md` and `gh issue list` for `[PRD]`   | `to-prd`               |
| plan file        | feature   | `ls ./plans/*.md`                                      | `to-plan`              |
| tickets          | feature   | `gh issue list` — any issue NOT titled `[PRD]`         | `to-tickets`           |

Feature-level artifacts are **countable, not binary**. Three plan files means three features are
mid-spine; it does not mean "the plan stage is done". Report counts and names, never a single tick.

## Procedure

1. **Scan local files** in one pass:
   ```bash
   ls ./ideas/reports/*-ideate.md ./docs/approach.md ./roadmap.html \
      ./docs/prd-*.md ./plans/*.md 2>/dev/null
   ```
2. **Scan GitHub issues** once, then split by title:
   ```bash
   gh issue list --state open --limit 100 --json number,title 2>/dev/null
   ```
   - PRDs: any title starting `[PRD]`.
   - Tickets: any open issue that is *not* a `[PRD]` issue.
   - If `gh` is unavailable or errors (no remote, not authed), mark the GitHub-backed artifacts
     **unknown**, not absent — never guess. Say so in the report. Local `./docs/prd-*.md` files
     still count as PRDs, so a repo with no remote can still show the PRD artifact present.
3. **Report what you observed**, citing the file path or issue number as evidence. Never report an
   artifact you did not see. Group by level so the user can tell portfolio work from feature work.
4. **Route.** Do not compute "the earliest gap". Ask what the user has and match it:

   | What the user has | Route to | Why |
   |---|---|---|
   | A scoped feature or a clear idea | `to-prd` | The spine starts here. Prior artifacts are optional accelerants |
   | Nothing — no idea what to build | `ideate` | The only door that generates an idea from the codebase |
   | A conversation from claude.ai or elsewhere | `chat-to-approach` | Captures the handoff into `approach.md` |
   | An approach doc, wants to prioritise | `to-roadmap` | Turns many features into Now / Next / Later |
   | A roadmap, wants to start building | Pick one card → `to-prd` | Each card is its own spine run |
   | A PRD for the feature in hand | `to-plan` | |
   | A plan for the feature in hand | `to-tickets` | |
   | Tickets already filed | `ticket-tracker` | Move them through the board |

   When the invocation itself makes the intent obvious, route on that and say so — a user who
   describes a feature is not asking to ideate, even on an empty repo.

   If you genuinely cannot tell which of two doors applies, ask **one** question via
   `AskUserQuestion` rather than guessing. Recommend exactly one skill.

## Report format

```
Product artifacts

  Portfolio (many features)
    [x] ideate report  — ./ideas/reports/2026-05-01-ideate.md
    [x] approach doc   — ./docs/approach.md
    [ ] roadmap        — not found

  Feature (per feature)
    [x] PRDs           — 2: ./docs/prd-search.md, #41 [PRD] Saved views
    [x] plans          — 1: ./plans/search.md
    [?] tickets        — unknown (gh unavailable)

Next: `to-plan` — "Saved views" has a PRD but no plan yet.
```

Use `[x]` present, `[ ]` absent, `[?]` unknown. Always list all six so the user sees the whole
suite. Close with one line naming exactly one skill and the reason.

Absent portfolio artifacts are **not** a finding. Do not tell the user to go back and write an
approach doc because one is missing — say what is missing, then let the routing table decide.

## Notes

- Read-only. This skill never creates or edits artifacts — it only reports and routes.
- When invoked as a preflight by another skill, return the artifact map and skip the prose framing
  and the routing verdict — the calling skill does its own routing.
