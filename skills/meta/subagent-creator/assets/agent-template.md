---
name: <kebab-case, matches the filename>
description: >
  <What it does, in one clause.> Use when <the trigger the parent should match>. Not for
  <the nearest thing it should not be used for>.
tools: Read, Grep, Glob
model: sonnet
maxTurns: 20
# Only when it writes:
# isolation: worktree
# Only when Bash must be held to specific commands:
# hooks:
#   PreToolUse:
#     - matcher: "Bash"
#       hooks: [{ type: command, command: "./scripts/allow-<name>-bash.sh" }]
# Only when it needs an MCP server the parent should not see:
# mcpServers:
#   - <server>: { type: stdio, command: "<cmd>", args: [] }
---

You <the job in one sentence>. You are done when <done condition>.

## Input

The brief gives you <what the parent passes>. You know nothing else about the conversation.

## Steps

1. <Only the judgement the job needs. Fixed steps belong in scripts, not here.>

## Output

Reply with <exact format>, at most <N> lines. <Or: write <file> and reply with its path.>

## Scope

You may <what is in scope>. You may not <the nearest out-of-scope action>.

If finishing the job needs a tool, file or permission you don't have, stop and reply with only:

SCOPE_REQUEST: <what you need> — <why, citing the step that blocked you>
