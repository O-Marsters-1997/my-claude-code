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
`go-idiomatic` skill (`references/database.md`, `references/enums.md`).
## Where to Look

| If you're working on…                                   | Load this                                        |
| ------------------------------------------------------- | ------------------------------------------------ |
| Go code that runs these queries (`database/sql`, pgx)   | `go-idiomatic` `references/database.md`          |
| A closed value set mirrored as a Go type                | `go-idiomatic` `references/enums.md`             |
| Writing or changing a migration, any tool               | `references/migrations.md`                       |
| Migrations in a repo that uses goose                    | `references/goose.md` (only when the repo uses goose) |
| Only if the repo has `sqlc.yaml`: queries, config, CI   | `references/sqlc.md`                             |

## Schema

- **Closed value sets are `text` + a named `CHECK`, never a Postgres `ENUM`.** Adding a value to
  an `ENUM` is a type migration and removing one is not supported; a `CHECK` is one `ALTER`.
  Never mix the two styles in one schema, and never leave a status column unchecked.
- **Every foreign key column has an index.** Postgres does not create one. Without it, deleting
  a parent scans the child table, and `ON DELETE CASCADE` turns that into one scan per row.
- **`timestamptz` for instants, `date` only for calendar dates.** Never store times as `text`.
- **`jsonb` for JSON, never `text`.**
- **Name every constraint and index, with one scheme** (`<table>_<cols>_key`, `_fkey`, `_check`,
  `<table>_<cols>_idx`). Names are what errors and migrations refer to.
- **One meaning for absence per column.** NULL or `''`, not both. Pick NULL unless the column is
  `NOT NULL DEFAULT ''` by design.

## Queries

- **One clock per comparison.** `DEFAULT now()` is fine for audit columns. Any time the
  application compares against, or a test asserts on, is passed in as a parameter. Mixing SQL
  `NOW()` with the app's `time.Now()` makes leases and expiries untestable and skewed.
- **Explicit column lists.** No `SELECT *` and no `RETURNING *`; adding a column must not change
  what callers receive.
- **No N+1 loops.** Fetch with `= ANY($1::uuid[])`, insert with multi-row `VALUES` or `COPY`.
- **Claiming work uses `FOR UPDATE SKIP LOCKED`** so concurrent workers take different rows
  instead of queueing on one. Update the claimed row in the same statement or transaction.
- **Read-then-write uses `SELECT ... FOR UPDATE`** in a transaction. A plain read races.
- **Match unique violations (SQLSTATE `23505`) by constraint name**, not by message text or by
  assuming which constraint fired. This is why constraints are named.
- **`ON CONFLICT` names its target** (`ON CONFLICT (col)` or `ON CONSTRAINT name`). A bare
  `ON CONFLICT DO NOTHING` silently swallows violations of constraints you did not mean.

## Altering tables

- **Add columns nullable or with a constant default**, then backfill, then add `NOT NULL`.
  Never rewrite a large table inside the migration that adds the column.
- **Backfill in batches**, keyed on the primary key, committing between batches. One giant
  `UPDATE` holds locks and bloats the table.
- **Create indexes on live tables with `CREATE INDEX CONCURRENTLY`**, outside a transaction.
- **Add constraints in two steps on big tables:** `ADD CONSTRAINT ... NOT VALID`, then
  `VALIDATE CONSTRAINT`, which does not block writes.
- **Migrations are forward-only once merged.** Fix mistakes with a new migration.
