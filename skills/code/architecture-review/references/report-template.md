# Report Template

Written to `./docs/architecture-reviews/<descriptive-name>.md`. Narrative and
evidence-grounded — no numeric scores, no severity tags. Triage lives in the
recommendations section, expressed as ranking and prose.

```markdown
# Architecture Review — [project / system name]

_Read-only assessment. Date: [YYYY-MM-DD]. Scope: [what was and wasn't covered]._

## Executive summary

The 2–3 biggest architectural risks, in plain language. What they are, why they matter.
One or two sentences on what the system does *well* architecturally — so deliberate
design is distinguishable from accident.

## System map

- **Entry points** — what drives the system.
- **Modules / services** — each with its responsibility in one line.
- **Boundaries & data flow** — how a representative request or job moves through.
- **State & persistence** — what's stored and who owns it.
- **External dependencies** — owned services vs. third parties.
- **Runtime shape** — monolith / services / serverless.

## Findings by dimension

One subsection per lens. Short overall read (including what's healthy), then findings.
Each finding: **what**, **where** (file-path evidence), **why it matters** architecturally,
**direction** for remediation. If a dimension is healthy, say so plainly.

### Simplicity & understandability
### Maintainability
### Extensibility
### Security
### Performance & scalability
### Modularity (deep modules + coupling & cohesion)

## Cross-cutting themes

Where several lenses independently landed on the same subsystem or boundary — name the
convergence. These themes usually drive the top of the roadmap.

## Prioritised recommendations

Ranked — most architectural impact first. For each:
- **What to change** and the outcome it buys.
- **Why it's ranked here** — cost of leaving it.
- **Next step** — concrete direction; the skill does not apply fixes.
```
