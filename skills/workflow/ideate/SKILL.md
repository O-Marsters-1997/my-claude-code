---
name: ideate
description: >
  Use whenever the user asks what to build next in this project: "what should I build next?",
  "any ideas for this project?", "what's missing?", "help me find gaps", "where could this go?",
  "what are competitors doing?", "where could we differentiate?", "how do I grow this?". Trigger for
  roadmap ideation, product discovery, gap analysis and competitive positioning in a codebase, even
  when the user never says "ideate". Inventories what the code already does, scans the real market,
  and produces a portfolio of 6–10 evidence-backed ideas (mostly deepening what exists, some
  net-new), each of which later runs /to-prd on its own. For "which skill should I run next?" use
  artifact-scan.
---

# Ideate

Produce a **portfolio** of what to build next in this codebase. The whole run is built on an
inventory of what the code already does, checked against a market matrix, with every claim tied to
evidence. It is not a single-idea design session. Each shortlisted idea later becomes its own
`to-prd` run. Nothing gates anything: a user who already knows what they want goes straight to
`to-prd`.

## Hard rules

- **Code before questions.** Never ask the user something the repo can answer. Ask nothing until
  the inventory exists and has been shown.
- **At most 3 `AskUserQuestion` calls before the report.** Options offered are about *process*
  ("looks right, carry on" / "go deeper on X area"), never about idea *content*. Anything left
  unasked becomes a labelled assumption in the report.
- **A focus re-weights, never deletes.** If the user names a focus (`/ideate alerts`), rank ideas
  inside it higher. At least 2 shortlisted ideas still sit outside it, marked `outside focus`.
- **Firewall.** Market findings decide what to ask and what to compare. They are never evidence
  about this codebase. The inventory is never evidence about the market.
- **Generate everything before cutting anything.** Cuts are visible with a reason. Leftovers are
  parked, not deleted.
- **No pre-selected default** at the final pick. "None" is a valid answer.
- **Tone:** curious, specific, calibrated. No hype. "Nothing surprising this run" is a normal result.

## Phase 1: Load memory

If `./ideas/CONTEXT.md` exists, read it in full (`references/context-template.md` describes its
sections). Use its `## Capability inventory` and `## Market matrix` as the baseline and update
only what changed: `git diff --stat <pinned-commit>..HEAD` tells you which inventory rows to
re-check, and matrix entries older than 3 months get refreshed. Collect every title under
Implemented, Accepted and Rejected. Those are never re-proposed, including under a new name. A
Proposed idea may come back only if labelled `iteration on: <title>`. Do not ask the user to
confirm the file.

An older CONTEXT.md may lack the inventory, matrix or Rejected sections. Treat those as empty and
add them when writing back.

Create a run-local work dir (`WORK=$(mktemp -d -t ideate)`). Keep the inventory, matrix, raw
candidates and critic inputs there, and pass subagents file paths rather than pasted text. A
fixed scratch path collides with any other ideate run going on at the same time.

## Phase 2: Capability inventory

Follow `references/inventory.md`. The output is a table of capabilities in the repo's own domain
terms. Each has a status of COMPLETE, PARTIAL, ORPHAN or GAP, `file:line` citations, the search
that proves every absence, a job-map overlay, and a pinned commit hash.

As soon as the capability rows exist, launch the market scan (Phase 4) in the background so it
runs while the user reads.

## Phase 3: Show, then ask

Show the inventory summary: the capability rows, their status, and the 3–5 sharpest signals.
Then ask what's wrong, missing, or more important than the summary suggests. Use at most 3
questions, each about something only the user knows: who else uses it, what they actually do day
to day, what they've tried and abandoned. A focus given with the invocation counts as an answer
already received, so don't re-ask it.

Record every answer and every unasked gap in an **assumptions log** (`A1: <assumption> — basis:
<code | user | none>`). Answers re-weight candidates later. They never remove areas from
generation.

## Phase 4: Market scan (background subagent)

Spawn one `Agent` (general-purpose, `run_in_background: true`) with the contract in
`references/market-scan.md`, filled in with the product summary, the capability names as matrix
rows, and any prior matrix from CONTEXT.md. `WebSearch` and `WebFetch` are used only inside this
subagent. Wait for it before generating. If it fails or comes back thin, say so in the report and
continue with what you have. Never fill the gap from memory.

## Phase 5: Generate wide

Follow `references/generate.md`. Derive 3–5 axes from the inventory. Spawn **3 generator
subagents in one message**, so they can't see each other, each with a different lens set. Aim for
25–40 raw candidates. Any axis that comes back empty gets one recovery pass. If a parallel spawn
is refused because of a concurrency limit, spawn the generators one at a time. Each still runs in
its own context and is given none of the others' output. Never fold them into your own context.

## Phase 6: Converge

Still in `references/generate.md`: merge duplicates explicitly, write the cut table (one reason
per cut), and pick 6–10 ideas:

- half or more are `deepen` or `extend` ideas on existing inventory rows;
- at least 2 are `net-new`, from an empty job step or a matrix gap;
- every idea cites inventory rows;
- with a focus, at least 2 are outside it.

Order the shortlist by Kano: must-be gaps, then performance, then at most a few delighters. Write
each idea with `references/idea-template.md`. Everything else goes to the Parked appendix.

## Phase 7: Critic

Spawn a fresh `Agent` that saw none of the generation, using the brief in `references/evidence.md`.
It tries to refute each shortlisted idea and every insight claim from code and sources, and writes
a one-line pre-mortem per idea. Act on its verdicts: drop refuted ideas into the cut table and
promote from Parked to keep 6–10, downgrade strength badges and insight tiers it weakens, and
record its strongest counter-argument against each surviving insight.

## Phase 8: Insight gate

Apply the gate in `references/evidence.md` to every "the market has this wrong" claim. Each one
ends as **Insight**, **Hypothesis** (with the cheapest build-and-measure check) or **Watch**.
"None this run; conventional wisdom looks sound because …" is a valid and common outcome.

## Phase 9: Report, memory, selection

1. Resolve the path and write the report with `references/report-template.md`:
   ```bash
   mkdir -p ./ideas/reports
   BASE="./ideas/reports/$(date -u +%Y-%m-%d)-ideate"; OUT="$BASE.md"; N=2
   while [ -e "$OUT" ]; do OUT="$BASE-$N.md"; N=$((N+1)); done; echo "$OUT"
   ```
2. Update or create `./ideas/CONTEXT.md` with `references/context-template.md`. Write back the
   inventory and matrix, append insights, add the shortlist to `## Proposed ideas (pending)`, and
   add critic-refuted ideas with load-bearing reasons to `## Rejected ideas (with reason)`. If
   `git log` since `last_updated` clearly shows an Accepted or Proposed idea shipped, move it to
   Implemented and note this in the assumptions log. Keep every existing heading. Move lines
   between sections verbatim: other skills and the user match on that exact text.
3. Present the shortlist in chat: one line per idea (`n. title · type · appetite · badge`) with
   the report path. Ask which one, if any, to take forward. Ask in plain text, because the
   shortlist exceeds `AskUserQuestion`'s four options and trimming it to fit would be a hidden
   pre-selection. Recommend nothing in the question.
4. On a pick: fill the report's `## Selection` (title plus first slice), move its line from
   Proposed to `## Accepted ideas`, and invoke `to-prd` with the idea's first slice as the brief.
   Ideas the user explicitly rejects go to Rejected with their reason. On "none" or no answer,
   leave everything as proposed. Selection can happen on a later pass.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke
skill-feedback-collector passing this skill's name and path.
