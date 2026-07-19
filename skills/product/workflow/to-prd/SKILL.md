---
name: to-prd
description: >
  Turn a feature idea into a PRD for one feature — the WHAT and WHY in product language — through
  interview, codebase exploration, and an MVP cut-line, saved to ./docs/prd-<feature>.md and filed
  as a GitHub issue. The per-feature spine (to-prd → to-plan → to-tickets) starts here once a feature
  is chosen. Covers exactly one feature: if handed a multi-feature source (approach doc, roadmap,
  ideate report), ask which feature this PRD is for — each roadmap card is its own run down the
  spine. Trigger on "write a PRD", "spec this out", "I want to build X", "I have an idea for a
  feature", "help me scope this feature", "let's plan a new feature", or "define the requirements
  for X". Reads ./docs/approach.md and ./ideas/reports/*-ideate.md as priors if they exist, so it
  doesn't re-ask what's settled — but needs neither, and interviews from nothing. Prefer over ideate
  when the user already knows what they want built; use artifact-scan if they don't know where they
  are in the workflow.
---

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-plan → to-tickets → ticket-tracker
                                 ^^^^^^ you are here
```

**Portfolio → feature is a fan-out, not a step.** A roadmap card is a *feature* — something to be
built, which then breaks down into several tickets further down the spine. Six cards means six PRDs,
not one; never write a single PRD covering a whole roadmap or approach doc.

**Nothing gates anything.** Upstream artifacts make this stage faster and better-informed; they never
make it required. Do not send the user backwards to create a missing artifact.

**PRD principle: product language only.**
A PRD describes WHAT and WHY — outcomes, behaviour, and requirements in product language. It must NOT contain technical design: no code or pseudo-code, no schema DDL, no class/function/API signatures, no file paths, no library/framework choices. All technical design belongs in the plan (the technical design document produced by to-plan). If a technical detail feels important, translate it into the product outcome it serves, or note it as a "Further Notes" item to be resolved during planning.

**Skip-if-answered precondition.**
Before asking any question in the steps below, treat two sources as priors and skip any question they already answer — only ask what is genuinely unknown:

1. **The conversation** — scan the full conversation for an answer already provided.
2. **On-disk priors** — load the portfolio artifacts if present: the approach doc (`./docs/approach.md`) and the ideate report (`./ideas/reports/*-ideate.md`). Use `artifact-scan` as a preflight to locate them. Read whichever exist and treat their captured decisions (problem, scope, audience, product decisions) as already-answered; do not re-interview on them, only confirm and grill deltas.

These priors are **accelerants, not prerequisites** — they exist to remove questions, never to gate this skill. If neither is on disk, fall back to the full interview unchanged. Do not send the user upstream to create one first.

Both priors are portfolio-level: they cover several features, while this PRD covers exactly one. If the prior holds more than one feature and the target isn't obvious from the invocation, ask which feature this PRD is for before going further.

## Process

You may skip steps if you don't consider them necessary.

### 1. Get the problem in the user's words

Ask for a long, detailed description of the problem they want to solve and any potential ideas for solutions.

### 2. Explore the repo

Verify their assertions and understand the current state of the codebase.

### 3. Grill for scope and requirements

Run the grilling session using the `grill-with-docs` skill — it owns the grilling loop; defer to it rather than re-inlining the questioning here. Keep the focus on product-level scope and behaviour (per the PRD principle above), not technical design.

### 4. Sketch the user-facing capabilities

Sketch out the major user-facing capabilities and behaviour that will be needed. Actively look for opportunities to identify deep, independently-valuable capabilities — those that deliver user value on their own and can be verified against the user stories.

Check with the user that these capabilities match their expectations.

### 5. Draw the MVP cut-line

With the capabilities agreed, ask the user which of them belong in the *first cut* — the smallest release that delivers the core outcome — versus what is deliberately deferred to a later pass. Sequence by user value, not technical dependency, and keep it in product language (no build phases, no architecture). Deferred items are staged, not rejected. Record the split in the PRD's **MVP Cut-Line** section (deferred-but-planned items) and in **Out of Scope** (things genuinely not being built).

### 6. Write the PRD

Once you have a complete understanding of the problem and solution, use the template below to write the PRD.

### 7. Save it, then file it

The Markdown file is the artifact; the GitHub issue is a copy of it for discussion. Write the file first so the PRD survives a repo with no remote, no `gh`, or no network — and so this stage can be re-run and iterated on without depending on the issue.

1. Create `./docs/` if it doesn't exist. Write the filled-in template to `./docs/prd-<feature>.md`, kebab-cased (e.g. `./docs/prd-saved-views.md`).
2. Show the rendered PRD to the user and confirm before going further.
3. File the issue from that file — do **not** re-render it into a temp file, or the two copies can drift:
   ```bash
   gh issue create --title "[PRD] <Feature Name>" --body-file ./docs/prd-<feature>.md
   ```
   Add `--label` flags if relevant labels exist in the repo.
4. Print both the file path and the issue URL. `to-plan` accepts either.

If `gh` is unavailable or fails, keep the file and say the issue wasn't created — the PRD still stands and `to-plan` can read the file directly. Do not treat a missing issue as a failed run.

When finished, hand off: `to-plan` turns this PRD into the technical design.

<prd-template>

## Problem Statement

The problem that the user is facing, from the user's perspective.

## Solution

The solution to the problem, from the user's perspective.

## User Stories

A LONG, numbered list of user stories. Each user story should be in the format of:

1. As an <actor>, I want a <feature>, so that <benefit>

<user-story-example>
1. As a mobile bank customer, I want to see balance on my accounts, so that I can make better informed decisions about my spending
</user-story-example>

This list of user stories should be extremely extensive and cover all aspects of the feature.

## Product Decisions

A list of product-level decisions that were made. This covers scope and behaviour choices, not technical implementation. Include:

- Scope boundaries: what is and is not included
- Behavioural choices: how the feature should act from a user's perspective
- Non-functional requirements stated as user-visible outcomes (e.g. "the search should return results within 2 seconds", not "use Redis")
- UX / interaction decisions
- Specific user-facing interactions or flows

Do NOT include schema changes, API contracts, architectural decisions, library choices, or any other technical design. Those are resolved during planning (to-plan), not here.

## Testing Decisions

A list of testing decisions that were made. Include:

- A description of what makes a good test (only test external behavior, not implementation details)
- Which user-facing behaviours will be tested
- Prior art for the tests (i.e. similar types of tests in the codebase)

## MVP Cut-Line

The first-cut boundary in product language: the capabilities that ship in the MVP versus those deliberately deferred to a later pass.

- **In the first cut:** the smallest set of capabilities that delivers the core user outcome.
- **Deferred (planned, not now):** capabilities that are agreed but intentionally held back to keep the first cut small — each with a one-line reason (e.g. "adds value but not needed to prove the core flow").

Deferred items are staged for a later release, not rejected — anything genuinely not being built belongs in Out of Scope below.

## Out of Scope

A description of the things that are out of scope for this PRD.

## Further Notes

Any further notes about the feature. Technical details that surfaced during the interview but belong in the plan can be listed here as "to be resolved during planning."

</prd-template>

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
