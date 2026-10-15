---
name: triage-reflections
description: >
  Routes pending REFLECT-marker learnings from the reflect ledger to GitHub issues. Use when the user
  says "triage reflections", "/triage-reflections", "process my REFLECT markers", "what have I
  learned", or when the statusline shows reflect:N. Infers missing skill and scope, splits notes
  into explicit and unclear, plays unclear ones back for correction, merges duplicates into one
  item with every example, drops lessons already covered by a rule or lint config, routes by a
  three-way test the user can correct, files one issue per skill, and writes every decision back to the ledger
  with `reflect marker mark`.
---

# Triage reflections

Turn pending learnings into filed issues, and leave none pending that were handled. Triage files
issues only; it never edits a skill (`skill-updater`, `/implement` and `/fleet` do that).

`reflect` is `~/.claude/bin/reflect`. Routing: `skills/meta/reflect/references/routing.md`. Issue
body, labels, dedupe and the `-R` repo flags: `skills/meta/reflect/references/issue-format.md`.

## Steps

1. **Load.** `reflect marker ls --json --status pending`. Empty means say so and stop.
2. **Infer.** For a learning with no `skill` or no `scope`, read its `text`, `before`, `after`
   and file, then match it to a repo skill (`<repo>/.claude/skills/`) or a library skill
   (`reflect status` gives the `library:` path). Resolve an installed copy to its library source
   as `routing.md` describes.
3. **Classify.** Split every note by one test:
   - **Explicit**: it has an `after`, or its text states a checkable rule ("inject the clock,
     don't sleep"). For a PR note (`source: pr`) the `after` is fetched now: diff the commented
     lines between the record's `commit` and the merged head of PR `pr` with `gh`. Write nothing
     back to the ledger for it.
   - **Unclear**: any other note ("not sure this is best", "could be simpler"). Read the code and
     reach a reading and a proposed rule.
4. **Merge and drop.** Merge pending learnings with the same skill and scope that say the same
   thing (judged by reading `text`, `before` and `after`) into one item, keeping every example
   (file, line, `before`/`after`) and a count such as "seen in 4 places": the repetition is the
   evidence, and each merged learning still counts toward a proposal's 3+.
   Mark `rejected` only a learning that an existing skill, `AGENTS.md`, rule or lint config
   already states. Check by reading the target; do not assume.
5. **Play back.** Send one message holding every unclear note (code, your reading, the proposed
   rule) and the route table for the explicit ones (ids, text, count, route, reason). Close with:
   "Correct anything I've got wrong. Anything you don't mention, I'll take as agreed." Silence on
   a note is agreement; never ask note by note. Re-verify all corrections together, then play back
   only the changed notes plus a one-line list of what is settled, until a reply carries none.
   "Go" settles the rest. Never file before this.
6. **File** the settled rows (below), then **mark** each.
7. **Propose upward** (below) over the deferred and promoted learnings.
8. **Report** the issue URLs, the proposals, and counts of promoted, deferred and rejected.

## Routes

Apply the three-way test; its result is the reason column. A `(skill)` hint that disagrees with it
appears in the play-back next to your choice.

| Learning | Action | Mark |
| --- | --- | --- |
| only about this code | `inbox` issue on the learning's `repo` via `file-issue` inbox mode | `promoted` |
| depends on this repo's terms, ADRs or layout | one issue on the learning's `repo` field (pass `-R <owner>/<repo>` from its `origin`) to amend the repo skill or `AGENTS.md` | `promoted` |
| holds in any repo using the same language or tool | one issue on the library skill's source repo | `promoted` |
| repo lesson with no skill (`REFLECT(repo)`), or no owner and no other learning shares its theme | file nothing | `deferred` |
| user rejects, or a drop from step 4 | file nothing | `rejected` |

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

After marking, run `reflect marker ls --json` and consider every learning with status `deferred` or
`promoted` and an empty `proposal` field. A learning already carrying a `proposal` URL belongs to a
filed group: skip it, so a group is never proposed twice.

| Group | Proposal |
| --- | --- |
| 3+ learnings in one repo on one theme, with an empty `skill` field and no repo skill in `<repo>/.claude/skills/` covering it | one issue on that repo suggesting a repo skill, an `AGENTS.md` line or an ADR |
| the same lesson in 2+ repos | one issue on the most likely library skill's source repo, with the `from:` line |

Judge "one theme" and "same lesson" by reading `text`, `before` and `after`; when unsure, leave the
learning out. Show the proposed groups to the user and wait for acceptance before filing, as in
step 5. A proposal lists each learning (id, text, repo, `before`/`after`) as evidence and stops
there: it contains no skill draft, no proposed wording and no patch. Use the reflect issue format
with label `reflect`, and the genericising rules above for global and public-repo issues.

After the proposal exists, for every learning in it, keep its status and record the URL:

```bash
reflect marker mark <id> <current-status> --proposal <proposal-url>
```

`--proposal` is separate from `--issue`, so a promoted learning keeps its own issue link.

## Mark

After each issue is filed, for every learning in it:

```bash
reflect marker mark <id> promoted --issue <issue-url> [--scope global|repo] [--skill name]
reflect marker mark <id> deferred
reflect marker mark <id> rejected
```

Mark only after the issue exists, so a failed filing leaves the learning pending for the next run.
Pass `--scope` and `--skill` on any mark when step 2 inferred them. Re-running `ls --status pending` must show
nothing already handled.
