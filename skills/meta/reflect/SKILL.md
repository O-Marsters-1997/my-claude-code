---
name: reflect
description: >
  Review a session's transcripts (the main agent and every subagent it spawned) for confusion
  and waste, such as hallucinated paths and symbols, repeated failures, edit churn, user
  corrections, expensive or redundant tool calls and long hunts for a file, then file GitHub
  issues for the fixes the user approves. Also reports instruction-file metrics. Use only when
  invoked as /reflect, /reflect <session id>, /reflect metrics or /reflect status.
disable-model-invocation: true
argument-hint: "[<session id>|metrics|status]"
allowed-tools: Bash(~/.claude/bin/reflect *), Bash(gh issue *), Bash(gh label *), Bash(git -C *), Read, Grep, Glob, Write, Agent
---

# Reflect

`~/.claude/bin/reflect` reads Claude Code's own transcripts, so nothing has to be switched on
first. This skill never edits instruction files: approved fixes become GitHub issues.

Arguments: `$ARGUMENTS`

## Metrics and status

- `metrics`: run `~/.claude/bin/reflect metrics --exclude ${CLAUDE_SESSION_ID}`.
- `status`: run `~/.claude/bin/reflect status`.

Print the output verbatim and stop.

## Review a session

1. **Scan.** Run `~/.claude/bin/reflect scan <sid>`, where `<sid>` is the argument if one was
   given, else `${CLAUDE_SESSION_ID}`. On a non-zero exit, relay the message and stop; live agents
   mean the user waits and reruns. The scan already leaves out this `/reflect` turn and earlier ones.
2. **Fan out.** In one message, spawn with `subagent_type: reflect-reviewer`:
   - one `Agent` per `cluster` line of the index, with the prompt `Session: <sid>`, `Cluster:
     <mechanism>` and one `<agent> L<n>` instance per line after it;
   - one `Agent` per index row whose verdict is `review`, with three lines: `Session: <sid>`,
     `Agent: <agent>` and `Digest: <digest path>`. Clustered events are already excluded from the
     verdicts, so agents left with no flags are `skip`.

   Do not read the digests yourself first.
3. **Synthesise** the reviewers' findings:
   - Drop any finding whose claim or rejected alternative carries no `evidence`.
   - Merge findings that share a cause and a target file into one item, keeping every `fp`.
     Per-agent findings that share a cause across two or more agents become a new cluster item.
   - Drop a finding no instruction, skill, agent definition, hook or check could have prevented.
   - Route each item per [references/routing.md](references/routing.md). A deterministic fix
     (hook, lint rule, CI job) beats a sentence of prose.
   - Repo-level checks run only when an item points at them: a mistake a linter would catch means
     checking for a pre-commit hook or CI job running lint, typecheck and tests; an ignored
     instruction means checking whether `AGENTS.md`/`CLAUDE.md` is oversized or the line is a no-op.
   - **Verify** every cluster fix: apply it to a copy of the mechanism's source in the scratchpad,
     then run `~/.claude/bin/reflect replay` on each instance against it. A fix that still blocks
     an instance goes back to its reviewer once for rework, else the item is Rejected.
   - Sort into **Accepted** (worth an issue now), **Backlog** (real but one-off or cheap; listed,
     not filed) and **Rejected** (with the reason).
4. **Report.** Write `<YYYY-MM-DD>-<first 8 of sid>.md` in the directory printed by
   `~/.claude/bin/reflect reports`: the scan index, then the three lists with evidence, then the
   `overflow` agents as not reviewed.
5. **Triage in chat.** Number the Accepted items, one line each with target file and owning repo.
   One line per Backlog and Rejected item, and a count of `skip` and `overflow` agents. Ask which
   Accepted items to file, then wait.
6. **File** the approved items per [references/issue-format.md](references/issue-format.md). Reply
   with one line per item: the issue URL, and whether it is new or a comment on an existing one.
