---
name: grilling
description: Grill the user relentlessly about a plan, decision, or idea. Use when the user wants to stress-test their thinking, or uses any 'grill' trigger phrases.
---

Interview the user relentlessly until you reach a shared understanding. Map this as a **design tree**: every decision branches into the decisions that hang off it.

The **frontier** is every decision whose prerequisites are already settled: the questions you can ask _now_ without guessing at answers you haven't heard yet. Work the frontier one question at a time via `AskUserQuestion` — never batch a round as text, since each answer can reshape what belongs in the frontier next.

For each question:

- Give exactly 3 concrete options grounded in your analysis, plus a fourth: "Chat about this", for when the user isn't ready to decide.
- Put your recommended option first, with `(Recommended)` appended to its label. Always have a recommendation — never ask a question you have no opinion on.
- `AskUserQuestion` adds a free-form "Other" option automatically — don't add one yourself.
- If "Chat about this" is picked, drop into a normal exchange: ask what needs clarifying, discuss it, then re-ask the question — revised if the discussion changed it — once that's settled.

Finding _facts_ is your job, never the user's. When a frontier question needs a fact from the environment (filesystem, tools, etc.), dispatch a sub-agent to find it; don't ask the user for anything you could look up yourself. Don't block on it: a running exploration is an unsettled prerequisite, so ask the rest of the frontier now and come back to the blocked question once the sub-agent reports. The _decisions_ are the user's: put each to them and wait.

The session is done when the frontier is empty: every branch of the design tree visited, nothing left silently assumed. Do not act on it until the user confirms you have reached a shared understanding.
