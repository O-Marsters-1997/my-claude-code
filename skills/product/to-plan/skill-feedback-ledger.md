# Feedback ledger — prd-to-plan

Each entry below is standalone. Do not edit or cross-reference prior entries — the ledger is an append-only log.

---

## 2026-06-14T00:00:00+00:00 — prd-to-plan vunversioned

**Summary**
Feedback session arising from a pipeline-boundary review across write-a-prd, prd-to-plan, and plan-to-issues. The core finding: because the PRD was updated to be strictly product-language-only (no technical design), the plan must compensate by owning the technical design firmly. The old prd-to-plan produced plans that were too abstract — phases described behaviour at a product level rather than specifying data models, contracts, and module boundaries concretely. This made plan-to-issues difficult: the new skill needs to re-derive the design from a plan that should already contain it. Two proposals were applied to close this gap.

**Selected proposals**
- [high | instructions] Reframe prd-to-plan as the technical design document: state explicitly that it owns the HOW the PRD leaves out, and expand the architectural decisions section to capture data models, schema shapes, API/interface contracts, module boundaries, integration points, and key flows — concretely enough that a developer can derive implementation tickets without re-deriving the design.
- [medium | instructions] Reconcile the volatility rule in `<vertical-slice-rules>`: distinguish between durable technical design (schema shapes, model/module names, route paths, contract shapes — INCLUDE) and volatile line-level detail (exact internal function names, file paths likely to churn — AVOID). The old rule said "do NOT include implementation details" without this distinction, which contradicted the goal of being ticket-ready.

**Discarded proposals**
(none)
