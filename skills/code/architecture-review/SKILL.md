---
name: architecture-review
description: Critical, evidence-grounded architectural assessment — boundaries, coupling, data flow, failure modes, information hiding. An orchestrator maps the system, detects the stack, fans out seven parallel reviewer subagents (one per quality dimension), and synthesises a consolidated report with an explicit verdict and a prioritised recommendations roadmap. The vocabulary is technology-agnostic but the probes adapt to the stack. Use when the user wants an architecture review, a design health check, a staff-level read on architectural risk, or wants to assess coupling and cohesion. Trigger on phrases like "review my architecture", "assess this system's design", "is this codebase well-architected", "architecture health check", "evaluate coupling and cohesion", "where's the design risk", or "design review before we scale". Diagnoses only — does not refactor or apply fixes.
---

# Architecture Review

Critical architectural assessment. Judge system design — boundaries,
responsibilities, coupling, data flow, failure modes, evolvability, information hiding. Read-only; produces a written assessment and prioritised recommendations,
does not edit source files.

**Stance.** This is an adversarial review, not a balanced scorecard. Default to
scepticism: hunt for what is wrong, weak, or risky. Do not manufacture strengths for the
sake of balance — praise is earned, not owed, and a dimension (or the whole system) is
allowed to come back net-negative. Counterbalance: stay high-signal. A long list of minor
findings is noise; surface the issues that actually matter and go *deep* on each — root
cause, blast radius, concrete evidence — rather than broad and shallow.

**Agnostic vocabulary, stack-specific probes.** The architectural vocabulary is
technology-agnostic (a god object is a god object in any language). Formatting, naming, and
language idioms stay out of scope. But the *probes* adapt to the stack: a frontend has
architectural concerns (design-system/theming architecture, UI/UX consistency, component-API
consistency) that a generic review misses. See
[`references/domain-probes.md`](references/domain-probes.md).

## Capabilities

A staff-engineer reviewer operating this skill:

- Identifies the architectural style in use and evaluates whether it fits the system's actual constraints
- Names patterns precisely — where they are applied correctly, misapplied, or conspicuously absent
- Traces data flow and transaction/consistency boundaries across service and module edges
- Locates resilience and scalability ceilings (missing circuit breakers, unbounded queues, stateful services)
- Maps trust boundaries and identity/secret flow; flags where authentication, authorisation, or secret handling is absent or porous
- Assesses module depth, coupling direction, cohesion, and information-hiding discipline
- Surfaces genuine quality-attribute trade-offs rather than treating all dimensions as simultaneously maximisable
- Grounds every finding in code evidence — and, where the repo provides it, in measured evidence (bundle reports, dependency graphs, coverage, profiles), not reasoning from reading alone
- Is willing to deliver a net-negative verdict; reserves praise for what is genuinely well-built

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

Also **detect the stack** (frontend, backend service, data/ML, infra/IaC, library, or a
mix) and note which build/analysis tooling the repo already provides — bundlers, dependency
analysers, test/coverage runners, profilers, linters with architectural rules. Both feed
Phase 2: the stack selects which stack-specific probes to inject (see
[`references/domain-probes.md`](references/domain-probes.md)), and the tooling list tells
reviewers what they can measure rather than estimate.

Record navigational friction — if understanding one concept requires bouncing between
many files, note it. The system map is both this phase's output and the shared context
for all reviewer subagents. Do not grade yet.

### Phase 2 — Fan out per-dimension reviewers

Spawn **seven reviewer subagents in parallel** — one message, seven Agent calls,
`subagent_type=Explore`. Each owns one lens:

1. **Simplicity & understandability**
2. **Maintainability**
3. **Extensibility**
4. **Security**
5. **Performance & scalability**
6. **Modularity** (deep modules + coupling & cohesion)
7. **Deployability** (release coupling, CI/build-matrix coupling, operational blast radius)

