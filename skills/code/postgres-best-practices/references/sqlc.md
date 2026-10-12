# sqlc Reference

Load only when the repo has a `sqlc.yaml` (or `sqlc.json`). Rules for writing the queries sqlc
compiles and for configuring and verifying generation. Keeping generated types inside the store
is covered in [`go-idiomatic` database.md](../../go-idiomatic/references/database.md); this file
does not restate it.

## Contents

- Query naming
- Parameters
- Config
- Generated code
- CI
- Shared schemas and enums

## Query naming

Every query name starts with a verb, reads included. Bare nouns (`Applications`) and a zoo of
write verbs (`Insert`, `Add`, `Save`, `Modify`) make the generated method list unscannable.

- Reads: `List` (many rows), `Get` (one row), `Count`.
- Writes: `Create`, `Update`, `Upsert`, `Delete`, `Set` (one column or flag).

```sql
-- name: ListApplicationsByStatus :many
-- name: GetApplication :one
-- name: CountApplications :one
-- name: UpsertApplication :one
-- name: SetApplicationStatus :exec
```

## Parameters

Name every parameter. Unnamed `$n` generates `Column2`, `Column3`, which say nothing at the call
site and silently reorder when the query changes.

- `sqlc.arg(name)` for required parameters.
- `sqlc.narg(name)` for optional filters, so the Go field is nullable and the query tests
  `IS NULL`. Never a sentinel such as `'' = $2`: it cannot tell "no filter" from "filter on
  empty".
- `sqlc.slice(name)` for `IN` lists on engines without arrays. On Postgres prefer
  `= ANY(sqlc.arg(ids)::uuid[])`.

```sql
-- name: ListInsights :many
SELECT id, kind, created_at
FROM insights
WHERE (sqlc.narg(kind)::text IS NULL OR kind = sqlc.narg(kind))
  AND created_at >= sqlc.arg(since)
ORDER BY created_at DESC;
```

## Config

Default to `sql_package: pgx/v5`. Choose `database/sql` only for a reason (a required
driver-agnostic seam, an existing `database/sql` codebase) and record it in an ADR.

- `overrides` map `uuid` to `github.com/google/uuid.UUID`, `timestamptz` to `time.Time`, and the
  nullable forms to pointers, so no `pgtype` value leaks into the store.
- `rename` fixes initialisms (`id: ID`, `url: URL`) so generated names pass lint.
- `emit_empty_slices: true` so `:many` returns `[]` rather than `nil`.

```yaml
version: "2"
sql:
  - engine: postgresql
    schema: db/migrations
    queries: db/queries
    gen:
      go:
        package: dbgen
        out: internal/store/dbgen
        sql_package: pgx/v5
        emit_empty_slices: true
        overrides:
          - db_type: uuid
            go_type: github.com/google/uuid.UUID
          - db_type: uuid
            nullable: true
            go_type: { import: github.com/google/uuid, type: UUID, pointer: true }
          - db_type: timestamptz
            go_type: time.Time
          - db_type: timestamptz
            nullable: true
            go_type: { import: time, type: Time, pointer: true }
        rename:
          id: ID
          url: URL
```

## Generated code

- Commit it. A build must not need sqlc installed, and reviewers see what a query change does.
- Only the store package imports it. Enforce with `depguard` so a handler cannot reach for
  `dbgen`.
- Pin the sqlc version in one place (a `go.mod` tool directive, or one version string shared by
  the Makefile and CI). Different versions emit different code and produce noisy diffs.

## CI

Regenerate, then fail on any change, untracked files included. `git diff --exit-code` misses a
newly generated file; `git status --porcelain` does not.

```yaml
- name: sqlc is up to date
  run: |
    go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate
    test -z "$(git status --porcelain)" || { git status --porcelain; exit 1; }
```

A repo that records a decision to commit generated code but has no such step is not enforcing it.

## Shared schemas and enums

When several sqlc packages share one schema, each package generates its own copy of every enum
type, and the copies are not interchangeable. Map each to the domain type in the store (see
`references/enums.md` in `go-idiomatic`). `text` + `CHECK`, per the Schema rules in `SKILL.md`,
generates plain `string` and avoids the duplication.
