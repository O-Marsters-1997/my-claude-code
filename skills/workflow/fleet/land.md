# Land

Run after the user has reviewed the wave's PRs and marked the good ones "Ready for review". A
draft PR means not approved, so landing never touches it. Everything mechanical lives in the
`fleet` CLI; this file only branches on its last stdout line.

```bash
fleet land "<label>" --state "<scratchpad>/fleet/state.md"
```

Every run, finished or stopped, appends one row to `.claude/fleet-metrics.jsonl` and prints it to
stderr. Pass each resolver return line as `--resolved "<line>"`. `fleet metrics [--label X]
[--last N]` shows the rows, a trend and any crossed threshold.

Progress goes to stderr and is for the user. Read only the one line on stdout:

- `done <n>`: every PR in the wave landed and the suite is green. Run reconcile
  ([reconcile.md](reconcile.md)) for the label.
- `drafts <PR>...`: everything marked ready has landed. Tell the user which PRs remain drafts,
  that they need review and "Ready for review", and to re-run `/fleet land <label>` afterwards.
  Do not reconcile.
- `needs-resolver <PR> <branch> <path>...`: that PR conflicts with the feature branch in the
  listed paths, and every later PR is still unlanded. Spawn exactly one fresh resolver for that PR
  (see Resolver below), never the ticket's own subagent. Do not reconcile yet.
- `gate-failed <PR>`: the resolver's `pre_push` gate failed for that PR. Handle it as `failed`
  under Resolver below.
- `suite-red #<n>`: the suite failed on the feature branch after landing, and bisecting the
  squash commits this run landed names PR `#<n>` as the culprit. Spawn a resolver subagent, giving
  it the culprit PR, the label and `feat/<slug>`, to fix the culprit in a follow-up PR into
  `feat/<slug>`; never open that PR by hand. Do not reconcile. Once the user has marked the
  follow-up PR "Ready for review", re-run `/fleet land <label>`: it lands the follow-up and
  re-runs the suite, and reconcile waits for `done`.

## Resolver

Brief the resolver with: both PRs' diffs (the conflicting PR and the PRs already landed on the
feature branch that touch the conflicted paths), both tickets' intent, the conflicted paths, its
worktree via `tp`, and `[checks] pre_push` from `.claude/fleet.toml` as the gate. It must:

1. Merge `origin/feat/<slug>` into the PR's branch and resolve the conflict once.
2. Keep both intents, or name what it dropped. Invent no behaviour.
3. Run every `pre_push` command; a red gate means it stops and reports `failed`.
4. Push without force, and comment on the PR the touched files and any trade-off.
5. Write a report and return exactly one line: `<PR> resolved|failed append|rewrite <report path>`
   (`append` when both sides only added code, `rewrite` when it had to change existing lines).

On `resolved`, re-run `fleet land "<label>" --resolved "<line>"`, passing the line verbatim; the
remaining PRs then continue in order. On `failed`, record it once with
`fleet land "<label>" --resolved "<line>"`, which lands nothing and prints `gate-failed <PR>`. Then
stop: name the PR, say the later PRs are still unlanded, hand the PR to the user, and do not
reconcile. Never spawn a second resolver for the same PR.

A non-zero exit is an error, not an outcome: show its stderr to the user and stop.
