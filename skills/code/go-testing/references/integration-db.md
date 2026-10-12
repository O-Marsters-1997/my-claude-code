# Integration tests against Postgres

Every database test gets its own database from `internal/pgtest`. This file covers installing the
helper in a repo that lacks it, using it, CI, and what to do when it misbehaves. The design record
is ADR 0001 in the skills repo (`docs/adr/0001-go-testing-skill-and-pgtest.md`).

## Contents

- [What pgtest does](#what-pgtest-does)
- [Installing it](#installing-it)
- [Using it](#using-it)
- [CI](#ci)
- [Rejected approaches](#rejected-approaches)
- [Troubleshooting](#troubleshooting)

---

## What pgtest does

- **One container per repo.** The first test binary to need Postgres starts a container named
  `<repo>-pgtest-v<N>`; every other package binary, every concurrent `go test` and every worktree
  reuses it by name. No `TestMain`, no per-package wiring.
- **A database per test.** `pgtest.New(t)` runs `CREATE DATABASE … TEMPLATE` from a template
  migrated once per migrations hash, and drops it `WITH (FORCE)` in `t.Cleanup`. Cloning takes
  tens of milliseconds, so tests stay parallel and never see each other's rows.
- **It cleans up after itself.** Ryuk is off for this container, because Ryuk reaps a reused
  container while another session still uses it (testcontainers-go #2445, #2654). Instead a
  watchdog inside the container stops Postgres after five minutes with no client connections,
  and `AutoRemove` deletes it. Each binary's startup drops `pgtest_*` databases with no
  connections older than five minutes, which covers a `kill -9`.
- **`TEST_DATABASE_URL`** (the same name in every repo) points it at an existing server instead
  of Docker. It must be a `postgres://` URL for a role allowed to `CREATE DATABASE`.

API:

```go
func New(tb testing.TB) string                                          // DSN; dropped in tb.Cleanup
func Create(ctx context.Context) (dsn string, drop func() error, err error) // for callers without a *testing.T
```

`Create` is for setup that has no `testing.TB`, such as a testscript `Setup` hook:
`env.Defer(func() { if err := drop(); err != nil { env.T().Log(err) } })`.

---

## Installing it

1. **Migrations in a leaf package.** pgtest imports the repo's migrations, so they must live in a
   package that imports nothing under test, typically `internal/migrations` (or wherever the repo
   keeps them), exposing an `fs.FS` of the migration files and a func
   `func(ctx context.Context, dsn string) error` that applies them to the database at `dsn`.
   Wrap the repo's existing runner (goose, migrate, a hand-rolled loop); don't write a second one.
   A store package that imports `pgtest` from an in-package (`package store`) test would cycle if
   pgtest imported the store; keep migrations out of the store package.
2. **Copy the helper.** Copy `assets/pgtest/pgtest.go`, `settings.go`, `pgtest_test.go` and
   `sweep_test.go` from this skill into `internal/pgtest/`. Change only `settings.go`, and the
   module path in `pgtest_test.go`:

   ```go
   const (
       containerName = "notes-pgtest-v1"       // <repo>-pgtest-v<N>; bump N when you change container options
       image         = "postgres:17.6-alpine"  // the exact major.minor production runs, never latest
   )

   var (
       migrationFiles = migrations.FS             // hashed to name the template
       migrate        = migrations.Up             // func(ctx, dsn string) error
       dsnParams      = url.Values{"pool_max_conns": {"4"}}
   )
   ```

   `dsnParams` goes only on the DSNs tests receive; `migrate` gets a plain DSN, so it can use
   any driver. `pool_max_conns` is a pgxpool parameter; the server rejects it from `pgx.Connect`
   or `database/sql`. If the repo opens databases through `database/sql` (or a bare `pgx.Conn`),
   set `dsnParams = nil` and call `db.SetMaxOpenConns(4)` where the test opens it. The cap
   matters: the container allows 500 connections, and fifty parallel tests with default pool
   sizes exhaust that.
3. **Dependencies.** `go get github.com/testcontainers/testcontainers-go/modules/postgres
   github.com/jackc/pgx/v5`, then `go mod tidy`. The `container.HostConfig` import comes from
   `github.com/moby/moby/api` on testcontainers-go v0.40+; older versions use
   `github.com/docker/docker/api/types/container`. Match what `go.mod` resolves.
4. **Lint.** Merge `assets/golangci-depguard.yml` into `.golangci.yml`. It keeps testcontainers
   imports inside `internal/pgtest`, so nobody adds a second container path later.
5. **Check.** `go test -race ./internal/pgtest/` must pass, then the repo's own tests.

If the repo already has a helper (`cctest`, `dbtest`, a `pgtest` with truncation), its
`testing-policy` says whether to migrate it. Don't build a second helper beside it.

---

## Using it

```go
package notestore_test

func newStore(t *testing.T) *notestore.Store {
    t.Helper()
    pool, err := pgxpool.New(t.Context(), pgtest.New(t))
    if err != nil {
        t.Fatalf("connect: %v", err)
    }
    t.Cleanup(pool.Close)
    return notestore.New(pool)
}

func TestGetMissingNote(t *testing.T) {
    t.Parallel()
    s := newStore(t)

    _, err := s.Get(t.Context(), 42)
    if !errors.Is(err, notestore.ErrNotFound) {
        t.Errorf("Get(42) error = %v, want %v", err, notestore.ErrNotFound)
    }
}
```

- Register the pool's `Close` after `pgtest.New`, so it runs first and the drop finds no open
  connections. `WITH (FORCE)` covers it if not.
- The database starts migrated and empty. Seed through the package's own API, not raw SQL,
  unless the test is about the SQL.
- Don't `t.Skip` when Docker is missing. A skipped database test is a test that silently
  stopped running.
- The store's tests are the contract suite's real half; the fake runs the same suite
  (see `examples.md` § Test Doubles).

---

## CI

Point `TEST_DATABASE_URL` at a service container and leave Docker out of the test job:

```yaml
services:
  postgres:
    image: postgres:17.6-alpine
    env:
      POSTGRES_PASSWORD: postgres
    ports: ["5432:5432"]
    options: >-
      --health-cmd "pg_isready -U postgres" --health-interval 2s --health-retries 30
env:
  TEST_DATABASE_URL: postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
```

Same image tag as `settings.go`. pgtest builds the template and the per-test databases on it
exactly as it does locally.

---

## Rejected approaches

| Approach                              | Why not                                                                                     |
| ------------------------------------- | ------------------------------------------------------------------------------------------- |
| `TestMain` per package                | Starts a container even for tests that never touch the DB; every package copies the wiring |
| `TRUNCATE` between tests              | Serialises tests that share the database; misses sequences and schema changes              |
| Snapshot / restore                    | Restores the whole database between tests; can't run in parallel                           |
| Container per test                    | Seconds per test and hundreds of containers under `t.Parallel()`                            |
| Ryuk for the shared container         | Ryuk reaps a reused container with the session that created it, mid-run for the others     |
| `sqlmock` / mocked driver             | Proves the query string was sent, not that it works                                         |
| A build tag or `-short` for DB tests  | DB tests stop running on the commands people actually type                                  |

---

## Troubleshooting

| Symptom                                                   | Cause and fix                                                                                    |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `start postgres (needs Docker running, or TEST_DATABASE_URL …)` | Start Docker, or export `TEST_DATABASE_URL`                                               |
| `unrecognized configuration parameter "pool_max_conns"`   | The test opened the DSN with `database/sql` or `pgx.Connect`; set `dsnParams = nil`             |
| `too many clients already`                                | A test opens pools without a cap; see `dsnParams` above                                          |
| Tests see an old schema                                   | A migration changed outside `migrationFiles`; include every file the migrator reads             |
| Changed `entrypoint`, `image` or Postgres flags have no effect | The old container is reused by name; bump `-v<N>` in `containerName`                       |
| `pgtest_*` databases pile up                              | Binaries killed mid-run; the next binary's startup sweep removes them after five minutes        |
| `TESTCONTAINERS_RYUK_DISABLED` seems ignored              | Something else in the binary called testcontainers first; keep all container code in pgtest   |
