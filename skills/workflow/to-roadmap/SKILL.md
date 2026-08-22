---
name: to-roadmap
description: >
  Prioritises settled features into a Now / Next / Later kanban board, rendered as a
  self-contained ./roadmap.html. Each card is a feature, not a ticket — one card later becomes
  several tickets via /to-prd -> /to-plan -> /to-tickets. Boards the features in
  ./docs/approach.md, or the "## Accepted ideas" in ./ideas/CONTEXT.md when there is no approach
  doc. The approach doc is settled truth: this assigns tiers and reasons about ordering, it
  never re-litigates what was decided.
disable-model-invocation: true
---

## Overview

Turn the features you have settled on into a visual product roadmap: a standalone HTML
kanban with three columns (Now / Next / Later) where each card is a **feature**, not a ticket — a
thing to be built, which the spine later breaks down into several tickets.

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./roadmap.html
                                                    ^^^^^^^^^^^^ you are here
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-plan → to-tickets → ticket-tracker
```

**Portfolio → feature is a fan-out, not a step.** This skill is portfolio-level: it prioritises many
features at once. Crossing to the spine means picking one card — six cards means six independent
runs of `to-prd → to-plan → to-tickets`, each producing its own several tickets.

**Nothing gates anything.** Board from `./docs/approach.md`, from `## Accepted ideas` in
`./ideas/CONTEXT.md`, or from a short list the user gives you. The one thing this skill will not do
is build a roadmap by parsing a raw conversation — `chat-to-approach` owns that (see Phase 0).

Solo/indie-builder scale throughout: no assignees, no story points, no sprints, no stakeholder
matrices.

---

## Phase 0 — Load the source

Two sources, in priority order. Run `artifact-scan` as a preflight if it's installed rather than
reimplementing its file checks.

**1. `./docs/approach.md`** — the alignment doc, and the richer source. Read it and extract,
treating every item as already agreed:

- All features (Name, Problem, Value, Dependencies) → cards
- Constraints → a note below the board, not cards
- Principles → use to inform tier suggestions (e.g. "ship weekly" → fewer Now items)
- Open questions → carry forward verbatim into the Open Questions section

**2. `## Accepted ideas` in `./ideas/CONTEXT.md`** — the ideas the user has decided to build,
whether `ideate` promoted them or they were hand-added. One line each, one card each. Accepting is
the selection, so board the whole section rather than asking which of them to include.

Do not modify either file.

**If both exist**, the approach doc wins: it has been grilled, and it carries the constraints,
principles and open questions a bare idea list doesn't.

**If neither exists**, ask the user for the features to board — a short list, one line each — and
mention that `ideate` or `chat-to-approach` leaves a durable source next time. What you must not do
is build the board by parsing a pasted conversation: `chat-to-approach` owns conversation →
structure, including the grilling that resolves ambiguity before anything is committed. Two skills
parsing transcripts differently is how a board ends up built on ungrilled assumptions.

A feature too vague to tier is not a licence to re-litigate it. Say so, and either ask the user for
a tier directly or point them back at the source to sharpen it.

---

## Phase 1 — Prioritise

Assign each feature a tier. Offer a suggested tier per feature — biasing dependency-blocking
features toward Now and anything speculative toward Later — and let the user confirm or override in
one pass.

**Now** — Committed, ready to start, essential for current momentum. If the user isn't working on
it this week or next, it's not Now.

**Next** — Decided but intentionally deferred. Depends on Now items finishing first, or deliberately
held back to maintain focus.

**Later** — Someday / low urgency / speculative. Worth keeping visible so nothing falls through the
cracks, but not near-term.

**Why this ordering — the prioritisation lens.** A tier is a decision, not a guess. When a feature's
tier is unclear, reason it out against four levers:

- **Value** — how much user pain it removes or upside it unlocks. High value pulls toward Now.
- **Effort** — rough size. High value + low effort is the classic Now; high effort earns Next/Later
  unless it's blocking.
- **Dependency** — does other work need this first? A blocker for several features earns Now even if
  its own value is modest; something that depends on unfinished work can't be Now.
- **Risk** — uncertainty or the cost of getting it wrong. De-risk early when a wrong call is
  expensive to unwind; defer speculative bets to Later.

State the *why* in one clause when you suggest a tier (e.g. "Now — blocks capture and auth"), so the
ordering is auditable rather than arbitrary.

Default to fewer Now items. A solo builder can realistically focus on 2–4 things at once. If
everything looks like Now, something is wrong — push back and ask.

---

## Phase 2 — Generate the board

Write a complete, self-contained HTML file to `./roadmap.html` in the current working directory.

**Always write the file.** Other skills detect this stage with `test -f ./roadmap.html`, so a board
that only ever existed as a chat artifact leaves the repo looking like no roadmap was made. Render
it as an artifact too if the interface supports it, but the file is the deliverable, not the copy.

Use `assets/roadmap-template.html` as the starting point — inline CSS, no external dependencies, no
JavaScript. Replace its placeholder comments with the actual cards and questions.

### Card format

Each card represents a **feature** and shows:

- A short noun-phrase title naming the feature (e.g. "User authentication", "Onboarding flow")
- One sentence describing the user problem it solves or the value it delivers
- A dependency line only if it blocks or is blocked by another feature on the board

Nothing else — no assignees, no tags, no dates, no story points, no implementation tasks.

### Open questions format

Below the board, list each open question from the approach doc as a numbered item: the question
concisely, plus a one-line note on why it matters or what's blocked on it.

```
Now (amber)  |  Next (blue)  |  Later (slate)
─────────────────────────────────────────────
[ Card ]       [ Card ]        [ Card ]
[ Card ]       [ Card ]
[ Card ]

Open Questions
──────────────
1. Should auth use JWT or sessions? — blocks mobile client work.
2. Is the pricing page in scope for v1?
```

---

## Finishing Up

After writing the file, a brief summary (3–5 lines max). It's a receipt, not a re-explanation —
don't repeat the roadmap verbatim.

- How many features are on the board and how they're distributed (N Now / N Next / N Later)
- How many open questions were flagged
- Which source was used, left unmodified — to change what's on the board, edit the source
  (`chat-to-approach` for the approach doc, or the `## Accepted ideas` list directly) and re-run
- The next step: pick one card and run `to-prd` on it. Each card is its own pass down the spine.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
