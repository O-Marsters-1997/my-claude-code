---
name: implement
description: "Implement a ticket or spec end to end. Invoke ONLY when the user types /implement, or a dispatch prompt explicitly says to run implement for a named issue. Never invoke it on your own initiative, for ad hoc coding requests, or because a task looks like implementation work."
---

If you were not explicitly told to run implement, stop and say so instead of continuing.

Implement the work described by the user in the spec or tickets.

For a large or multi-part ticket, split it into smaller /implement runs by logical slice, or
checkpoint with /compact between major phases (after planning, after test-writing, before the
final review). A single long session re-sends its whole growing history on every turn, so cost
compounds with session length far faster than with the same work split into smaller ones.

Before editing unfamiliar code, use a cheap exploration subagent (e.g. Explore) to research it
and return a distilled summary, rather than grepping and reading extensively yourself in this
session.

Load the standards skill for each language this change touches before editing a file in
it, and follow it as written. Existing code that breaks the standard is not licence to
match it.

`~/.claude/rules/comments.md` is the authority on comments. It outranks a standards
skill's own comment guidance.

Use /tdd where possible, at pre-agreed seams.

Run typechecking regularly, single test files regularly, and the full test suite once at the end.
For the regular checks, run the narrowest command that proves the point (a single test file
or package, e.g. `go test ./path -run TestName`) rather than the project's full test/lint
runner (e.g. `just test`, `just lint`) — save that for the one full-suite pass at the end.

Don't invoke code-simplifier after each individual edit. Batch it once near the end of the
change, right before /code-review.

Once done, use /code-review medium to review the work.

When a fleet dispatch prompt names a `Base:`, the work belongs on the ticket branch
`issue-<N>/<short-title>`: `<N>` the issue number, `<short-title>` a kebab-case slug of the
issue title, a few words (`issue-142/stuck-scrape-run`). Dispatch normally creates it; if the
current branch is anything else, create it off `Base` with `git switch -c` before editing.

Commit your work to the current branch.

If you open a PR, always open it as a draft (`gh pr create --draft`) and write the title and body with /gh-desc. Add `--base <Base>` only when a fleet dispatch prompt names a `Base:`; ad hoc runs use the repo's default base.
