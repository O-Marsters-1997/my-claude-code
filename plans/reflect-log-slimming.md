# Plan: Reflect log slimming

> Source: conversation analysing `job-scraper/.claude/reflect/events.jsonl` (2026-09-29)

## Technical design decisions

- **Problem.** 95% of the sample log is the `files` map on the `session` event (133 paths and hashes, ~12.5 KB per session, read by nothing). Once real sessions run, `edit` events (~500 bytes each, ~45% of it repeated `cwd` and `transcript`) become the bulk.
- **Schema.** Bump `schemaVersion` to 2. The reader accepts v1 and v2 lines. Fields a v2 writer drops are resolved by the reader, and v1 values win when present.
- **Transcript resolution.** `transcript` is stored only on `session` events; sweep events carry `agent_id`, which is enough to derive their path. For any other event the reader derives it from the session's `session` event plus `agent_id`, using `transcript.SubagentPath`. `report.Show` prints the resolved path.
- **Manifest.**
  - `session` carries `instr_hash` only.
  - New kind `manifest`: `{instr_hash, prev, added, changed, removed}`, where each of the three is a map or list of path to hash12. It is written only when `instr_hash` differs from the repo's last recorded hash.
  - The first manifest for a repo has `prev` empty and lists everything under `added`.
  - Roots are deduplicated by resolved path, so `.claude/skills/x` and `.agents/skills/x` symlinks or mirrors count once, and the hash is computed over the deduplicated set.
- **Module boundaries.**
  - `logstore` owns the schema, append, read and transcript resolution.
  - `hook` owns capture and decides what is worth writing.
  - `report` owns derivation and rendering (`show`, `metrics`, new `log`).
  - `install` is untouched unless phase 6 changes the hook event list.
- **Size guard.** A test builds one of each event kind with worst-case truncated fields and asserts the marshalled line is under 2 KB.
- **`tool_error` policy.**
  - `input` is truncated to 150 characters.
  - `error` keeps its last 300 characters.
  - After the third event with the same `fp` in a session, further events are written bare (no `input`, no `error`).
  - `Append` callers dedupe by `(session_id, tool_use_id)` for `tool_error`.
- **Compatibility rule.** Every phase ships with a test that reads a fixture of v1 lines and produces the same `show` and `metrics` output as before.
- **Rollout.** `./setup.sh --reflect` rebuilds `~/.claude/bin/reflect` in place and existing repo hooks pick it up. Only a change to `install.go`'s event list needs `/reflect on` rerun in each repo and a session restart.

---

## Phase 1: Slim per-event context

**User stories**: as an agent reading the log, I don't pay for the same paths on every line.

### What to build

`event()` stops setting `Cwd` and `Transcript` for every kind except `session` (and sweep-recorded `tool_error`, which carries its own resolved path). `logstore` gains a resolver that fills `Transcript` on read for events that lack one, from the session's `session` event and `agent_id`. Add the 2 KB per-line size-guard test. Bump `schemaVersion` to 2.

### Acceptance criteria

- [ ] New `edit`, `correction`, `tool_error` (hook-written) and `session_end` lines contain no `cwd` or `transcript` keys
- [ ] `reflect show` prints the same `transcript=` path for a v2 event as it did for the equivalent v1 event, including subagent paths
- [ ] Size-guard test fails when a field is added that pushes any kind over 2 KB
- [ ] A v1 fixture log produces unchanged `show` and `metrics` output

---

## Phase 2: Diff-based instruction manifest

**User stories**: as a maintainer, I can see which instruction files changed between two hashes without every session paying for the full list.

### What to build

Remove `Files` from `session` events. `instrFiles` deduplicates roots by resolved path. On SessionStart, compare against the repo's last recorded hash (read from the log) and, if it differs, append one `manifest` event with added, changed and removed entries. `reflect show` and `metrics` gain an `instr` line naming the files that changed since the previous hash.

### Acceptance criteria

- [ ] A session with an unchanged hash writes a `session` event under 500 bytes and no `manifest`
- [ ] Editing one skill file produces exactly one `manifest` event listing one `changed` entry
- [ ] Mirrored `.claude/skills/x` and `.agents/skills/x` count once, and the combined hash is stable when only the mirror changes
- [ ] The first manifest for a repo lists every file under `added`
- [ ] v1 `session` events that still carry `files` are read without error

---

## Phase 3: `tool_error` trimming and dedupe

**User stories**: as an agent reading the log, failures are compact and appear once.

### What to build

