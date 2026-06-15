---
name: architecture-review
description: Technology-agnostic architectural assessment — boundaries, coupling, data flow, failure modes, information hiding. An orchestrator maps the system, fans out six parallel reviewer subagents (one per quality dimension), and synthesises a consolidated report ending in a prioritised recommendations roadmap. Use when the user wants an architecture review, a design health check, a staff-level read on architectural risk, or wants to assess coupling and cohesion. Trigger on phrases like "review my architecture", "assess this system's design", "is this codebase well-architected", "architecture health check", "evaluate coupling and cohesion", "where's the design risk", or "design review before we scale". Diagnoses only — does not refactor or apply fixes.
---

# Architecture Review

Technology-agnostic architectural assessment. Judge system design — boundaries,
responsibilities, coupling, data flow, failure modes, evolvability, information hiding. Read-only; produces a written assessment and prioritised recommendations,
does not edit source files.

## Capabilities

A staff-engineer reviewer operating this skill:

- Identifies the architectural style in use and evaluates whether it fits the system's actual constraints
- Names patterns precisely — where they are applied correctly, misapplied, or conspicuously absent
- Traces data flow and transaction/consistency boundaries across service and module edges
- Locates resilience and scalability ceilings (missing circuit breakers, unbounded queues, stateful services)
- Maps trust boundaries and identity/secret flow; flags where authentication, authorisation, or secret handling is absent or porous
- Assesses module depth, coupling direction, cohesion, and information-hiding discipline
- Surfaces genuine quality-attribute trade-offs rather than treating all dimensions as simultaneously maximisable
- Grounds every finding in code evidence; calls out strengths, not only problems

The vocabulary for naming patterns precisely lives in
[`references/knowledge-base.md`](references/knowledge-base.md).
The per-dimension rubric (what good looks like, checklists, anti-patterns) lives in
[`references/dimensions.md`](references/dimensions.md).

## Process

### Phase 1 — Discover and map

Use the Agent tool with `subagent_type=Explore` to build a **system map**:

- Entry points and how the system is driven
- Top-level modules / services and their responsibilities
- Major boundaries and seams
- Data flow through a representative request or job
- State, persistence, and ownership
- External dependencies
- Deployment / runtime shape

Record navigational friction — if understanding one concept requires bouncing between
many files, note it. The system map is both this phase's output and the shared context
for all reviewer subagents. Do not grade yet.

### Phase 2 — Fan out per-dimension reviewers

Spawn **six reviewer subagents in parallel** — one message, six Agent calls,
`subagent_type=Explore`. Each owns one lens:

1. **Simplicity & understandability**
2. **Maintainability**
3. **Extensibility**
4. **Security**
5. **Performance & scalability**
6. **Modularity** (deep modules + coupling & cohesion)

Read [`references/dimensions.md`](references/dimensions.md) and
[`references/knowledge-base.md`](references/knowledge-base.md). For each reviewer, paste
(a) their dimension's section from `dimensions.md` and (b) the knowledge-base domains
listed for their lens in the **Lens → domains mapping** table at the end of
`knowledge-base.md`. Subagents do not share your context.

**Subagent prompt template** — fill the five `{{...}}` slots per reviewer:

```
You are a staff engineer performing a focused architecture review of the codebase at
{{REPO_PATH}}. Your lens: {{DIMENSION NAME}}.

Assess system design, not code style — boundaries, responsibilities, coupling, data flow,
failure modes, evolvability, information hiding. Ignore language idioms, formatting, and
naming unless they reveal a structural problem. Technology-agnostic.

System map (your ground truth — read the real code to confirm or challenge it):
{{SYSTEM MAP}}

Rubric for your lens (definition, what good looks like, checklist, anti-patterns):
{{PASTE THIS DIMENSION'S SECTION FROM references/dimensions.md}}

Architectural vocabulary relevant to your lens:
{{PASTE THE KNOWLEDGE-BASE DOMAINS FOR THIS LENS FROM references/knowledge-base.md}}

Name patterns from this vocabulary precisely where they apply, are misapplied, or are
conspicuously absent. Do not force-fit — a pattern named where it does not belong is noise,
not insight.

Method: explore the codebase read-only. Ground every finding in concrete evidence —
cite file paths (and line ranges where useful). Distinguish durable architectural issues
from cosmetic ones. Call out genuine strengths, not only problems.

Return narrative prose, no scores or severity tags. Structure:
- 2–3 sentence overall read for this dimension (including what's working well).
- Findings list. For each: (a) what, (b) where — file-path evidence, (c) why it matters
  architecturally, (d) a direction for remediation (direction, not implementation).
If the system is healthy on your lens, say so plainly.
```

### Phase 3 — Synthesise

Collect the six reports:

- **Dedupe.** The same root issue often surfaces across multiple lenses. Merge it.
- **Surface cross-cutting themes.** Where several lenses independently land on the same
  subsystem or boundary, that convergence is the highest-signal result — lead with it.
- **Rank by impact.** Be opinion-forward. Name genuine strengths too.

### Phase 4 — Deliver

Write the report to `./docs/architecture-reviews/<descriptive-name>.md` (create the
directory if needed), following
[`references/report-template.md`](references/report-template.md). The directory allows
multiple reviews over time.
