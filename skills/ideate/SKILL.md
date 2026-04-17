---
name: ideate
description: >
  Surface product opportunities for this codebase — new features, competitive gaps, strategic bets —
  grounded in what the code actually does and calibrated against the real market. Use whenever the
  user asks "what should I build next?", "any ideas for this project?", "where should this go from
  here?", "what could I add?", "what are competitors doing?", "how do I grow this?", "help me find
  gaps", or "what's missing?". Also trigger for roadmap ideation, product discovery, opportunity
  analysis, or competitive positioning. Trigger even if the user doesn't say "ideate" — if the
  question is *what to build next in this project*, this skill applies.
---

# Ideate

Help the user discover what to build next in *this* codebase: a set of high-conviction product
ideas, grounded in the code's actual shape and calibrated against the real market.

## Operating principles

- **Delight is the currency** — what would make someone say "whoa"?
- **Ship something you can show people.** The best version of anything is the one that exists.
- **The best side projects solve your own problem.** If the user is building for themselves,
  trust that instinct and sharpen it.
- **Explore before you optimise.** Try the weird idea first. Polish later.
- **End with concrete build steps**, not business validation tasks. The deliverable is *what to
  build next*, not *who to interview*.

## Response posture

Be an enthusiastic, opinionated collaborator. Riff on the user's ideas. Bring adjacent ideas,
unexpected combinations, "what if you also…" suggestions. Help them find the most *exciting*
version of their idea — don't settle for the obvious version.

## How ideas go wrong

Generic SaaS ideation — "add AI analytics", "build a plugin marketplace" — is the failure mode.
The filter is: *why is this codebase specifically positioned to win at this idea?* If that answer
isn't concrete and grounded in observed code, the idea gets dropped.

---

## Phase 0 — Load canonical context

Before anything else, check whether `./ideas/CONTEXT.md` exists in the target repo.

**If it exists:**
1. Read it in full.
2. Tell the user in two lines: *"Loaded ideation context (last updated [date]). My current
   read: [one-sentence positioning from CONTEXT.md]. [N] ideas implemented, [M] proposed."*
3. Ask via `AskUserQuestion`: *"Does this still look right, or anything to correct before we
   continue?"* — update your working model with any corrections before proceeding.
4. With CONTEXT.md loaded:
   - **Phase 1** skips re-derivation of characterisation — use the `## Positioning` section
     as the starting point, verify it isn't contradicted by very recent commits, note deltas.
   - **Phase 3** uses CONTEXT.md's `## Competitive landscape` as the starting point unless
     the user requests a refresh or `last_updated` is more than 6 months ago — if stale, do
     a targeted refresh of only what looks outdated.
   - **Phase 5** dedupes against `## Implemented ideas` and `## Proposed ideas (pending)` —
     do not re-propose these. You may iterate on a pending idea (extend, split, combine) but
     name that explicitly.

**If it does not exist**, proceed to Phase 1 as normal. Phase 7 handles the bootstrap offer.

---

## Phase 1 — Explore the codebase

Build a real understanding of what this tool is, who it's for, and what it's trying to do.

Gather context in roughly this order:

1. **Docs first.** Read `README.md`, `CLAUDE.md`, `TODOS.md`, and any top-level `docs/` if they
   exist. These encode intent.
2. **Recent history.** `git log --oneline -30` and `git diff origin/main --stat` to see where
   work is currently flowing and what's in progress.
3. **Prior ideation.** Surface any existing design / ideas / RFC docs so you don't duplicate
   ground the user has already covered. Run:
   ```bash
   ls -t ./ideas/reports/*.md ./docs/design-*.md ./docs/rfc/*.md 2>/dev/null
   ```
   For each hit, note title + date in the draft report. (`./ideas/CONTEXT.md` is handled
   separately in Phase 0 — do not include it here.)
4. **The code.** Use the Agent tool (`subagent_type=Explore`) to map entry points, public
   interfaces, top-level modules, and what the product actually exposes to its users. You're
   not doing a code review — you're answering: what does this thing *do*?

Then write a short answer to these three questions into the draft report:

- **What is this?** (type of product + what it does)
- **Who is it for?**
- **What is its implicit thesis?** (the bet this product is making)

Surface this characterisation early so the user can correct the framing before ideation runs
off the wrong premise.

---

## Phase 2 — Interview

Pick 3–5 questions from `references/interview-questions.md` that are genuinely unresolved after
Phase 1. Good picks are questions where different answers would meaningfully change which ideas
are worth generating.

**Ask one question at a time** via `AskUserQuestion`. Never batch. Each answer can shift what you
ask next — batching wastes that signal and makes the user feel interrogated rather than heard.

If an answer changes your Phase 1 characterisation, update the draft report before continuing.

---

## Phase 3 — Search Before Building

*If `./ideas/CONTEXT.md` was loaded in Phase 0, start from its `## Competitive landscape`
section rather than a blank canvas. Verify the entries look current, then augment or refresh
as needed — don't redo work that's already done.*

Before inventing, search. Most good ideas already partially exist somewhere — the question is
whether the existing answers are any good, and where they fall short.

Think in three layers:

- **Layer 1 — tried and true.** Mature solutions. Don't reinvent.
- **Layer 2 — new and popular.** Scrutinise. Hype isn't value.
- **Layer 3 — first principles.** Prize above all. Where real differentiation lives.

Use `WebSearch` to identify 2–3 direct competitors and 1–2 adjacent tools (tools in a related
niche). For each: what does it do, who uses it, where is it strong, *where is it weak*. The
weak spots are the gap-finding fuel.

Don't invent competitor features. If you're unsure whether a feature is real, say so or skip it.

