# Errors Reference

The decisions behind the core error rules in `SKILL.md`: which kind of error to create, where a
sentinel lives, how cleanup errors are kept, and when to panic.

## Contents

- [Wrap With Context](#wrap-with-context)
- [Sentinel, Type or Plain](#sentinel-type-or-plain)
- [Sentinel Layering](#sentinel-layering)
- [Handle Once](#handle-once)
- [Cleanup Errors](#cleanup-errors)
- [Panic](#panic)

---

## Wrap With Context

Each layer adds what it knows, and the chain reads as one sentence:

```text
create order: charge card: dial payment gateway: connection refused
```

```go
// Bad: capitalised, punctuated, says "failed" at every layer
return fmt.Errorf("Failed to charge card: %w.", err)

// Good
return fmt.Errorf("charge card: %w", err)
```

Use `%v` instead of `%w` only where callers must not depend on the underlying error: an exported API
that would otherwise leak a dependency's error types into its contract.

---

## Sentinel, Type or Plain

| Caller needs to…                      | Use                                   |
| ------------------------------------- | ------------------------------------- |
| Branch on a specific condition        | Sentinel (`errors.New` package var)   |
| Read structured data out of the error | Custom type, matched with `errors.As` |
| Only report it                        | `fmt.Errorf` with `%w`                |
| Report several independent failures   | `errors.Join`                         |

Most errors are the last two rows. A sentinel or type is API: each exported one promises callers the
condition stays detectable. A custom type that wraps another error implements `Unwrap`.

---

## Sentinel Layering

Without an owner, sentinels drift into every package in several naming styles.

1. **Export only when a caller in another package branches on it.** Otherwise make it unexported,
   or use a plain `fmt.Errorf`.
2. **Define it in the package that detects the condition.** Never collect sentinels in a shared
   `errs` package: that couples every package to every condition.
3. **Translate at the seam.** The store turns `sql.ErrNoRows` into its own `ErrNotFound`, so no
   caller imports `database/sql` to ask "was it missing?". The HTTP layer turns that into a 404.
4. **Name it `ErrXxx`, and prefix the message with the package name**, as the standard library
   does (`sql: no rows in result set`). A sentinel is often returned unwrapped, so its message must
   say where it came from.

```go
// package store

var ErrNotFound = errors.New("store: not found")

func (s *Store) Task(ctx context.Context, id int64) (Task, error) {
    var t Task
    err := s.db.QueryRowContext(ctx, taskQuery, id).Scan(&t.ID, &t.Title)
    if errors.Is(err, sql.ErrNoRows) {
        return Task{}, fmt.Errorf("task %d: %w", id, ErrNotFound)
    }
    if err != nil {
        return Task{}, fmt.Errorf("query task %d: %w", id, err)
    }
    return t, nil
}
```

```go
// package web

t, err := h.store.Task(ctx, id)
switch {
case errors.Is(err, store.ErrNotFound):
    http.Error(w, "task not found", http.StatusNotFound)
    return
case err != nil:
    h.logger.ErrorContext(ctx, "load task", "task_id", id, "err", err)
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
}
```

---

## Handle Once

Logging and returning the same error produces one log line per layer for a single failure:

```go
// Bad
if err := s.charge(ctx, order); err != nil {
    s.logger.ErrorContext(ctx, "charge card", "order_id", order.ID, "err", err)
    return fmt.Errorf("charge card: %w", err)
}

// Good: the top of the chain logs it
if err := s.charge(ctx, order); err != nil {
    return fmt.Errorf("charge card for order %s: %w", order.ID, err)
}
```

The top of the chain is wherever the error stops travelling: an HTTP handler, a worker loop, a
deferred cleanup with nothing to return to, or `main`.

---

## Cleanup Errors

A discarded `Close` on a written file is lost data; a discarded `Remove` is a leaked file nobody
hears about.

**Something to return to:** join the cleanup error into a named return. `errors.Join` drops nils,
so the success path stays `nil`. The same applies to files opened only for reading:

```go
func WriteReport(path string, r Report) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create report: %w", err)
    }
    defer func() { err = errors.Join(err, f.Close()) }()

    if err := json.NewEncoder(f).Encode(r); err != nil {
        return fmt.Errorf("encode report: %w", err)
    }
    return nil
}
```

**Nothing to return to:** a background loop, or a rollback after the real error was already
returned. Log at `Warn`, so it is visible but pages nobody:

```go
if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
    s.logger.WarnContext(ctx, "remove stale run file", "path", path, "err", err)
}
```

The one bare close is `defer resp.Body.Close()` on a fully read HTTP response: its error carries
nothing. Infallible writes (`bytes.Buffer`, `strings.Builder`, `hash.Hash`) need no handling at all.

---

## Panic

Panic only when the program itself is wrong, never because the world is:

- A `MustXxx` helper on constant input at package init: `regexp.MustCompile`, `template.Must`.
- A `default:` branch for a case the type system already rules out.
- A broken invariant that makes every later result wrong.

A missing row, a timeout or bad input is an error, not a panic. Recover only at a goroutine
boundary you own (HTTP middleware, a worker), log the value with `debug.Stack()`, and fail that
request or job rather than the process.
