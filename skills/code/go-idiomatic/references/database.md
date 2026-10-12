# Database Reference

The Go side of a SQL database. It holds for any driver and any query tool (`database/sql`, pgx,
sqlc, sqlx): the rules are about where types live and how errors, NULLs, time and transactions
cross a package boundary. For the schema, migrations and SQL itself, use the `postgres` skill.

## Contents

- [The Store Owns the Database](#the-store-owns-the-database)
- [Transactions](#transactions)
- [Read Then Write](#read-then-write)
- [NULL at the Boundary](#null-at-the-boundary)
- [Constraint Violations](#constraint-violations)
- [Time](#time)
- [Connection Setup](#connection-setup)
- [Testing](#testing)

---

## The Store Owns the Database

The store package is the only package that imports the driver or the generated query package.
These never appear in an exported signature, a return value or another package's imports:

- driver null wrappers (`sql.Null*`, `pgtype` values)
- generated row and parameter structs
- transaction handles
- driver "no rows" errors and driver error types (such as the Postgres error carrying a SQLSTATE
  code)

Each store method takes and returns the domain types the caller already uses, and translates
errors at the seam as [errors.md](errors.md#sentinel-layering) describes. That file owns the
absence contract (`ErrNotFound`); do not restate it here or check a driver's no-rows error
outside the store.

Enforce the boundary with the linter rather than review: a `depguard` rule that denies the
generated package and the driver to every package except the store.

Smells this prevents: a store method returning a generated row, a web handler comparing against
the driver's no-rows error, a service that accepts a transaction handle as a parameter.

---

## Transactions

Keep one helper that runs a function inside a transaction, and have every multi-statement method
go through it:

```go
func (s *Store) inTx(ctx context.Context, fn func(q queries) error) error {
    tx, err := s.db.begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    if err := fn(tx.queries()); err != nil {
        return errors.Join(err, rollback(ctx, tx))
    }
    if err := tx.commit(ctx); err != nil {
        return errors.Join(fmt.Errorf("commit tx: %w", err), rollback(ctx, tx))
    }
    return nil
}

func rollback(ctx context.Context, tx transaction) error {
    if err := tx.rollback(context.WithoutCancel(ctx)); err != nil {
        return fmt.Errorf("rollback tx: %w", err)
    }
    return nil
}
```

Rules the shape encodes:

1. Roll back on every path that did not commit: an error from `fn` or a failed commit. Return
   from each branch explicitly; a deferred rollback that tests a named `err` return is silently
   defeated by a shadowed `err`. If `fn` can panic, recover, roll back and re-panic.
2. Roll back with `context.WithoutCancel(ctx)`. A cancelled request context makes the rollback
   fail, and the connection returns to the pool with the transaction still open.
3. Wrap the commit error once, here. `fn` returns its own errors with their own context.
4. The callback receives the query handle, not the transaction. The transaction never leaves the
   helper, so no service can commit or roll back behind the store's back.

Copy-pasted begin/defer/commit blocks drift into several styles, each with a different bug. Write
the helper once per store.

---

## Read Then Write

A read that decides a write belongs in the same transaction as the write. Reading the old state
outside the transaction and then writing inside it lets another writer change the row in between,
so the decision is made on stale data:

```go
func (s *Store) SetStatus(ctx context.Context, id ApplicationID, to Status) error {
    return s.inTx(ctx, func(q queries) error {
        old, err := q.status(ctx, id)
        if err != nil {
            return fmt.Errorf("read status of %d: %w", id, err)
        }
        if !old.CanMoveTo(to) {
            return fmt.Errorf("move %d from %s to %s: %w", id, old, to, ErrInvalidTransition)
        }
        if err := q.setStatus(ctx, id, to); err != nil {
            return fmt.Errorf("set status of %d: %w", id, err)
        }
        return nil
    })
}
```

When the read must also block concurrent writers, lock the row in the query (`FOR UPDATE`, see the
`postgres` skill) or make the write conditional on the old value.

---

## NULL at the Boundary

A nullable column is an optional value: `Option[T]` in the domain type (see
[optional.md](optional.md)). The store converts in both directions, in one place:

```go
func toOption[T any](v nullable[T]) option.Option[T] {
    if !v.Valid {
        return option.None[T]()
    }
    return option.Some(v.Value)
}

func fromOption[T any](o option.Option[T]) nullable[T] {
    if o.IsNone() {
        return nullable[T]{}
    }
    return nullable[T]{Value: o.Unwrap(), Valid: true}
}
```

1. Never read a null wrapper's value field (`.String`, `.Int64`) without checking `.Valid`. An
   invalid wrapper holds the zero value, which is indistinguishable from real data.
2. Never write `''` or `0` to mean NULL in one direction while reading NULL back as absent in the
   other. That round trip turns "not set" into "set to empty" and the data silently changes.
3. A column that is `NOT NULL` is a plain `T`, not an `Option`. Fix the column or the type, not
   the mapping.

---

## Constraint Violations

A unique violation becomes a domain error. Match it on the constraint name, not only the error
code: the code says some unique constraint failed, the name says which one, and a table with two
unique constraints needs two different answers.

```go
var ErrEmailTaken = errors.New("store: email taken")

func (s *Store) CreateUser(ctx context.Context, u User) error {
    err := s.insertUser(ctx, u)
    if isUniqueViolation(err, "users_email_key") {
        return fmt.Errorf("create user %q: %w", u.Email, ErrEmailTaken)
    }
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }
    return nil
}
```

`isUniqueViolation` is the only function that inspects the driver error type; it checks the code
and the constraint name and lives in the store. Name constraints explicitly in the migration so
the string in Go is stable. The same applies to foreign-key and check violations.

---

## Time

The store never calls `time.Now()` and SQL never calls `now()` for a value the application
compares. Take the instant as a parameter from the injected clock and pass it to the query.
A test then controls time, and every comparison in one request uses one instant. The
`postgres` skill states the one-clock-per-comparison rule from the SQL side.

---

## Connection Setup

Configure the pool and prove the database is reachable before serving traffic:

1. Set the pool limits explicitly (max open and idle connections, connection lifetime). The
   defaults are unlimited open connections, which exhausts the server under load.
2. Ping with a context deadline at startup and fail `main` if it errors, so a bad DSN is a crash
   at boot and not a 500 on the first request.
3. Set a statement timeout (per connection, or per query through the context) so one slow query
   cannot hold a connection forever.
4. Close the pool in `main`'s shutdown path and join its error as [errors.md](errors.md#cleanup-errors)
   describes.

---

## Testing

Test the store against real Postgres, never a mocked driver (see
[testing.md](testing.md#databases)). The shape that stays fast and isolated:

1. One shared server per test binary, started once on first use and stopped at exit. Tests that
   don't touch the database don't pay for it.
2. An environment variable that points the tests at an already running server instead, for CI
   services and a warm local container. When it is set, start nothing.
3. Migrate one template database once, then create a fresh database per test by cloning the
   template. Cloning takes milliseconds and every test starts from a known schema.
4. Drop the cloned database in `t.Cleanup`. Do not `TRUNCATE` between tests: it serialises the
   suite on one database, leaks rows when a test forgets, and blocks `t.Parallel()`.

```go
func newTestStore(t *testing.T) *Store {
    t.Helper()
    dsn := cloneTemplate(t)
    s, err := Open(t.Context(), dsn)
    if err != nil {
        t.Fatalf("Open() error = %v", err)
    }
    t.Cleanup(func() {
        if err := s.Close(); err != nil {
            t.Errorf("Close() error = %v", err)
        }
    })
    return s
}
```

Run the contract suite from [testing.md](testing.md#test-doubles-fakes-and-contract-suites)
against the real store built this way, and against the in-memory fake.
