---
name: to-prd
description: >
  Turn a feature idea into a project doc (PRD) for one feature — what we're building, how it behaves,
  and the scope boundary — through interview, codebase exploration and a phased cut-line, saved to
  ./docs/prd-<feature>.md and published to Notion. The per-feature spine (to-prd → to-plan →
  to-tickets) starts here once a feature is chosen. Covers exactly one feature: if handed a
  multi-feature source (approach doc, roadmap, ideate report), ask which feature this doc is for —
  each roadmap card is its own run down the spine. Trigger on "write a project doc", "project doc for
  X", "write a PRD", "spec this out", "scope this project", "I want to build X", "I have an idea for
  a feature", "help me scope this feature", "let's plan a new feature", or "define the requirements
  for X". Reads the Linear project brief, ./docs/approach.md and ./ideas/reports/*-ideate.md as
  priors if they exist, so it doesn't re-ask what's settled — but needs none of them, and interviews
  from nothing. Prefer over ideate when the user already knows what they want built; use
  artifact-scan if they don't know where they are in the workflow.
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
built, which then breaks down into several tickets further down the spine. Six cards means six docs,
not one; never write a single project doc covering a whole roadmap or approach doc.

**Nothing gates anything.** Upstream artifacts make this stage faster and better-informed; they never
make it required. Do not send the user backwards to create a missing artifact.

## What this doc is

A project doc describes **what we're building and how it behaves**, opinionated enough that a reader
knows why the scope is drawn where it is. Written as the shipped product's own documentation:
present tense, second person, naming the concepts the feature introduces.

**One solution, argued** — no option matrices. **Relevant technical context belongs in it** — the
concepts, columns, events, mutations, flows and diagrams a reader needs, because the scope boundary
is usually technical. **Implementation sequencing and file-level design stay out** — build order,
tracer bullets, module boundaries and per-file changes are `to-plan`'s job, and for a large project
the convention is a separate technical-proposal doc alongside this one.

Voice rules, the Scenarios format, the full template and the Notion mechanics:
[references/project-doc.md](references/project-doc.md). Read it before writing.

## Skip-if-answered precondition

Before asking any question in the steps below, treat these as priors and skip any question they
already answer — only ask what is genuinely unknown:

1. **The conversation** — scan the full conversation for an answer already provided.
2. **The Linear project brief** — if a Linear project exists for this feature, read its description.
   It answers problem, why, who (with named customers), risks and shipping sequence. Carry the named
   customers through as evidence rather than re-asking who wants this.
3. **On-disk priors** — the approach doc (`./docs/approach.md`) and the ideate report
   (`./ideas/reports/*-ideate.md`). Use `artifact-scan` as a preflight to locate them. Treat their
   captured decisions (problem, scope, audience, product decisions) as already-answered; confirm and
   grill deltas only.

These priors are **accelerants, not prerequisites** — they exist to remove questions, never to gate
this skill. If none is available, fall back to the full interview unchanged. Do not send the user
upstream to create one first.

The approach doc and ideate report are portfolio-level: they cover several features, while this doc
covers exactly one. If a prior holds more than one feature and the target isn't obvious from the
invocation, ask which feature this doc is for before going further.

## Process

You may skip steps if you don't consider them necessary.

### 1. Get the problem in the user's words

Ask for a long, detailed description of the problem they want to solve and any potential ideas for
solutions.

### 2. Explore the repo

Verify their assertions and understand the current state of the codebase. This read has a second job:
it feeds the doc's **What we already have (reuse)** section — the existing event, the existing
pipeline, the behaviour that half-exists already. A doc that reads as all-net-new usually hasn't
looked.

### 3. Grill for scope and behaviour

Run the grilling session using the `grilling` skill — it owns the grilling loop; defer to it
rather than re-inlining the questioning here. Keep the focus on scope, behaviour and the concepts
being introduced, not on build order.

### 4. Sketch the capabilities and name the concepts

Sketch out the major user-facing capabilities and behaviour that will be needed. Actively look for
opportunities to identify deep, independently-valuable capabilities — those that deliver user value
on their own.

**Decide which shape this doc is** — a *concept doc* introducing a new noun into the product, or a
*change doc* altering an existing surface. That choice decides most of the structure, and getting it
wrong is what produces a doc padded with sections that don't apply. Both shapes are set out in
[references/project-doc.md](references/project-doc.md).

If it's a concept doc, name the **concepts** it introduces and what's on each. Trying to write them up
as though they already exist is what exposes the decisions you hadn't made yet.

Ask whether a **prototype** exists or would help — five of eight project docs open with a screen
recording or a prototype link, and it is far cheaper to settle a state model or a layout in a
prototype than in review. The `prototype` skill builds one; link it near the top of the doc either
way.

Check with the user that these match their expectations.

### 5. Draw the phase cut-line

With the capabilities agreed, ask which belong in the *first phase* — the smallest release that
delivers the core outcome — versus what is deliberately deferred. Sequence by user value, not
technical dependency. Deferred items are staged, not rejected: they go in **Out of scope for Phase
1** with a one-line reason (a follow-up project link beats a promise), while things genuinely not
being built go in **Out of scope**.

Keep **Nice to haves** separate from both. Those are the details that would spark joy ✨ if there's
room — polish, not commitments.

### 6. Write the doc

Use the template in [references/project-doc.md](references/project-doc.md). It is a **menu, not a
skeleton** — no real project doc uses every section, and the shortest good one uses three. Drop what
doesn't apply rather than padding it with restated filler, and don't carry risks or a shipping
sequence into this doc: those belong to the Linear brief, and two copies drift.

### 7. Save it, then publish it

The Markdown file is the offline twin; the Notion page is where the team reads and comments. Write
the file first so the doc survives no network and this stage can be re-run without depending on the
page.

1. Create `./docs/` if it doesn't exist. Write the filled-in template to `./docs/prd-<feature>.md`,
   kebab-cased (e.g. `./docs/prd-saved-views.md`).
2. Show the rendered doc to the user and confirm before going further.
3. Create the page in the **📜 Project Docs** database with `Status: WIP`, `Lead` and `Team` — fetch
   the data source for the current options rather than guessing.
   Mechanics: [references/project-doc.md](references/project-doc.md).
4. Print both the file path and the Notion URL. `to-plan` accepts either.
5. If the Linear project brief has empty `🧩` / `🚨` / `👥` sections, **offer** to fill them from what
   the interview surfaced. Offer — never write to Linear unprompted.

If Notion is unavailable, keep the file and say the page wasn't created — the doc still stands and
`to-plan` can read the file directly. Do not treat a missing page as a failed run.

**Outside Plain** — a personal repo with no Notion home — the same doc files as a GitHub issue
instead, but only when the user asks for it:

```bash
gh issue create --title "[PRD] <Feature Name>" --body-file ./docs/prd-<feature>.md
```

File it from the Markdown file, not a re-render, or the two copies drift.

**Picking a doc back up.** Review happens in the Notion page's inline comments. Re-fetch it with
`include_discussions: true`, read the threads and fold them in — don't re-derive the doc and
overwrite what reviewers wrote.

Then hand off: `to-plan` turns this doc into the implementation plan.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke
skill-feedback-collector passing this skill's name and path.
