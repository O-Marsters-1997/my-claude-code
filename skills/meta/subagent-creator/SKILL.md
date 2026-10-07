---
name: subagent-creator
description: >
  Designs a lean Claude Code subagent, or widens one on request. Use when the user wants to
  create, scope or audit a subagent or custom agent, or asks to give an existing one more
  tools, turns or access. Grills on whether it earns its place before writing anything. Not
  for writing skills.
---

# Subagent creator

A subagent is a context firewall with a tool budget: worth having when it keeps noise out of
the parent, enforces a restriction, or runs independent work in parallel. It is not a persona.
This skill grills first and writes the smallest definition that does the job. It never widens
an agent on its own initiative: scope grows only when the user asks, and only with evidence.

Three modes. Pick from the user's words:

- **create** (default): grill, write, check.
- **expand `<name>`**: only when the user asks. Widen the agent by the smallest step.
- **audit `<name>|all`**: check existing agents and report what to cut.

## The rules

Each is enforced in the grill, the template, or the step 4 checklist.

1. **Earn it.** A subagent needs one of four reasons: verbose output the parent doesn't need,
   an enforced tool restriction, self-contained work that returns a summary, or independent
   parallel work. No reason means it belongs in the main thread, a skill, a hook or a script.
2. **Code owns control flow.** Fixed steps (lint, push, label, run CI) are scripts or hooks.
   The agent covers only the open-ended step.
3. **Smallest box.** An explicit `tools:` allowlist, read-only by default, MCP servers scoped
   inline, never inherited wholesale. Every visible tool is a decision the agent can get wrong.
4. **Single writer.** An agent that writes owns its files and runs in `isolation: worktree`.
   Reviewers, diagnosers and researchers don't write.
5. **Fresh means blind.** A fresh agent knows only its prompt and the brief. Say what the brief
   carries. If it really needs the whole conversation, use a fork instead of a custom agent.
6. **Artifacts, not transcripts.** The agent returns a short, fixed-format result or a file path.
7. **Bounded.** `maxTurns` is always set. Use a cheaper model only for bounded, checkable work,
   never for work that must judge its own limits.
8. **Few hops.** A chain of more than four agent-to-agent handoffs fails. Split by phase, not
   by domain.
9. **Enforce, don't ask.** A rule that matters is a hook or a tool restriction, not a sentence
   in the prompt.
10. **Ask, don't improvise.** When the agent hits its scope, it stops and returns a
    `SCOPE_REQUEST`. That request is the only way scope grows.

## Create

### 1. Look before asking

Read the existing agents in `agents/` (if this repo syncs a top-level `agents/` directory) or
`.claude/agents/`, plus `~/.claude/agents/`. Note any that already overlap the request. That is
a fact you look up yourself, never a question for the user.

### 2. Grill

Invoke the `grilling` skill and pass it this agenda. Work the value branch first. If the value
branch fails, recommend the alternative and stop: declining to create an agent is a good
outcome of this skill.

**Value branch**
- Which of the four reasons applies? Push back on "it's a specialist" with no restriction or
  isolation behind it.
- Does an existing agent already do 90% of this? If so, widen that agent or branch inside its
  prompt instead.
- Is any of this fixed steps that a script or hook could own (rule 2)?
- Would a skill (reusable knowledge in the parent context) or a fork (needs the conversation)
  do better?

**Scope branch**
- The job in one sentence, plus what "done" looks like.
- What the brief hands it, and what it must not assume it knows.
- The output contract: format, length cap, or the file it writes.
- Does it write? Default no. If yes, to which files, and confirm worktree isolation.
- The smallest tool set that finishes the job. For Bash, which commands, and whether a
  `PreToolUse` hook should hold it to them.
- Turn budget and model tier.
- Will it run in parallel with copies of itself? If so, who owns which files?

Recommend the leanest answer on every question. Widening later is cheap through **expand**, so
start small. Don't offer extra tools or turns "just in case".

### 3. Write

Copy `assets/agent-template.md` into the agent directory as `<name>.md` and fill it in. Keep the
body under 40 lines. The description says when to delegate *and* when not to, because it is
what the parent reads to decide.

### 4. Check

Read the finished file against this checklist and fix any failure before showing it:

- `tools` is an explicit allowlist: no wildcard, no omission, and nothing the job doesn't use.
- `maxTurns` is set.
- Write tools appear only with `isolation: worktree`.
- Open `Bash` is held by a `PreToolUse` hook, or justified.
- The body has an `## Output` section with a fixed format or file path.
- The body tells the agent to reply `SCOPE_REQUEST: <what> — <why>` when blocked.
- The description says when to delegate and when not to.
- No template placeholders remain, and the body is under 40 lines.

Show the user the final file and the grill answer behind each field.

## Expand

Run this only when the user asks to widen an agent. Never propose it unprompted, and never
widen an agent as a side effect of another task. A `SCOPE_REQUEST` is a report to the user,
not permission to act on it.

1. **Get the evidence.** A `SCOPE_REQUEST` the agent returned, a denied tool call, a
   `maxTurns` cut-off, or a result that was wrong because the agent lacked access. "It might
   need it" is not evidence. Ask for the run, or find it in the transcript.
2. **Take the smallest step that fixes the observed failure**, trying these in order:
   - a better brief or preloaded skill (the agent lacked knowledge, not access)
   - a narrower tool form: one MCP tool instead of a whole server, `Bash` behind a hook
     allowlist instead of open `Bash`
   - one more tool
   - a higher `maxTurns`, in steps of about 50%
   - write access, which also brings `isolation: worktree` and a named file set
   - a stronger model
3. **If the need is wider than the job**, the agent's scope was wrong. Split it into a second
   agent rather than widening the first.
4. **Re-run the step 4 checklist**, then commit with a message that names the evidence:
   `agent(<name>): allow Bash(go test *) — SCOPE_REQUEST in run on #142`. Git history is the
   scope ledger, so the agent's prompt stays lean.

## Audit

Read each agent in `agents/` and `.claude/agents/` and run the step 4 checklist on it. Also
report what a checklist can't see: no clear reason to exist under rule 1, overlap with another
agent, or tools never used in practice. Propose cuts, not additions.
