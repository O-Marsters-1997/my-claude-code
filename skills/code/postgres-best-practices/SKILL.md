---
name: postgres-best-practices
description: >
  Postgres schema and query rules, independent of the calling language. TRIGGER — load BEFORE
  writing or editing a .sql file, a migration, or any schema, index, constraint or query change
  against Postgres: new tables, foreign keys, status columns, claim or lease queries, backfills,
  ON CONFLICT upserts. SKIP for MySQL, SQLite and other engines, and for code that never touches
  SQL.
---

# Postgres

Rules for schema and queries on Postgres. They are language-agnostic; the Go half lives in the
`go-idiomatic` skill (`references/database.md`, `references/patterns.md`).

## Where to Look

| If you're working on…                                   | Load this                                        |
| ------------------------------------------------------- | ------------------------------------------------ |
| Go code that runs these queries (`database/sql`, pgx)   | `go-idiomatic` `references/database.md`          |
| A closed value set mirrored as a Go type                | `go-idiomatic` `references/patterns.md` § Typed Value Sets |
| Writing or changing a migration, any tool               | `references/migrations.md`                       |
| Converting an existing Postgres `ENUM` to `text` + `CHECK` | `references/migrations.md` § ENUM Conversion  |
| Migrations in a repo that uses goose                    | `references/goose.md` (only when the repo uses goose) |
| Only if the repo has `sqlc.yaml`: queries, config, CI   | `references/sqlc.md`                             |

## Schema

- **Closed value sets are `text` + a named `CHECK`, never a Postgres `ENUM`.** Adding a value to
  an `ENUM` is a type migration and removing one is not supported; a `CHECK` is one `ALTER`.
  Never mix the two styles in one schema, and never leave a status column unchecked. An existing
  `ENUM` is a conversion to schedule, not a style to extend.
- **Every foreign key column has an index.** Postgres does not create one. Without it, deleting
  a parent scans the child table, and `ON DELETE CASCADE` turns that into one scan per row.
- **`timestamptz` for instants, `date` only for calendar dates.** Never store times as `text`.
- **`jsonb` for JSON, never `text`.**
- **Name every constraint and index, with one scheme** (`<table>_<cols>_key`, `_fkey`, `_check`,
  `<table>_<cols>_idx`). Names are what errors and migrations refer to.
- **One meaning for absence per column.** NULL or `''`, not both. Pick NULL unless the column is
  `NOT NULL DEFAULT ''` by design.
- **An append-only table gets a retention policy in the migration that creates it.** Outboxes,
  run logs and audit trails only grow; decide how long rows live, index the column the cleanup
  filters on, and delete expired rows in batches (`DELETE ... WHERE id IN (SELECT id ... LIMIT n)`)
  from a scheduled job. Added later, the first cleanup is one huge delete on a huge table.

## Queries

- **One clock per comparison.** `DEFAULT now()` is fine for audit columns. Any time the
  application compares against, or a test asserts on, is passed in as a parameter. Mixing SQL
  `NOW()` with the app's `time.Now()` makes leases and expiries untestable and skewed.
- **Explicit column lists.** No `SELECT *` and no `RETURNING *`; adding a column must not change
  what callers receive.
- **Every list query has a hard `LIMIT`.** A list without one returns the whole table once the
  table grows. On growing tables page by keyset (`WHERE (created_at, id) < ($1, $2) ORDER BY
  created_at DESC, id DESC LIMIT $3`), not `OFFSET`, which reads and discards every skipped row.
- **No N+1 loops.** Fetch with `= ANY($1::uuid[])`, insert with multi-row `VALUES` or `COPY`.
- **Claiming work uses `FOR UPDATE SKIP LOCKED`** so concurrent workers take different rows
  instead of queueing on one. Update the claimed row in the same statement or transaction.
- **Read-then-write uses `SELECT ... FOR UPDATE`** in a transaction. A plain read races.
- **Match unique violations (SQLSTATE `23505`) by constraint name**, not by message text or by
  assuming which constraint fired. This is why constraints are named.
- **`ON CONFLICT` names its target** (`ON CONFLICT (col)` or `ON CONSTRAINT name`). A bare
  `ON CONFLICT DO NOTHING` silently swallows violations of constraints you did not mean.

## Building SQL

- **Never build SQL, or shell that carries SQL, with `fmt.Sprintf` or string concatenation.**
  Pass values as bind parameters (`$1`) or, for a script, as environment variables the script
  reads. Pass identifiers through `pgx.Identifier{"schema", "table"}.Sanitize()`. Interpolating
  even an int starts the habit that becomes injection. Enable gosec G201/G202
  (`lint/golangci.yml`); they catch Sprintf and concatenation that reach a query call, not SQL
  inside a shell string, so review has to.

Before, SQL baked into a Sprintf'd entrypoint:

```go
var entrypoint = fmt.Sprintf(`psql -c "ALTER SYSTEM SET idle_in_transaction_session_timeout = %d"`, idleTimeout)
```

After, a static script that reads its value from the environment (`IDLE_TIMEOUT` set via the
container's `Env`):

```go
//go:embed entrypoint.sh
var entrypoint string
```

```sh
psql -v t="$IDLE_TIMEOUT" -c "ALTER SYSTEM SET idle_in_transaction_session_timeout = :'t'"
```

## Altering tables

- **Add columns nullable or with a constant default**, then backfill, then add `NOT NULL`.
  Never rewrite a large table inside the migration that adds the column.
- **Backfill in batches**, keyed on the primary key, committing between batches. One giant
  `UPDATE` holds locks and bloats the table.
- **Create indexes on live tables with `CREATE INDEX CONCURRENTLY`**, outside a transaction.
- **Add constraints in two steps on big tables:** `ADD CONSTRAINT ... NOT VALID`, then
  `VALIDATE CONSTRAINT`, which does not block writes.
- **Migrations are forward-only once merged.** Fix mistakes with a new migration.
