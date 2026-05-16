---
name: chat-to-approach
description: Convert a pasted AI conversation summary or transcript into a canonical alignment document at ./docs/approach.md — the single source of truth for what we are building and why. Grills for alignment on ambiguities and conflicts, then writes or merges the result. If approach.md already exists, reads it first and only grills on deltas and conflicts, not settled decisions. Downstream skills (approach-to-roadmap, prd-to-plan, prd-to-issues) consume this doc without re-grilling. Use this skill when the user pastes a summary and wants to capture alignment, extract decisions, write an approach doc, or record what was agreed on. Trigger on phrases like "turn this chat into an approach", "extract our alignment", "write an approach doc from this", "what did we decide (save it)", "convert this summary into approach.md", "capture our decisions", or "create an alignment doc". Use it even if the user just pastes a conversation summary and says "save this" or "let's capture this".
---

## Overview

Parse a pasted AI conversation summary into a structured, persistent alignment document: `./docs/approach.md`. This is the canonical record of **what we are building and why** — agreed features, constraints, principles, and open questions. It is not a roadmap (no priority tiers) and not a PRD (no implementation detail). It is the alignment checkpoint between ideation and execution.

If `./docs/approach.md` already exists, it is treated as prior settled truth. Only net-new items and conflicts require grilling — not things the doc already covers.

Once this doc exists, run `approach-to-roadmap` to generate a kanban board without repeating the grilling.

Solo/indie-builder scale throughout.

---

## What Is a Feature (at Alignment Level)?

A feature is a coherent, user-facing capability that delivers value on its own. It is not an implementation task, a ticket, or a technical decision.

**Anatomy of an alignment-level feature:**

| Field            | What it captures                                  | Example                                                      |
| ---------------- | ------------------------------------------------- | ------------------------------------------------------------ |
| **Name**         | Short, user-facing label — what it does, not how  | "User authentication"                                        |
| **Problem**      | The user pain or need it addresses (one sentence) | "Users need to log in securely to access their data"         |
| **Value**        | What it unlocks or enables                        | "Unblocks all personalisation and data persistence features" |
| **Dependencies** | Other features that must exist first              | "Requires: database setup"                                   |

Only Name + Problem are required. Value and Dependencies are included when they matter for understanding what we're building.

**How to group chat content into a feature:**

- A feature = a cluster of related decisions that together describe one user-facing capability
- Multiple small decisions about the same area → one feature, not many entries
- Ask: "Could a user describe what they'd get from this in one sentence?" If yes, it's a feature
- Distinguish features from:
  - _Constraints_ — architectural or technical decisions that shape features but aren't features themselves (e.g. "use Postgres") — record in Constraints, not Features
  - _Tasks_ — implementation steps inside a feature (e.g. "wire up the ORM") — not captured here; `prd-to-issues` handles this later
  - _Principles_ — values the user expressed ("keep it simple", "mobile-first") — capture in Principles, not Features

**Anti-patterns to avoid:**

- Don't split one feature into multiple entries because the conversation touched it in different ways
- Don't create an entry for every distinct sentence — synthesise first
- Don't mistake a technical decision for a feature ("use JWT" is a constraint; "user sessions" is a feature)

---

## Phase 1 — Parse the Conversation

Read the pasted transcript or summary carefully. Classify everything into four buckets:

**Decided/committed** — Both parties agreed and committed to something. Look for language like "we'll do X", "let's go with Y", "agreed: Z", "the plan is", "I'll build", "we'll ship". These become candidates for approach.md.

**Discussed but not resolved** — Ideas that came up, were explored, but didn't reach a clear conclusion. Do not silently drop these. Surface them for grilling in Phase 3. **An idea not explicitly accepted is not the same as an idea that was rejected.**

**Explicitly discarded** — Ideas the user or conversation actively rejected or dismissed in favour of something else. Look for language like "no, let's not", "actually scratch that", "we're not doing X", "that's out of scope". These go into the "Out of scope" section of the doc, not into Features.

**Open questions** — Things that need a decision before work can proceed. These go into the "Open questions" section verbatim.

**Default to inclusion over exclusion.** When in doubt, treat something as either a candidate feature (pending grilling) or an open question. An approach doc that surfaces ambiguity honestly is better than one that silently drops things the user cared about.

---

## Phase 2 — Load Existing approach.md (If Present)

