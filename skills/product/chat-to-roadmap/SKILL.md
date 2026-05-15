---
name: chat-to-roadmap
description: Parse a pasted AI conversation summary or transcript into a lightweight product roadmap. Extracts only agreed-upon decisions and committed action items — ignoring discarded ideas and unresolved tangents — then organises them into Now / Next / Later priority tiers rendered as a visual kanban HTML board, with open questions flagged separately at the bottom. Solo/indie-builder scale — no enterprise PM overhead. Use this skill whenever the user pastes a conversation and wants a roadmap, action items, or a kanban board extracted from it. Trigger on phrases like "turn this into a roadmap", "what did we decide?", "extract actions from this chat", "make me a kanban", "pull the decisions from this", "what should I build next based on this conversation", or "roadmap from this transcript". Use it even if the user just drops a wall of AI conversation text without explicit instructions — if they seem to want structure from a chat, this skill applies.
---

## Overview

Turn a messy AI conversation into a clean, visual product roadmap. The goal is to extract *signal* — only things that were genuinely decided and committed to — and discard the noise (explored ideas, dead ends, unresolved discussions). The output is a standalone HTML kanban board with three columns (Now / Next / Later) and an open questions section at the bottom.

Solo/indie-builder scale throughout: no assignees, no story points, no sprints, no stakeholder matrices.

---

## Phase 1 — Parse the Conversation

Read the pasted transcript or summary carefully. Classify everything into four buckets:

**Decided/committed** — Both parties agreed and committed to doing something. Look for language like "we'll do X", "let's go with Y", "agreed: Z", "the plan is", "I'll build", "we'll ship". These become roadmap items.

**Explored but discarded** — Ideas considered and then rejected or deprioritised in favour of something else. Do not include these in the roadmap.

**Unresolved tangents** — Topics that got discussed but trailed off without a conclusion. Treat as open questions unless there's a clear decision.

**Open questions** — Things that need a decision before work can proceed. Surface these at the bottom of the board.

Be conservative: when in doubt whether something was decided, treat it as an open question rather than a committed item. A roadmap full of things that were only half-agreed is worse than a short, honest one.

---

## Phase 2 — Alignment Grilling (Safeguard)

Before generating the roadmap, run a focused grilling session inspired by the `grill-with-docs` pattern. This is the safeguard: you are stress-testing your extraction before committing it to a visual artifact. Don't skip this phase.

Scan your extracted items for:
- Items where it's genuinely unclear if a decision was reached vs. just discussed
- Items with ambiguous priority (could reasonably be Now or Next)
- Items that seem to contradict each other — were both agreed, or did one supersede the other?
- Items with fuzzy scope — "improve X" or "think about Y" without any concrete shape

For each ambiguous item, ask one targeted clarifying question. Provide your recommended interpretation and wait for the user's confirmation before moving on. One question at a time.

**How to ask good grilling questions:**
- Lead with what you found in the text: "You discussed two auth approaches — JWT and sessions — but I couldn't tell which was chosen."
- Give your best read: "My read is you landed on JWT because of the mobile client. Is that right?"
- Stay concrete: don't ask "what's your priority?" — ask "Should 'wire up payments' be Now (this week) or Next (after auth is solid)?"

Keep this lean. Only ask if the answer would meaningfully change what ends up on the board. If the conversation was explicit and decisive, tell the user and skip straight to Phase 3.

**Examples of when to grill vs. skip:**
- Grill: "You mentioned analytics twice but the second mention seemed to dismiss it — was that a backlog item or a cut?"
- Skip: "You said 'we'll ship the MVP by end of month' — I'll put the MVP items in Now."
- Grill: "Onboarding was described as 'important' but no concrete action was attached — what's the task?"
- Skip: "You agreed to use Postgres — I'll include that as a decided constraint, not a roadmap item."

---

## Phase 3 — Prioritise

Once decisions are confirmed, assign each item to a tier:

**Now** — Committed, ready to start, essential for current momentum. If the user isn't working on it this week or next, it's not Now.

**Next** — Decided but intentionally deferred. Depends on Now items finishing first, or deliberately held back to maintain focus.

**Later** — Someday / low urgency / speculative. Worth keeping visible so nothing falls through the cracks, but not near-term.

Default to fewer Now items. A solo builder can realistically focus on 2–4 things at once. If everything looks like Now, something is wrong — push back and ask.

---

## Phase 4 — Generate the HTML Kanban Board

Generate a complete, self-contained HTML file. Deliver it as an artifact (using `<artifact>` tags if the interface supports it) or write it to `roadmap.html` in the current working directory.

### Card format

Each card shows:
- A short imperative title (e.g. "Ship auth", "Wire up Stripe", "Validate landing page copy")
- One sentence of context extracted from the conversation — enough to remember why it was decided
- Nothing else — no assignees, no tags, no dates, no story points

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
      <!-- <div class="card"><div class="card-title">Title</div><div class="card-context">Context from chat.</div></div> -->
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
- How many items are on the board and how they're distributed (N Now / N Next / N Later)
- How many open questions were flagged
- A one-liner on any items you deliberately excluded (discarded ideas, unresolved tangents) — so the user knows you didn't miss them

Don't repeat the roadmap verbatim. The summary is a receipt, not a re-explanation.
