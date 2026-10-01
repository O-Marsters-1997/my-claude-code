# Ideate — Evals

Benchmark cases for `ideate`, run with the `skill-creator` loop.

## Target

Every eval runs against a fresh `git clone` of `~/Documents/coding/job-hunting/job-scraper`, with
cwd at the clone root. `ideas/` is gitignored, so a clone starts with no ideation memory. Eval 4
copies the real `ideas/CONTEXT.md` (written in the pre-rewrite format) and the 2026-09-30 report into
the clone before running, which also tests that older context files are upgraded in place.

## Cases

| # | Name | Prompt shape | What it pins |
|---|---|---|---|
| 1 | what-next-casual | low-information "what next?" | the full contract: inventory, matrix, quotas, critic, cut table, no default |
| 2 | competitive-gap-hunt | "where do we differentiate?" | market-scan depth: 5 buckets, vendor docs plus user voice, labels, absence wording, insight gate |
| 3 | focus-arg-alerts | `/ideate alerts` | the focus re-weights but doesn't narrow: ≥2 ideas outside the focus, focus not re-asked |
| 4 | rerun-with-context | "what should i build next?" plus existing CONTEXT.md | memory: no re-proposing Accepted, Implemented or Rejected ideas; headings intact; matrix extended, not rebuilt |

## Runner protocol

A subagent can't reach a real user, so each run:

- logs every question it would ask to `outputs/process-log.md`, with the tool, question and options,
  and answers neutrally ("looks right, carry on", "no strong view"). The final pick is answered
  "none for now", so `to-prd` never fires;
- logs each phase, every subagent it spawned, and its web-call count;
- copies the clone's `ideas/` to `outputs/ideas/`, and writes `outputs/final-message.md`.

The question-count, critic-ran and no-default assertions are graded from the process log and the
final message. Everything else is graded from the files.

## What the assertions deliberately reward

The old suite rewarded the failure modes: two or more questions, confirmed premises, named
Eurekas. This suite rewards the opposite:

- inventory with `file:line` evidence and at least one proven absence;
- 50% or more deepen/extend, and at least 2 net-new;
- 4 of the 5 market buckets, with no listicle cited as a fact;
- every Insight backed by 2 or more independent sources, or downgraded;
- critic ran, cut table and Parked appendix present;
- 3 or fewer questions before the report, and no pre-selected default;
- with a focus, 2 or more ideas outside it;
- on a rerun, no re-proposals and headings intact.

## Running

```
/skill-creator:skill-creator run evals for ideate
```

Baseline is the previous skill version (snapshot in `ideate-workspace/skill-snapshot/`). Results go
to `ideate-workspace/iteration-N/<eval>/{with_skill,old_skill}/`.

## Not covered

Whether the ideas are actually good. Read the reports in the viewer's Outputs tab. Triggering
accuracy is tuned separately with skill-creator's description optimiser.
