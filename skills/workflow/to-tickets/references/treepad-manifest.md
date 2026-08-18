# The treepad Batch Manifest

`tp batch sync` reconciles a Manifest into a fleet of stacked worktrees — one worktree per ticket,
each branched from the one before it, with agents launched into them. Treepad never writes a
Manifest and never reads a Tracker, so the file is this skill's to produce.

## Where it goes

`<git-common-dir>/treepad/batches/<feature>.toml`, where `<feature>` is a slug of the feature the
run cut tickets for. Resolve the directory with `git rev-parse --git-common-dir` — in a linked
worktree `.git` is a file, not a directory, so `.git/treepad/...` is wrong.

Treepad unions every `*.toml` in that directory, so one file per feature is the shape.

## Shape

```toml
name = "silent-refresh"

[[chain]]
tickets = ["ENG-12", "ENG-13", "ENG-14"]

[[chain]]
tickets = ["ENG-20"]

[[chain]]
tickets = ["ENG-21"]
```

| Field | Default | Set it when |
| --- | --- | --- |
| `name` | filename stem | Always — cheap, and it labels the Batch in `tp ui` |
| `branch_prefix` | `feat/` | The repo uses something else |
| `base` | `main` | The repo's default branch is not `main` |
| `[[chain]]` | — | One per chain; `tickets` is ordered, bottom first |

`tickets` holds bare Refs — `ENG-12` on Linear, the issue number on GitHub. Treepad renders them
through `[from_spec] ticket_url` in `.treepad.toml` and derives branch names from them
(`feat/eng-12`). **If `ticket_url` is not configured in the repo, `tp batch sync` errors** — say so
when you write the file rather than leaving it to fail later. A full URL works in place of a Ref and
needs no config, but yields worse branch names, so prefer the Ref.

## Deriving chains from the blocking graph

A Chain is a **maximal linear run** of the step 4 blocking graph. Walk it:

- A slice with no blockers starts a new chain.
- A slice blocked by exactly one other joins that blocker's chain, directly after it.
- Slices with no blocking relationship between them go in **separate chains**, never the same one.

That last rule is the one that matters. Chains within a Batch run in parallel and have no ordering
between them; putting independent tickets in one chain serialises their reviews for nothing.

**Chain depth is a review-latency multiplier.** Layer five cannot land until four reviews complete
below it, and every merge below rewrites the base under the agents still working above. Shallow and
wide beats deep and narrow. If the graph produces a chain deeper than about four, flag it to the
user rather than writing it silently.

A chain of one ticket is normal and correct — it just never becomes a GitHub Stack.

## The case treepad cannot model

A Chain is strictly linear, so a slice blocked by **two or more** others has no faithful placement.
Put it at the end of the chain holding its longest blocker path, and tell the user which dependency
the Manifest drops. Nothing is lost from the record — the Linear `blockedBy` relation still holds
every blocker; only treepad's worktree base is approximate.

If that case is common in a run, the slices are probably too coarse. Say so.

## Never silently overwrite

`gh stack link` is additive only: treepad can build a Chain into a Stack but cannot reorder or
unlink one. **A Manifest edited after its Chain has been linked leaves a Stack on GitHub that no
treepad command can correct** — it has to be fixed by hand on github.com.

So if `<feature>.toml` already exists, stop and show the user the diff before writing. Only
overwrite on an explicit yes.
