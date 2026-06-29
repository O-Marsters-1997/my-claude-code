---
name: find-skills
description: >
  Not for building skills — for finding them. Use when the intent is to discover, search,
  or inventory what agent skills already exist for a topic. The user wants to know "what's
  out there for X" rather than "make me something." Triggers on: "find a skill for X",
  "skill search: X", "what skills exist for X", "discover skills for X", "is there already
  a skill for X", "before I build this, has someone done it?", or any pre-build audit.
  Also trigger proactively before any skill-creator session when the user hasn't searched
  yet — finding what exists prevents duplicate effort. Output is a ranked table plus a
  thematic synthesis document, never a new skill.
compatibility:
  tools:
    - WebSearch
    - WebFetch
    - Bash
---

# Find Skills

You've been asked to find existing skills relevant to a description. Search the sources below in order, rank by relevance and quality, produce a discovery document, then hand off to skill-creator on confirmation.

## Sources

Work through these in order — earlier sources carry richer quality signals:

| # | Source | How |
|---|--------|-----|
| 1 | **skills.sh** (SkillsMP) | Fetch `https://skills.sh/` and search the leaderboard; this is the primary marketplace with install counts |
| 2 | **Skills CLI** | Run `bunx skills find {keywords}` (fall back to `npx skills find {keywords}` if Bun isn't available) |
| 3 | **GitHub code search** | `WebSearch` for `site:github.com filename:SKILL.md {keywords}` — finds skills not yet in the marketplace |
| 4 | **GitHub collections** | `WebSearch` known skill collections: `ComposioHQ/awesome-claude-skills`, `vercel-labs/agent-skills`, `xixu-me/skills`, `openai/skills` |
| 5 | **Local** | Scan installed skills in `~/.claude/skills/`, `.claude/skills/`, and the project's `skills/` directory |

From each skill found, extract:
- **name** — skill identifier from frontmatter
- **author** — publisher, GitHub org, or username
- **url** — skills.sh page or direct link to SKILL.md
- **installs** — from skills.sh if available, otherwise N/A
- **description** — one sentence from frontmatter `description:` field

## Quality signals

Use these to rank, not to hard-exclude:
- **Install count** — prefer 1K+; flag anything under 100 as low-confidence
- **Publisher reputation** — known publishers (vercel-labs, anthropics, openai, microsoft, xixu-me) rank higher
- **Repository quality** — check GitHub stars; treat skills from repos with <100 stars with skepticism

A niche but highly relevant skill from a small repo beats a popular but off-topic one.

## Ranking

Score each candidate against the user's description. Weight:
1. Relevance to the task domain (highest weight)
2. Quality signals above
3. Breadth — broader coverage is more reusable

Pick the top 10. If fewer than 10 exist, include all.

## Output

### 1. Write the discovery document

Write to `./docs/skill-discovery.md` (create `./docs/` if needed):

```markdown
# Skill Discovery: {user's description}

_Sources: skills.sh · Skills CLI · GitHub code search · GitHub collections · Local_  
_Add comments anywhere in this file, then reply "build it" to proceed._

---

## Top 10 Relevant Skills

| # | Skill | Author | Source | Installs | Description |
|---|-------|--------|--------|----------|-------------|
| 1 | [skill-name](url) | Author | skills.sh / GitHub / Local | 12K / N/A | One sentence |
...

---

## Synthesis

[3–5 thematic paragraphs in the style of a literature review. Group skills by what they
emphasise rather than listing them one by one. Note variation in approach, scope, or
workflow philosophy across the set. Write in flowing prose — no sub-headers, no bullets,
no gaps or recommendation section. The goal is a readable overview of the landscape.]

---

## Your notes

<!-- Add comments here, or inline above any row in the table -->
_Nothing yet._
```

### 2. Tell the user

- Confirm the file was saved
- Name the top 3 picks with one sentence each in the conversation
- Say: "Full table and synthesis in `./docs/skill-discovery.md` — add notes directly there, then reply **build it** to hand off to skill-creator."

## On confirmation

When the user replies "build it" (or equivalent):

1. Re-read `./docs/skill-discovery.md` to pick up any notes they added
2. Build a skill-creator prompt from:
   - The original description
   - Patterns and approaches from the synthesis worth borrowing
   - Specific techniques to adopt, with source URLs
   - Any user notes from the document

Handoff prompt structure:
```
Build a skill for: {description}

Patterns worth borrowing from existing skills:
- {pattern} — from [{skill-name}]({url})

User notes: {anything from the discovery document}
```

3. Invoke `/skill-creator` with that prompt.

## What not to do

- Don't skip skills.sh — install counts are the strongest signal of real-world adoption.
- Don't run `bunx skills add` — this skill only discovers; the user decides what to install.
- Don't fabricate URLs — if a URL can't be confirmed, mark it `(unverified)`.
- Don't make the synthesis a disguised bullet list or a rehash of the table — it should surface patterns that only emerge when you read the skills as a set.
