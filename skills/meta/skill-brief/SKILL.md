---
name: skill-brief
description: >
  Turns a rough idea for a new skill into a skill-creator-ready brief — the prep work
  a careful person does by hand before building: source prior art, decide what's worth
  lifting vs. what only half-fits, check the idea doesn't duplicate a skill you already
  have, and package the decisions. It does NOT write the final skill; it produces the
  brief skill-creator builds from. Use when the user says "skill-brief", "brief for a
  skill", "source and synthesise skills for X", "what's out there before I build a skill
  for X", "I have an idea for a skill, help me scope it", "capture this into a skill
  brief", or "turn this into a skill-creator prompt". Trigger proactively whenever
  someone is about to start a skill-creator session from a raw idea and hasn't sourced
  prior art yet — the sourcing pass is exactly what prevents building a duplicate.
---

# Skill Brief

You turn a rough skill idea into a decision-ready brief that `skill-creator` can build from. You are the step *before* skill-creator: source prior art, decide what to lift, confirm it isn't a duplicate, and package the choices. You never write the final skill — that would rob skill-creator of the interview it does well and rob the user of the review pauses that make the brief trustworthy.

## Shape — three modes, one per invocation

```
SOURCE  →  (user reviews)  →  CAPTURE  →  (user reviews)  →  HANDOFF  →  skill-creator
```

Each mode does one thing, writes one file, and **stops**. The review pauses between modes are the whole point — the user reads what you found, adds notes, and only then do you advance. Never chain source → handoff in a single run; a brief nobody reviewed is just a guess with extra steps.

Route by what the user gives you:

| Input | Mode |
|-------|------|
| A fresh skill idea / description | **SOURCE** |
| "capture" (or notes on `sources.md`) | **CAPTURE** |
| "handoff" (or "turn this into a skill-creator prompt") | **HANDOFF** |

All working files live in `.claude/skill-brief/` in the current repo — these are personal scratch, not committed artifacts. Don't put them in `docs/`. (The user can `.gitignore .claude/skill-brief/` if they want; you don't need to.)

---

## SOURCE mode

Goal: a `sources.md` that tells the user whether to build at all, and if so, what prior art to steal from.

### 1. Delegate discovery to find-skills

Invoke the existing **`find-skills`** skill for the idea. It already searches skills.sh, the Skills CLI, GitHub code search, GitHub collections, and local installs, ranks candidates, and produces a table plus a landscape synthesis. Do **not** reimplement any of that.

When find-skills finishes it will offer its own "reply build it" CTA — **ignore that CTA** and continue with skill-brief's flow below. You want its table and synthesis, not its handoff.

### 2. Add "what to lift" per candidate — the part find-skills skips

find-skills ranks by relevance; it does not read each candidate and tell you what to take. That's your job. For each candidate worth considering, read its `SKILL.md` and write a blunt note keyed by the skill's catalog number (see step 4):

- **Partial match** → "lift X, ignore the rest" (name the one technique or structure worth taking).
- **Full match** → "grab wholesale" — and that's a signal the idea might be a duplicate; see step 3.

Don't pad this list. A candidate with nothing worth lifting doesn't belong in the Candidates list — drop it (it still appears in the catalog above). Three real lifts beat ten polite ones.

### 3. Compliance review vs. skills the user already has

Scan the user's own installed skills — not just what's on the internet:

```bash
ls skills/ ~/.claude/skills/ .claude/skills/ 2>/dev/null
```

