# Plan: reflect from transcripts

> Source: grilling session 2026-10-07 (conversation), redesigning `skills/meta/reflect` and `tools/reflect`
> Label: reflect-transcripts

`/reflect` stops depending on a per-repo hook log. It reads Claude Code's own transcripts for a
session, fans out one reviewer per agent that ran, and turns approved findings into GitHub issues.
`/reflect metrics` survives on one global SessionStart hook.

## Technical design decisions

### Transcript layout (verified on Claude Code 2.1.282–2.1.285)

- Main agent: `~/.claude/projects/<slug>/<sid>.jsonl`. Locate by globbing
  `~/.claude/projects/*/<sid>.jsonl`; the session id is unique, so the slug never has to be derived.
- Subagent: `<sid>/subagents/agent-<aid>.jsonl` plus `agent-<aid>.meta.json`
  (`agentType`, `description`, `toolUseId`, `parentAgentId`, `spawnDepth`, `isFork`, `model`,
  `worktreePath`, `spawnedWithWorktree`). Nested agents sit flat in the same directory.
- Worktree-isolated subagent: the parent's `subagents/` holds only the `.meta.json`; the transcript
  is under the slug of `worktreePath` (with `/` and `.` mapped to `-`), with its own session id.
- Compaction stays in the same file and session id, marked by a `system` line with subtype
  `compact_boundary`. `/clear` starts a new session id.
- Assistant lines carry `message.usage` (`input_tokens`, `output_tokens`,
  `cache_read_input_tokens`, `cache_creation_input_tokens`).

### Agent linking, in order of preference

1. `meta.toolUseId` matches the parent's `tool_use.id`. The parent is `meta.parentAgentId`, or main when it's absent.
2. The parent's `tool_result` `toolUseResult.agentId`. This covers forked skills and older meta files without `toolUseId`.
3. The directory path alone. The agent is attached to main with no spawning call.

Every agent records which method linked it. Method 3 is reported as a warning.

### Session boundaries

- **Live agents.** An agent counts as live when its launch result is async and no later `<task-notification>` for its id reports a terminal status. If any agent is live, `reflect scan` exits non-zero and names the live agents.
- **Self cut-off.** The main transcript is read up to the user line that invoked `/reflect`. Earlier `/reflect` turns in the same session, and the agents they spawned, are excluded. A turn ends at the next real user prompt.

### Signals (deterministic, reusing `internal/detect`)

| Tag | Fires on |
|---|---|
| `halluc` | Tool error classified as a hallucination by `detect.IsHallucination`, including `<tool_use_error>` edit misses |
| `fail` | Any other failed tool call |
| `repeat` | The 3rd or later failure with the same `detect.Fingerprint` |
| `churn` | The 4th or later Edit/Write to the same file |
| `revert` | An Edit whose `new_string` equals an earlier `old_string` on the same file |
| `reread` | A Read of a file already read, with no edit to it in between |
| `big` | A tool result over 20k chars |
| `correction` | A user prompt (main agent only) with `detect.Correction` confidence ≥ 0.6 |

Each tagged line carries its fingerprint. The issue dedupe in phase 2 uses this fingerprint.

### Module boundaries (`tools/reflect`)

- `internal/transcript`: existing JSONL reader, extended with usage, compaction markers and meta files.
- `internal/session`: a deep module. `Load(sid) (Session, error)` locates every file, builds the agent tree, applies the cut-off and the live check. It is tested against a fixture session under `testdata/` that includes a nested fork, a worktree agent, a meta file without `toolUseId`, a compaction and an earlier `/reflect`.
- `internal/digest`: `Render(Agent) string` and `Signals(Agent) []Signal`, plus the triage rule. It is tested against the same fixture.
- `internal/detect`: unchanged.
- `internal/hook`: shrinks to a single SessionStart handler.
- `internal/metrics`: reads the session record and the transcripts.
- Removed: `internal/install`, `internal/logstore`, `internal/report`, and the rest of `internal/hook`.

### CLI contract (`~/.claude/bin/reflect`)

| Command | Contract |
|---|---|
| `scan <sid> [--cap 8]` | Writes one digest per agent to `$TMPDIR/reflect/<sid>/<aid>.txt` (main is `main.txt`). Prints an index: agent, type, model, depth, link method, calls, tokens, signal counts, `review` or `skip`, digest path. Exits non-zero on live agents |
| `slice <sid> <aid\|main> <line> [-C 20]` | Prints the raw transcript lines around a line, redacted (`internal/redact`) and clipped |
| `metrics` | As today: per instruction hash, confusion per 100 tool calls, correction rate and repeat failures per session, no verdict below 5 sessions, plus a change summary between hashes |
| `status` | Global hook installed?, number of session records, `cleanupPeriodDays` (warns below 90), library path |
| `hook` | SessionStart entry point (stdin payload) |

### Digest shape (one file per agent)

```
agent a7eb… Explore haiku depth=1 link=toolUseId parent=main
brief: <spawning prompt, clipped to 600 chars>
outcome: <parent's next action after the result, clipped>
 L12  Bash  rg blocker skills/                   ok    3.1k
 L14  Read  skills/tickets/x.md                  FAIL  [halluc fp=3f9a…]
 ──── compact_boundary (pre 180k → post 22k) ────
totals: 41 calls, 3 fail, 120k tok in / 8k out
```

### Triage rule

An agent gets a reviewer if it has any tag other than `big`/`reread`, or if it is in the top 25% of the session by tokens. Main is always reviewed. Above `--cap`, agents are ranked by signal count, then tokens; the rest are listed as not reviewed.

### Session record (for metrics)

