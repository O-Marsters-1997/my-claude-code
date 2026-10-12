---
name: go-testing
description: >
  How Go tests are written in these repos: assertions, tables, test doubles, helpers, async
  waits, golden files, fuzzing, benchmarks, and integration tests against a real Postgres via
  the bundled pgtest helper (one shared testcontainer, a fresh database per test). TRIGGER —
  load BEFORE writing, editing or reviewing any *_test.go file, testdata/ golden file, fuzz
  target or benchmark, or when the task says go test, test helper, fake, mock, stub, table test,
  testcontainers, TestMain, test database, flaky test or coverage in a Go repo. Plain requests
  count: "write the tests", "add a test for this", "fix the failing test", or a feature task
  whose last step is testing it. Load alongside go-idiomatic and tdd, not instead of them.
  SKIP when no Go test code is written or judged.
allowed-tools: Read, Edit, Write, Glob, Grep, Bash(go:*), Bash(golangci-lint:*), Bash(docker:*), Bash(git:*)
---

# Go Testing

How a Go test is written, in every repo. What to test is the repo's `testing-policy` when it
ships one: the kinds of test it uses, what each is worth, which a feature needs, and the reusable
patterns to copy as the codebase grows. It never restates how a test is written, so the two do not
overlap. For the red-green-refactor loop, use `tdd`.

## Your defaults that are wrong here

These are the habits that show up unprompted. Check your test against this table before
writing it, and again before finishing, because they creep back in late in a long task.

| You reach for                                         | Do this instead                                                                 |
| ----------------------------------------------------- | ------------------------------------------------------------------------------- |
| testify `assert`/`require`, any assertion library     | Plain `if` plus `t.Errorf`; `cmp.Diff` for anything not a scalar                |
| `reflect.DeepEqual`                                   | `cmp.Diff(want, got)`, reported as `(-want +got)`                                |
| A table with `wantErr bool` and an `if tt.wantErr`    | Two tests: a success table and an error table checked with `errors.Is`          |
| gomock, mockery, `EXPECT().Times(1)`                  | A stateful in-memory fake; assert on the state it ends in                       |
| `sqlmock` or a mocked `database/sql` driver           | A real Postgres from `internal/pgtest`                                          |
| `TestMain` starting a container per package           | `pgtest.New(t)`: one container per repo, a database per test                    |
| `TRUNCATE` between tests, shared rows                 | A fresh database per test, so `t.Parallel()` is safe                            |
| `time.Sleep` to wait for a goroutine                  | A channel the code signals, a deadline poll, or `testing/synctest`              |
| `strings.Contains(err.Error(), "...")`                | `errors.Is` / `errors.As` against an exported sentinel or type                  |
| A new test dependency                                 | Only `github.com/google/go-cmp`; nothing else without asking                    |

## Rules

1. Test behaviour through the exported API, in the black-box `package foo_test`. Use
   `package foo` only when the unit is unexported by design, and `export_test.go` as a last resort.
2. Assert on outputs and state, not on which calls were made. A test that breaks on a refactor
   that changes no behaviour is a change-detector: delete it.
3. Pick the most faithful double: real, then fake, then stub, then mock. Mock only a
   state-changing call across a system boundary (an email sent, a card charged), and then a
   hand-written recorder, not a mocking library.
4. A fake is stateful and honest: the same sentinels, the same not-found behaviour. One contract
   suite runs against the fake and the real implementation. Force an error path with a
   one-method stub that embeds the interface, never with failure knobs on the fake.
5. Test SQL against a real Postgres through `internal/pgtest`. If the repo has no
   `internal/pgtest`, install it from this skill's assets: load `references/integration-db.md`.
6. Compare with `cmp.Diff(want, got)` and report `(-want +got)`. If `go.mod` lacks
   `github.com/google/go-cmp`, add it with `go get` on the first test you touch, and fix any
   lint rule or doc in the repo that contradicts this in the same change. When the user or the
   repo has explicitly ruled out new dependencies ("stdlib only"), compare scalars with `==`,
   slices and maps with `slices.Equal`/`maps.Equal`, structs field by field. Never
   `reflect.DeepEqual` either way.
7. Failure messages read `Func(in) = got, want want`. Match errors with `errors.Is`/`errors.As`;
   a caller that needs to tell errors apart gets an exported sentinel or type.
8. Use a table only when every case is checked the same way: a slice of structs with a `name`
   field run with `t.Run(tt.name, ...)`. A branch on a case field means split the test.
9. `t.Error` to keep going, `t.Fatal` when the rest can't run. Never `t.Fatal` from a goroutine
   the test didn't start on; send the error back on a channel.
10. Helpers take `t`, call `t.Helper()`, fail on setup errors and register `t.Cleanup`. They set
    up; they don't assert. No generic `assertEqual`.
11. No `time.Sleep` to wait. Wait on a channel, poll a condition with a deadline, or run the test
    in `synctest.Test` (Go 1.25+) when it involves timers.
12. `t.Parallel()` only for tests that share no state: no globals, `t.Setenv`, `t.Chdir` or
    shared rows. A `pgtest` database is never shared, so DB tests can be parallel.
13. Compare JSON by decoding it, never as bytes. Normalise golden output before comparing; a
    human reviews every `-update` diff.
14. Fuzz parsers of untrusted input with a property (no panic, round trip); commit
    `testdata/fuzz/`.
15. Test HTTP handlers through the real router and middleware, built the way `main` builds them.

## Running

Plain `go test ./...` runs everything, database tests included: no build tag, no `-short` skip,
no make target. Build tags are only for end-to-end suites. Without Docker and without
`TEST_DATABASE_URL`, `pgtest` fails with a message naming both; don't paper over it with
`t.Skip`. Run `go test -race ./...` and `golangci-lint run` before finishing.

## Where to look

| If you're writing…                                                              | Load                           |
| ------------------------------------------------------------------------------- | ------------------------------ |
| A test that touches Postgres, or a repo with no `internal/pgtest` yet           | `references/integration-db.md` |
| CI for database tests (`TEST_DATABASE_URL`, service containers)                 | `references/integration-db.md` |
| A unit, table, fake plus contract suite, HTTP, async, golden, fuzz or benchmark test | `references/examples.md`  |

`assets/pgtest/` holds the helper and its own tests; `assets/golangci-depguard.yml` holds the
lint rule that keeps testcontainers imports inside `internal/pgtest`.

## See also

| Skill            | When                                                                |
| ---------------- | ------------------------------------------------------------------- |
| `go-idiomatic`   | Everything about the code under test                                |
| `tdd`            | Writing the test first, one behaviour at a time                     |
| `testing-policy` | What to test: test kinds, their value, reusable patterns (per repo) |
