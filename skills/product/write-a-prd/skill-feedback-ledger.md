# Feedback ledger — write-a-prd

Each entry below is standalone. Do not edit or cross-reference prior entries — the ledger is an append-only log.

---

## 2026-06-14T00:00:00+00:00 — write-a-prd vunversioned

**Summary**
Audit-based feedback session on write-a-prd, conducted without running the skill live. The skill was reviewed against a known failure: the last PRD produced contained Go code, leaking technical specification into a document that should be purely product-language. Four candidate improvements were surfaced — three focused on product-boundary enforcement and interaction quality, one on discoverability of downstream skills. The user selected four proposals (including the technical-boundary fix added during review) and declined the handoff-to-next-skills proposal. Changes were applied inline as part of the same session.

**Selected proposals**
- [high | instructions] Add a product-only boundary principle near the top: PRD stays outcome/behaviour-only, and reframe the `Implementation Decisions` section so schema/API/architecture decisions explicitly defer to the plan — stops technical leakage (e.g. Go code appearing in a PRD).
- [high | instructions] Add a skip-if-answered precondition: before asking any question, scan the full conversation for an existing answer and skip if found — eliminates repeat questions when context is already in the conversation.
- [medium | instructions] Expand step 5 with concrete `gh issue create` mechanics: title convention `[PRD] <Feature Name>`, confirm with the user before creating, use `--body-file`, print the resulting issue URL — removes ambiguity about how the PRD is "submitted as a GitHub issue".
- [low | structure] Append the feedback opt-in line, wiring write-a-prd into skill-feedback-collector at the end of a run.

**Discarded proposals**
- [low | instructions] End-of-skill handoff pointing to prd-to-plan / prd-to-issues as next steps — declined by user.
