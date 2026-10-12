---
name: handoff
description: Compact the current conversation into a handoff document that another agent picks up in a fresh context window — including an agent targeting a different repo. Use when the user says "hand off", "write a handoff", "pass this to another agent", "continue in another repo", or wants to preserve context before switching sessions/repos.
---

# Handoff

Compact the current conversation into a self-contained handoff document so a *fresh* agent — with none of this conversation's context — can pick up exactly where you left off. Work often crosses repos (backend work in one repo, then a separate agent for the frontend), and the receiving agent must be able to find the handoff without being told where to look.

## Where the handoff lives (READ THIS FIRST)

The single most important thing about this skill: **the handoff must be trivially findable by an agent in a brand-new context window.** Do not save it to a temp dir — it would be invisible to the next agent.

**Always write to `.claude/handoffs/` at the current git repo's root**, and tell the user the exact absolute path.

- Create the directory if it does not exist: `mkdir -p <repo-root>/.claude/handoffs`
- **Filename convention:** `<from-repo>__to__<to-repo>__<short-slug>.md`
  - e.g. `services__to__web__teammate-invite-api.md`
  - If the handoff is not repo-to-repo (same repo, just a new session), use `<repo>__<short-slug>.md`.
- Also maintain a pointer to the most recent handoff at `<repo-root>/.claude/handoffs/LATEST.md` — overwrite it each time with a one-line link to the newest handoff file plus its one-line summary. A new agent that knows nothing can always start by reading `LATEST.md`.

For a cross-repo handoff, the receiving agent works in a different checkout, so give it the absolute path to the file.

### Tell the user how to resume

After writing the file, end your response with the **exact** copy-pasteable instruction for the next agent, e.g.:

> Start the next agent in `web/` and paste:
> `Read /abs/path/to/services/.claude/handoffs/services__to__web__teammate-invite-api.md and continue from there.`

## For the receiving agent (how to find a handoff)

If you were told to continue from a handoff, or you suspect one exists:

1. Read `<repo-root>/.claude/handoffs/LATEST.md` for the newest handoff.
2. Or `ls <repo-root>/.claude/handoffs/` and pick the file whose `__to__<your-repo>__` segment matches the repo you're working in.

## What to include

Keep it tight — this is a compaction, not a transcript. Prefer links over copies.

1. **Goal** — one or two sentences: what we're ultimately trying to achieve.
2. **Target repo & where to work** — which repo the next agent should `cd` into, the branch/worktree to use, and how to run/build/test it there. Be explicit; the next agent may never have opened that repo.
3. **State so far** — what's done, what's in progress, what's verified vs. assumed. Note the current branch and whether there are uncommitted changes.
4. **Cross-repo contract** — for backend→frontend (or any cross-repo) handoffs, spell out the *interface* between the repos: API endpoints/shapes, types, event names, feature flags, env vars, generated clients, or schema the next agent depends on. This is the context most expensive to rediscover — make it precise.
5. **Next steps** — an ordered, concrete task list for the receiving agent.
6. **Gotchas / decisions** — non-obvious constraints, dead ends already tried, and decisions made (with the *why*).
7. **Suggested skills** — name the skills the next agent should use (e.g. `tdd`, `code-review`, `triage-issue`), so it doesn't reinvent workflow.

## Reference, don't duplicate

Do **not** paste large artifacts into the handoff. Link to them by path or URL instead:

- specs, plans, PRDs, ADRs → repo path
- issues / PRs → GitHub URL or number
- commits / diffs → SHA (and `git show <sha>` hint)
- prior handoffs → path under `.claude/handoffs/`

Use clickable `file_path:line_number` references so the next agent can jump straight to the code.

## Security

Before writing the file, **redact** all secrets and PII: API keys, tokens, passwords, connection strings, customer data, personal emails/names. Replace with `<REDACTED: description>`. The handoff lives on disk — treat it as if it could be read by anyone with repo access.

## Tailoring

If the user passed arguments describing the next session's focus (e.g. "handing off the frontend work" or "only the migration"), scope the handoff to that focus: lead with the relevant next steps and trim state that doesn't serve it.

## Template

```markdown
# Handoff: <short title>

**From:** <repo> (<branch>)  →  **To:** <repo> (<branch/worktree>)
**Date:** <YYYY-MM-DD>  ·  **One-line summary:** <what the next agent is picking up>

## Goal
<1–2 sentences>

## Where to work
- Repo: `<absolute path to repo>`
- Branch/worktree: `<branch>`
- Run / test: `<commands>`

## State so far
- Done: …
- In progress: …
- Verified vs assumed: …
- Uncommitted changes: <yes/no — what>

## Cross-repo contract  (backend ↔ frontend)
- Endpoints / types / events / flags / env the other side depends on
- Links: <paths, URLs, SHAs>

## Next steps
1. …
2. …

## Gotchas & decisions
- …

## Suggested skills
- `<skill>` — why
```