Apply the `tool_error` policy above. Start with a failing test that reproduces the suspected double-write: a `Read` of a missing file recorded by the PostToolUseFailure hook and again by the sweep. Then dedupe by `(session_id, tool_use_id)`. Confirm `repeat_fail` still fires on the third occurrence and `hallucination` qualification is unchanged.

### Acceptance criteria

- [ ] The same failure reaching both hook and sweep produces one event
- [ ] `input` is at most 150 characters and `error` at most 300, keeping the tail
- [ ] The fourth identical `fp` in a session is written without `input` or `error`, and `show` still reports `x4`
- [ ] Redaction still runs before truncation

---

## Phase 4: Session event hygiene

**User stories**: as a maintainer, long sessions don't accumulate redundant session lines, and `session_end` tells me whether the session was quiet.

### What to build

Skip the `session` event for source `compact` when the same session already logged one with the same branch and `instr_hash`. `clear` is always logged: it may mint a new session id, and every session id needs a `session` event to resolve its transcript and hash. Extend `session_end` with a count of `tool_error` events for the session, alongside the existing `tool_calls` and `prompts`.

### Acceptance criteria

- [ ] A compact with unchanged branch and hash writes nothing
- [ ] A compact after a branch switch or a hash change, and any `clear`, still writes a `session` event
- [ ] `session_end` carries `errors` and `metrics` ignores its absence on old lines

---

## Phase 5: `reflect log` pretty-printer

**User stories**: as a user, I can read the log in the terminal without `jq`.

### What to build

`reflect log [--session id] [--kind k] [--last n]` prints one line per event: time, kind, agent, class or file, and a truncated detail. It reads through the resolver so v1 and v2 lines look the same. Document it in the README command table.

### Acceptance criteria

- [ ] Output is one line per event and fits 120 columns by default
- [ ] `--session` and `--kind` filter correctly, and `--last` limits from the end
- [ ] Malformed lines are skipped, matching `Read`
- [ ] README lists the command

---

## Phase 6: Edit-event reduction

**User stories**: as a maintainer, the highest-volume event kind is written only when it can matter.

### What to build

Implemented as option A without the measuring spike, because it is lossless: the first edit to a file is held in `edits/<session>/` and is written to the log only when a second edit to the same file arrives, so `churn` and `revert` see exactly the events they did before. Files edited once, which can never be churn or revert, are never logged. State is deleted at SessionEnd and by `prune`. The bytes saved depend on how many files a session edits once, so measure that on a week of real logs. The options considered were:

- **A.** Keep a per-session state file and write an `edit` event only from the second edit to a file onward, plus the first when its hashes are needed for `revert`.
- **B.** Remove the PostToolUse hook and derive edit events in the sweep from the transcript.

Whichever is chosen must produce identical `churn` and `revert` results on a recorded fixture. Option B changes the hook event list, so it needs `/reflect on` rerun in each repo.

### Acceptance criteria

- [ ] A written measurement summary is attached to the plan before implementation starts
- [ ] `churn` (4+ edits to a file) and `revert` derive the same results as before on a fixture session
- [ ] Median `edit` bytes per session drop by at least half against the measured baseline
- [ ] If the hook list changed, `README.md` says to rerun `/reflect on`

---

## Phase 7: `reflect prune`

**User stories**: as a user, the log has a bounded size.

### What to build

`reflect prune [--older-than days]`, defaulting to `cleanupPeriodDays`. It rewrites the log, replacing raw events for sessions older than the cutoff with one `summary` event per session: signal counts by kind, `instr_hash`, `tool_calls`, `prompts`. It writes to a temp file and renames, and it keeps the most recent sessions untouched. `metrics` reads `summary` events as well as raw ones.

### Acceptance criteria

- [ ] Pruning never changes `metrics` output for the pruned sessions beyond dropping per-event detail
- [ ] An interrupted prune leaves the original log intact
- [ ] Sessions inside the cutoff are byte-identical after pruning
- [ ] `status` reports the log size

---

## Phase 8: Rollout docs

**User stories**: as the maintainer, I know exactly how to ship a change to this tool.

### What to build

Update `tools/reflect` README and `skills/meta/reflect/SKILL.md` for the new commands, the v2 schema and the manifest event. Add an "Updating" section: `go test ./...`, `./setup.sh --reflect`, when `/reflect on` must be rerun, reinstalling the skill with `npx skills add` from merged main, and how to verify with `reflect status` and the line-size check.

### Acceptance criteria

- [ ] README documents `log` and `prune`
- [ ] The Updating section states that only hook-list changes need `/reflect on` rerun
- [ ] SKILL.md still describes the analysis flow correctly against the v2 output
