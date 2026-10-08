---
name: fix-ci
description: >
  Fix failing GitHub Actions CI on pull requests fast: read the failed job logs, make the smallest
  fix, run one cheap local check, push, and let GitHub CI be the judge while it watches in the
  background. Takes an optional target: a run or job URL/ID (fix only that run's failures), a PR
  number/URL (fix every failing check on it), or nothing (the PRs in the conversation, else the
  current branch's PR). Handles stacked PRs by fixing the base first and rebasing dependents.
  Use whenever the user says "fix ci", "ci is failing", "checks failed", "the build is red",
  "fix the failing PRs", "both prs fail ci", pastes a GitHub Actions run link, or otherwise asks
  to get red checks green, even without saying "CI". Not for writing new CI workflows.
---

# Fix CI

The user wants red checks turned green with as little waiting as possible. GitHub CI is the
judge, not the local machine. Every minute spent re-running suites locally is a minute the
user waits for something CI is about to do anyway. Speed comes from three habits: diagnose from
the logs, make a narrow fix, push quickly.

## 1. Resolve the target

The argument decides what to fix:

| Given | Scope |
|---|---|
| Run URL or ID (`.../actions/runs/<run>`) | Only the failed jobs in that run. Find the PR from the run's head branch. |
| Job URL (`.../runs/<run>/job/<job>`) | Only that one job. |
| PR number or URL | Every failing check on the PR. |
| Nothing | The PRs named or discussed earlier in this conversation. If there are none, the current branch's PR. |

Never go through the user's other open PRs. If nothing resolves to a PR, say so and stop.

```bash
gh run view <run> --json headBranch,url,jobs      # run -> branch, failed jobs
gh pr view <branch> --json number,headRefName,baseRefName,url
gh pr checks <pr> --json name,bucket,link         # bucket=fail -> failing; link holds run/job ids
```

## 2. Diagnose from the logs

```bash
gh run view --job <job-id> --log-failed
```

Read the failure, then the code it points at. That is usually enough. Before deciding the cause,
check whether the failure is related to the PR's diff (`gh pr diff <pr> --name-only`).

**Looks flaky?** Signs are a timeout or race in code the PR didn't touch, a network or runner
error, or the same test passing on the base branch. If so, rerun once and change no code:

```bash
gh run rerun <run> --failed
```

Tell the user it looked flaky and offer to file an issue for it (the `file-issue` skill).
Hardening a flaky test inside this PR is scope creep the reviewer didn't ask for.

**Needs a deny-listed file?** Read the repo's `CLAUDE.md` (and any it links) for files agents may
not push, such as `.github/**`, `go.mod`, `go.sum`, `.golangci.yml`. If the only real fix touches
one of them, stop on that PR. Name the file and the change it needs, then carry on with the other
PRs. Do not work around the deny-list. The user has to make that change themselves.

## 3. Get a checkout

Never switch the branch in the user's main checkout, because they may have work there.

- `git worktree list`: if a worktree already has the PR's branch (e.g. from `/fleet`), work there
  and `git pull --ff-only` first.
- Otherwise create one. Use `tp` if it is on PATH (see the `treepad` skill), else
  `git worktree add`.

## 4. Fix narrowly

Fix only what the targeted checks reported. If you notice anything else (other lint, a smell,
a flaky neighbour), it goes in the final report as one line. Don't edit it.

Follow the repo's own rules from `CLAUDE.md` and `~/.claude/rules/comments.md`. A CI fix rarely
needs a comment.

## 5. One cheap check, then push

Run the narrowest command that would catch a typo-level mistake. It should take seconds, not
minutes:

- lint failure: the same linter on the touched package/files
- unit test failure: that one test, once (`go test ./pkg -run '^TestName$' -count=1`, or the
  repo's equivalent)
- build/type failure: build or vet the affected package
- slow e2e/integration failure: compile it (`go vet ./e2e/...` or `go test -run XXX ./e2e/...`).
  Do not run the suite. CI will.

Never run the full suite, never loop a test for stability, never "make sure" with repeated runs.
If the cheap check fails, fix it and run it again. That is still seconds.

Commit as a new commit, one per PR per round:

```
Fix CI: <what failed, briefly>

Co-Authored-By: <the trailer from the session's attribution reminder>
```

Then `git push` normally. Don't amend: the repo squash-merges, and a plain push keeps what
reviewers already saw.

## Stacked PRs

If one target PR's base branch is another target PR's head branch, work bottom-up:

1. Fix and push the base PR.
2. Rebase the dependent onto the base's new head and push with `--force-with-lease`.
3. Look at the dependent's failures again. Many were inherited and are now gone. Fix only what
   is still its own.

## Inline vs subagents

Work inline, one PR after another. Delegating adds start-up latency, and subagents drift: one
spent 15 minutes looping the e2e suite before pushing. Only fan out when there are 3 or more
independent (unstacked) PRs. Give each subagent this skill's rules word for word, especially
"one cheap check, never the full suite, push immediately".

## 6. Report, then watch in the background

Report straight away, one line per PR:

```
#439: e2e timeout in TestCrossRepoSlice → fixed in abc1234, CI watching
#440: rebased onto #439, golangci-lint unused param → fixed in def5678, CI watching
#441: blocked — needs go.mod bump (golang.org/x/sync v0.9), deny-listed
#442: TestPoll flaked (runner timeout) → rerun, no code change
Noticed, not fixed: #439 internal/web has a vet shadow warning
```

Then start one background watch per pushed PR (Bash with `run_in_background`), and end your turn:

```bash
gh pr checks <pr> --watch --interval 30
```

When a watch finishes:

- **Green**: report one line, e.g. `#439: CI green`.
- **Still red on the targeted checks**: go back to step 2 by yourself, without asking. Allow
  at most 2 automatic rounds per PR after the first push. After that, stop and report what
  keeps failing and what you tried. A third identical failure means you are guessing.
- **Red only on checks outside the target** (e.g. you were given one job): report them and
  don't fix them.
