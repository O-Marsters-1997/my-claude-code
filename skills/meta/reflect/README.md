# reflect

## Setup

Once per machine, from a checkout of this repo (needs Go):

```bash
./setup.sh --reflect
```

Builds the machine-wide command `~/.claude/bin/reflect`. Nothing in `~/.claude/settings.json` changes.

Install the skill (project-level: drop `-g`):

```bash
npx skills add O-Marsters-1997/my-claude-code --skill reflect -g -y
```

Once per repo, from inside it, then restart the Claude session:

```
/reflect on
```

This writes the hooks into that repo's `.claude/settings.local.json`, pointing at the machine-wide command. Logs go to `.claude/reflect/` in the main checkout (a self-ignoring dir); git worktrees resolve to it. No other repo is affected.

## Commands

| Command | Does |
| --- | --- |
| `/reflect on` / `off` | Add / remove this repo's hooks in `.claude/settings.local.json` (`off` keeps the log) |
| `/reflect status` | State, log path, event and session counts, library path; warns if `cleanupPeriodDays` < 90 |
| `/reflect` | Analyse the current session and its subagents |
| `/reflect metrics` | Per instruction-file hash: confusion per 100 tool calls, correction rate, repeat failures per session; no verdict below 5 sessions |

Two more commands run from the shell as `~/.claude/bin/reflect`, not through the skill:

| Command | Does |
| --- | --- |
| `reflect log [--session id] [--kind k] [--last n]` | One readable line per event; `--session` takes an id prefix |
| `reflect prune [--older-than days]` | Replace sessions older than the cutoff (default `cleanupPeriodDays`) with one `summary` event each; run it when no session is active |

Deleting `.claude/reflect/events.jsonl` or the whole directory is safe; it is recreated on the next event. Hooks only exist while `settings.local.json` has them, so `off` or deleting that file stops logging.

## Use

Run `/reflect` at the end of a session. It writes at most 5 lean diff proposals to:

- `.claude/reflect/proposals/` for local files
- the skills library's `.claude/reflect/proposals/` for library files (skills, agents, rules, hooks)

It never edits instruction files. Apply a diff with `git apply` in the owning repo.

## How it works

Plain async command hooks in `tools/reflect`. No model calls, no output, always exit 0, so zero tokens.

| Hook event | Logged |
| --- | --- |
| SessionStart | commit, branch, dirty flag, instruction-file hash; a `manifest` event listing added, changed and removed instruction files when the set changed |
| UserPromptSubmit | prompts that look like corrections, with confidence |
| PostToolUseFailure | tool, error class, fingerprint, input, error |
| PostToolUse (Edit, Write) | file path, hashes of old and new content; the first edit to a file is held back until a second one arrives |
| Stop, SubagentStop | transcript sweep for missed tool errors |
| SessionEnd | sweep, plus tool-call, prompt and error counts |

Edit misses fire no hook, so Stop, SubagentStop and SessionEnd sweep the transcript for them.

Signals are derived when read:

| Signal | Fires when |
| --- | --- |
| hallucination | missing path, command or symbol, edit target not found, edit before read; qualifies if it recurs in-session or matches 2+ earlier sessions |
| repeat_fail | same command fails 3+ times |
| churn | same file edited 4+ times |
| revert | an edit is undone by a later one |
| correction | user corrections at confidence >= 0.6; qualifies at 2+ in session or 2+ earlier sessions |

Secrets are redacted before truncation. The log is append-only JSONL at `.claude/reflect/events.jsonl`, kept until deleted. Correction patterns are adapted from [claude-reflect](https://github.com/BayramAnnakov/claude-reflect) (MIT).

## Updating

After pulling changes to `tools/reflect`, run `./setup.sh --reflect` from the repo. Repos with reflect on use the rebuilt command from their next hook event; rerun `/reflect on` only if the hook events themselves changed. Skill text updates only when you rerun the `npx skills add` command above.
