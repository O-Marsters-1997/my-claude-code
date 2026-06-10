# Feedback ledger — teach

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

---

## 2026-06-10T10:59:11+01:00 — teach unversioned

**Summary**
The user has been using /teach over multiple sessions to learn chess openings and is largely happy (75/100): the multi-file workspace structure fits their learning method, the in-browser interactivity lands well, and the recently-reliable Drill section is a clear win. Friction is concentrated in three areas. (1) Embedded videos still won't play from local file:// lessons despite existing embed guidance, forcing a manual YouTube hop. (2) The end-of-lesson "what's next" is too open-ended — they want a concrete default next lesson proposed every time. (3) Biggest: lessons are created strictly one-at-a-time, which is slow, token-heavy, and high-friction; they want an upfront course curriculum (provisional table of contents with per-lesson summaries and a depth choice), then lesson 1 built first as a reference template, confirmed, after which remaining lessons are batch-created to match — with feedback that updates NOTES.md and revises lessons together (their chess-board styling was wrong in lesson 1 and ideally fixed once for all). They also flagged a meta-constraint: any update to teach must preserve its leanness rather than bloat it.

**Selected proposals**
- [high | tools] Serve lessons over a local HTTP server (e.g. `python3 -m http.server` in ./lessons, open http://localhost:PORT/...) instead of opening file:// directly, so embedded videos have a real origin and play inline; keep the fallback link. Also add an easy invocable command to open a lesson by its number in the series (e.g. `teach open 3`) that starts the server and opens that lesson. — expected impact: fixes the dead-video-on-file:// problem and makes reopening any lesson a one-liner.
- [high | structure] Add a curriculum phase on first /teach for a subject: produce a provisional CURRICULUM.md table of contents (one-line summary per lesson) and offer ~3 depth levels (overview = fewer, grouped, quicker / deeper = more, tightly-scoped) that drive lesson count and scope; user confirms or amends before lessons begin. — expected impact: gives the course a shape up front, removes per-lesson planning overhead, sets up batch creation.
- [high | instructions] After curriculum confirmation, build lesson 1 alone as the reference template; add an explicit in-skill step/command to confirm lesson 1's structure is correct, then batch-create the remaining lessons to that template in one pass. — expected impact: cuts the gap between lessons, is more token-efficient, lets styling fixes (e.g. the chess board) be made once and inherited.
- [high | instructions] Replace the open-ended end-of-lesson "what's next" with a mandatory concrete default: always name the specific next lesson that will be created unless the user amends it; leave room to redirect but never leave it unspecified. — expected impact: eliminates open-ended drift; accept-by-default or correct.
- [medium | instructions] Add a feedback-and-revise mechanism over the batch-created lessons: capture lesson feedback into NOTES.md (as today) and allow the generated lessons to be edited together so a single correction propagates. — expected impact: makes batch creation safe — early/template mistakes fixable across the whole set.
- [high | structure] Bake a leanness constraint into the skill-updater handoff: integrate changes by editing/condensing existing sections in place (not appending), delete redundant guidance, and push detailed/expandable material into references/ rather than SKILL.md. — expected impact: keeps teach lean and well-structured as it gains the above capabilities.

**Discarded proposals**
- (none — all surfaced proposals were selected)
