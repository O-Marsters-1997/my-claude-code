---
name: reflect
description: >
  Review this session's logged agent confusion (hallucinated paths and symbols,
  repeated failures, edit churn, user corrections) and propose amendments to
  AGENTS.md, skills, agents, rules or hooks. Also switches per-repo logging on
  and off. Use only when invoked as /reflect, /reflect on, /reflect off,
  /reflect status or /reflect metrics.
disable-model-invocation: true
argument-hint: "[on|off|status|metrics]"
allowed-tools: Bash(~/.claude/bin/reflect *), Read, Grep, Glob, Write
---

# Reflect

Logging is capture-only and runs in hooks (`tools/reflect`). This skill reads the log for
the current session and writes proposals. It never edits instruction files.

Arguments: `$ARGUMENTS`

## Switch and metrics

If the first argument is `on`, `off`, `status` or `metrics`, run
`~/.claude/bin/reflect <argument>`, print its output verbatim and stop. Logging is off
in every repo until `/reflect on`.

## Analyse this session

1. Run `~/.claude/bin/reflect show ${CLAUDE_SESSION_ID}`. It covers this session and every
   subagent it spawned, and nothing else. If it prints `no qualifying signals`, say so and
   stop. If the log is missing, tell the user to run `/reflect on` and stop.
2. For each finding, read the transcript around its `tool_use_id`. Grep the id in the
   `transcript=` path and read about 20 lines either side. Transcripts older than
   `cleanupPeriodDays` are gone; work from the logged input and error and say so.
3. Infer why the agent struggled. Ask what an instruction, skill or hook would have had to
   say to prevent it. A finding with no such answer is dropped, not padded.
4. Route each amendment per [references/routing.md](references/routing.md).
5. Write proposals per [references/proposal-format.md](references/proposal-format.md).
6. Reply with the proposal file paths and one line per item. Do not apply anything.
