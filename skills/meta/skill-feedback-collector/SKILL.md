---
name: skill-feedback-collector
description: Collect structured feedback on a skill session and convert it into improvement proposals, then hand them off to skill-updater. Use when a skill finishes and asks whether you want to log feedback, or when the user says "log feedback on the X skill", "record feedback on skill Y", "capture feedback for skill Z", "give feedback on the X skill", or "I want to leave feedback on skill X". Make sure to use this skill whenever feedback, impressions, or a rating on a skill are mentioned — even if the user doesn't explicitly say "feedback".
---

# Skill Feedback Collector

Captures feedback immediately after a skill runs, while recall is fresh. Appends a standalone ledger entry to the invoking skill's ledger, synthesises proposals, lets the user pick which ones to act on, then hands the selection to skill-updater.

**Hard constraints — do not break these:**

- This skill never creates a ledger inside its own directory. Ledgers live next to each opted-in skill's SKILL.md, never here. (Exception: when collecting feedback on skill-feedback-collector itself, the ledger necessarily lives in its own directory — this is intentional.)
- Never call `skill-creator` at runtime. Diagnosis reasoning is embedded in this skill.
- Never run evals or benchmarks.

---

## Inputs

Expected when invoked programmatically by another skill:

- `skill_name` — the name of the skill that just ran (string)
- `skill_path` — absolute path to that skill's SKILL.md

When invoked directly by the user without these (e.g. "give feedback on the clean-comments skill"), discover the target skill first. Scan:

- `~/.claude/skills/`
- `<cwd>/.claude/skills/`
- `<cwd>/skills/`
- Plugin cache: read `enabledPlugins` from `~/.claude/settings.json`, then find SKILL.md files under `~/.claude/plugins/cache/`

If the user named a skill, locate it. If ambiguous, list matching skills and ask them to confirm.

---

## Step 1 — Read the invoking skill

Read the target SKILL.md in full. Extract `name` from frontmatter. If a `version` field is present, extract it. If absent (the common case — most skills don't carry one), record `unversioned`. Do not fabricate a version.

---

## Step 2 — Confirm the ledger exists

Check for `skill-feedback-ledger.md` as a sibling to the target skill's SKILL.md (i.e., in the same directory).

If it doesn't exist: copy `skill-feedback-ledger-template.md` from this skill's own directory (`~/.claude/skills/skill-feedback-collector/skill-feedback-ledger-template.md`) to `<target-skill-dir>/skill-feedback-ledger.md`. Tell the user: "Created a new feedback ledger for <skill-name>."

---

## Step 3 — Ask three questions, one at a time

Wait for each answer before asking the next. Use these exact phrasings:

1. "What worked well?"
2. "What was clunky or missing?"
3. "On a scale of 1–100, how well did the skill fit your workflow?"

Don't rush or bundle them — asking one at a time invites better answers.

---

## Step 4 — Append the ledger entry

Append a new entry to the bottom of the target skill's `skill-feedback-ledger.md`. Keep it self-contained — no references to other entries. Use this schema exactly:

```
---

## <ISO 8601 timestamp, local time> — <skill-name> v<version or "unversioned">
```

Leave a `<!-- proposals pending -->` marker below the heading — you'll replace it in Step 8.

---

## Step 5 — Synthesise

Produce two artefacts:

**1. Plain-language summary** — one paragraph describing what the user was doing, what landed, what didn't, and the overall sentiment. Write for a future reader who hasn't seen the conversation.

**2. Structured proposals** — a list, ordered high → low by priority. Each proposal:

```
{ priority: "high" | "medium" | "low",
  category: "instructions" | "tools" | "examples" | "error_handling" | "structure" | "references",
  suggestion: <string>,
  expected_impact: <string> }
```

**Diagnosis rule.** When the user describes a symptom, find the underlying instruction weakness — not a surface fix. Then propose fixing the root cause:

| User says                                      | Weak proposal         | Diagnosis                                                               | Strong proposal                                                                                                                                |
| ---------------------------------------------- | --------------------- | ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| "It kept asking me things I'd already told it" | "Ask fewer questions" | Skill lacks a 'skip clarification when context is already present' rule | "Add a precondition check: before asking any clarifying question, scan the conversation for an existing answer and skip the question if found" |
| "It buried the output in caveats"              | "Write less"          | Skill doesn't specify a results-first structure                         | "Restructure output section: lead with the result, put caveats in a collapsible or trailing paragraph"                                         |
| "I had to re-explain the context twice"        | "Read context better" | Skill reads context lazily — misses early messages                      | "Explicitly instruct: read the full conversation from the top before taking any action; don't rely on recency alone"                           |

Use this diagnostic approach for every proposal, even when the user's description is brief. You have the skill in full — you can read it and spot which instructions are absent, vague, or contradictory.

---

## Step 6 — Present and select

Show:

1. The one-paragraph summary.
2. A numbered proposal list, each tagged `[priority | category]`.

Example:

```
1. [high | instructions] Add a precondition check: before asking any clarifying question…
   Expected impact: eliminates repeat questions when context is already in the conversation.

2. [medium | examples] Add a before/after example showing results-first output structure…
   Expected impact: anchors the output format for edge cases where the skill currently buries results.
```

Then ask: "Which proposals would you like to apply? Reply with comma-separated numbers (e.g. '1,3'), 'all', or 'none'."

After the user selects, ask: "Anything you'd like to add that wasn't in the list? If so, describe it briefly and I'll structure it — or just say 'no'."

If the user adds something in prose, infer `priority` and `category` from context and fold it into the selection.

---

## Step 7 — Package and hand off

Build `improvement_suggestions[]` from the selected proposals. Then invoke skill-updater with:

```
{ skill_name, skill_path, improvement_suggestions: [...] }
```

---

## Step 8 — Finalise the ledger entry

After handing off, replace the `<!-- proposals pending -->` marker in the ledger entry with:

```
**Summary**
<one paragraph from Step 5>

**Selected proposals**
- [priority | category] <suggestion> — expected impact: <expected_impact>
(repeat for each selected)

**Discarded proposals**
- [priority | category] <suggestion> — expected impact: <expected_impact>
(repeat for each not selected, including any the user explicitly rejected)
```

This makes each entry a complete, self-contained record — anyone reading the ledger later can see what was surfaced and what was acted on.

---

## Opting a skill in (documentation — do not do this for any skill during this build)

To enable the feedback prompt at the end of a skill, two things must happen:

1. Create its ledger: copy `~/.claude/skills/skill-feedback-collector/skill-feedback-ledger-template.md` to `<skill-dir>/skill-feedback-ledger.md`.
2. Append this exact line to the skill's SKILL.md:
   > When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
