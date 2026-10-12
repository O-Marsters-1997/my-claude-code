---
name: go-idiomatic
description: >
  Go engineering standards for this repo — package layout, error handling, logging, testing,
  concurrency and design patterns. TRIGGER — load BEFORE writing or editing Go, not after: any
  .go, go.mod, go.sum or go.work file is touched, Go or Golang is named, or the working directory
  contains go.mod and the task is implementation, refactor, review or test-writing. Plain feature
  requests in a Go repo ("add an endpoint", "extract a helper", "write a test for this") count;
  the user will not say "idiomatic". Load it even when another skill (tdd, simplify) is active.
  SKIP only when no Go file is involved and Go is not named.
allowed-tools: Read, Edit, Write, Glob, Grep, Bash(go:*), Bash(golangci-lint:*), Bash(git:*), Agent
---

# Go Idiomatic

This repo's rules for Go, where they differ from or sharpen general Go practice. The rules below
apply to every change; the references hold the detail and worked examples.

## Where to Look

Route to the right reference before reading further:

| If you're working on…                                                          | Load this                        |
| ------------------------------------------------------------------------------ | -------------------------------- |
| Goroutines, channels, sync primitives, errgroup, leaks, races                  | `references/concurrency.md`      |
| Adding a sentinel or error type, translating errors between packages, cleanup in `defer`, `panic` | `references/errors.md` |
| Giving `main` or a component a logger, choosing a level, CLI output vs logs    | `references/logging.md`          |
| Interface design, zero values, identifier types, typed value sets, anti-patterns | `references/patterns.md`         |
| Writing or reviewing tests, benchmarks, fuzzing, HTTP tests                    | `references/testing.md`          |
| Optional fields or return values — when to use `Option[T]`                     | `references/optional.md`         |
| Pointer vs value receivers, mutation vs optionality                             | `references/value-vs-pointer.md` |
| Scaffolding a Go HTTP client for a third-party API, writing a `NewXxxClient`, or attaching auth to outbound HTTP calls | `references/http-client.md` |
| Shelling out with `os/exec` to `git`, `gh` or another CLI, parsing its `--json` output, faking it in tests | `references/exec.md` |
| Storing data in SQL: store package boundary, transactions, NULLs, constraint errors, pool setup, testing against Postgres | `references/database.md` |
| Rendering `html/template` for htmx: page vs fragment vs redirect, `HX-*` headers, `.tmpl` files | the `htmx-go-templates` skill |

---

## Errors and Logging

Nearly every change returns an error or reports one, so these rules apply to all Go work, not only
to a task described as error handling:

1. Wrap with `fmt.Errorf("<verb> <noun>: %w", err)`. Error strings are lowercase, with no
   punctuation and no "failed to".
2. Handle an error once: return it with context, or log it at the top of the chain. Never both.
3. Export a sentinel only when a caller in another package branches on it. Define it where the
   condition is detected, and translate a dependency's errors into your own at the package seam.
4. Never discard an error with `_ =`. Join a cleanup error into the returned error, or log it at
   `Warn` when there is nothing to return to.
5. `panic` only for programmer errors: `Must` helpers on constant input, impossible cases.
6. Only components where an error stops travelling (handlers, workers, long-lived services) take
   an injected `*slog.Logger`, as a constructor field. Code below them returns errors and takes no
   logger. `main` builds the one logger; nothing calls `log.*`, `slog.Info` or `slog.Default()`.
7. Log messages are constant; variable data goes in snake_case key-value attributes, with `err`
   for the error. Never log secrets or personal data.
8. Output the user asked for goes to an injected `io.Writer`, not to the logger.

A routine `return fmt.Errorf(...)` or log line needs only these rules. Load `references/errors.md`
when adding a sentinel or error type, or handling a cleanup error, and `references/logging.md` when
wiring a logger, choosing a level or separating CLI output.

---

## Testing

This section is the authority on how a Go test is written. Which tests a change needs, and how
much coverage is enough, is the repo's policy, not this skill's. For the TDD loop, use the `tdd`
skill.

1. Test behaviour through the exported API. Default to the black-box `package foo_test`; use
   `package foo` only when the unit is unexported by design, and `export_test.go` as a last resort.
2. Assert on outputs and state, not on which calls were made. A test that breaks when the code is
   refactored without a behaviour change is a change-detector: delete it.
3. Pick the most faithful double: real, then fake, then stub, then mock. Mock only state-changing
   calls across a system boundary (sending email, charging a card).