Read [`references/dimensions.md`](references/dimensions.md) and
[`references/knowledge-base.md`](references/knowledge-base.md). For each reviewer, paste
(a) their dimension's section from `dimensions.md`, (b) the knowledge-base domains
listed for their lens in the **Lens → domains mapping** table at the end of
`knowledge-base.md`, and (c) any stack-specific probes for their lens from
[`references/domain-probes.md`](references/domain-probes.md) matching the stack detected in
Phase 1. Subagents do not share your context.

**Measure, don't estimate.** `Explore` agents can run read-only shell commands. Where Phase
1 found analysis tooling, tell the relevant reviewer to run it and cite the output — e.g. a
bundle/size report for Performance, a dependency-graph/cycle tool for Modularity, a coverage
report for Maintainability. This is stack-agnostic: use whatever the project already
provides. Prefer measured numbers over reasoning from reading; say so plainly when a claim is
an estimate because no tooling existed.

**Subagent prompt template** — fill the five `{{...}}` slots per reviewer:

```
You are a staff engineer performing a focused architecture review of the codebase at
{{REPO_PATH}}. Your lens: {{DIMENSION NAME}}.

Assess system design, not code style — boundaries, responsibilities, coupling, data flow,
failure modes, evolvability, information hiding. Formatting, naming, and language idioms are
out of scope unless they reveal a structural problem — but stack-specific *architectural*
concerns (e.g. for a frontend: design-system/theming architecture, UI/UX consistency,
component-API consistency) are in scope.

System map (your ground truth — read the real code to confirm or challenge it):
{{SYSTEM MAP}}

Rubric for your lens (definition, what good looks like, checklist, anti-patterns):
{{PASTE THIS DIMENSION'S SECTION FROM references/dimensions.md}}

Architectural vocabulary relevant to your lens:
{{PASTE THE KNOWLEDGE-BASE DOMAINS FOR THIS LENS FROM references/knowledge-base.md}}

Stack-specific probes for your lens (skip any that don't apply to this stack):
{{PASTE THIS LENS'S PROBES FOR THE DETECTED STACK FROM references/domain-probes.md, or "none"}}

Available measurement tooling (run it and cite the output where it strengthens a finding):
{{TOOLING FOUND IN PHASE 1 RELEVANT TO THIS LENS, or "none found — reason from the code and say so"}}

Name patterns from this vocabulary precisely where they apply, are misapplied, or are
conspicuously absent. Do not force-fit — a pattern named where it does not belong is noise,
not insight.

Method: explore the codebase read-only (you may run read-only shell commands and the
tooling above). Ground every finding in concrete evidence — cite file paths (and line ranges
where useful) and measured output where you have it.

Stance: be critical. Your job is to find what is wrong, weak, or risky on this lens, not to
reassure. Do NOT invent strengths for balance — mention something positive only if it is
genuinely well-built and load-bearing to the assessment. Stay high-signal: a pile of minor
nitpicks is noise. Surface the issues that matter and go deep on each (root cause, blast
radius), rather than listing many shallow ones.

Return narrative prose, no scores or severity tags. Structure:
- A 2–3 sentence overall verdict for this dimension — blunt, and allowed to be net-negative
  (e.g. "this dimension is in poor health because…"). Only note strengths here if they are real.
- Findings list, ranked most-serious first. For each: (a) what, (b) where — file-path/measured
  evidence, (c) why it matters architecturally, (d) a direction for remediation (direction,
  not implementation).
If the system is genuinely healthy on your lens, say so plainly — but only if the evidence
supports it, not as a default.
```

### Phase 3 — Synthesise

Collect the seven reports:

- **Dedupe.** The same root issue often surfaces across multiple lenses. Merge it.
- **Surface cross-cutting themes.** Where several lenses independently land on the same
  subsystem or boundary, that convergence is the highest-signal result — lead with it.
- **Rank by impact.** Be opinion-forward and critical. The synthesis carries an explicit
  overall verdict that may be net-negative — do not soften it with manufactured positives.
  Note strengths only where they are genuine and material to the verdict.

### Phase 4 — Deliver

Write the report to `./docs/architecture-reviews/<descriptive-name>.md` (create the
directory if needed), following
[`references/report-template.md`](references/report-template.md). The directory allows
multiple reviews over time.
