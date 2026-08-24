---
name: implement
description: "Implement a piece of work based on a spec or set of tickets."
disable-model-invocation: true
---

Implement the work described by the user in the spec or tickets.

Load the standards skill for each language this change touches before editing a file in
it, and follow it as written. Existing code that breaks the standard is not licence to
match it.

`~/.claude/rules/comments.md` is the authority on comments. It outranks a standards
skill's own comment guidance.

Use /tdd where possible, at pre-agreed seams.

Run typechecking regularly, single test files regularly, and the full test suite once at the end.

Once done, use /code-review to review the work.

Commit your work to the current branch.

If you open a PR, always open it as a draft (`gh pr create --draft`) and write the title and body with /gh-desc.