**Eureka callout.** If first-principles reasoning contradicts conventional wisdom — if the
market has converged on an answer that seems wrong given what this codebase can do — name it
explicitly as a **Eureka** in the report. These are often the highest-value opportunities.

---

## Phase 4 — Premise Challenge

Before generating solutions, challenge the premises. This is where shallow ideation gets caught.
Consider:

- Is this the right problem framing? Would a different framing yield a dramatically simpler or
  more impactful solution?
- What happens if we do nothing? Is this real pain or hypothetical?
- What existing code already partially solves this? Map existing patterns, utilities, and flows
  that could be reused or extended — often the best idea is *finishing something that's 60%
  there*.
- If the deliverable is a new artifact (CLI, library, package, app), how will users actually
  get it? Distribution is part of the design.

Write the premises as clear statements:

```
PREMISES:
1. [statement] — agree/disagree?
2. [statement] — agree/disagree?
3. [statement] — agree/disagree?
```

Confirm each premise via `AskUserQuestion`, **one at a time**. If the user disagrees with a
premise, revise your understanding and loop back — don't paper over a broken premise to keep
momentum.

---

## Phase 5 — Generate ideas

*If `./ideas/CONTEXT.md` was loaded in Phase 0, check `## Implemented ideas` and
`## Proposed ideas (pending)` before generating. Do not re-propose either list. You may
iterate on a pending idea — extend, split, or combine it — but label it "iteration on [title]"
so the user knows.*

Structure ideation as outcome → opportunities → solutions (opportunity-solution tree framing):

- **Desired outcome**: what this tool should enable that it doesn't fully enable today.
- **Opportunities**: unmet needs, pain points, or friction observed — each a distinct angle.
- **Solutions**: concrete ideas that address the opportunities.

Produce 5–8 solution ideas total, tiered:

- **Adjacent** — small extensions to existing surface, low risk, fast to ship
- **Expansion** — new surface, builds on existing architecture, medium lift
- **Moonshot** — significant bet, requires investment, high upside if it lands

Aim for at least one of each tier. Strong ideas beat a padded list.

Each idea follows `references/idea-template.md`. The mandatory field is **why this codebase is
uniquely positioned** — real observed architecture, not aspiration. If that field is weak,
drop the idea.

Feasibility is expressed as observed architectural leverage: "extends the existing X module",
"builds on the plugin system", "requires a new persistence layer not currently present".

---

## Phase 6 — Write the report

Write to `./ideas/reports/YYYY-MM-DD-ideate.md` using `references/report-template.md`.

Resolve the path deterministically:

```bash
mkdir -p ./ideas/reports
DATE=$(date -u +%Y-%m-%d)
BASE="./ideas/reports/${DATE}-ideate"
OUT="${BASE}.md"
N=2
while [ -e "$OUT" ]; do OUT="${BASE}-${N}.md"; N=$((N+1)); done
echo "$OUT"
```

Then write the report to the path echoed by that block.

The report should read in 5 minutes and be directly actionable. End with **concrete build
steps** — the single idea to start on, the tracer-bullet first slice, and a pointer to
`write-a-prd` for the next step. Not business validation. Not "go interview 10 users". Build steps.

---

## Phase 7 — Update canonical context

After writing the report, update `./ideas/CONTEXT.md` if it exists, or offer to bootstrap it
if it doesn't.

**If `./ideas/CONTEXT.md` exists:**
1. Ask the user via `AskUserQuestion`: *"Before I update the context file — were any previously
   proposed ideas actually shipped since our last session? I'll move them to Implemented."*
   Update `## Implemented ideas` and `## Proposed ideas (pending)` accordingly.
2. Bump `last_updated` in the YAML frontmatter.
3. Append any new competitor weaknesses or Eureka moments to `## Competitive landscape` and
   `## First-principles insights`.
4. Append the run's recommended idea to `## Proposed ideas (pending)`:
   `- [YYYY-MM-DD] [Idea title] — [one-sentence summary]`

**If `./ideas/CONTEXT.md` does not exist**, ask the user via `AskUserQuestion`:

*"Want me to save a canonical ideation context for this repo? Future runs will start with full
context — positioning, competitive landscape, and idea history — instead of from scratch."*

If they agree:
1. Run:
   ```bash
   mkdir -p ./ideas
   gh pr list --state merged --limit 30 --json title,mergedAt,body 2>/dev/null \
     || git log --merges --since=6.months --pretty=format:"- %s (%ad)" --date=short
   ```
2. Write `./ideas/CONTEXT.md` using `references/context-template.md` as the skeleton,
   populated with:
   - Phase 1 characterisation → `## Positioning`
   - Phase 3 competitive findings → `## Competitive landscape`
   - Phase 4 Eureka moments → `## First-principles insights`
   - Plausible feature-work PRs/merges from the output above → `## Implemented ideas`
     (conservative: only items that clearly look like feature work; skip chores and fixes)
   - This run's recommended idea → `## Proposed ideas (pending)`

---

## Quality bar (self-check before finalising)

- [ ] Every idea has a non-generic "uniquely positioned" field grounded in observed code
- [ ] All three tiers represented
- [ ] No two ideas are variants of the same concept
- [ ] All competitor observations are verifiable — nothing invented
- [ ] Feasibility references real observed modules, not assumed ones
- [ ] Premises were confirmed with the user before ideation
- [ ] The characterisation in the report matches what you actually found
- [ ] The recommended next step is a concrete build, not a validation task
- [ ] Any Eureka moments (first-principles beats convention) are named explicitly
- [ ] If CONTEXT.md existed: it was read in Phase 0 and the user confirmed the characterisation
- [ ] CONTEXT.md was updated (or bootstrap was offered) at the end of the run
