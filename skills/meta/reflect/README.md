# reflect

## Setup

Once per machine, from a checkout of this repo (needs Go):

```bash
./setup.sh --reflect
```

This builds `~/.claude/bin/reflect` and links the `reflect-reviewer` agent into `~/.claude/agents/`.
The SessionStart hook that records which instruction files each session started with, and
`cleanupPeriodDays: 90`, come from `settings.shared.json`, which `setup.sh` merges into
`~/.claude/settings.json`.

Install the skill (project-level: drop `-g`):

```bash
npx skills add O-Marsters-1997/my-claude-code --skill reflect -g -y
```

Nothing is set up per repo. If a repo still has hooks from the old `/reflect on`, run
`~/.claude/bin/reflect uninstall-legacy` inside it once.

## Commands

| Command | Does |
| --- | --- |
| `/reflect` | Review the current session: the main agent and every subagent, nested agent, forked skill and worktree agent it ran |
| `/reflect <session id>` | Review a past session, as long as its transcripts are still kept |
| `/reflect metrics` | Per instruction-file hash: confusion per 100 tool calls, correction rate, repeat failures per session; no verdict below 5 sessions |
| `/reflect status` | Whether the SessionStart hook is installed, session record count, transcript retention, library path |

The binary also runs directly:

| Command | Does |
| --- | --- |
| `reflect scan <sid> [--cap 8]` | Link the session's agents, write one digest per agent to `$TMPDIR/reflect/<sid>/`, print the index with each agent's triage verdict |
| `reflect slice <sid> <agent\|main> <line> [-C 20]` | Print redacted transcript lines around a line |
| `reflect reports` | Print (and create) the report directory |
| `reflect uninstall-legacy` | Strip old per-repo hooks from `.claude/settings.local.json` and delete the old event log |

## How it works

1. `scan` finds `~/.claude/projects/*/<sid>.jsonl` and every agent under `<sid>/subagents/`. Each
   agent links to the tool call that spawned it by exact id: the `.meta.json` `toolUseId` first, then
   the parent's `toolUseResult.agentId`, then the directory alone (reported as a warning).
   Worktree agents' transcripts are found under their worktree's project folder.
2. It refuses while any agent is still running, and drops every `/reflect` turn along with the agents
   those turns spawned.
3. Each digest is a timeline with deterministic tags, token totals and compaction markers:

   | Tag | Fires on |
   | --- | --- |
   | `halluc` | missing path, command or symbol, edit target not found, edit before read |
   | `fail` / `repeat` | a failed call; the third or later failure with the same fingerprint |
   | `churn` / `revert` | the fourth or later edit to a file; an edit undoing an earlier one |
   | `reread` / `big` | a Read of an unchanged file already read; a result over 20k chars |
   | `correction` | a user prompt that reads as a correction, confidence ≥ 0.6 |

4. Main, plus every agent with an actionable tag or in the top quarter by tokens, gets a
   `reflect-reviewer` agent (Sonnet, read-only, Bash held to `reflect slice` by
   `hooks/allow-reflect-slice.sh`), up to the cap. Reviewers return structured findings.
5. The skill merges and routes them, writes `.claude/reflect/reports/<date>-<sid8>.md`, shows
   Accepted / Backlog / Rejected, and files issues labelled `reflect` only for what you approve. A
   finding whose fingerprint already sits on an open issue becomes a comment there.

Secrets are redacted in digests and slices. Correction patterns are adapted from
[claude-reflect](https://github.com/BayramAnnakov/claude-reflect) (MIT).

## Updating

After pulling changes to `tools/reflect`, rerun `./setup.sh --reflect`. Skill text updates only when
you rerun the `npx skills add` command above.
