# The alignment doc — stage 3b's output, and `shape` mode's input

The artifact the 3b gate produces: a Notion page holding the agreed problem space for one flow. It exists so alignment survives the conversation that produced it — the team reads one statement of what's wrong, and `shape` re-enters from it as many times as the ideas need reworking, without re-deriving the problems each time.

## Where it goes

A **new child page** under the growth team's Notion home:

> **📈 Home** — `https://app.notion.com/p/3b3810036f9d8246bca5011749c51244`

One page per audited flow. Create a child page; never edit Home itself — it's the team's landing page (roadmap, dashboards, Slack and Linear links), not a place to append findings to.

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
- Which of these get design work, and which is load-bearing? They're rarely peers, and stage 5's budget is set by this answer.

Edit the page rather than the conversation, so the agreement and the artifact never diverge. Iterate until the user agrees.

## Goal metrics — fetched here, never carried

The team's current goal metrics live in the **Q3 Roadmap** — `https://app.notion.com/p/team-plain/Q3-Roadmap-3a5810036f9d8027905ce48603a7d09f`. Fetch them at this gate. **Don't copy the definitions into this skill** — they're quarterly and Plain-specific, nothing here would signal they'd gone stale, and a stale metric that reads as authoritative silently reframes an audit; a pointer stays correct, a copy doesn't. And **don't fetch them before the walk** — handed "valuable set-up means two channels connected and a workflow live" at scope, the audit measures the flow against that definition instead of seeing it, and the finding that the definition *is* the problem never happens.

They enter as context, not a rubric: what the business is trying to move says which problems matter, and feeds the budget question above — a problem sitting on the primary goal earns weight one outside it doesn't. **A metric can be wrong, and the evidence may say so.** Where a metric assumes something the walk contradicts, state it plainly in the problem it belongs to: a definition of set-up that counts configuration describes the starting point rather than the goal, and a workspace can satisfy it while being unable to reply to anyone. Surface the contradiction; redesigning the metric belongs to the team.

## Re-entry context

`shape` may run days later, in a fresh context, possibly from a different repo. It needs enough to start correctly and no more — the failure mode on both sides is real: too little and it designs blind against a surface it can't see; too much and you've rebuilt the audit transcript inside a product doc, and the page stops being readable by the humans it was written for.

Append as a collapsed section at the foot of the page:

- **Audited surface** — the URL walked, the driver used (`playwright-cli` or `claude-in-chrome` against a live session), and the date.
- **Code** — the audited repo, and the key `file:line` refs from 3a corroboration, one line each on what they establish.
- **Per problem** — its finding IDs with a one-line summary of the evidence behind each.
- **Current design** — a short read of the affected surfaces: the regions, primary actions and copy as they stand today, enough that a wireframe can be a modification of something real.
- **Artifacts** — the path to the run's `findings.md`.

Deliberately excluded: the step log, the full findings prose, and any prior fix directions. The first two are in `findings.md` if a reader wants them. The third is excluded on purpose — carrying old directions forward makes them the ceiling on the new ones, which is the same trap [fresh runs](../SKILL.md#fresh-runs) exists to avoid.
