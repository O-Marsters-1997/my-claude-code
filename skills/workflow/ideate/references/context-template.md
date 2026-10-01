# `./ideas/CONTEXT.md` template

The ideation memory for the repo. Each `ideate` run reads it first and writes it back last.
Humans can edit it between runs. Other skills match its headings by name, so never rename or
drop one:

- `to-roadmap` reads `## Accepted ideas`
- `capture-idea` writes `## Accepted ideas`

Idea lines always use the form `- [YYYY-MM-DD] Title — summary`.

When an older file lacks a section below, add it in place and leave the others untouched. Its
older sections (`## Positioning`, `## Competitive landscape`, `## First-principles insights`) can
stay. Fold their content into the new sections when it's still current.

```markdown
---
last_updated: YYYY-MM-DD
---

# <Project> — Ideation Context

## Positioning

<What it is, who it's for, the bet it makes. Three lines.>

## Capability inventory

_Pinned at `<hash>` on YYYY-MM-DD. The next run re-checks rows whose files changed since then._

| # | Capability | Status | Evidence | Job step |
|---|---|---|---|---|

## Market matrix

_Entries are dated. Re-verify entries older than 3 months; append, don't overwrite._

### YYYY-MM-DD
<matrix table, including this product's column and the evidence-quality row>

## Insights

- [YYYY-MM-DD] **Insight | Hypothesis | Watch:** <claim in gate form> — sources: <URLs> —
  check: <the cheapest test, for a Hypothesis>

## Implemented ideas

- [YYYY-MM-DD] Title — summary

## Accepted ideas

- [YYYY-MM-DD] Title — summary

## Proposed ideas (pending)

- [YYYY-MM-DD] Title — summary

## Rejected ideas (with reason)

- [YYYY-MM-DD] Title — reason (who rejected: user | critic; evidence)
```

A later run can downgrade an insight when its sources go stale or are contradicted. Edit its tier
in place and add the date. Don't delete the line.
