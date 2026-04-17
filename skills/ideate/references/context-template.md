# Ideation Context Template

Use this skeleton when bootstrapping `./ideas/CONTEXT.md` for a new repo. The file is
**human-editable between runs** — the `ideate` skill reads it at the start of each session
and updates it at the end. Edit it directly to correct stale entries, add context the skill
missed, or remove ideas that were abandoned rather than shipped.

Sections are stable by heading — the skill matches them by name. Don't rename them.

---

```markdown
---
last_updated: YYYY-MM-DD
---

# [Project name] — Ideation Context

*This file is the canonical ideation memory for this repo. It is read at the start of every
`ideate` run and updated at the end. Edit it freely between runs.*

---

## Positioning

**What is this?**
[One sentence: type of product and what it does]

**Who is it for?**
[Target user — primary persona]

**Implicit thesis**
[The bet this product makes — what it optimises for that alternatives don't]

---

## Problems it solves

- [Core pain point 1]
- [Core pain point 2]
- [Core pain point 3]

---

## Competitive landscape

*Updated [YYYY-MM-DD]. Refresh if more than 6 months old.*

**[Competitor 1]**
Strong at: [what it does well]
Weak at: [where it falls short — this is the gap map]

**[Competitor 2]**
Strong at: [...]
Weak at: [...]

**[Adjacent tool]** *(related niche, not direct)*
Strong at: [...]
What it reveals: [adjacent user need this exposes]

---

## First-principles insights

*Eureka moments: where first-principles reasoning contradicts conventional wisdom in this
space. These are often the highest-value opportunities. Add new ones as they emerge.*

- [YYYY-MM-DD] **[One-line insight]** — [short explanation: what the market assumes and why
  this codebase is positioned to challenge it]

---

## Implemented ideas

*Running log of ideas that shipped. Seeded on bootstrap from recent merged PRs; extended
by each ideation run.*

- [YYYY-MM-DD] [Feature / idea title] — [one-sentence description of what was built]

---

## Proposed ideas (pending)

*Ideas recommended by past ideation runs that haven't shipped yet. The `ideate` skill will
not re-propose these — it may iterate on them. Move to Implemented when shipped.*

- [YYYY-MM-DD] [Idea title] — [one-sentence summary]
```
