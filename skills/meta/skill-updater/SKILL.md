---
name: skill-updater
description: Update an existing skill with new context, patterns, or improvements. Edits the skill at its source repo (not the installed copy) and opens a draft PR for the change. Use when user wants to update a skill, modify a skill, add a pattern to a skill, improve an existing skill, edit a SKILL.md, tweak skill instructions, change skill behaviour, or mentions "update skill", "edit skill", "change skill", "add to skill". Also use when invoked programmatically with a skill name, path, and improvement_suggestions[] payload from skill-feedback-collector, with a GitHub issue number, or when /implement hands over an issue whose changes are all under `skills/`. Do NOT use for creating new skills from scratch — use skill-creator for that.
---

# Skill Updater

In-place updates to an existing skill's SKILL.md. Edits live in the source repo and are shipped as a draft PR — never to the installed copy. Every change is verified before it ships: the old and new skill run the same scenario, and the new one must do better. For creating new skills, use skill-creator instead.

## Programmatic input path

When invoked with a GitHub issue number, read the issue (`gh issue view <N>`): "Where to look" names the skill, the acceptance criteria are the change, and the scenario in them is the one to rerun. Skip discovery and selection.

When invoked with a structured payload — `skill_name`, `skill_path`, and `improvement_suggestions[]` — skip the discovery and selection steps and go straight to applying the suggestions.

Each suggestion in `improvement_suggestions[]` has this shape:

```
{ priority: "high" | "medium" | "low",
  category: "instructions" | "tools" | "examples" | "error_handling" | "structure" | "references",
  suggestion: string,
  expected_impact: string }
```

Apply suggestions in priority order (high first). Even on the programmatic path, run the preview-and-confirm step before writing.

## Process

### 0. Resolve source repo

Read `~/.claude/skill-updater.config`. If it exists and is non-empty, use its contents (trimmed) as `<source-repo>`.

If missing or empty: ask the user for the absolute path to their skills source repo, validate it is a git repo containing a `skills/` directory, then write the path to `~/.claude/skill-updater.config`. Report the resolved path in one line so the user can catch a wrong config immediately.

### 1. Identify the target skill

Enumerate `<source-repo>/skills/**/SKILL.md` (skills nest, e.g. `skills/meta/<name>/`). Record each skill's name (from frontmatter) and absolute path.

> This skill only edits skills in `<source-repo>/skills/`. Plugin skills or `~/.claude/skills/` entries are out of scope — those require a fork-and-PR workflow against the plugin's upstream repo.

If the user didn't name a skill, list what was found and ask which to update:

```
1. clean-comments — <source-repo>/skills/code/clean-comments/SKILL.md
2. tdd — <source-repo>/skills/code/tdd/SKILL.md
```

### 2. Read and understand

Read the target SKILL.md in full. Also read any bundled reference files relevant to the requested change.

Before editing, tell the user: what the skill currently does (one sentence), which section(s) the change affects, and your proposed edit. Wait for confirmation — the user may have a different mental model than what's written.

### 3. Apply the update

Before editing, copy the current skill directory to `<scratchpad>/<name>-baseline/` so the old version can be rerun. Write changes to the skill's own SKILL.md (and its `references/`) only. Never edit `~/.claude/skills/...` or `.agents/skills/...`.

- **Preserve voice and style.** Don't rewrite sections the user didn't ask to change.
- **Keep it lean.** If adding content pushes SKILL.md past 100 lines, split into a `references/` file.
- **Frontmatter hygiene.** Update `description` if the change affects triggers or behaviour — stale descriptions cause misfires.
- **Remove dead lines.** Delete lines that change nothing about behaviour, not only add new ones.
- **Explain the why.** New instructions should include reasoning, not just directives.
- **Imperative form.** Write "Do X" not "You should do X".

### 4. Verify old vs new

The scenario is the issue's acceptance criterion, or a case from the skill's `evals/evals.json` when there is no issue. Run it twice in fresh subagents, one pointed at the baseline copy and one at the edited skill, with the same prompt. Compare the outputs against the criterion.

- New passes and baseline fails: continue.
- New does not pass: rework the edit once and rerun. If it still fails, reject the change and report why; open no PR.
- Baseline also passes: the scenario does not test the change; tighten it and rerun, or report that the change is unverified.

Append the scenario to `<skill>/evals/evals.json` (create it as `{"skill_name": "<name>", "evals": []}` if absent) as `{"id": <next>, "prompt": ..., "expected_output": ...}`, and keep the baseline and new outputs for the PR body.

### 5. Open a draft PR

See [`references/pr-workflow.md`](references/pr-workflow.md) for the full git and gh commands.

1. Fetch origin and create branch `skill-update/<name>-<YYYYMMDD>` from `origin/main`.
2. Stage changed files: the skill's SKILL.md, `evals/evals.json` and any `references/` additions.
3. Draft commit message: `chore(skills/<name>): <one-line summary>`.
4. Draft PR body using the template in `references/pr-workflow.md`.
5. **Show the commit message and PR body to the user. Wait for confirmation.** Then push and open the draft PR.
6. Return the PR URL.

### 6. Confirm and report

Output the edited file path and the PR URL:

```
Updated: <source-repo>/skills/<name>/SKILL.md
PR: <url>
```

Offer to read back the updated skill so the user can review it in context.
