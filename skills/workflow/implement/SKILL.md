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

Before editing unfamiliar code, explore through `codegraph_explore` (load it via ToolSearch if
deferred) when the repo has a `.codegraph/` index. Pass the symbols or files you expect to touch;
the result is verbatim source, so don't `cat`, `sed -n` or `Read` what it already returned. `Read`
only files you are about to edit or that it didn't surface. If a fleet brief is named, start from
it. For research broader than one query, use a cheap exploration subagent (e.g. Explore) and take
its distilled summary rather than grepping and reading extensively yourself in this session.

Load the standards skill for each language this change touches before editing a file in
it, and follow it as written. Existing code that breaks the standard is not licence to
match it.

`~/.claude/rules/comments.md` is the authority on comments. It outranks a standards
skill's own comment guidance.

Use /tdd where possible, at pre-agreed seams.

A ticket carrying `size:xs`, or a dispatch prompt saying `Size: xs`, is a small fix: keep the
diff to what the ticket asks, with no adjacent refactors. Skip /tdd unless the fix changes
observable behaviour, and then add one regression test. Skip code-simplifier. Review with
/code-review low instead of medium.

Given several issues at once (a fleet bundle), work them in the order given, one commit per
issue whose message ends `Closes #<N>`, and repeat each `Closes #<N>` line in the PR body. Run
the review passes once, over the whole bundle.

Run typechecking regularly, single test files regularly, and the full test suite once at the end.
For the regular checks, run the narrowest command that proves the point (a single test file
or package, e.g. `go test ./path -run TestName`) rather than the project's full test/lint
runner (e.g. `just test`, `just lint`) — save that for the one full-suite pass at the end.

Don't invoke code-simplifier after each individual edit. Batch it once near the end of the
change, right before /code-review.

Once done, use /code-review medium to review the work, or low for a small fix. Pass it the
worktree path and the diff against `Base:` (`git -C <worktree> diff origin/<base>...HEAD`). If
that diff is empty, stop and report; never fall back to `HEAD~1`.

When a fleet dispatch prompt names a `Base:`, the work belongs on the ticket branch
`issue-<N>/<short-title>`: `<N>` the issue number, `<short-title>` a kebab-case slug of the
issue title, a few words (`issue-142/stuck-scrape-run`). Dispatch normally creates it; if the
current branch is anything else, create it off `Base` with `git switch -c` before editing.

Commit your work to the current branch.

If you open a PR, always open it as a draft (`gh pr create --draft`) and write the title and body with /gh-desc. Add `--base <Base>` only when a fleet dispatch prompt names a `Base:`; ad hoc runs use the repo's default base.
