# Goose Reference

Load this only when the repo uses goose (`github.com/pressly/goose`). The rules that hold for
any tool (immutability, real Downs, expand/contract, locks) live in
[migrations.md](migrations.md); this file is goose syntax and wiring.

## Contents

- [File Format](#file-format)
- [Transactions](#transactions)
- [Irreversible Downs](#irreversible-downs)
- [File Names](#file-names)
- [Embedding and Startup](#embedding-and-startup)

---

## File Format

A SQL migration is one file with an Up and a Down section:

```sql
-- +goose Up
CREATE TABLE tasks (id uuid CONSTRAINT tasks_pkey PRIMARY KEY);

-- +goose Down
DROP TABLE tasks;
```

Goose splits statements on semicolons. A statement containing semicolons of its own (a
`CREATE FUNCTION` body, a `DO` block) must be wrapped so it is sent whole:

```sql
-- +goose StatementBegin
CREATE FUNCTION touch() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
```

## Transactions

Goose wraps each migration in a transaction. `CREATE INDEX CONCURRENTLY` and other statements
Postgres refuses inside one need the opt-out, placed at the top of the file before the Up marker:

```sql
-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY tasks_project_id_idx ON tasks (project_id);

-- +goose Down
DROP INDEX CONCURRENTLY tasks_project_id_idx;
```

Keep a no-transaction migration to one statement; there is no rollback if a later one fails.

## Irreversible Downs

Goose has no built-in irreversible marker. Make the Down fail loudly:

```sql
-- +goose Down
-- +goose StatementBegin
SELECT 'irreversible: dropped column cannot be restored'::int;
-- +goose StatementEnd
```

The cast raises an error, so `goose down` stops instead of reporting success. Use one form
across the repo so it can be grepped.

## File Names

Goose orders by the numeric prefix. Sequential (`00017_add_tasks.sql`) reads cleanly and
makes a merge collision obvious; timestamps (`20261012093000_add_tasks.sql`) avoid collisions
between branches but sort out of merge order. Pick one per repo and keep it.

The part after the version must be unique across the directory. A repo that had two files named
`wis_region_to_loc` could not tell which one an error came from. Name by what the migration
does.

## Embedding and Startup

Embed the migrations so the binary carries them, and point goose at the embedded filesystem:

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

var migrateOnce sync.Once

func Migrate(ctx context.Context, db *sql.DB) error {
	var err error
	migrateOnce.Do(func() {
		goose.SetBaseFS(migrationsFS)
		if err = goose.SetDialect("postgres"); err != nil {
			return
		}
		err = goose.UpContext(ctx, db, "migrations")
	})
	return err
}
```

`goose.SetBaseFS` and `SetDialect` set package-level state, so run them once behind a
`sync.Once`. Apply at startup before the server accepts traffic, and fail startup on error.
If several replicas start together, serialise with `goose.WithSessionLocker` or run migrations
as a separate deploy step.
