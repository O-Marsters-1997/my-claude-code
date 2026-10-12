# 1. A go-testing skill with a shared, self-stopping Postgres helper

**Status:** accepted · 2026-10-12 · [#210](https://github.com/O-Marsters-1997/my-claude-code/issues/210)

## Context

Go test guidance lived in `go-idiomatic` § Testing and told agents to test SQL against a real
database without saying how. A baseline run of the Postgres store eval produced a per-package
`TestMain` container, no CI override, no startup timeout and no template database. job-scraper
(`internal/pgtest`) and command-center (`internal/cctest`) had each solved half of the problem:
job-scraper a reused, self-stopping container with a hashed migrated template, command-center a
fresh database per test.

## Decision

A new `go-testing` skill owns how Go tests are written; `go-idiomatic` keeps a pointer. Which
tests a layer needs stays in each repo's `testing-policy`, which wins on conflict and names the
deviation.

The skill ships `assets/pgtest.go`, copied to `internal/pgtest`:

- **One container per repo**, reused by name (`<repo>-pgtest-v<N>`) across package binaries,
  concurrent runs and worktrees. Pinned to production's Postgres version, tmpfs `PGDATA`,
  `AutoRemove`, `max_connections=500`.
- **Ryuk off, idle watchdog on.** A shell loop inside the container stops Postgres after ~5 min
  with no client connections. Ryuk was rejected: the session that created a reused container
  reaps it while other sessions still use it (testcontainers-go #2445, #2654).
- **A database per test**, `CREATE DATABASE … TEMPLATE` from a template migrated once per
  migrations hash under an advisory lock, dropped `WITH (FORCE)` in `t.Cleanup`. Safe under
  `t.Parallel()`. Truncation, snapshot/restore and a container per test were rejected.
- **Startup sweep** in each binary drops `pgtest_*` databases with no connections older than a few
  minutes, covering `kill -9`.
- **`TEST_DATABASE_URL`**, the same name in every repo, skips the container for CI service
  containers. Without Docker or the variable, tests fail naming both.
- Plain `go test`, no build tag for DB tests. `cmp.Diff` is mandatory; go-cmp is added on first
  test touch. A `depguard` rule keeps testcontainers imports inside `internal/pgtest`.

## Consequences

Each adopting repo carries a copy of `pgtest.go` until a third repo adopts it, when it becomes a
shared module. command-center migrates `cctest` first; job-scraper decides in job-scraper#859.
A stopped container leaves nothing behind, but a killed run's databases linger until the next
binary's sweep or the watchdog stops the container.
