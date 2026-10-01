# Generate wide, then converge

LLM idea pools duplicate themselves and collapse toward the same generic ideas ("add AI insights",
"browser extension"). Breadth has to be built into the process, because asking for "more ideas"
doesn't produce it. That means isolated generators, forced lenses, axes taken from this codebase,
and a dedupe step.

## Axes

Derive 3–5 orthogonal **axes** from the inventory and the market matrix: *what* to think about.
Good axes are user-facing tensions or flows, such as "time from signal to action", "trust in
automated output", "what accumulates over months of use", or a specific persona's day. If the axes
read like a directory listing of the code, redo them. Every inventory row should fall under at
least one axis.

## Lenses (how to think), split across three generators

| Generator | Lenses |
|---|---|
| G1: finish what exists | **Deepen a module.** Add behaviour without adding surface: one smart default instead of three settings, one pipeline absorbing manual steps. Use the `codebase-design` vocabulary: depth, leverage, locality. **Kano must-be gaps.** Missing edges, error states, the absent "second half". **Data leverage.** What the system stores or computes that no user sees. |
| G2: walk the journey | **Friction along the user's journey.** Walk the main flow step by step: unmet expectations, waits, manual hand-offs, dead ends. **One ordinary persona.** Day 1 versus week 8, the user whose output was wrong, the user who stopped using it. Never a celebrity. **Reverse an assumption.** Name something the product takes for granted, then invert it. |
| G3: look outward | **Steal from the matrix.** `steal` verdicts and rows where an alternative is Full and this product is Partial or None. **Empty job steps.** GAP rows on the job map. This generator owns most net-new ideas. **Reverse a market assumption.** What every alternative does that this product could deliberately not do. |

## Generator brief

Spawn all three in **one message**, so none can see another's output. Ideas that can see each
other anchor each other. Pass each the same inputs: the inventory table, the signals, the job map,
the market matrix and user-voice clusters, the axes, the assumptions log, and the exclusion list
(Implemented / Accepted / Rejected titles). Give each only its own lenses.

```text
Generate 10–14 product ideas for this codebase using only your lenses: <lenses>.
Cover every axis: <axes>. The first few ideas you think of will be the obvious ones; keep going.
For each, also give your probability (0–1) that a typical AI assistant would propose the same idea
for any app in this category. Keep the low-typicality ideas that are still grounded.
Do not propose anything in the exclusion list, or a renamed version of it.
You may read code to ground an idea. Do not search the web.

Return one line per idea:
<title> | axis | lens | deepen/extend/net-new | inventory rows | basis | typicality
basis is one of:
  direct: file:line (quote the line)
  external: URL from the matrix
  reasoned: a one-sentence argument
```

- `deepen`: more behaviour behind an existing capability's current surface.
- `extend`: a new surface on an existing capability.
- `net-new`: a capability or job step that doesn't exist yet.

**Recovery.** After merging, count ideas per axis. Any axis with zero ideas gets one more small
generator call, with that axis alone and all lenses. If it still comes back empty, report the axis
as a deliberate gap.

## Converge

1. **Merge.** Group near-duplicates, then keep the best-argued version and list what was merged
   into it ("C7 ← G1-3, G3-8"). If most ideas fall into one cluster, divergence was too narrow, so
   run recovery before converging.
2. **Cut table.** Every raw candidate that isn't merged or shortlisted gets a row: `idea | reason`.
   Reasons are specific: "already exists at `x.go:40`", "rests on Unknown matrix cells only", "a
   Module 2.0 grab-bag, not bounded", "duplicates Rejected: <title>", "commodity: use a library".
3. **Select 6–10.** Score each survivor on evidence strength (direct > external > reasoned),
   codebase leverage (rows it builds on), Kano class and appetite. Shuffle the order before you
   compare, so generation order and write-up length don't win. Then enforce:
   - half or more are `deepen` or `extend`;
   - at least 2 are `net-new`;
   - every idea cites inventory rows;
   - with a focus, at least 2 are outside it and labelled.
   If a quota can't be met from the pool, say so rather than inventing to fill it.
4. **Parked.** Everything plausible that wasn't selected goes in the appendix, one line each:
   "Interesting, maybe some day".
5. **Order** the shortlist by Kano: must-be, then performance, then delighters (two at most).
