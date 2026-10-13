# Migrations Reference

Rules for schema migrations on Postgres, independent of the migration tool. For tool syntax
(file markers, embedding, startup), load the tool's own reference, such as [goose.md](goose.md).

## Contents

- [Applied Migrations Are Immutable](#applied-migrations-are-immutable)
- [Down Migrations](#down-migrations)
- [Breaking Changes](#breaking-changes)
- [Locks and Indexes](#locks-and-indexes)
- [ENUM Conversion](#enum-conversion)
- [One Schema Source of Truth](#one-schema-source-of-truth)
- [Tests](#tests)

---

## Applied Migrations Are Immutable

Never edit a migration that has run anywhere (another developer's database, CI, staging,
production). The tool records it as done and will not rerun it, so the edit reaches fresh
databases only and the schemas silently diverge. Add a new migration instead, even for a typo.

Only a migration that has never left your branch may be edited.

## Down Migrations

Every migration has either a real Down or an explicit irreversible marker. An empty Down is
neither: it reports success and changes nothing.

- A real Down restores the previous schema. If it cannot restore data (a dropped column), the
  migration is irreversible.
- An irreversible migration says so in the file, in one form used across the repo, and its Down
  fails loudly.
- Offer a `migrate-down` or `migrate-redo` recipe (Justfile, Makefile, script) only when the
  Downs are real. A recipe over empty Downs reads as a safety net and is not one. One repo had
  0 real Downs out of 17 while its Justfile offered both recipes.

## Breaking Changes

Use expand/contract. Never rename or drop in the same release that stops using the old shape.

1. **Expand:** add the new column, table or constraint (nullable, or a constant default).
2. **Migrate:** deploy code that writes both, then backfill old rows.
3. **Contract:** once nothing reads the old shape, drop it in a later migration.

A backfill is its own migration, never bundled with the DDL that adds the column. DDL takes
strong locks and a backfill runs long; together they hold the lock for the backfill's duration
and a failure rolls back the schema change too. Batch large backfills as
[SKILL.md](../SKILL.md#altering-tables) describes.

## Locks and Indexes

- **`CREATE INDEX CONCURRENTLY` cannot run inside a transaction.** Put it in a migration that
  opts out of the tool's transaction wrapper, alone, so a failure leaves nothing half-applied.
  A failed concurrent build leaves an invalid index; drop it before retrying.
- **Set `lock_timeout` on DDL against busy tables** (for example `SET lock_timeout = '5s'`).
  `ALTER TABLE` queues for an exclusive lock and every query behind it queues too; without a
  timeout, one long transaction elsewhere stalls the whole table. A timeout fails the migration
  fast, and you retry it.
- Keep DDL migrations short: one table, one concern.

## ENUM Conversion

An existing Postgres `ENUM` becomes `text` + a named `CHECK` in one migration per type:

```sql
ALTER TABLE jobs ALTER COLUMN status TYPE text USING status::text;
ALTER TABLE jobs ADD CONSTRAINT jobs_status_check
    CHECK (status IN ('queued', 'running', 'done', 'failed'));
DROP TYPE job_status;
```

Convert every column using the type before the `DROP TYPE`, and recreate any default that names an
enum value (`DEFAULT 'queued'::job_status`) as plain text, or the drop fails.

`ALTER COLUMN ... TYPE` rewrites the table under an `ACCESS EXCLUSIVE` lock. On a small table that
is fine with a `lock_timeout`. On a big one use expand/contract: add a `text` column, write both,
backfill in batches, swap readers over, then drop the enum column and the type.

## One Schema Source of Truth

The migrations are the schema. If the repo also keeps a hand-written `schema.sql` (read by
sqlc, docs or a test fixture), CI must check it against the migrations: apply the migrations to
a scratch database, dump it, and diff the result against the mirror. An unchecked mirror
drifts; one repo's mirror was missing two tables.

Prefer generating the mirror from the migrations over maintaining it by hand.

## Tests

Tests do not hard-code the latest migration version. An assertion like `version == 17` fails on
every new migration for no reason. Assert that the migrator reports no pending migrations, or
compare against the highest version found in the migrations directory.
