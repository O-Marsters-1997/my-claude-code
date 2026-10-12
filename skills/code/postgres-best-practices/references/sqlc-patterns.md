# sqlc Patterns

Load from [`sqlc.md`](sqlc.md) when writing transactions, joins, bulk writes or lint rules.
Anything not covered here: <https://docs.sqlc.dev/en/latest/> (how-to guides, `reference/config`,
`reference/query-annotations`). Check it only when something below is unclear.

## Contents

- Annotations
- Transactions
- Joins with sqlc.embed
- Bulk writes
- Lint and verify

## Annotations

| Annotation | Generated return |
| --- | --- |
| `:one` | `(T, error)` |
| `:many` | `([]T, error)` |
| `:exec` | `error` |
| `:execrows` | `(int64, error)`, rows affected |
| `:execresult` | `(sql.Result, error)` |
| `:copyfrom` | `(int64, error)`, pgx only |
| `:batchexec` `:batchone` `:batchmany` | `*<Name>BatchResults`, pgx only |

Use `:execrows` when the caller must know whether an `UPDATE` or `DELETE` matched anything;
`:exec` hides a no-op.

## Transactions

`WithTx` returns a `Queries` bound to the transaction. Every query in the unit of work runs on it;
the original `Queries` stays outside.

```go
tx, err := pool.Begin(ctx)
if err != nil {
	return err
}
defer tx.Rollback(ctx)

qtx := q.WithTx(tx)
if _, err := qtx.GetAccountForUpdate(ctx, id); err != nil {
	return err
}
return tx.Commit(ctx)
```

The sqlc docs' example reads then writes with no lock, which loses updates under concurrency.
Lock the row (`SELECT ... FOR UPDATE`, see `SKILL.md`) or use one `UPDATE x = x + 1`.

## Joins with sqlc.embed

`sqlc.embed(table)` keeps each joined table as one struct field instead of flattening columns.

```sql
-- name: ListScoresWithStudent :many
SELECT sqlc.embed(students), sqlc.embed(test_scores)
FROM test_scores
JOIN students ON students.id = test_scores.student_id;
```

Rows come back as `row.Student.Name`, `row.TestScore.Score`. Prefer it over `students.*`, which
generates a wide row type that changes whenever either table gains a column. The docs do not
cover `LEFT JOIN` nullability with `embed`; test that generated shape before relying on it.

## Bulk writes

- `:copyfrom` uses the Postgres COPY protocol and is the fast path for loading many rows. It
  needs `sql_package: pgx/v5` and takes `[]Params`. COPY has no `RETURNING` or `ON CONFLICT`; use
  a multi-row `INSERT` or `:batchexec` when you need either.
- `:batchexec` / `:batchone` / `:batchmany` queue queries on one round trip and return a
  `BatchResults` whose callback receives `(index, error)`. Close it, or the connection stays busy.

## Lint and verify

- `sqlc vet` runs CEL rules from a top-level `rules:` block. A rule does nothing until a `sql`
  package lists it under `rules:`. There is no built-in `SELECT *` rule; write one:

```yaml
rules:
  - name: no-select-star
    message: list columns explicitly
    rule: query.sql.contains("SELECT *")
```

- `sqlc/db-prepare` is a built-in rule that prepares each query against a live database
  (`POSTGRESQL_SERVER_URI`). It catches errors static analysis misses.
- Opt one query out with `/* @sqlc-vet-disable no-select-star */`; no names disables all rules.
- `sqlc diff` fails when generated code is stale and needs no database. Keep the
  `git status --porcelain` check from `sqlc.md` if `diff` misses newly generated files.
- `sqlc verify` (needs sqlc Cloud and `sqlc push`) checks existing queries against a proposed
  migration. Deploy the fixed query before the migration, not after.
