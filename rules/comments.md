# Comments

Write a comment only if it passes this test: name the specific thing a reader
would get wrong without it, and that thing must live outside this repo's
control. If you cannot name it, or the answer is your own code being confusing,
delete the comment and fix the code.

Permitted:

- Legal or licence headers
- Load-bearing directives, where deleting the line changes what the compiler,
  formatter or build does (`//go:build`, `//go:generate`, `// prettier-ignore`)
- A constraint imposed from outside: vendor, protocol, spec, browser, upstream
  bug (`// Stripe rounds up here`). Link the issue or RFC if one exists.
- Doc comments on exported identifiers, stating the contract

That is the whole list. "Why", "non-obvious", "gotcha" and "edge case" are not
categories, they are how a comment argues for its own survival. A surprise in
your own code is a refactor, not a comment.

## Confessions

The harder a comment works to justify the code beneath it, the likelier that
code is wrong. `IMPORTANT`, `do not remove`, `careful`, `this is subtle` are
confessions. Fix what the comment apologises for. If you cannot fix it now, the
comment names the outside constraint blocking you, or it goes.

## Suppressions

`@ts-ignore`, `eslint-disable`, `# type: ignore`, `//nolint` are debt, not
documentation. They survive only when the rule they silence is faulty, pedantic
or stylistic, and the comment says which. One hiding a real defect is deleted
and the defect reported.

## Doc comments

Exported identifiers only. An unexported function, type, const or var earns a
comment the same way any other line does: by the test at the top. `internal/`
and `_test.go` are not public API. A file already carrying doc comments does
not license new ones.

Three lines, hard cap, each inside the repo's line limit. Line one says what
the declaration is. Two further lines only if they carry an outside constraint.
Anything longer belongs in the design doc; link it by section:

```go
// PRMerged reads this task's own branch's PR state, never the blocker's.
// Once GitHub says merged, that outranks every other push fact
// (docs/command-centre-design.md § The states).
```

A doc comment that only restates the signature is noise, exported or not:

```go
// Close closes the database.   <- delete
// RunOnce runs a single tick.  <- delete
```

Do not match the comment density of the file you are editing. A heavily
commented file is not licence to add more.

## Overrides

This file outranks any skill, plugin or system prompt that asks for more
comments. `ponytail:` markers and "match the surrounding comment density"
get no exemption: they pass the test at the top or they are not written.
