---
name: capture-idea
description: >
  Capture an idea you just had — shape it into one concrete line, then append it to
  ## Accepted ideas in ./ideas/CONTEXT.md. Use for "capture this idea", "add that to
  the roadmap", "note this down before I forget".
disable-model-invocation: true
---

# Capture an idea

A 30-second capture, not a design review. The one line you write is a *feature* — something to be
built, later broken into several tickets — so it needs a title and a summary someone could act on,
and nothing more.

## Process

**1. Shape it.** Most ideas arrive concrete enough already; if so, skip straight to step 2. Only if
the idea is too vague to write as one line, ask:

- Is there a concrete feature here, or is this still a hunch?
- What is the one-sentence version?

**2. Read `./ideas/CONTEXT.md`.** If a near-duplicate already sits in Accepted, Proposed or
Implemented, say which one and let the user decide: merge, sharpen, or add anyway. If the file
doesn't exist, create it with just this stub and say so:

```markdown
# Ideation Context

## Accepted ideas

*Ideas you've decided to build, not yet shipped. `to-roadmap` boards this section when there is no
approach doc, so accepting is what puts an idea on the board.*
```

**3. Confirm the line, then append.** Show the exact line first — one line is cheap to correct, and
a wrong one silently becomes a roadmap card:

```
- [YYYY-MM-DD] [Idea title] — [one-sentence summary]
```

Append it to the end of `## Accepted ideas`. No provenance field: nothing downstream should need to
know whether an idea came from you or from a discovery run. Change nothing else in the file.

## Finishing up

Echo the line you wrote and where.
