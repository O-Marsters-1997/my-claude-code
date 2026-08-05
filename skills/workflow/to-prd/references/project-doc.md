# The project doc

The artifact this skill produces: one doc describing one feature, opinionated enough that a reader
knows what is being built and why the scope is drawn where it is. At Plain it lives in the
**📜 Project Docs** database in Notion, with the local Markdown file as its offline twin.

## Two shapes of doc

Docs in this database come in two shapes, and which one you're writing decides most of the
structure. Pick before you start.

**Concept docs** introduce a new noun into the product — broadcasts, tenants, timelines. They read
like the product's own help-centre documentation: define the noun, list what's on it, say how
instances come into being and what they relate to, then close with **Scenarios**. Broadcasts and
Merging companies and tenants are the models.

**Change docs** alter an existing surface — a new entry point, a redesign, a builder docked next to
an editor. They open with a **TL;DR** or **Problem**, argue the **Why**, state the **Current state**,
then walk the **Proposed experience** step by step under sub-headings. Terser and more bulleted;
Sidekick in Workflows and Help Center Redesign are the models. No Scenarios and no Concepts section —
those would be filler.

Most features are one or the other. A few are both: a new noun *and* a new flow onto it, in which
case run the concept sections then the flow ones.

## House voice

**One solution, argued.** *"Be opinionated and only outline one solution."* No option matrices, no
"we could either A or B". If two approaches were genuinely live, pick one and say in a line why the
other lost — then move on.

**Concept sections read as the shipped product's own documentation.** Present tense, second person,
describing the thing as though it already exists: *"An audience is a reusable way to describe who
should receive a broadcast."* Not *"we will build an audience concept"*. This forces you to name the
concepts and notice the ones you hadn't decided. It applies to the sections defining the product, not
to the whole doc — the Why and Current state sections argue in your own voice, and change docs stay
in it throughout.

**Technical context belongs in it.** Columns, event names, mutations, tables, projections, locks,
flows. The reader is an engineer, and the scope boundary usually *is* technical — you cannot explain
why test threads skip SLA tracking without naming `is_test` and the subscribers that check it. What
stays out is *implementation sequencing and file-level design*: build order, tracer bullets, module
boundaries, per-file changes. That's `to-plan`, and for larger projects Plain's own convention is a
separate technical-proposal doc.

