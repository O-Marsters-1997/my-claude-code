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
  listed paths, and every later PR is still unlanded. Hand it to the user: name the PR, branch and
  paths, ask them to merge the feature branch into it and push, then re-run `/fleet land <label>`.
  Do not reconcile.
- `suite-red`: the suite failed on the feature branch after landing. Report it to the user and
  stop. Do not reconcile.

A non-zero exit is an error, not an outcome: show its stderr to the user and stop.
