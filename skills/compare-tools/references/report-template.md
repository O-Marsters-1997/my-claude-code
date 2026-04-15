# Report Template

Use this skeleton when writing `./ideas/YYYY-MM-DD-tech-<slug>.md`. Adapt headings as
the content demands — the goal is a report someone can scan in 5 minutes and act on.

---

```markdown
# Technology Selection: [Topic]
*Generated: [date] · Codebase: [project name or repo path]*

---

## Context

[One paragraph: codebase characterisation (language, runtime, what's already installed, what
the project does) + the confirmed use case. If grill-me ran, include 1–2 sentences summarising
the key constraints that shaped the shortlist — e.g. "self-hosted required, < 10k req/day,
Go primary language".]

---

## Priorities & Rubric

These priorities were aligned with the user before research. They determine the compatibility
scores below.

| Priority | Weight | What 0 looks like | What 100 looks like |
|----------|--------|-------------------|---------------------|
| [Priority 1] | [W1]% | [anchor] | [anchor] |
| [Priority 2] | [W2]% | [anchor] | [anchor] |
| [Priority 3] | [W3]% | [anchor] | [anchor] |
| **Total** | **100%** | | |

*Weights can be adjusted: if your priorities have shifted, re-run the arithmetic with new
weights — the raw scores in the comparison table are independent of weighting.*

---

## Options

### [Tool A]

[One-sentence summary of what it is and what it's for.]

- **Strengths**: [list]
- **Weaknesses**: [list]
- **License**: [MIT/Apache/GPL/commercial/etc.]
- **Last release**: [date or "active" if recent]
- **Docs**: [url]

---

### [Tool B]

[...]

- **Strengths**: [...]
- **Weaknesses**: [...]
- **License**: [...] | **Last release**: [...] | **Docs**: [url]

---

### [Tool C]

[...]

---

## Comparison Table

Rows = tools. Columns = each priority (with weight) + final weighted score. Raw scores are
0–100 on each priority; the final column is the weighted sum divided by 100.

| Tool | [Priority 1] ([W1]%) | [Priority 2] ([W2]%) | [Priority 3] ([W3]%) | Score /100 |
|------|----------------------|----------------------|----------------------|-----------|
| [Tool A] | [score] | [score] | [score] | **[total]** |
| [Tool B] | [score] | [score] | [score] | **[total]** |
| [Tool C] | [score] | [score] | [score] | **[total]** |

*Arithmetic: Score = (P1_score × W1 + P2_score × W2 + P3_score × W3) / 100*

---

## Ranked Recommendations

> **Top recommendation: [Tool X] ([score]/100)**

### 1. [Tool X] — [score]/100

[2–3 sentences: why this tool ranked first. Tie the score directly back to the codebase
facts and the user's stated priorities. E.g. "BullMQ scored highest on community support
(75/100) and ecosystem fit (90/100) because the project is Node.js-based and BullMQ is
the de-facto standard in the Node job queue space — it's already in thousands of
production deployments and integrates natively with the existing Redis instance noted
in docker-compose.yml."]

### 2. [Tool Y] — [score]/100

[2–3 sentences: what this tool does well, where it fell short, and under what conditions
it would become the top recommendation — e.g. "would win if the team weighted cost over
community support".]

### 3. [Tool Z] — [score]/100

[2–3 sentences.]

---

**Next step:** feed this decision into `write-a-prd` or open a design document to capture
the integration approach, migration plan, and any open questions (configuration, hosting,
monitoring).
```
