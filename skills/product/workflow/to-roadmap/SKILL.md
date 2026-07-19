---
name: to-roadmap
description: >
  Prioritise the features in ./docs/approach.md into a Now / Next / Later kanban board, rendered as
  a self-contained ./roadmap.html. Each card is a feature, not a ticket. The approach doc is settled
  truth — this skill assigns tiers and reasons about ordering, it never re-litigates what was
  decided. Use whenever the user wants to prioritise, sequence, or visualise what to build: "make me
  a roadmap", "turn the approach into a roadmap", "generate a kanban", "what should I build first",
  "prioritise these features", "refresh the roadmap", or "what's Now vs Later". Requires
  ./docs/approach.md — if it is missing, route to chat-to-approach first rather than interviewing
  here.
---

## Overview

Turn the settled features in `./docs/approach.md` into a visual product roadmap: a standalone HTML
kanban with three columns (Now / Next / Later) where each card is a **feature**, not a ticket.
`to-tickets` breaks features into tickets later.

This skill is **portfolio-level** — it prioritises many features at once. The per-feature spine
(`to-prd → to-plan → to-tickets`) starts from a single card. See `../README.md`.

Solo/indie-builder scale throughout: no assignees, no story points, no sprints, no stakeholder
matrices.

---

## Phase 0 — Load the approach doc

Run `artifact-scan` as a preflight to detect `./docs/approach.md` — defer to it, don't reimplement
its file checks.

**If `./docs/approach.md` exists**, read it and extract, treating every item as already agreed:

- All features (Name, Problem, Value, Dependencies) → cards
- Constraints → a note below the board, not cards
- Principles → use to inform tier suggestions (e.g. "ship weekly" → fewer Now items)
- Open questions → carry forward verbatim into the Open Questions section

Do not modify `./docs/approach.md`.

**If it does not exist**, stop and route — do not interview, and do not build a roadmap from a
conversation:

> "There's no `./docs/approach.md` yet. Run `chat-to-approach` first to capture what you're
> building — paste the conversation and it'll write the alignment doc. Then re-run this and I'll
> prioritise it."

Offer to invoke `chat-to-approach` there and then. **Why route instead of absorbing the
conversation:** `chat-to-approach` owns conversation → structure, including the grilling that
resolves ambiguity before anything is committed. Duplicating that here means two skills parsing
transcripts differently and drifting apart, and a roadmap built on ungrilled assumptions. One source
of alignment truth, one skill that produces it.

If the user pasted a conversation *and* an approach doc exists, the doc wins — point them at
`chat-to-approach` to fold the new material into alignment first, then re-run this.

A feature too vague to tier is not a licence to re-litigate it. Say so, and either ask the user for
a tier directly or point them back to `chat-to-approach` to sharpen the doc.

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
- The source was `./docs/approach.md`, left unmodified — to change what's on the board, run
  `chat-to-approach` to update alignment, then re-run this to refresh
- The next step: pick one card and run `to-prd` on it. Each card is its own pass down the spine.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