Read the `description` of anything that looks close. If the idea duplicates a skill the user already has, **say so plainly** and recommend not building it (or building a narrower thing the existing skill doesn't cover). Catching a duplicate here is the single most valuable thing this mode does — a "don't build, you already have X" verdict just saved the user an afternoon.

For a full health audit of their whole skill set (conflicts, stale descriptions), point them at **`review-skills`** — that's a different job and not something you inline here.

### 4. Write `sources.md`

Write `.claude/skill-brief/sources.md`, tables over prose. Assign every skill find-skills surfaced a **stable incrementing number** in the catalog — that number is the skill's canonical ID for the rest of the brief's life. Reference it (not the name) everywhere downstream: candidates, compliance, notes, and any follow-up command.

```markdown
# Skill brief — sources: {idea}

**Verdict:** {build / build narrower (duplicates parts of X) / don't build (X already covers this)}

## Similar skills found
Reproduce find-skills' ranked results — do NOT pre-digest them away; this is the browsable catalog the user reads from. One table per capability area of the brief. Number skills continuously across all tables (1, 2, 3 …).

### {capability area}
| # | Skill | Installs | Source | Relevance | Link |
|---|-------|----------|--------|-----------|------|
| 1 | ... | 95.1K | ✅ Microsoft | ... | url |
| 2 | ... | 13.5K | community | ... | url |

## Candidates — what to lift
Only the numbered skills worth stealing from, by their catalog number — drop any with nothing worth lifting (they stay in the catalog above, just not here):

3. **{skill-name}**
   - lift: {technique / structure worth taking}
   - leave: {what to ignore}

## Compliance — vs. skills you already have
| # | Existing skill | Overlap | Verdict |
|---|----------------|---------|---------|
| ... | ... | none / partial / full | safe to build / narrow it / skip |

## Landscape
{the synthesis paragraphs from find-skills — the readable overview of the field}

## Your notes
<!-- For each catalogued skill, add bullets you want captured. Leave a skill's bullets empty to capture nothing for it. Then run skill-brief in "capture" mode. -->
1. {skill-name} —
2. {skill-name} —
```

List every catalogued skill in "Your notes" by its number, so the user has one slot per skill; an empty slot means "capture nothing for this one".

### 5. Tell the user and stop

Give the one-line verdict, name the top 2–3 lifts in the conversation, then: "Full breakdown in `.claude/skill-brief/sources.md` — add notes there, then say **capture** to lock in the decisions." Do not advance on your own.

---

## CAPTURE mode

Goal: a tight decision record so skill-creator inherits *decisions*, not a research dump.

Re-read `.claude/skill-brief/sources.md`, **including the user's per-skill notes** — those override your original recommendations. If any note raises a clarifying question or contradicts a recommendation (e.g. asks to lift from a skill you flagged as a full-match duplicate), invoke **`/grilling`** to stress-test the decision *before* writing `capture.md` — don't silently bake the contradiction in. Preserve the catalog numbers from `sources.md` verbatim when referencing skills. Then write `.claude/skill-brief/capture.md`:

```markdown
# Skill brief — capture: {name}

## Decision
{build / build narrower / don't build} · **name:** {proposed-name} · **modes:** {if any}

## What it does
- **In scope:** ...
- **Out of scope:** ...

## Lifted from prior art
| Take | From (#) | How it's adapted |
|------|----------|------------------|
| ... | #3 [skill](url) | ... |

## Deliberately not doing
- {thing we could have built but chose not to, and why}

## Open questions for skill-creator
- ...
```

The "deliberately not doing" list is as valuable as the lifts — it stops skill-creator from re-adding scope the user already rejected. Then tell the user it's ready and to say **handoff** when they want the skill-creator prompt. Stop.

---

## HANDOFF mode

Goal: a self-contained prompt skill-creator can run cold.

Re-read `.claude/skill-brief/capture.md` and emit a prompt with: what to build, proposed name, shape/modes, patterns to borrow (with source links and their catalog numbers, preserved verbatim from `sources.md`), out-of-scope list, and open questions. Print it in the conversation so the user can eyeball it, then offer to run `/skill-creator` with it (invoke the `skill-creator` skill on confirmation).

---

## What not to do

- **Don't reimplement find-skills' discovery.** Marketplace/GitHub/local search is its job; call it, don't rebuild it.
- **Don't force-fit adjacent skills.** `source-synthesis` digests URLs, `review-skills` audits skill-set health, `handoff` compacts a conversation for another agent — none is the sourcing-a-brief job. Point the user at them where relevant; don't wire them in.
- **Don't skip the review pauses.** One mode per invocation. A brief the user never read is worthless.
- **Don't write the final skill.** That's skill-creator's job, and it does the interview better with a clean brief.
- **Don't pad "what to lift".** A candidate with nothing to steal gets dropped, not a courtesy row.
