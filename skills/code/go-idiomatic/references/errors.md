# Errors Reference

How errors are created, wrapped, matched, handled and cleaned up. Logging an error at the top of
the chain is covered in `references/logging.md`.

## Contents

- [Error Strings](#error-strings)
- [Wrap With Context](#wrap-with-context)
- [Sentinel, Type or Plain](#sentinel-type-or-plain)
- [Sentinel Layering](#sentinel-layering)
- [Handle Once](#handle-once)
- [Cleanup Errors](#cleanup-errors)
- [Panic](#panic)
- [Lint Enforcement](#lint-enforcement)

---

## Error Strings

Lowercase, no trailing punctuation, no "failed to" or "error:" prefix. Wrapped errors are joined
with `: `, so each layer reads as one link in a chain:

```text
create order: charge card: dial payment gateway: connection refused
```

```go
// Bad: capitalised, punctuated, says "failed" at every layer
return fmt.Errorf("Failed to charge card: %w.", err)

// Good
return fmt.Errorf("charge card: %w", err)
```

---

## Wrap With Context

Each layer adds what it knows and nothing the layer below already said. The form is
`"<verb> <noun>: %w"`, with an identifier when one is in scope:

```go
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("load config %s: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config %s: %w", path, err)
    }

    return &cfg, nil
}
```

Use `%w` by default so callers can match with `errors.Is` / `errors.As`. Use `%v` only where a
caller must not depend on the underlying error: an exported API that would otherwise leak a
dependency's error types into its contract.

Return a bare `err` only when the function adds nothing a reader of the chain would want, such as
a one-line passthrough.

---

## Sentinel, Type or Plain

| Caller needs to…                       | Use                                   | Example                                        |
| -------------------------------------- | ------------------------------------- | ---------------------------------------------- |
| Branch on a specific condition         | Sentinel (`errors.New` package var)   | `var ErrNotFound = errors.New("store: not found")` |
| Read structured data out of the error  | Custom type, matched with `errors.As` | `type ValidationError struct{ Field, Msg string }` |
| Only report it                         | `fmt.Errorf` with `%w`                | `fmt.Errorf("dial %s: %w", addr, err)`         |
| Report several independent failures    | `errors.Join`                         | `errors.Join(errs...)`                         |

Most errors are the last row. A sentinel or type is API: every exported one is a promise to
callers that the condition stays detectable.

A custom type that wraps another error implements `Unwrap` so the chain stays inspectable:

```go
type QueryError struct {
    Query string
    Err   error
}

func (e *QueryError) Error() string { return fmt.Sprintf("query %q: %v", e.Query, e.Err) }
func (e *QueryError) Unwrap() error { return e.Err }
```

---

## Sentinel Layering

Sentinels drift into every package in several naming styles unless each one has a clear owner.
The rules:

1. **Export only when a caller in another package branches on it.** If nothing outside the
   package calls `errors.Is` on it, make it unexported, or drop it for a plain `fmt.Errorf`.
2. **Define it in the package that detects the condition**, next to the code that returns it.
   Never collect sentinels in a shared `errs` or `errors` package: that couples every package to
   every condition.
3. **Translate at the seam.** A package never lets a dependency's sentinel become part of its own
   contract. The store turns `sql.ErrNoRows` into its own `ErrNotFound`, so no caller imports
   `database/sql` to ask "was it missing?". The HTTP layer turns `store.ErrNotFound` into a 404.
4. **Name it `ErrXxx` (or `errXxx`), and prefix the message with the package name**, as the
   standard library does (`sql: no rows in result set`). A sentinel is often returned unwrapped,
   so its own message must say where it came from.

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

func (h *Handler) task(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        http.Error(w, "invalid task id", http.StatusBadRequest)
        return
    }

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
    writeJSON(w, t)
}
```

Never compare errors with `==` or a type assertion: wrapping breaks both. Use `errors.Is` and
`errors.As`.

---

## Handle Once

An error is either returned or handled, never both. Handling means logging it, turning it into a
response, or deciding it doesn't matter. Logging and returning produces one log line per layer for
a single failure, and the top layer logs it again anyway.

```go
// Bad: logged here, then logged again by whoever receives it
if err := s.charge(ctx, order); err != nil {
    s.logger.ErrorContext(ctx, "charge card", "order_id", order.ID, "err", err)
    return fmt.Errorf("charge card: %w", err)
}

// Good: add context and return; the top of the chain decides
if err := s.charge(ctx, order); err != nil {
    return fmt.Errorf("charge card for order %s: %w", order.ID, err)
}
```

The top of the chain is wherever the error stops travelling: an HTTP handler, a worker loop, a
deferred cleanup with nothing to return to, or `main`. That is the only place it is logged.

---

## Cleanup Errors

Never discard an error with `_ =`. A discarded `Close` on a written file is lost data; a discarded
`Remove` is a leaked file nobody hears about. The choice is between returning it and logging it.

**Something to return to:** join the cleanup error into the function's result with a named
return. `errors.Join` drops nils, so the success path stays `nil`:

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

The same pattern applies to a file opened only for reading. Prefer `os.ReadFile` when the whole
file is wanted anyway.

**Nothing to return to:** a background loop, a rollback after the real error has already been
returned, a temp file removed during shutdown. Log it at `Warn` through the injected logger, so it
is visible but does not page anyone:

```go
for _, path := range stale {
    if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
        s.logger.WarnContext(ctx, "remove stale run file", "path", path, "err", err)
    }
}
```

**The one bare-close exception:** `defer resp.Body.Close()` on an HTTP response body. The body has
been read; its `Close` error carries nothing. The lint config below excludes exactly that call.

Infallible writes (`bytes.Buffer`, `strings.Builder`, `hash.Hash`) need no handling and no `_ =`:
errcheck already knows they never fail.

---

## Panic

Panic only when the program itself is wrong, never because the world is:

- A `MustXxx` helper called with constant input at package init: `regexp.MustCompile`,
  `template.Must`.
- A `default:` branch for a case the type system already rules out.
- A broken invariant that means every later result would be wrong.

```go
var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// Bad: a missing row is an expected failure, and callers cannot handle a panic
func GetUser(id string) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err)
    }
    return user
}
```

Recover only at a goroutine boundary you own (an HTTP middleware, a worker), log it with the stack,
and turn it into a failed request or job rather than a dead process:

```go
func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if v := recover(); v != nil {
                logger.ErrorContext(r.Context(), "panic", "panic", v, "stack", string(debug.Stack()))
                http.Error(w, "internal error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## Lint Enforcement

Merge into the repo's golangci-lint (v2) config:

```yaml
linters:
  enable:
    - errcheck
    - errorlint
  settings:
    errcheck:
      check-blank: true
      exclude-functions:
        - (io.ReadCloser).Close
```

- `errcheck` with `check-blank`: flags `_ = f()` as well as a bare `f()`.
- `exclude-functions`: allows a bare `Close` on a value typed `io.ReadCloser`, which is what
  `resp.Body` is. An `*os.File` is still checked.
- `errorlint`: flags `==` and type assertions on errors, and `%v` where `%w` was meant.