**Show it, don't just describe it.** Five of eight docs open with something you can watch or click: a
[supercut](https://supercut.ai) screen recording, or a Vercel prototype link. Put it near the top, not
buried in the scope section — a reader who has seen the thing reads the rest of the doc far faster.

**Diagrams earn their place.** Mermaid sequence diagrams are the house form and render natively in
Notion. Reach for one when a flow crosses more than two systems; a table when behaviour varies by
case (which subscribers fire, which events are included, what each field type does).

**Name the customers, and cite the competition.** *"Our biggest customers want this, we have
committed to Vercel that they will be able to do this by the end of June."* Named accounts with
Linear customer links, not "users have asked for this". Where the argument is about positioning,
link the comparison directly — Help Center Redesign makes its whole case by pointing at
`cursor.com/docs` versus `cursor.com/help`.

**Be honest about the bet.** A goal that doesn't move a number is allowed, if you say so. *"People
don't build workflows every day, so this won't move Sidekick usage on its own. It compounds two
slower ways… and that's the real bet."* That reads as confident; a metric retrofitted onto an
indirect payoff reads as spin.

## Scenarios, not user stories

**Never** `As a <actor>, I want <feature>, so that <benefit>` — that appears in none of these docs, in
any shape. When behaviour needs illustrating, the house format is a Scenario.

Scenarios belong in **concept docs**, where the noun is flexible enough that how teams reach for it
isn't obvious — Broadcasts and Merging tenants both close with a set. Two to four covering the real
ways it gets used, not an exhaustive enumeration. In a change doc, a step-by-step
**Proposed experience** does this job better and Scenarios would just restate it.

```markdown
### A customer writes in for the first time

**You want:** zero setup. Someone emails support and your team should immediately see who they
are and where they're from.

**How it works:** Out of the box, Plain matches the customer's email domain to a tenant, creating
one if it doesn't exist yet. Your team sees the organisation, its other open threads, and its
30-day support volume without you configuring anything.
```

```markdown
### A targeted message to one tier

**You want:** to reach only your Enterprise customers — say, about a pricing change that only
affects them.

**How it works:** Build an audience filtered to the Enterprise tier and send to that. Because the
audience is a filter, there's no list to maintain — any customer who's Enterprise at the moment
you send is included, and anyone who's since churned isn't.
```

`**You want:**` states the job in the reader's words. `**How it works:**` answers it in the shipped
product's voice, and is where a scenario either holds together or exposes a gap you hadn't decided.

## Template

**This is a menu, not a skeleton.** No doc in the database uses every section, and the shortest good
one — Help Center Redesign — uses three. Section *names* flex too (`Problem` or `Current state`,
`Scope` or `Proposed scope` or `Proposed experience`, `Parts` or `Phases`). Drop what doesn't apply
rather than filling it with restated filler; a doc padded to hit every heading reads worse than a
short one that covers what matters.

**Both shapes want the first four.** Everything after that is picked to fit.

<project-doc-template>

## TL;DR

Two or three sentences a reader can stop after, plus the recording or prototype. State what this
docks, changes or introduces, and for whom. Worth writing even when the doc is short.

## Problem  *(or: Current state)*

Why this isn't already solved — the specific things that are hard, missing or confusing today. Short,
concrete, no throat-clearing. Named customers, their requests, and competitor comparisons go here.

## Why does this matter?

The argument, as bolded claims each with a line or two under it. This is where an indirect payoff gets
stated honestly rather than dressed up.

## Goal  *(or: Goals)*

What this achieves, in one to three lines. Name **the metric it moves** where there is one — *"% of
workspaces using workflows, and workflows-per-workspace average"*. Where the payoff is genuinely
indirect, say that instead of retrofitting a number.

---
*From here on, pick what fits the doc's shape.*

## Proposed scope  *(change docs: Proposed experience)*

What we're building, at a level someone can hold in their head. Lead with the simple non-technical
description, then go as deep as the reader needs.

For a change doc, walk it under sub-headings that follow the user's path — `### Entry point`,
`### Build with Sidekick`, `### Editing an existing one` — with the flow as short bullets. Say
outright where something is still open ("still open how this should feel") rather than inventing a
decision to fill the gap.

## Concepts  *(concept docs)*

The named things this feature introduces, and what's on each. The "What's on a broadcast" / "What's on
a tenant" shape: define the noun, list its fields, say how instances come into being and what they
relate to. Skip entirely when the feature introduces no new noun.

## Scenarios  *(concept docs)*

Two to four, in the format above.

## What we already have (reuse) / What's net-new

What already exists in the codebase that this builds on, from the repo exploration — the existing
event, the existing pipeline, the behaviour that half-exists already — then the genuinely new surface,
grouped the way the work divides (Backend / SDK / UI, or by system). Naming the reuse is what keeps
the scope honest; a doc that reads as all-net-new usually hasn't looked.

## Important technical details

The context that doesn't belong to any one section: where data lives, which events fire, what's
served from where. A named home for this beats scattering it through the scope.

## Phases  *(or: Parts)*

Scope split into shippable steps. Each carries:

- **Requirements** — what this part must do.
- **Proposed flow** — how it behaves, with a diagram or screenshots where they help.
- **Nice to haves** — details that would spark joy ✨ but aren't load-bearing. Polish you take if
  there's room, not commitments.

Then **Out of scope for Phase 1** — agreed but deliberately held back, each with a one-line reason.
A follow-up project link beats a promise.

## Dependencies

Anything outside this project that has to land first — another team's project, an integration, a
migration. Name it, link it, and say what state it's in, including "no target date" when that's the
answer.

## Out of scope

Things genuinely not being built, as opposed to deferred. One line each on why.

## Open questions

Keep these in the doc rather than resolving them away — they're what a reviewer engages with.

- **Product questions:** decisions still open about behaviour or scope.
- **Design questions:** decisions still open about how it looks or feels.

Resolved questions stay too, in a collapsed toggle, with the answer that settled them. The record of
what was considered and rejected is worth more than a tidy doc.

## Feedback

What came back in review, captured as it was said, in a collapsed toggle once it gets long. Reviewers
can see their point landed, and the doc keeps the reasoning that changed it.

</project-doc-template>

**Risks and shipping sequence live in the Linear brief, not here** (`⚠️ Risks` and
`🚢 How do we ship this` — see below). None of the docs in this database carry them; duplicating them
means two copies that drift. If a risk is really an unresolved technical question, it's an open
question and belongs in this doc as one.

## Notion mechanics

The doc's home:

> **📜 Project Docs** — `https://app.notion.com/p/39f810036f9d803da599fa9672912dcd`
> Data source: `collection://39f81003-6f9d-805a-8609-000b0c4d77f9`
> Default page template: `39f810036f9d80c0aa0ccbe5b61f5b51`

Create a new page in that data source. Properties: `Name` (title), `Lead` (person), `Team` (select),
`Status` (status — `WIP` · `Review` · `Ready` · `Archived`). A new doc is `WIP`.

**Fetch the data source before setting `Team`** and read the options off it rather than working from
a list written down here. Team names churn, nothing in this file would signal the list had gone
stale, and a stale option that reads as authoritative silently files the doc under the wrong team.
Same for `Lead` — resolve the user against Notion's users rather than guessing an ID.

**Iteration happens in the page's inline comments.** When picking a doc back up, re-fetch it with
`include_discussions: true`, read the threads, and fold them in — don't re-derive the doc from the
conversation and overwrite what reviewers wrote. Docs in this database carry their review history
inline, and a `## Feedback` section capturing what came back is normal.

**No Notion access?** Keep the Markdown file, tell the user the page wasn't created and needs
publishing by hand, and carry on. The doc matters more than where it lives — a missing page is not a
failed run.

## The upstream Linear brief

Separate artifact, upstream of this one: the **Linear project**, whose description follows a fixed
template and syncs into Notion's *Product Projects DB*.

```
🧩 What problem are we solving?   — the problem, not the solution, nothing technical
🚨 Why does this matter?
👥 Who is this for?               — named customers, linked to their Linear customer records
👷 Project docs                   — links to this doc
⚠️ Risks
🚢 How do we ship this            — milestone sequencing
```

Read it as a prior when one exists. It answers problem, why, who, risks and shipping sequence, and
re-asking those wastes the user's time. It is also frequently half-empty, which is a gap worth
offering to fill from what the interview surfaced — offered, never written unprompted.
