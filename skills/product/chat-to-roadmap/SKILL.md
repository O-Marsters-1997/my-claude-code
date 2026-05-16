---
name: chat-to-roadmap
description: Parse a pasted AI conversation summary or transcript into a lightweight product roadmap. Extracts agreed-upon features and surfaces ambiguous items for clarification — only explicitly rejected ideas are silently dropped. Organises output into Now / Next / Later priority tiers rendered as a visual kanban HTML board where each card represents a feature (not a ticket), with open questions flagged separately at the bottom. Solo/indie-builder scale — no enterprise PM overhead. Use this skill whenever the user pastes a conversation and wants a roadmap, action items, or a kanban board extracted from it. Trigger on phrases like "turn this into a roadmap", "what did we decide?", "extract actions from this chat", "make me a kanban", "pull the decisions from this", "what should I build next based on this conversation", or "roadmap from this transcript". Use it even if the user just drops a wall of AI conversation text without explicit instructions — if they seem to want structure from a chat, this skill applies.
---

## Overview

Turn a messy AI conversation into a clean, visual product roadmap. The goal is to extract *signal* — things that were genuinely decided, committed to, or worth clarifying — and only discard noise that was explicitly rejected in the conversation. The output is a standalone HTML kanban board with three columns (Now / Next / Later) where each card represents a **feature**, not a ticket. A separate `prd-to-issues` skill handles breaking features into tickets later.

Solo/indie-builder scale throughout: no assignees, no story points, no sprints, no stakeholder matrices.

---

## What Is a Feature (at Roadmap Level)?

A feature is a coherent, user-facing capability that delivers value on its own. It is not an implementation task, a ticket, or a technical decision.

**Anatomy of a roadmap-level feature:**

| Field | What it captures | Example |
|---|---|---|
| **Name** | Short, user-facing label — what it does, not how | "User authentication" |
| **Problem** | The user pain or need it addresses (one sentence) | "Users need to log in securely to access their data" |
| **Value** | What it unlocks or enables | "Unblocks all personalisation and data persistence features" |
| **Dependencies** | Other features that must exist first | "Requires: database setup" |

At roadmap level, only Name + Problem are required on the card. Value and Dependencies are surfaced when they matter for prioritisation.

**How to group chat content into a feature:**

- A feature = a cluster of related decisions that together describe one user-facing capability
- Multiple small decisions about the same area → one feature, not many cards
- Ask: "Could a user describe what they'd get from this in one sentence?" If yes, it's a feature
- Distinguish features from:
  - *Constraints* — architectural or technical decisions that shape features but aren't features themselves (e.g. "use Postgres") — record as a note, not a card
  - *Tasks* — implementation steps inside a feature (e.g. "wire up the ORM") — don't put these on the kanban; `prd-to-issues` handles this later
  - *Principles* — values or preferences the user expressed ("keep it simple", "mobile-first") — capture as context, not a card

**Anti-patterns to avoid:**

- Don't split one feature into multiple cards because the conversation touched it in different ways
- Don't create a card for every distinct sentence — synthesise first
- Don't mistake a technical decision for a feature ("use JWT" is a constraint; "user sessions" is a feature)

---

## Phase 1 — Parse the Conversation

Read the pasted transcript or summary carefully. Classify everything into four buckets:

**Decided/committed** — Both parties agreed and committed to something. Look for language like "we'll do X", "let's go with Y", "agreed: Z", "the plan is", "I'll build", "we'll ship". These become roadmap features.

**Discussed but not resolved** — Ideas that came up, were explored, but didn't reach a clear conclusion. Do not silently drop these. Bring them to Phase 2 for grilling. **An idea not explicitly accepted is not the same as an idea that was rejected.**

**Explicitly discarded** — Ideas the user or conversation actively rejected or dismissed in favour of something else. Look for language like "no, let's not", "actually scratch that", "we're not doing X", "that's out of scope". Only these get dropped without grilling.

