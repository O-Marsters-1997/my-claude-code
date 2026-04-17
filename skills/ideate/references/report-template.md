# Report Template

Use this skeleton when writing `./ideas/reports/YYYY-MM-DD-ideate.md`. Adapt headings as the
content demands — the goal is a report someone can scan in 5 minutes and act on.

_If `./ideas/CONTEXT.md` was loaded: the Codebase Characterisation below can be a one-line
delta ("Same as CONTEXT.md — no material changes") rather than a full re-statement, unless
the characterisation changed during this run._

---

```markdown
# Ideation Report

_Generated: [date] · Codebase: [project name or repo path]_

---

## Codebase Characterisation

**What is this?**
[One sentence: type of product and what it does]

**Who is it for?**
[Inferred or stated target user]

**Implicit thesis**
[The bet this product makes — what it optimises for that alternatives don't]

_If any of the above looks wrong, correct it before reading further — every idea downstream
depends on this framing._

---

## Interview Summary

_3–5 bullet points capturing what the user told you that shaped this ideation — goals,
constraints, context, things to avoid. Skip if the interview didn't move your priors._

---

## Market Landscape

_2–3 direct competitors + 1–2 adjacent tools. One paragraph each. Focus on where they're
weak — that's the gap map._

**[Competitor 1]**
[What it does · who uses it · where it's strong · where it falls short]

**[Competitor 2]**
[...]

**[Adjacent tool]**
[What it does · why it's adjacent, not direct · what it reveals about adjacent user needs]

### Eureka moments

_If first-principles reasoning contradicted conventional wisdom during the market scan, name
those moments here. Each one is often an outsized opportunity. Skip the section if there
weren't any — don't invent a Eureka to fill space._

- **[One-line summary of the insight]** — [short explanation of what the market assumes and
  why this codebase is positioned to challenge it]

---

## Premises

_The premises confirmed with the user before ideation. Recording them makes the ideas'
logic auditable._

1. [statement] — agreed ✓
2. [statement] — agreed ✓
3. [statement] — agreed ✓

---

## Desired Outcome

_One or two sentences on what this tool should enable that it doesn't fully enable today.
This is what the ideas below are trying to advance._

---

## Opportunities → Solutions

_Each opportunity is an unmet need, pain point, or friction the product could address.
Under each, one or more solution ideas._

### Opportunity A: [one-line framing of the gap / pain / need]

#### 1. [Solution idea title] · _Tier: adjacent / expansion / moonshot_

[Idea content following `idea-template.md`]

#### 2. [Solution idea title] · _Tier: ..._

[...]

### Opportunity B: [...]

#### 3. [Solution idea title] · _Tier: ..._

[...]

### Opportunity C: [...]

#### 4. [Solution idea title] · _Tier: ..._

[...]

---

## Recommended Next Step

**Start with idea #[N]: [title].**

_One or two sentences on why this idea now — best risk/reward, most feasible, most urgent,
most differentiated, or most excitingly weird._

**Concrete build steps:**

1. [First tangible step — what to stand up, stub, or sketch]
2. [Second step]
3. [Third step]

**Then:** feed idea #[N] into `write-a-prd` using the tracer-bullet slice as the starting brief.
```
