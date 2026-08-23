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

A fan-out, where `ENG-13`, `ENG-14` and `ENG-15` all wait on `ENG-12` and on nothing else:

```toml
[[chain]]
tickets = ["ENG-12", "ENG-13"]

[[chain]]
base    = "feat/eng-12"
tickets = ["ENG-14"]

[[chain]]
base    = "feat/eng-12"
tickets = ["ENG-15"]
```

| Field | Default | Set it when |
| --- | --- | --- |
| `name` | filename stem | Always — cheap, and it labels the Batch in `tp ui` |
| `branch_prefix` | `feat/` | The repo uses something else |
| `base` | `main` | The repo's default branch is not `main` |
| `[[chain]]` | — | One per chain; `tickets` is ordered, bottom first |
| `[[chain]].base` | `base` | This chain roots on another chain's branch (a fan-out) |

`tickets` holds bare Refs — `ENG-12` on Linear, the issue number on GitHub. Treepad renders them
through `[from_spec] ticket_url` in `.treepad.toml` and derives branch names from them
(`feat/eng-12`). **If `ticket_url` is not configured in the repo, `tp batch sync` errors** — say so
when you write the file rather than leaving it to fail later. A full URL works in place of a Ref and
needs no config, but yields worse branch names, so prefer the Ref.

## Deriving chains from the blocking graph

A Chain is a **maximal linear run** of the step 4 blocking graph. Walk it:

- A slice with no blockers starts a new chain.
- A slice blocked by exactly one other joins that blocker's chain, directly after it.
- **Several slices blocked by the same one** is a fan-out. One of them continues the parent's chain.
  Each of the others starts a new chain with `base` set to the parent's **branch**, not its Ref:
  `base = "feat/70"` for a parent whose ticket is `70` under `branch_prefix = "feat/"`.
- Slices with no blocking relationship between them go in **separate chains**, never the same one.

The fan-out rule is the one that decides whether the Batch is worth running. Three siblings on one
parent flattened into `70 → 71 → 72 → 74` invents two dependencies and drops the width to 1. Split
into chains rooted with `base`, the same graph starts three agents the moment the parent's branch
exists.

One caveat to state when you write it: position 0 of every chain is unconditionally ready, so a
sibling chain does not wait for its parent's pull request. The first `tp batch sync` creates the
parent and errors on the siblings, whose base does not exist yet. The next tick creates them. Sync is
idempotent and a failed member stops only its own chain, so it resolves itself, but tell the user to
expect one tick of errors rather than letting it look like a broken Manifest.

That last rule is the one that matters. Chains within a Batch run in parallel and have no ordering
between them; putting independent tickets in one chain serialises their reviews for nothing.

**Chain depth is a review-latency multiplier.** Layer five cannot land until four reviews complete
below it, and every merge below rewrites the base under the agents still working above. Shallow and
wide beats deep and narrow. If the graph produces a chain deeper than about four, flag it to the
user rather than writing it silently.

A chain of one ticket is normal and correct — it just never becomes a GitHub Stack.

## The case treepad cannot model

One shared parent is modellable now, via the fan-out rule above. What is left is the join: a slice
blocked by **two or more** others has no faithful placement, because `base` takes one branch.
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
