---
name: approach-to-roadmap
description: Convert an existing ./docs/approach.md alignment doc into a visual HTML kanban roadmap at ./roadmap.html. No grilling — alignment is assumed to be done. Only asks light prioritisation questions to assign each feature to Now / Next / Later tiers. If ./docs/approach.md is missing, tells the user to run chat-to-approach first. Use this skill when the user wants to generate or refresh a roadmap from an existing approach doc. Trigger on phrases like "turn approach into roadmap", "generate roadmap from approach.md", "kanban from our approach doc", "prioritise the approach", "make a roadmap from the approach", "refresh the roadmap", or "approach to kanban". Also trigger when the user says "roadmap" and ./docs/approach.md exists in the repo — prefer this skill over chat-to-roadmap when alignment is already captured.
---

## Overview

Read `./docs/approach.md` — the canonical alignment doc — and produce a self-contained visual HTML kanban board at `./roadmap.html`. The approach doc contains the settled features, constraints, principles, and open questions. This skill's only job is to assign each feature to a priority tier (Now / Next / Later) and render the board.

**No grilling.** The approach doc is treated as settled truth. This skill does not invoke `grill-with-docs`, re-parse source conversations, or modify `./docs/approach.md`. Alignment was already done. If a feature in the doc looks unclear or incomplete, point the user back to `chat-to-approach` to update the doc — then re-run this skill.

---

## Phase 1 — Load the Approach Doc

Read `./docs/approach.md`.

**If the file does not exist:** stop immediately with a clear message:

> `./docs/approach.md` not found. Run `chat-to-approach` first to capture alignment, then come back here for the roadmap.

Do not proceed past this point without the file.

**If the file exists:** extract:
- All features (Name, Problem, Value, Dependencies)
- Constraints — carry forward as a note below the board, not as cards
- Principles — use to inform tier suggestions (e.g. "ship weekly" → bias toward Now)
- Open questions — carry forward verbatim into the Open Questions section of the board

---

## Phase 2 — Prioritisation Quiz

For each feature extracted from approach.md, ask the user to assign a tier. This is the only interaction in this skill — lean, targeted, no alignment re-litigation.

**Present features as a list** and let the user respond in one pass, e.g.:

> Here are the features from your approach doc. For each one, tell me: **Now** (this week / actively working on it), **Next** (decided but deferred — after Now items land), or **Later** (someday / low urgency / speculative).
>
> 1. User authentication — Secure login for access to personal data
> 2. Onboarding flow — Guide new users through setup
> 3. Payments integration — Accept subscriptions
>
> You can reply with a list: "1=Now, 2=Next, 3=Later" or just tell me in plain English.

**Default tier suggestions** — offer a suggested tier for each feature based on:
- Dependencies: a feature that blocks others → bias toward Now
- Principles in the doc (e.g. "ship weekly" → fewer Now items)
- Explicit language in the approach doc ("we're starting with X" → Now)
- Anything marked as speculative or "someday" → Later

Always offer the suggestion as a prompt, not a decision. The user confirms or overrides.

**Hard limit on Now:** default to 2–4 features max. If the user wants more than 4 in Now, push back: "A solo builder can realistically focus on 2–4 things at once. Which of these are you actually working on this week?"

**If a feature is ambiguous or incomplete in the doc:** do not re-litigate. Say: "I can't confidently tier 'Improve analytics' — it's a bit vague in the approach doc. Update it with `chat-to-approach` or tell me a tier and we'll move on."

---

## Phase 3 — Generate the HTML Kanban Board

Generate a complete, self-contained HTML file at `./roadmap.html`. No external dependencies. Deliver it as an artifact if the interface supports it, otherwise write it directly.

### Card format

Each card represents a **feature** and shows:
- A short noun-phrase title (from approach.md feature name)
- One sentence — the Problem or Value from approach.md (whichever is more informative at a glance)
- A dependency line only if it blocks or is blocked by another feature on the board

Nothing else — no assignees, no tags, no dates, no story points, no implementation tasks.

### Open questions format

Below the board, list each open question from `./docs/approach.md` as a numbered item. No fresh extraction — copy them verbatim. Add a one-line "why it matters" note if the approach doc includes one.

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

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Roadmap</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
           background: #f5f5f4; color: #1c1917; padding: 2rem; min-height: 100vh; }

    header { margin-bottom: 2rem; }
    h1 { font-size: 1.5rem; font-weight: 700; }
    .subtitle { color: #78716c; font-size: 0.875rem; margin-top: 0.25rem; }

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

    .card { background: #fafaf9; border: 1px solid #e7e5e4; border-radius: 0.5rem;
             padding: 0.75rem; margin-bottom: 0.75rem; }
    .card:last-child { margin-bottom: 0; }
    .card-title { font-weight: 600; font-size: 0.9rem; margin-bottom: 0.25rem; }
    .card-context { font-size: 0.8rem; color: #78716c; line-height: 1.4; }
    .card-dep { font-size: 0.75rem; color: #a8a29e; margin-top: 0.35rem;
                padding-top: 0.35rem; border-top: 1px solid #e7e5e4; }

    .open-questions { background: #fefce8; border: 1px solid #fde68a;
                       border-radius: 0.75rem; padding: 1.25rem; }
    .open-questions h2 { font-size: 0.8rem; font-weight: 700; letter-spacing: 0.08em;
                          text-transform: uppercase; color: #b45309; margin-bottom: 1rem; }
    .open-questions ol { padding-left: 1.25rem; }
    .open-questions li { font-size: 0.875rem; margin-bottom: 0.5rem; line-height: 1.5; }
    .open-questions .q-note { color: #92400e; font-size: 0.8rem; }

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
    <p class="subtitle">Generated from approach.md · <!-- date --></p>
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

Replace placeholder comments with actual cards and questions. No JavaScript.

---

## Finishing Up

After writing `./roadmap.html`, send a brief receipt (3–5 lines max):
- Feature distribution: N Now / N Next / N Later
- Number of open questions carried forward from approach.md
- A reminder that `./docs/approach.md` was not modified — run `chat-to-approach` to update alignment, then re-run this skill to refresh the roadmap