4. A fake is stateful and honest. One contract suite runs against both the fake and the real
   implementation. Force error paths with a one-method stub, not failure knobs on the fake.
5. Test SQL against a real database, never a mocked driver. `references/database.md` has the
   mechanics: one shared server, a template database cloned per test.
6. Compare with `cmp.Diff(want, got)` from `github.com/google/go-cmp` and report `(-want +got)`.
   Without go-cmp, compare scalars with `==` and structs field by field. Never `reflect.DeepEqual`.
7. Failure messages read `Func(in) = got, want want`. Match errors with `errors.Is`/`errors.As`,
   never by string.
8. Use a table only when every case is checked the same way: a slice of structs with a `name`
   field, run with `t.Run(tt.name, ...)`. An `if tt.wantErr` branch means split the test.
9. `t.Error` to keep going, `t.Fatal` when the rest of the test can't run. Never call `t.Fatal`
   from a goroutine the test didn't start on.
10. Helpers take `t`, call `t.Helper()`, fail on setup errors and register `t.Cleanup`. They set
    up; they don't assert. `TestMain` only for expensive setup shared by the whole package.
11. No `time.Sleep`. Poll a condition with a deadline or wait on a channel.
12. `t.Parallel()` only for tests that share no state: no globals, env, working dir or shared rows.
13. Compare JSON by decoding it, never as bytes. Normalise golden output before comparing, and a
    human reviews every `-update` diff.
14. Fuzz parsers of untrusted input, and commit the `testdata/fuzz/` corpus.
15. Test HTTP handlers through the real router and middleware.

Load `references/testing.md` for worked examples of each kind of test (unit, fake plus contract
suite, HTTP, golden, fuzz, benchmark).

---

## Code Organisation

1. Applications keep `main` in `cmd/<app>/main.go`, which only wires dependencies, and everything
   else under `internal/`. `pkg/` is for library modules imported from outside the repo; in an
   application it signals a public API that doesn't exist.
2. Extract shared logic on the **third** repetition across meaningfully different packages, not
   the second. A helper coupled to one domain lives in `internal/` next to its primary caller.
3. Name a package by what it does (`stringx`, `idgen`, `timeutil`). `utils` is allowed only when
   it is used across many packages and imports nothing from `internal/`. `helpers` and `common`
   are never allowed: a real name is waiting to be found.
4. No package-level state such as a global `*sql.DB` set in `init()`. Pass dependencies through
   constructors.

---

## Function Design

1. Past 3–4 arguments, group them into a struct, or use functional options (`opts ...Option`)
   for a constructor that will grow; a new option is backward-compatible, a new positional
   argument is not.
2. Declare complex arguments as named locals before the call instead of nesting constructor
   calls inline:

   ```go
   sentAt := time.Now().Add(-24 * time.Hour)
   priority := Priority(3)
   client.Send(NewMessage(sender, recipient, sentAt, priority))
   ```

3. Guard and return early; the happy path stays at the bottom, unindented, with no `else` after a
   `return`. Use `switch` over `if`/`else if` chains.
4. Use pointers for mutation and measurably large structs, values for small immutable types.
   Never use `*T` to mean "might not be set"; that is `Option[T]`.

---

## Naming and Comments

Name functions by what they return or do (`ParseConfig`, `IsExpired`) and booleans as predicates
(`canRetry`). Abbreviate only by Go convention (`ctx`, `err`, `id`, `r`, `w`).

`~/.claude/rules/comments.md` decides what earns a comment. It is the authority and this
skill does not restate or soften it. Read it before editing a file and apply it as written,
including its line cap and its exported-identifiers-only rule for doc comments. `internal/`
is not public API. When reviewing Go code for comment quality, use the `clean-comments` skill.
Do not instruct it to match the file's existing comment density.

---

## Tooling

Run `gofmt`, `goimports`, `go vet ./...` and `golangci-lint run` before finishing. Leave
mechanical rules to the linter rather than restating them here.

---

## See Also

| Skill                           | When to reach for it                                                        |
| ------------------------------- | --------------------------------------------------------------------------- |
| `tdd`                           | TDD cycle, red-green-refactor, writing the test first                       |
| `clean-comments`                | Audit and remove explanatory comments across Go code                        |
| `improve-codebase-architecture` | Identify structural improvements, shallow modules                           |
| `simplify`                      | Review changed code for quality, reuse, and efficiency after implementation |
