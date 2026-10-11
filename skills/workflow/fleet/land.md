# Land

Run after the user has reviewed the wave's PRs and marked the good ones "Ready for review". A
draft PR means not approved, so landing never touches it. Everything mechanical lives in the
`fleet` CLI; this file only branches on its last stdout line.

```bash
fleet land "<label>"
```

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
- `suite-red #<n>`: the suite failed on the feature branch after landing, and bisecting the
  squash commits this run landed names PR `#<n>` as the culprit. Spawn a resolver subagent, giving
  it the culprit PR, the label and `feat/<slug>`, to fix the culprit in a follow-up PR into
  `feat/<slug>`; never open that PR by hand. Do not reconcile. Once the user has marked the
  follow-up PR "Ready for review", re-run `/fleet land <label>`: it lands the follow-up and
  re-runs the suite, and reconcile waits for `done`.

A non-zero exit is an error, not an outcome: show its stderr to the user and stop.