- One global SessionStart hook in `~/.claude/settings.json`, installed by `setup.sh --reflect`.
- It appends `{sid, cwd, ts, instr_hash, files: {path: hash}, commit, branch, dirty}` to `~/.claude/reflect/sessions.jsonl`.
- It skips `source: compact` when the hash hasn't changed.
- It reuses today's `instrFiles` and `combinedHash`.

### Reviewer finding schema (YAML list in the reviewer's reply)

`category`, `claim`, `evidence` (`aid`, `L<line>`, `tool_use_id`), `cost` (calls, tokens), `cause`
(`brief` | `agent-def` | `skill` | `repo` | `agent`), `fix`, `fp`. A reviewer that finds nothing replies `clean: <one line>`.

### Issues

- **Repo.** Routed by the existing `references/routing.md`: library files go to the library repo, everything else to the current repo (`gh issue create --repo`).
- **Body.** `file-issue`'s three sections (Context, Where to look, Acceptance criteria) plus an Evidence section and a footer line `reflect-fp: <fp>`.
- **Labels.** `reflect` and `status:ready`.
- **Dedupe.** `gh issue list --state open --label reflect --search "reflect-fp: <fp> in:body"`. On a match, add a comment ("seen again in session `<sid8>`: n×") instead of filing a new issue.

---

## Phase 1: Scanner

**User stories**: no setup per repo; whole session including every subagent, nested agent and
worktree agent, linked by exact ids; compaction visible; refuse while agents run; never review the reflection itself; digest plus slices for reviewers.

### What to build

Add `internal/session` and `internal/digest`, plus the `scan` and `slice` commands, alongside the existing code. Nothing is removed yet. Build `testdata/` from a real session, trimmed and redacted, covering every case listed under Module boundaries.

### Acceptance criteria

- [ ] On the fixture, `reflect scan <sid>` links every agent and prints which method linked each one. The worktree agent's transcript is found under its own slug.
- [ ] Each digest shows the brief, the outcome, tagged lines with fingerprints, the compaction marker and token totals.
- [ ] Lines after the `/reflect` invocation, and the agents from an earlier `/reflect`, are absent.
- [ ] A fixture variant with an unfinished async agent makes `scan` exit non-zero and name that agent.
- [ ] `slice` prints redacted lines around a given line for main and for a subagent.
- [ ] Running `scan` on this repo's own sessions prints a sensible index in under 5s for the largest session (44 subagents, ~11MB main).

---

## Phase 2: The `/reflect` skill

**User stories**: one Sonnet reviewer per agent that needs one; reviewers see the brief and the outcome; error, economy and gap smells; repo-level checks only when a finding points at them; triage in chat; a report file; issues only on approval, deduped by fingerprint.

### What to build

Rewrite `skills/meta/reflect/SKILL.md`:

1. Run `reflect scan ${CLAUDE_SESSION_ID}`, or the given sid. On a non-zero exit, relay the error and stop.
2. In one message, spawn a `reflect-reviewer` agent (`agents/reflect-reviewer.md`, Sonnet) for each `review` row, passing the session id, agent and digest path.
3. Synthesise the findings: merge them by cause and fingerprint, rank by cost and recurrence, route them using `references/routing.md`, and run the repo-level checks (guardrail, oversized AGENTS.md, no-op instructions) only for findings that point at them. Present Accepted / Rejected / Backlog.
4. Write `.claude/reflect/reports/<date>-<sid8>.md` (the directory ignores itself in git) with all three lists and the agent index.
5. On approval, file or comment on issues as described under Issues.

Replace `references/proposal-format.md` with `references/issue-format.md`. The reviewer's category list, each with a "use when" trigger, lives in `agents/reflect-reviewer.md`.

### Acceptance criteria

- [ ] `/reflect` on a real session with subagents spawns exactly one reviewer per `review` row and none for `skip` rows.
- [ ] Every Accepted item cites an agent, a line and a fingerprint, and names the target file and the owning repo.
- [ ] Nothing is filed until the user approves. Approved items appear as issues labelled `reflect` in the right repo.
- [ ] Running `/reflect` again on the same session comments on the existing issues instead of filing duplicates.
- [ ] The report file exists, and a fresh repo needs no setup beyond the installed binary and skill.

---

## Phase 3: Swap the hooks

**User stories**: keep `/reflect metrics`; remove on/off and the per-repo hooks; discard old logs; raise transcript retention to 90 days; drop `corrections`, `log` and `prune`.

### What to build

- Add the global SessionStart hook: `setup.sh --reflect` builds the binary and merges the hook entry into `~/.claude/settings.json` idempotently.
- Move `metrics` onto the session record plus transcripts. It filters records to the current repo's checkout root (worktrees resolve to the main checkout) and computes counts from `internal/digest` signals.
- Rework `status`.
- Delete `on`, `off`, `show`, `proposals`, `log`, `prune`, `corrections`, and the removed packages.
- Add a one-shot `reflect uninstall-legacy`, which strips the old hook groups from `.claude/settings.local.json` and deletes `.claude/reflect/events.jsonl`, `edits/` and `instr.json`.
- Set `cleanupPeriodDays: 90` in the settings that `setup.sh` manages.
- Update the README and SKILL.md argument hint to `[<sid>|metrics|status]`.

### Acceptance criteria

- [ ] After `setup.sh --reflect`, a new session in any repo appends one record to `~/.claude/reflect/sessions.jsonl`. A second run of `setup.sh` doesn't duplicate the hook.
- [ ] `reflect metrics` groups sessions by instruction hash, using counts from transcripts, and keeps the below-5-sessions rule.
- [ ] `reflect uninstall-legacy` leaves `settings.local.json` with no reflect hooks and other entries untouched.
- [ ] No code path reads or writes `.claude/reflect/events.jsonl`, and `go test ./...` passes.
- [ ] The README describes only `/reflect [sid]`, `metrics`, `status` and the one-time setup.