Check for an existing `./docs/approach.md`. Two paths:

**No existing file:** proceed directly to Phase 3 with all "decided/committed" items as candidates.

**Existing file found:** read it fully, then compute the delta against the parsed chat:

- **Net new** — features, constraints, or principles in the chat that aren't yet in the doc → grill if ambiguous, otherwise add directly
- **Conflicts** — the chat contradicts something the doc already records (e.g. doc says JWT, chat says sessions) → always grill; present both versions
- **Supersedes** — the chat clearly updates or replaces a prior decision → treat as conflict requiring explicit confirmation before updating
- **Already covered** — the chat repeats what the doc already has → ignore; do not re-grill

Items already settled in the existing doc are **treated as prior truth**. Do not re-open them unless the new chat explicitly contradicts them.

---

## Phase 3 — Alignment Grilling (Safeguard)

Before writing anything, run a focused grilling session using the `grill-with-docs` skill. This is the safeguard: stress-test the extraction before committing it to a durable file.

Grill only on:

- Items in the "discussed but not resolved" bucket
- Conflicts between the chat and the existing doc
- Items with ambiguous scope ("improve X", "think about Y") — is there a concrete feature here?
- Items where it's unclear whether something is a feature, constraint, or principle
- Items that could be the same feature described differently at two points in the conversation

Do NOT grill on:

- Items clearly decided in the conversation
- Items already settled in the existing approach.md (unless the new chat conflicts with them)
- Items that are clearly constraints, principles, or out-of-scope — classify and move on

**How to ask good grilling questions:**

- Lead with what you found: "You discussed two auth approaches — JWT and sessions — but I couldn't tell which was chosen."
- Give your best read: "My read is you landed on JWT because of the mobile client. Is that right?"
- For discussed-but-not-resolved items: "You mentioned analytics a couple of times but didn't land on a decision — was this cut, deferred, or still live?"
- For conflicts with the existing doc: "The doc currently records X, but this conversation suggests Y. Did you change your mind, or am I misreading?"

One question at a time. Keep grilling lean — only ask if the answer would meaningfully change what ends up in the doc. If the conversation was explicit and decisive throughout, tell the user and skip straight to Phase 4.

---

## Phase 4 — Write / Merge ./docs/approach.md

Write or update the file. If `./docs/` doesn't exist, create it.

**When creating a new file:** write all sections, even if some are empty (e.g. "No constraints recorded yet.").

**When updating an existing file:** only replace the content of sections that changed. Preserve sections untouched by the new chat verbatim. Update the "Last updated" line.

### Document structure

```markdown
# Approach

_Last updated: <YYYY-MM-DD> · Canonical alignment doc — what we are building and why._

## Problem

<One paragraph: who has the pain, what the pain is, why it matters now.>

## Goals

- <Outcome-oriented goal, one per line>

## Non-goals

- <Explicitly out of scope, one per line — so they don't re-surface in future chats>

## Features

Each feature is a user-facing capability. No priority tiers — use approach-to-roadmap for Now/Next/Later.

### <Feature name>

- **Problem:** <one sentence — the user pain this solves>
- **Value:** <what it unlocks or enables>
- **Dependencies:** <other features required first, if any>

## Constraints

Technical and architectural decisions that shape the work but aren't features.

- <e.g. Use Postgres for persistence>
- <e.g. JWT-based auth, stateless sessions>

## Principles

Values and preferences that should shape all decisions.

- <e.g. Keep it simple — no enterprise PM overhead>
- <e.g. Ship weekly — no multi-month batches>

## Open questions

Things that still need a decision before downstream work can proceed.

1. <Question>? — <One line on why it matters or what's blocked on it.>

## Out of scope

Ideas explicitly rejected in conversation, recorded so they don't resurface.

- <Rejected idea> — <brief note on why it was cut>
```

**What does NOT belong here:** Now/Next/Later tiers, story points, assignees, implementation tasks, ticket breakdowns, acceptance criteria. These are downstream concerns.

---

## Finishing Up

After writing the file, send a brief receipt (3–5 lines max):

- Number of features captured
- Number of conflicts resolved during grilling (if any)
- Number of open questions outstanding
- A one-liner on anything placed in "Out of scope" — so the user knows you didn't silently drop things they cared about
- Reminder: run `approach-to-roadmap` to generate a visual kanban from this doc without repeating grilling
