# Land

Run once the wave's PRs are all open, before the user merges anything. Land never merges. It makes
every open PR of the wave (drafts included) mergeable by the user, one after another in order,
with no conflicts and green CI, by merging the right branches into the PR branches. Everything
mechanical lives in the `fleet` CLI; this file only branches on its output and spawns subagents.

```bash
fleet land "<label>" --state "<scratchpad>/fleet/state.md" [--resolved "<line>"]...
```

Each run waits for every PR's checks to finish, simulates merging the PRs in order onto
`feat/<slug>`, regenerates generated-only conflicts itself, and writes the plan JSON (default
`.claude/fleet-land-plan.json`; `--plan` moves it). Every run, finished or stopped, appends one row
to `.claude/fleet-metrics.jsonl`. `fleet land "<label>" --check` only reads and simulates, spawns
nothing and pushes nothing, and prints `land` or `reconcile`.

Progress goes to stderr. Stdout has one line per PR, in merge order, then one terminal line:

```
#<PR> <branch> ci=green|red|pending|none conflicts=<paths|none> pull-in=<#PR,...|none> <status>
```

`<status>` is `ok`, `needs-work`, `gate-failed <reason>` or `blocked-by #<PR>`. Read the terminal
line:

- `needs-work #<PR>...`: those PRs each need exactly one fresh subagent (see Subagent below). Run
  them in the order of the plan's `pull_in` graph: a PR's subagent starts only after the subagents
  of every PR it pulls in have returned, so it merges their final tips. PRs with no pull-in
  relation run in parallel. If a pulled-in PR's subagent returns `failed`, do not start the
  dependent PR's subagent. Then re-run `fleet land` with every return line as `--resolved`.
- `ready <n>`: the final ordered simulation is conflict-free and the suite is green on it. Tell the
  user the order to squash-merge the PRs into `feat/<slug>`, and that `/fleet reconcile <label>`
  follows. Land has already commented a compare link on each PR that pulled in earlier ones, so
  its diff shows only its own changes.
- `gate-failed #<PR> <reason>`: `<reason>` is `resolver-failed`, `ci-red` (still red after its
  subagent), `ci-pending` (checks did not finish within `--ci-timeout`; re-run later) or
  `conflicts-remain`. Name the PR and the reason, list the PRs the plan shows as `blocked-by` it,
  and hand them to the user. No retries and no second subagent. Reconcile is still next.
- `gate-failed #<PR> <cmd>`: any other text after the PR is a `pre_push` command: a
  regenerate-only fix failed it before anything was pushed. No per-PR lines precede it. Report it
  and stop.
- `regen-stale <path>...`: re-running the `[regen]` commands on the final simulated result changed
  these generated files. Report the paths and stop.
- `suite-red #<PR>`: the suite failed on the final simulated result, and bisecting over the chain
  names this PR. Report it; do not spawn another subagent.

After every outcome, recommend `/fleet reconcile <label>` once the user has merged what they
chose to merge. A non-zero exit is an error, not an outcome: show its stderr (for example a
missing `feat/<slug>`, or open tickets whose PRs target another base) and stop.

## Subagent

One fresh subagent per `needs-work` PR, never the ticket's own, and never a second for the same PR.
It fixes CI and conflicts together. Brief it with the plan path and its PR number (it reads the
plan for failing checks, conflicting paths and the PRs to pull in), the PR's diff and ticket
intent, the intent of each pulled-in PR, its worktree via `tp`, and `[checks] pre_push` as the
gate. It must:

1. Merge `origin/feat/<slug>` and then each `pull_in` branch's tip into the PR branch. Never
   rebase and never force-push. Resolve each conflict once.
2. Fix every failing check listed for the PR.
3. Keep both intents, or name what it dropped. Invent no behaviour.
4. Run every `pre_push` command; a red gate means it stops and reports `failed`.
5. Push without force, and comment on the PR the touched files and any trade-off.
6. Write a report and return exactly one line: `<PR> resolved|failed append|rewrite <report path>`
   (`append` when both sides only added code, `rewrite` when it had to change existing lines).

Pass each return line verbatim as `--resolved "<line>"` on the re-run.
