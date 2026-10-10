---
name: triage-learnings
description: >
  Routes pending LEARN-marker learnings from the reflect ledger to GitHub issues. Use when the user
  says "triage learnings", "/triage-learnings", "process my LEARN markers", "what have I learned",
  or when the statusline shows learn:N. Infers missing skill and scope, drops duplicates and
  lessons already covered by a rule or lint config, proposes routes for the user to accept, files
  one issue per skill, and writes every decision back to the ledger with `reflect learn mark`.
---

# Triage learnings

Turn pending learnings into filed issues, and leave none pending that were handled. Triage files
issues only; it never edits a skill (`skill-updater`, `/implement` and `/fleet` do that).

`reflect` is `~/.claude/bin/reflect`. Routing: `skills/meta/reflect/references/routing.md`. Issue
body, labels, dedupe and the `-R` repo flags: `skills/meta/reflect/references/issue-format.md`.

## Steps

1. **Load.** `reflect learn ls --json --status pending`. Empty means say so and stop.
2. **Infer.** For a learning with no `skill` or no `scope`, read its `text`, `before`, `after`
   and file, then match it to a repo skill (`<repo>/.claude/skills/`) or a library skill
   (`reflect status` gives the `library:` path). Resolve an installed copy to its library source
   as `routing.md` describes.
3. **Drop.** Mark `rejected` any learning that duplicates another pending one (keep the one with
   the fullest `before`/`after`; mark the rest `rejected`), or that an existing skill, `AGENTS.md`,
   rule or lint config already states. Check by reading the target; do not assume.
4. **Propose.** Show a table, one row per learning: id, text, route, reason. Wait for the user to
   accept, change or reject each row. Never file before this.
5. **File** the accepted rows (below), then **mark** each.
6. **Report** the issue URLs and counts of promoted, deferred and rejected.

## Routes

| Learning | Action | Mark |
| --- | --- | --- |
| scope global, names a library skill | one issue on the skill's source repo | `promoted` |
| scope repo, names a repo skill | one issue on the learning's repo to amend that skill | `promoted` |
| kind `later`, scope repo | `inbox` issue on that repo via `file-issue` inbox mode | `promoted` |
| `LEARN(repo)`, or no owner and seen once | file nothing | `deferred` |
| user rejects, or a drop from step 3 | file nothing | `rejected` |

Group by skill: one issue per skill per run, with every learning for that skill as an item. Use
the reflect issue format with labels `reflect` and `status:ready`, and add the `from:` line for
library issues. Put each learning's `before`/`after` under Evidence.

- **Mechanical lessons** (a script or linter could catch it): the issue also asks for a snippet in
  the skill's `lint/` folder, in the house forms (forbidigo, errorlint for Go).
- **Public repos** (`gh repo view <repo> --json isPrivate`): rewrite each `before`/`after` as a
  generic example, and leave out other repos' names, paths and private identifiers. An issue on the
  private repo the code came from keeps the code as written.

## Mark

After each issue is filed, for every learning in it:

```bash
reflect learn mark <id> promoted --issue <issue-url> [--scope global|repo] [--skill name]
reflect learn mark <id> deferred
reflect learn mark <id> rejected
```

Mark only after the issue exists, so a failed filing leaves the learning pending for the next run.
Pass `--scope` and `--skill` when step 2 inferred them. Re-running `ls --status pending` must show
nothing already handled.
