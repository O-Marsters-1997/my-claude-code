---
name: artifact-scan
description: >
  Scan the repo for the six product-lifecycle artifacts and report which stages are satisfied and
  where to start. Detects the ideate report, approach doc, roadmap, PRD issue, plan file, and
  tickets, then recommends the next skill to run. Use whenever the user asks "where do I start with
  this project?", "what product docs already exist?", "where are we in the pipeline?", "what should
  I run next?", "do we have a roadmap/PRD/plan yet?", or "what's left before I start coding?". Also
  invoke this as a preflight routine from other product skills that need to know which lifecycle
  artifacts are already present before acting.
---

# Artifact Scan

A fast, read-only routine. Scan the repo for the six product-lifecycle artifacts, report which
lifecycle stages are already satisfied, and point the user at a single entry point — the next skill
to run. No interview, no grilling, no writing. Just look and report.

The product lifecycle is a strict sequence. Each stage produces one artifact and feeds the next:

```
ideate  →  approach  →  roadmap  →  PRD  →  plan  →  tickets
```

## The six artifacts

| Stage    | Artifact             | How to detect it                                       | Skill that produces it |
|----------|----------------------|--------------------------------------------------------|------------------------|
| ideate   | ideate report        | `ls ./ideas/reports/*-ideate.md`                       | `ideate`               |
| approach | approach doc         | `test -f ./docs/approach.md`                           | `chat-to-approach`     |
| roadmap  | roadmap board        | `test -f ./roadmap.html`                               | `approach-to-roadmap`  |
| PRD      | GitHub `[PRD]` issue | `gh issue list` — any issue whose title starts `[PRD]` | `write-a-prd`          |
| plan     | plan file            | `ls ./plans/*.md`                                      | `prd-to-plan`          |
| tickets  | GitHub issues        | `gh issue list` — any issue NOT titled `[PRD]`         | `plan-to-issues`       |

## Procedure

1. **Scan local files** in one pass:
   ```bash
   ls ./ideas/reports/*-ideate.md ./docs/approach.md ./roadmap.html ./plans/*.md 2>/dev/null
   ```
2. **Scan GitHub issues** once, then split by title:
   ```bash
   gh issue list --state open --limit 100 --json number,title 2>/dev/null
   ```
   - PRD issue satisfied if any title starts with `[PRD]`.
   - Tickets satisfied if any open issue is *not* a `[PRD]` issue.
   - If `gh` is unavailable or errors (no remote, not authed), mark both GitHub-backed stages
     **unknown**, not satisfied — never guess. Say so in the report.
3. **Mark each stage** satisfied / not satisfied / unknown, citing the file path or issue number as
   evidence. Do not report an artifact you did not observe.
4. **Pick the entry point.** The recommended next step is the skill for the *earliest* unsatisfied
   stage — the lifecycle is sequential, so the first gap is where the user should start. If every
   stage is satisfied, the entry point is "start building the tickets".

## Report format

```
Product lifecycle scan

  [x] ideate    — ./ideas/reports/2026-05-01-ideate.md
  [x] approach  — ./docs/approach.md
  [ ] roadmap   — not found
  [ ] PRD       — no [PRD] issue found
  [?] plan      — unknown (gh unavailable)
  [ ] tickets   — no non-PRD issues found

Start here: approach → roadmap is the first gap. Run the `approach-to-roadmap` skill.
```

Use `[x]` satisfied, `[ ]` not satisfied, `[?]` unknown. Always list all six stages in lifecycle
order so the user sees the whole pipeline. Close with one line naming the recommended skill.

## Notes

- Read-only. This skill never creates or edits artifacts — it only reports and routes.
- When invoked as a preflight by another skill, return the stage map and skip the prose framing.
