---
name: sweep
description: >
  Turn a pile of small fixes into a batch /fleet can dispatch. Takes the open `inbox` issues and
  any items pasted with the command, explores them in one pass, rejects anything too big for a
  small fix and takes it out of the inbox, merges items that touch the same files into one
  ticket, and files the tickets under one dated `sweep-YYYY-MM-DD` label. Use for "sweep the
  inbox", "batch these fixes", "turn these into tickets and fleet them". For one ad-hoc ticket,
  use file-issue; for a feature, use /to-tickets.
disable-model-invocation: true
---

# Sweep

Small fixes cost little to make and a lot to run one at a time. Sweep pays the exploration and
review once for the whole batch: the label it files under is a feature label as far as
`/fleet` is concerned, so dispatch, integration and the single PR into `main` all apply.

## 1. Gather

- `gh issue list --state open --label inbox --json number,title,body`
- Any items in the user's message, split into one item per distinct fix.

Nothing from either source: say the inbox is empty and stop.

## 2. Explore once

Locate every item in a single pass, never one per item. With a `.codegraph/` index, one
`codegraph_explore` naming every item's symbols and files. Otherwise one `Explore` subagent with
`model: "haiku"`, given the whole list, returning per item: the files and symbols involved, and a
rough size (lines, files).

## 3. Reject what isn't small

An item stays in the sweep only if all hold:

- it changes existing behaviour or code, not a new feature;
- about 80 lines or fewer, across 1–3 files;
- no schema, public API or dependency change;
- the cause is known. A bug whose cause still needs finding is triage, not a sweep item.

Rejects are listed with their route: `triage-issue` for an unexplained bug, `capture-idea` or
`to-prd` for a feature, and `needs-design` for a design or refactor concern with no known fix
(such as a `LEARN later:` marker). Name no skill for `needs-design`: deciding the fix is the
open work. A rejected inbox issue stays open but leaves the inbox in step 6, or every later
sweep explores and rejects it again.

## 4. Cluster by file

Items that touch a common file go into one ticket, transitively: if A shares a file with B and B
with C, all three are one ticket. Disjoint items stay separate tickets, so they run in parallel
without conflicting.

If a cluster grows past about 3 files or 150 lines, split it back into tickets and chain them
with `## Blocked by` in item order, so `/fleet` runs them in successive waves instead of
colliding.

Size each ticket: `size:xs` when it touches one file and needs at most one new test case,
otherwise `size:s`.

## 5. Confirm the batch

Show one table and wait for a yes. The clustering is the decision, so this is the one confirm:

```
Label: sweep-2026-10-07
| Ticket | Items (inbox #) | Files | Size |
Rejected: <item> (inbox #) → <route>
```

## 6. File

Label: `sweep-YYYY-MM-DD` with today's date. If that label already has open issues, append
`-2`, `-3`, … so a dispatched batch never gains tickets mid-flight. Create the label with
`gh label create "<label>" -d "Sweep of small fixes"`, and `size:xs` / `size:s` if missing.

Each ticket uses file-issue's body (`## Context`, `## Where to look`, `## Acceptance criteria`)
with step 2's findings as "Where to look", plus:

```markdown
## Items

- <item> (from #<inbox issue>)

## Blocked by

None
```

Acceptance criteria: one or two per item.

```bash
gh issue create --title "<title>" --body-file <tmpfile> \
  --label "<sweep label>" --label "status:ready" --label "size:<xs|s>"
```

A chained ticket from step 4 gets `status:backlog` and a `- Blocked by #<N>` line instead.
Then close every absorbed inbox issue: `gh issue close <N> --comment "Swept into #<ticket>"`.

Take every reject that came from an inbox issue out of the inbox with
`gh issue edit <N> --remove-label inbox`. A `needs-design` reject also gets
`--add-label needs-design`; create that label first if missing:
`gh label create needs-design -d "Needs a design decision before it can be fixed"`.

## 7. Report

The tickets as `#<N> <title> (size)`, one per line, then the rejects with their route, then the
command on its own line:

```
/fleet dispatch <sweep label>
```
