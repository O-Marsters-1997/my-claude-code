# Flow friction audit — {app name / surface}

**Run goal:** {friction discovery | conversion | concept validation}
**Audited:** {url} · {date}
**Flow walked:** {flow name}
**Driven in:** {playwright-cli | claude-in-chrome against the user's authenticated session — and why}
**Not exercised:** {steps deliberately not completed, and why — so a reader can tell observed claims from inferred ones}

## Summary

{One paragraph: what a first-time user actually experienced walking this flow, and the verdict. Written for someone who will read nothing else.}

**Problems**

1. **{problem name}** — {one line}
2. **{problem name}** — {one line}

{n} findings ({n} blocker · {n} major · {n} minor)

---

## Problems

The aligned problems — what the findings *mean*, agreed with the user before any design work. These are what the fix directions are designed against.

### 1. {Problem name}

{Two or three lines: the shape of the problem and what it costs the user. Not a list of symptoms — the thing underneath them.}

**Evidence:** `{finding-id}`, `{finding-id}`, `{finding-id}`

<!-- repeat per problem, 2–4 total. Each finding appears under exactly one problem. -->

---

## Findings

The observation layer the problems are drawn from. Ranked most severe first. Each ID describes the friction, not its position.

### {flow}-{what}  ·  {blocker | major | minor}

- **Flow:** {which flow, which step}
- **What happened:** {the element → expected → actual triple that exposed it}
- **Evidence:** {step-log row(s), and `file:line` for anything read from source}
- **Fails:** {rubric heuristic, e.g. "6 · Recognition rather than recall"}
- **Why it's friction:** {one line — the user's felt experience}

<!-- repeat per finding. Defects and automation artifacts do not belong in this file at all. -->

---

## Evidence — step logs

The raw walk each finding is drawn from. Kept so the critique and any re-run can trace a finding back to what actually happened.

| # | Element (ref · label) | Expected | Actual | Wait |
|---|-----------------------|----------|--------|------|
| 1 | e15 · "Continue" | advance to payment | advanced, no confirmation shown | a beat |
| 2 | … | … | … | … |

---

## Critique

{impeccable's critique, folded in — not linked. Design Health score, the specificity verdict, the P0s, and any check that failed critically. State whether it ran with isolated sub-agents; if not, say that its agreement with these findings is not corroboration.}

---

## Fix directions

{Three divergent directions per problem. Each names the problem and the critique point it answers, with its low-fi wireframe — `![{direction-id} — {problem}]({direction-id}.png)`. Fidelity per references/wireframe-style.md: one direction per frame, one frame per screen.}

---

## Inspiration — where to look next

Products that solve these problems well, for the reader to explore further.

**{Problem name}**

- **{Product · surface}** — {one line on what it does right}

<!-- repeat per problem -->