**Open questions** — Things that need a decision before work can proceed. Surface at the bottom of the board.

**Default to inclusion over exclusion.** When in doubt about whether something was decided, treat it as either a candidate for the board (pending grilling) or an open question. A roadmap that surfaces ambiguity honestly is better than one that silently drops things the user cared about.

---

## Phase 2 — Alignment Grilling (Safeguard)

Before generating the roadmap, run a focused grilling session using the `grill-with-docs` skill. This is the safeguard: stress-test your extraction before committing it to a visual artifact. Don't skip this phase. `grill-with-docs` owns the resolution of ambiguity — defer to it rather than making a judgement call yourself.

Scan your extracted items for:
- Items in the "discussed but not resolved" bucket — were these dropped intentionally, or just deferred?
- Items where it's unclear if a decision was reached vs. just explored
- Items with ambiguous priority (could reasonably be Now or Next)
- Items that seem to contradict each other — were both agreed, or did one supersede the other?
- Items with fuzzy scope — "improve X" or "think about Y" — is there a concrete feature here?
- Items that could be the same feature described differently at different points in the conversation

For each ambiguous item, ask one targeted clarifying question. Lead with what you found in the text, give your best read, and wait for the user's confirmation. One question at a time.

**How to ask good grilling questions:**
- Lead with what you found: "You discussed two auth approaches — JWT and sessions — but I couldn't tell which was chosen."
- Give your best read: "My read is you landed on JWT because of the mobile client. Is that right?"
- For discussed-but-not-resolved items: "You mentioned analytics a couple of times but didn't land on a decision — was this cut, deferred, or still live?"
- Stay concrete: don't ask "what's your priority?" — ask "Should 'onboarding flow' be Now (this week) or Next (after auth is solid)?"

**Examples of when to grill vs. skip:**
- Grill: "You mentioned a dashboard twice but never committed to building it — is this a Later item or cut entirely?"
- Grill: "Onboarding was described as 'important' but no concrete action was attached — is there a feature here or just a principle?"
- Grill: "You considered two pricing models but I couldn't tell which you chose — did you land on one?"
- Skip: "You said 'we'll ship the MVP by end of month' — I'll put the MVP features in Now."
- Skip: "You agreed to use Postgres — I'll note this as a constraint, not a feature card."

Keep this lean. Only ask if the answer would meaningfully change what ends up on the board. If the conversation was explicit and decisive throughout, tell the user and skip straight to Phase 3.

---

## Phase 3 — Prioritise

Once features are confirmed, assign each to a tier:

**Now** — Committed, ready to start, essential for current momentum. If the user isn't working on it this week or next, it's not Now.

**Next** — Decided but intentionally deferred. Depends on Now items finishing first, or deliberately held back to maintain focus.

**Later** — Someday / low urgency / speculative. Worth keeping visible so nothing falls through the cracks, but not near-term.

Default to fewer Now items. A solo builder can realistically focus on 2–4 things at once. If everything looks like Now, something is wrong — push back and ask.

---

## Phase 4 — Generate the HTML Kanban Board

Generate a complete, self-contained HTML file. Deliver it as an artifact (using `<artifact>` tags if the interface supports it) or write it to `roadmap.html` in the current working directory.

### Card format

Each card represents a **feature** and shows:
- A short noun-phrase title naming the feature (e.g. "User authentication", "Payments integration", "Onboarding flow")
- One sentence describing the user problem it solves or the value it delivers — extracted from the conversation
- A dependency line only if it blocks or is blocked by another feature on the board

Nothing else — no assignees, no tags, no dates, no story points, no implementation tasks.

### Open questions format

Below the board, list each open question as a numbered item:
- State the question concisely
- Add a one-line note on why it matters or what's blocked on it

### Visual design

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

### HTML template

