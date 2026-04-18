# Feedback ledger — <skill-name>

This file is a **template only**. Real ledgers live next to each opted-in skill's SKILL.md, never inside the skill-feedback-collector directory. When creating a ledger for a participating skill, copy this file to `<skill-dir>/skill-feedback-ledger.md` and update the heading.

Each entry below is standalone. Do not edit or cross-reference prior entries — the ledger is an append-only log.

---

## 2026-01-15T14:32:00+00:00 — example-skill vunversioned

**Summary**
The user was batch-converting markdown docs and found the file discovery reliable, but the skill missed an early context signal and asked a redundant clarifying question mid-run. Overall the skill covered the core task but created friction around context awareness.

**Selected proposals**
- [high | instructions] Add a precondition check: before asking any clarifying question, scan the full conversation for an existing answer and skip the question if found — expected impact: eliminates repeat questions when the user has already provided the answer upfront.
- [medium | examples] Add a before/after example showing how to handle a partially-specified input (hidden files case) without prompting — expected impact: anchors the no-redundant-prompt rule for edge cases the instructions currently leave ambiguous.

**Discarded proposals**
- [low | structure] Move the file-filter options to a reference file to reduce SKILL.md length — expected impact: minor readability improvement, not worth splitting at this size.
