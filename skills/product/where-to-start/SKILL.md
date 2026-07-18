---
name: where-to-start
description: >
  Front door to the product-lifecycle skill suite — runs the artifact scan, then tells the user
  where to enter the lifecycle and which single skill to reach for next. Use whenever the user asks
  "where do I start with this project?", "which skill should I use?", "what should I run next?",
  "I don't know where to begin", "how do I get going here?", or is staring at a repo unsure which
  product skill applies. Trigger even if the user names no skill — if the question is *which skill
  do I run and where do I jump in*, this is the router that answers it.
---

# Where to Start

The single front door to the product suite. The user knows they want to move the product forward
but not *which* skill does it or *where* they already are. This skill answers both in one shot:
run the scan, read the result, point at exactly one next step.

It does no product work itself — no ideation, no writing, no grilling. It only routes.

## Procedure

1. **Run the artifact scan.** Invoke the `artifact-scan` skill against the current repo. It detects
   the six lifecycle artifacts and returns which stages are satisfied, not satisfied, or unknown.
   Do **not** reimplement the detection here — the scan owns that logic; this skill consumes its
   result. If you need the detection details, read `artifact-scan`'s SKILL.md.

2. **Find the earliest unsatisfied stage.** The lifecycle is strictly sequential, so the first gap
   is where the user should enter:

   ```
   ideate  →  approach  →  roadmap  →  PRD  →  plan  →  tickets
   ```

3. **Map that stage to the skill that produces it:**

   | Earliest gap | Enter here → run this skill              |
   |--------------|------------------------------------------|
   | ideate       | `ideate`                                 |
   | approach     | `chat-to-approach`                       |
   | roadmap      | `to-roadmap`                             |
   | PRD          | `to-prd`                                 |
   | plan         | `to-plan`                                |
   | tickets      | `to-tickets`                             |

   `to-roadmap` auto-detects its source (an existing approach doc vs a pasted conversation), and
   `to-tickets` auto-detects its source (a plan file vs a PRD issue vs a conversation) — so there
   is a single skill to route to for each stage.

4. **If every stage is satisfied**, there is nothing left to route — tell the user the pipeline is
   complete and the next step is to start building the open tickets.

5. **If a stage is unknown** (e.g. `gh` unavailable, so PRD/tickets can't be checked), route to the
   earliest *definitely* unsatisfied stage and flag the unknown ones so the user can re-check.

## Report format

Show the full scan so the user sees the whole pipeline, then close with one recommendation:

```
Product lifecycle scan

  [x] ideate    — ./ideas/reports/2026-05-01-ideate.md
  [x] approach  — ./docs/approach.md
  [ ] roadmap   — not found
  [ ] PRD       — no [PRD] issue found
  [ ] plan      — no plan file found
  [ ] tickets   — no non-PRD issues found

Start here: roadmap is your first gap. Run the `to-roadmap` skill
(you already have an approach doc, so alignment is done).
```

Recommend exactly one skill. If the user wants the full lifecycle picture without a routing verdict,
that's the `artifact-scan` skill directly — this one always ends with a single next step.
