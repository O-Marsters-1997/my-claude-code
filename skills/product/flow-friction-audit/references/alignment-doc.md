# The alignment doc — stage 3b's output, and `shape` mode's input

The artifact the 3b gate produces: a Notion page in the **Team Growth project home** holding the agreed problem space for one flow. It exists so alignment survives the conversation that produced it — the team reads one statement of what's wrong, and `shape` re-enters from it as many times as the ideas need reworking, without re-deriving the problems each time.

`findings.md` in the audited repo stays the evidence layer. This page is the product layer. Link each to the other once and don't restate one inside the other.

## The page

```markdown
# <descriptive title — name the flow and the shape of the problem>

<The goal of this flow — what a user is here to achieve — and an honest
assessment of how well it currently does that, with the high-level why.
A few sentences. Someone who has never used the product should finish
this paragraph knowing whether the flow is broadly working or broadly not.>

## problems

1. **<problem name>** — <two or three lines: the shape of the problem and
   what it costs the user. The thing underneath the symptoms, not a list
   of them.>
2. …
```

Two to four problems. Each is either something to solve outright or an area to mitigate and improve.

**No solution ideas on this page.** The temptation is to write "needs a progress indicator" instead of "the user can't tell whether setup is finished". An idea phrased as a problem gets agreed by proxy, and from then on it's the only direction anyone designs — which is the entire value of the gate, spent before the research has happened. If a proposal is the only way you can express the problem, you haven't found the problem yet.

## Aligning on it

Publish the draft, then ask:

- Anything missing — friction you know about that the walk didn't reach?
- Anything you disagree with, or that looks mis-rated?
- Anything mis-clustered — one problem that's really two, or two that are really one?

Edit the page rather than the conversation, so the agreement and the artifact never diverge. Iterate until the user agrees.

## Re-entry context

`shape` may run days later, in a fresh context, possibly from a different repo. It needs enough to start correctly and no more — the failure mode on both sides is real: too little and it designs blind against a surface it can't see; too much and you've rebuilt the audit transcript inside a product doc, and the page stops being readable by the humans it was written for.

Append as a collapsed section at the foot of the page:

- **Audited surface** — the URL walked, the driver used (`playwright-cli` or `claude-in-chrome` against a live session), and the date.
- **Code** — the audited repo, and the key `file:line` refs from 3a corroboration, one line each on what they establish.
- **Per problem** — its finding IDs with a one-line summary of the evidence behind each.
- **Current design** — a short read of the affected surfaces: the regions, primary actions and copy as they stand today, enough that a wireframe can be a modification of something real.
- **Artifacts** — the path to the run's `findings.md`.

Deliberately excluded: the step log, the full findings prose, and any prior fix directions. The first two are in `findings.md` if a reader wants them. The third is excluded on purpose — carrying old directions forward makes them the ceiling on the new ones, which is the same trap [fresh runs](../SKILL.md#fresh-runs) exists to avoid.
