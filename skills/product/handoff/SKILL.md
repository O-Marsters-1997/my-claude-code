---
name: handoff
description: Compact the current conversation into a handoff document that another agent picks up in a fresh context window — including an agent targeting a different repo in the plain multirepo (e.g. backend work in services/ handing off to frontend work in support-app/). Use when the user says "hand off", "write a handoff", "pass this to another agent", "continue in another repo", or wants to preserve context before switching sessions/repos.
---

# Handoff

Compact the current conversation into a self-contained handoff document so a *fresh* agent — with none of this conversation's context — can pick up exactly where you left off. This is optimised for the **plain multirepo**, where work often crosses repos (do backend work in one repo, then hand off to a separate agent to do the frontend) and the receiving agent must be able to find the handoff without being told where to look.

## Where the handoff lives (READ THIS FIRST)

The single most important thing about this skill: **the handoff must be trivially findable by an agent in a brand-new context window, working in a different repo.** Do not save it to a temp dir or into the current repo — those are invisible to the next agent.

**Canonical location — always write here:**

```
/Users/ollymarsters/Documents/plain/.claude/handoffs/
```

This directory sits at the **workspace root**, above every sub-repo (`aws/`, `infrastructure/`, `services/`, `support-app/`, `toolbox/`, and any `support-app-*` worktree). Every agent working anywhere under `plain/` can reach it with the same absolute path, regardless of which repo it was started in.

- Create the directory if it does not exist: `mkdir -p /Users/ollymarsters/Documents/plain/.claude/handoffs`
- **Filename convention:** `<from-repo>__to__<to-repo>__<short-slug>.md`
  - e.g. `services__to__support-app__teammate-invite-api.md`
  - If the handoff is not repo-to-repo (same repo, just a new session), use `<repo>__<short-slug>.md`.
- Also maintain a pointer to the most recent handoff at `/Users/ollymarsters/Documents/plain/.claude/handoffs/LATEST.md` — overwrite it each time with a one-line link to the newest handoff file plus its one-line summary. A new agent that knows nothing can always start by reading `LATEST.md`.

> If the user is working **outside** `plain/` (e.g. a personal repo), fall back to a `.claude/handoffs/` directory at the current git repo's root, and tell the user the exact path. But default to the plain workspace path above unless it's clearly not a plain-workspace session.

### Tell the user how to resume

After writing the file, end your response with the **exact** copy-pasteable instruction for the next agent, e.g.:

> Start the next agent in `support-app/` and paste:
> `Read /Users/ollymarsters/Documents/plain/.claude/handoffs/services__to__support-app__teammate-invite-api.md and continue from there.`

## For the receiving agent (how to find a handoff)

If you were told to continue from a handoff, or you suspect one exists:

1. Read `/Users/ollymarsters/Documents/plain/.claude/handoffs/LATEST.md` for the newest handoff.
2. Or `ls /Users/ollymarsters/Documents/plain/.claude/handoffs/` and pick the file whose `__to__<your-repo>__` segment matches the repo you're working in.

Shell shortcuts are available (defined in dotfiles `zsh/functions/handoff-nav.sh`) and usable via Bash:

- `handoffs` — list all handoffs, newest first
- `handoffs-for <repo>` — only handoffs targeting `<repo>` (e.g. `handoffs-for support-app`)
- `hgrep <pattern>` — grep the text of every handoff
- `handoff-latest` — print the `LATEST.md` pointer

## What to include

Keep it tight — this is a compaction, not a transcript. Prefer links over copies.

1. **Goal** — one or two sentences: what we're ultimately trying to achieve.
2. **Target repo & where to work** — which sub-repo the next agent should `cd` into, the branch/worktree to use, and how to run/build/test it there. Be explicit; the next agent may never have opened that repo.
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

Before writing the file, **redact** all secrets and PII: API keys, tokens, passwords, connection strings, customer data, personal emails/names. Replace with `<REDACTED: description>`. The handoff lives on disk in a shared location — treat it as if it could be read by anyone with repo access.

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
- Repo: `/Users/ollymarsters/Documents/plain/<repo>`
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
