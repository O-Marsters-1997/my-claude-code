---
name: triage-learnings
description: >
  Routes pending LEARN-marker learnings from the reflect ledger to GitHub issues. Use when the user
  says "triage learnings", "/triage-learnings", "process my LEARN markers", "what have I learned",
  or when the statusline shows learn:N. Infers missing skill and scope, merges duplicates into one
  item with every example, drops lessons already covered by a rule or lint config, proposes routes
  for the user to accept, files one issue per skill, and writes every decision back to the ledger with `reflect learn mark`.
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
3. **Merge and drop.** Merge pending learnings that say the same thing into one item, keeping
   every example (file, line, `before`/`after`) and a count such as "seen in 4 places": the
   repetition is the evidence, and each merged learning still counts toward a proposal's 3+.
   Mark `rejected` only a learning that an existing skill, `AGENTS.md`, rule or lint config
   already states. Check by reading the target; do not assume.
4. **Propose.** Show a table, one row per item: ids, text, count, route, reason. Wait for the user
   to accept, change or reject each row. Never file before this.
5. **File** the accepted rows (below), then **mark** each.
6. **Propose upward** (below) over the deferred and promoted learnings.
7. **Report** the issue URLs, the proposals, and counts of promoted, deferred and rejected.

## Routes

| Learning | Action | Mark |
| --- | --- | --- |
| scope global, names a library skill | one issue on the skill's source repo | `promoted` |
| scope repo, names a repo skill | one issue on the learning's `repo` field (pass `-R <owner>/<repo>` from its `origin`) to amend that skill | `promoted` |
| kind `later`, scope repo | `inbox` issue on that repo via `file-issue` inbox mode | `promoted` |
| scope repo with no skill (`LEARN(repo)`), or no owner and no other learning shares its theme | file nothing | `deferred` |
| user rejects, or a drop from step 3 | file nothing | `rejected` |

A merged item takes one route, and every learning in it gets that item's mark and `--issue` URL.

Group by skill: one issue per skill per run, listing every item for that skill. Use
the reflect issue format with labels `reflect` and `status:ready`, and add the `from:` line for
library issues. Put each item's count and every example's `before`/`after` under Evidence.

- **Mechanical lessons** (a script or linter could catch it): the issue also asks for a snippet in
  the skill's `lint/` folder, in the house forms (forbidigo, errorlint for Go).
- **Global issues** always carry generic examples: rewrite each `before`/`after`, and leave out
  repo names, paths and private identifiers, whether or not the library repo is public.
- **Repo issues** on a public repo (`gh repo view <repo> --json isPrivate`) are genericised too. An
  issue on the private repo the code came from keeps the code as written.

## Proposals

After marking, run `reflect learn ls --json` and consider every learning with status `deferred` or
`promoted` and an empty `proposal` field. A learning already carrying a `proposal` URL belongs to a
filed group: skip it, so a group is never proposed twice.

| Group | Proposal |
| --- | --- |
| 3+ learnings in one repo on one theme, with an empty `skill` field and no repo skill in `<repo>/.claude/skills/` covering it | one issue on that repo suggesting a repo skill, an `AGENTS.md` line or an ADR |
| the same lesson in 2+ repos | one issue on the most likely library skill's source repo, with the `from:` line |

Judge "one theme" and "same lesson" by reading `text`, `before` and `after`; when unsure, leave the
learning out. Show the proposed groups to the user and wait for acceptance before filing, as in
step 4. A proposal lists each learning (id, text, repo, `before`/`after`) as evidence and stops
there: it contains no skill draft, no proposed wording and no patch. Use the reflect issue format
with label `reflect`, and the genericising rules above for global and public-repo issues.

After the proposal exists, for every learning in it, keep its status and record the URL:

```bash
reflect learn mark <id> <current-status> --proposal <proposal-url>
```

`--proposal` is separate from `--issue`, so a promoted learning keeps its own issue link.

## Mark

After each issue is filed, for every learning in it:

```bash
reflect learn mark <id> promoted --issue <issue-url> [--scope global|repo] [--skill name]
reflect learn mark <id> deferred
reflect learn mark <id> rejected
```

Mark only after the issue exists, so a failed filing leaves the learning pending for the next run.
Pass `--scope` and `--skill` on any mark when step 2 inferred them. Re-running `ls --status pending` must show
nothing already handled.