Write inline CSS only — no external dependencies. Use this structure as your starting point:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Roadmap</title>
  <style>
    /* Reset + base */
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
           background: #f5f5f4; color: #1c1917; padding: 2rem; min-height: 100vh; }

    /* Header */
    header { margin-bottom: 2rem; }
    h1 { font-size: 1.5rem; font-weight: 700; }
    .subtitle { color: #78716c; font-size: 0.875rem; margin-top: 0.25rem; }

    /* Board */
    .board { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1.25rem; margin-bottom: 2.5rem; }
    @media (max-width: 640px) { .board { grid-template-columns: 1fr; } }

    .column { background: #fff; border-radius: 0.75rem; padding: 1.25rem;
               box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
    .column h2 { font-size: 0.8rem; font-weight: 700; letter-spacing: 0.08em;
                  text-transform: uppercase; margin-bottom: 1rem; padding-bottom: 0.5rem;
                  border-bottom: 2px solid; }
    .column.now   h2 { color: #d97706; border-color: #fbbf24; }
    .column.next  h2 { color: #2563eb; border-color: #93c5fd; }
    .column.later h2 { color: #64748b; border-color: #cbd5e1; }

    /* Cards */
    .card { background: #fafaf9; border: 1px solid #e7e5e4; border-radius: 0.5rem;
             padding: 0.75rem; margin-bottom: 0.75rem; }
    .card:last-child { margin-bottom: 0; }
    .card-title { font-weight: 600; font-size: 0.9rem; margin-bottom: 0.25rem; }
    .card-context { font-size: 0.8rem; color: #78716c; line-height: 1.4; }
    .card-dep { font-size: 0.75rem; color: #a8a29e; margin-top: 0.35rem;
                padding-top: 0.35rem; border-top: 1px solid #e7e5e4; }

    /* Open questions */
    .open-questions { background: #fefce8; border: 1px solid #fde68a;
                       border-radius: 0.75rem; padding: 1.25rem; }
    .open-questions h2 { font-size: 0.8rem; font-weight: 700; letter-spacing: 0.08em;
                          text-transform: uppercase; color: #b45309; margin-bottom: 1rem; }
    .open-questions ol { padding-left: 1.25rem; }
    .open-questions li { font-size: 0.875rem; margin-bottom: 0.5rem; line-height: 1.5; }
    .open-questions .q-note { color: #92400e; font-size: 0.8rem; }

    /* Dark mode */
    @media (prefers-color-scheme: dark) {
      body { background: #1c1917; color: #fafaf9; }
      .column { background: #292524; box-shadow: 0 1px 3px rgba(0,0,0,0.3); }
      .card { background: #1c1917; border-color: #44403c; }
      .card-context { color: #a8a29e; }
      .card-dep { color: #78716c; border-top-color: #44403c; }
      .open-questions { background: #1c1409; border-color: #78350f; }
    }
  </style>
</head>
<body>
  <header>
    <h1>Roadmap</h1>
    <p class="subtitle">Extracted from conversation · <!-- date --></p>
  </header>

  <div class="board">
    <div class="column now">
      <h2>Now</h2>
      <!-- <div class="card">
             <div class="card-title">Feature name</div>
             <div class="card-context">Problem or value sentence.</div>
             <div class="card-dep">Needs: other feature</div>
           </div> -->
    </div>
    <div class="column next">
      <h2>Next</h2>
    </div>
    <div class="column later">
      <h2>Later</h2>
    </div>
  </div>

  <section class="open-questions">
    <h2>Open Questions</h2>
    <ol>
      <!-- <li>Question text? <span class="q-note">— Why it matters.</span></li> -->
    </ol>
  </section>
</body>
</html>
```

Replace the placeholder comments with the actual cards and questions. Do not include any JavaScript.

---

## Finishing Up

After delivering the HTML, write a brief summary (3–5 lines max):
- How many features are on the board and how they're distributed (N Now / N Next / N Later)
- How many open questions were flagged
- A one-liner on anything explicitly excluded (only items the conversation actively rejected) — so the user knows you didn't silently drop things they cared about

Don't repeat the roadmap verbatim. The summary is a receipt, not a re-explanation.
