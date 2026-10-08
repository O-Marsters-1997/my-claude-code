---
name: skill-feedback-collector
description: Collect structured feedback on a skill session and convert it into improvement proposals, then file them as reflect-format issues or hand them to skill-updater. Use when a skill finishes and asks whether you want to log feedback, or when the user says "log feedback on the X skill", "record feedback on skill Y", "capture feedback for skill Z", "give feedback on the X skill", or "I want to leave feedback on skill X". Make sure to use this skill whenever feedback or impressions on a skill are mentioned — even if the user doesn't explicitly say "feedback".
---

# Skill Feedback Collector

Captures feedback immediately after a skill runs, while recall is fresh. Synthesises proposals, lets the user pick which ones to act on, then writes each as a reflect-format issue or hands it straight to skill-updater.

**Hard constraints — do not break these:**

- Never call `skill-creator` at runtime. Diagnosis reasoning is embedded in this skill.
- Every proposal carries a scenario the old skill fails; skill-updater reruns it, this skill does not.

---

## Inputs

Expected when invoked programmatically by another skill:

- `skill_name` — the name of the skill that just ran (string)
- `skill_path` — absolute path to that skill's SKILL.md

When invoked directly by the user without these (e.g. "give feedback on the clean-comments skill"), discover the target skill first. Scan:

- `~/.claude/skills/`
- `<cwd>/.claude/skills/`
- `<cwd>/skills/**/SKILL.md`
- Plugin cache: read `enabledPlugins` from `~/.claude/settings.json`, then find SKILL.md files under `~/.claude/plugins/cache/`

If the user named a skill, locate it. If ambiguous, list matching skills and ask them to confirm.

---

## Step 1 — Read the invoking skill

Read the target SKILL.md in full. Extract `name` from frontmatter.

---

## Step 2 — Ask two questions, one at a time

Wait for each answer before asking the next. Use these exact phrasings:

1. "What worked well?"
2. "What was clunky or missing?"

Don't rush or bundle them — asking one at a time invites better answers.

---

## Step 3 — Synthesise

Produce a one-paragraph summary (what the user was doing, what landed, what didn't) and a list of proposals ordered high → low. Each proposal has `priority` (high | medium | low), `category` (instructions | tools | examples | error_handling | structure | references), `suggestion` and `expected_impact`.

**Diagnosis rule.** When the user describes a symptom, find the underlying instruction weakness, not a surface fix. Read the skill in full and spot which instructions are absent, vague, or contradictory. "It kept asking me things I'd already told it" is not "ask fewer questions"; it is a missing rule to scan the conversation for an existing answer before asking.

**Scenario.** Give each proposal a scenario the old skill fails: a concrete prompt plus the observable behaviour that counts as passing. skill-updater reruns it against old and new.

---

## Step 4 — Present and select

Show the summary and a numbered proposal list, each tagged `[priority | category]` with its scenario. Then ask: "Which proposals would you like to apply? Reply with comma-separated numbers (e.g. '1,3'), 'all', or 'none'."

If the user adds something in prose, structure it the same way and fold it into the selection.

---

## Step 5 — Check for duplicates

For each selected proposal, list open reflect issues that touch the same skill:

```bash
gh issue list --label reflect --state open --limit 200 --json number,url,title,body
```

If one already covers it, comment on it with the new evidence instead of filing.

---

## Step 6 — Write the issue body

Use the format in `skills/meta/reflect/references/issue-format.md` (Context, Where to look, Acceptance criteria, Evidence). The acceptance criterion is the scenario: "Given <prompt>, the skill <passing behaviour>; the current skill <failing behaviour>". "Where to look" names the skill's SKILL.md path and the section to change.

---

## Step 7 — File it or run it

Ask: "File as an issue, or run skill-updater now?"

- Issue: create it per issue-format.md, with the `reflect` and `status:ready` labels, and return the URL.
- Now: invoke skill-updater with `{ skill_name, skill_path, improvement_suggestions: [...], scenario }`.
