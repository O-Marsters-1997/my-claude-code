# Ideate — Evals

This directory holds the benchmark test cases for the `ideate` skill. The evals follow the
approach described in the `skill-creator` skill.

## What's here

- `evals.json` — 3 test prompts with assertions, covering different axes:
  - **what-next-casual** — low-information prompt; tests basic flow and report structure
  - **competitive-gap-hunt** — competitive framing; tests market scan rigor and gap identification
  - **personal-to-product** — strategic framing; tests premise challenge, distribution thinking,
    and outcome framing

## Assumptions the prompts make

Every eval assumes the subagent runs with `cwd` set to a real codebase checkout. The default
target is this skills repo itself (`my-claude-code`) so ideation has concrete code to ground
against. If you want to benchmark against a different codebase, point the subagent's working
directory there.

## Running the benchmark

Invoke skill-creator and hand it this evals file:

```
/skill-creator:skill-creator run evals for ideate
```

The skill-creator flow will:

1. Spawn one subagent per eval **with** the skill and one **without** (baseline), in parallel
2. Save outputs to `<skill-path>-workspace/iteration-N/eval-<name>/{with_skill,without_skill}/outputs/`
3. Grade each run's outputs against the assertions in `evals.json`
4. Aggregate into `benchmark.json` + `benchmark.md` (pass rate, time, tokens per config, with deltas)
5. Launch the review viewer (`eval-viewer/generate_review.py`) so you can click through outputs
   and leave feedback

## Iterating

After each iteration:

1. Review outputs in the viewer; leave feedback on anything that's off
2. Read `feedback.json` when the viewer exports it
3. Update `SKILL.md` / references based on the feedback + any patterns in the benchmark data
4. Rerun the benchmark into `iteration-<N+1>/` and compare against the previous iteration

The quantitative signal to watch: pass rate on the "uniquely positioned" assertion (idea #8 in
eval 1, #5 in eval 2, #7 in eval 3). This is the hardest thing for the skill to do well and
the thing most worth optimising.

## What this benchmark doesn't cover

- **Idea quality** beyond structural checks. Whether the ideas are actually *good* is a
  subjective judgment — that's what the viewer's Outputs tab is for. Read the reports yourself.
- **Description-optimisation triggering.** Once the skill body is stable, run the description
  optimisation loop from skill-creator separately to tune the frontmatter description for
  trigger accuracy.
