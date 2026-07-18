---
name: write-a-prd
description: Create a PRD through user interview, codebase exploration, and module design, then submit as a GitHub issue. Use when user wants to write a PRD, create a product requirements document, or plan a new feature.
---

**PRD principle: product language only.**
A PRD describes WHAT and WHY — outcomes, behaviour, and requirements in product language. It must NOT contain technical design: no code or pseudo-code, no schema DDL, no class/function/API signatures, no file paths, no library/framework choices. All technical design belongs in the plan (the technical design document produced by prd-to-plan). If a technical detail feels important, translate it into the product outcome it serves, or note it as a "Further Notes" item to be resolved during planning.

**Skip-if-answered precondition.**
Before asking any question in the steps below, treat two sources as priors and skip any question they already answer — only ask what is genuinely unknown:

1. **The conversation** — scan the full conversation for an answer already provided.
2. **On-disk priors** — load the earlier product-lifecycle artifacts if present: the approach doc (`./docs/approach.md`) and the ideate report (`./ideas/reports/*-ideate.md`). These are the upstream artifacts the `artifact-scan` skill locates — defer to it for how the lifecycle fits together rather than re-deriving it here. Read whichever exist and treat their captured decisions (problem, scope, audience, product decisions) as already-answered; do not re-interview on them, only confirm and grill deltas.

If neither file is on disk, fall back to the full interview unchanged.

This skill will be invoked when the user wants to create a PRD. You may skip steps if you don't consider them necessary.

1. Ask the user for a long, detailed description of the problem they want to solve and any potential ideas for solutions.

2. Explore the repo to verify their assertions and understand the current state of the codebase.

3. Grill the user to nail down the PRD's scope and requirements. Run the grilling session using the `grill-with-docs` skill — it owns the grilling loop; defer to it rather than re-inlining the questioning here. Keep the focus on product-level scope and behaviour (per the PRD principle above), not technical design.

4. Sketch out the major user-facing capabilities and behaviour that will be needed. Actively look for opportunities to identify deep, independently-valuable capabilities — those that deliver user value on their own and can be verified against the user stories.

Check with the user that these capabilities match their expectations.

5. Draw the MVP cut-line. With the capabilities agreed, ask the user which of them belong in the *first cut* — the smallest release that delivers the core outcome — versus what is deliberately deferred to a later pass. Sequence by user value, not technical dependency, and keep it in product language (no build phases, no architecture). Deferred items are staged, not rejected. Record the split in the PRD's **MVP Cut-Line** section (deferred-but-planned items) and in **Out of Scope** (things genuinely not being built).

6. Once you have a complete understanding of the problem and solution, use the template below to write the PRD.

7. Show the rendered PRD to the user and confirm before creating the issue. Then create the GitHub issue:
   - Title: `[PRD] <Feature Name>`
   - Body: the filled-in PRD template (write to a temp file and use `gh issue create --title "[PRD] <Feature Name>" --body-file <tmpfile>`)
   - Optionally add `--label` flags if relevant labels exist in the repo
   - Print the resulting issue URL — the next step (prd-to-plan or prd-to-issues) will reference it

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

Do NOT include schema changes, API contracts, architectural decisions, library choices, or any other technical design. Those are resolved during planning (prd-to-plan), not here.

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
