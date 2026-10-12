---
name: to-prd
description: >
  Turns one feature idea into a PRD built on journeys: a one-line problem, journeys (each ending in
  an observable state) and delivery slices that say what is usable after each. Interviews, explores
  the codebase, and saves to docs/prds/<feature>.md. Covers exactly one feature: given a
  multi-feature source, ask which one. Reads any approach doc or ideate report as priors so it
  doesn't re-ask what's settled, but needs none. Starts the per-feature spine
  (/to-prd -> /to-scaffold -> /to-tickets).
disable-model-invocation: true
---

## Where this sits

```
PORTFOLIO — many features        ideate  or  chat-to-approach
                                        ↓
                            ./docs/approach.md  →  ./ideas/roadmap.html
                                                        ↓  pick ONE card
FEATURE — one feature per run    to-prd → to-scaffold → to-tickets → ticket-tracker
                                 ^^^^^^ you are here
                                 non-code work: to-prd → to-tickets
```

**Portfolio → feature is a fan-out, not a step.** A roadmap card is a *feature*. Six cards means six
PRDs, never one covering a whole roadmap or approach doc.

**Never send the user backwards.** Upstream artifacts make this stage faster. A missing one is
written inline or offered, never a reason to stop and run another skill first.

## What a PRD is

Every PRD has a **core**: a one-line problem, journeys, and delivery slices. A small change is the
core and nothing else. Journeys and slices are defined in
[references/project-doc.md](references/project-doc.md); read it before writing.

Build order, tracer bullets, module boundaries and per-file design stay out: they are
`to-scaffold`'s and `to-tickets`' job.

## Priors

Before asking a question, skip it if one of these already answers it:

1. **The conversation.**
2. **On-disk priors** — `./docs/approach.md` and `./ideas/reports/*-ideate.md` (use `artifact-scan`
   to locate them if installed, else `ls`). Confirm and grill deltas only.

Priors are accelerants, not prerequisites: with none, run the full interview. They are portfolio
level; if one holds several features and the target isn't obvious, ask which one this PRD is for.

## Process

Skip steps that aren't necessary.

### 1. Get the problem in the user's words

Ask for a long, detailed description of the problem and any ideas for solutions.

### 2. Explore the repo

Verify their assertions and learn the current state. Note what already half-exists, so the journeys
build on it rather than reading as all-net-new.

### 3. Grill for scope and behaviour

Run the `grilling` skill; it owns the loop. Focus on who does what and what they see, not build order.

### 4. Write the journeys

For each way a user gets value, draft a journey: title, surface (CLI, API or UI), 2–5 steps, and a
final `→ <observable state>` line. Check the set with the user: does it cover the problem, and does
every journey end somewhere you could point at?

If a prototype exists or would help, link it; the `prototype` skill builds one.

### 5. Cut delivery slices

Ask what is the smallest release that delivers the core outcome, then what follows. Each slice says
what is usable after it, in product terms, and lists the journey numbers it delivers. Sequence by
user value, not technical dependency.

### 6. Offer the optional layer

For a bigger feature, offer to add why, goal, concepts, out of scope and open questions. Skip it for
a small change.

The format of this layer is the user's call and not fixed by this skill. If they want it, ask:
"How should the optional layer look: prose sections, a table, bullets, or something else?" Use their
answer; do not choose for them.

### 7. Save it

1. Create `docs/prds/` if missing. Write the PRD to `docs/prds/<feature>.md`, kebab-cased.
2. Show it to the user and confirm.
3. Print the file path. `to-scaffold` and `to-tickets` read it from there.

Filing a GitHub issue is opt-in, only when the user asks, and from the file so the copies don't drift:

```bash
gh issue create --title "[PRD] <Feature Name>" --body-file docs/prds/<feature>.md
```

Then hand off: `/to-scaffold` for code work; `/to-tickets` straight from the PRD for non-code work.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke
skill-feedback-collector passing this skill's name and path.
