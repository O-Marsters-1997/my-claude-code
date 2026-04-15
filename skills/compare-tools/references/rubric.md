# Rubric Design

This file explains how to translate user priorities into a weighted scoring rubric and
how to score each tool on each priority so the final compatibility number is auditable.

---

## Step 1: Assign weights

The user picks 3–5 priorities and ranks them. Convert the ranking into weights that sum to 100.

### Default distributions

| # priorities | Suggested weights (rank 1 → rank N) |
|---|---|
| 3 | 50 / 30 / 20 |
| 4 | 40 / 30 / 20 / 10 |
| 5 | 35 / 25 / 20 / 12 / 8 |

These are defaults. Adjust when the user signals a dominant priority (e.g. "license is
non-negotiable" → give it 50+; redistribute the rest).

**Always state the weights explicitly in the report.** The user should be able to see
immediately how much each priority influenced the final score.

---

## Step 2: Write scoring anchors per priority

For each priority, define what a score of 0, 50, and 100 looks like. This makes scoring
consistent and explainable. You don't need to use exact values — interpolate.

### Example anchors

**Community Support**
- 100: >10k GitHub stars, weekly or monthly releases, active Discord/forum, well-covered on
  Stack Overflow, multiple conference talks per year
- 50: Healthy but niche — a few thousand stars, responsive maintainers, some tutorials,
  most common questions answered somewhere
- 0: Abandoned or effectively unmaintained — last release >2 years ago, unanswered issues,
  no community forum

**Maintenance Activity**
- 100: Multiple active maintainers, regular releases (< 3 months between releases), issues
  triaged promptly, active roadmap or changelog
- 50: One or two maintainers, semi-regular releases (< 12 months), issues eventually
  addressed
- 0: Single maintainer inactive, no releases in >2 years, issues ignored

**Avoiding Vendor Lock-in**
- 100: Fully open-source, standard protocols (SQL, HTTP, S3-compatible), data is portable,
  no proprietary SDK required
- 50: Open-source core but hosted offering adds proprietary features; migration possible but
  non-trivial
- 0: Proprietary format, closed-source, no migration path without full rewrite

**Extensibility / Composability**
- 100: First-class plugin or middleware system, hooks at every lifecycle point, well-documented
  extension points, examples of community extensions
- 50: Some extension points but limited; workarounds exist for most customisation needs
- 0: Monolithic — customisation requires forking or monkey-patching internals

**Ecosystem Fit**
- 100: Official SDK in the project's language, idiomatic API that matches the project's
  existing patterns, compatible with frameworks already in use
- 50: Community SDK available, or the official SDK exists but requires some adaptation
- 0: No SDK in the project's language; requires HTTP calls or a polyglot shim

**License**
- 100: MIT or Apache 2.0 — maximally permissive, compatible with commercial closed-source use
- 50: LGPL or MPL — permissive for libraries but read the fine print; usually fine for
  application use
- 0: GPL (copyleft — may require open-sourcing the app), AGPL (same + network use),
  or commercial license required for production use

**Performance**
- 100: Benchmarks show it leads the field for this use case; low latency, high throughput,
  efficient resource usage
- 50: Performance is adequate for the expected load; not the fastest, but not the bottleneck
- 0: Known performance issues at the expected scale; would require significant tuning or
  horizontal scaling to compensate

**Learning Curve**
- 100: Excellent docs with tutorials and examples, simple API, mental model matches the
  team's existing knowledge
- 50: Docs exist but are incomplete or dense; some trial-and-error required
- 0: Poor or no docs, steep conceptual learning curve, requires expert knowledge to use correctly

**Cost**
- 100: Free and open-source with no usage-based pricing; self-hostable at near-zero cost
- 50: Generous free tier covers expected usage; paid tier reasonably priced for growth
- 0: No free tier for production use, or usage-based pricing that scales uncomfortably with
  expected load

**Operational Burden**
- 100: Fully managed, no infrastructure to run; or trivially self-hosted with zero-config
- 50: Self-hostable with moderate ops work; managed option available for a fee
- 0: Complex to operate — requires dedicated ops knowledge, custom monitoring, manual failover

**Maturity / Production-Proven**
- 100: Used in production at scale by multiple well-known organisations; stable API (v1+);
  failure modes well-documented
- 50: Used in production but at smaller scale; v1+, but less public evidence of large-scale
  deployments
- 0: Pre-v1, experimental, or limited production evidence; unknown failure modes

---

## Step 3: Score each tool

For each tool on each priority:
1. Read your research notes
2. Match against the 0 / 50 / 100 anchors and interpolate
3. Write a brief evidence note (1 sentence) explaining the score

**Example:**
```
BullMQ — Community Support: 75 (active, ~4k stars, regular releases, but smaller Discord
than alternatives)
```

---

## Step 4: Compute the compatibility score

```
score = Σ (weight_i × raw_score_i) / 100
```

Show this arithmetic in the comparison table so the user can audit it. Round to the
nearest integer.

**Example (3 priorities, weights 50/30/20):**

| Tool | Community (50) | Maintenance (30) | Ecosystem (20) | Score /100 |
|------|---------------|-----------------|---------------|-----------|
| BullMQ | 75 | 80 | 90 | (75×50 + 80×30 + 90×20) / 100 = **79** |
| Bee Queue | 40 | 35 | 70 | (40×50 + 35×30 + 70×20) / 100 = **45** |

---

## Presenting the rubric in the report

Always include:
1. The weights table (priority → weight)
2. The scoring anchors for each priority (summarised, not the full definitions above)
3. The comparison table with per-priority scores and the final weighted total
4. A note that the user can adjust weights if their priorities have shifted

The goal is that a sceptical reader can look at the numbers and say "yes, I can see why
it got that score."
