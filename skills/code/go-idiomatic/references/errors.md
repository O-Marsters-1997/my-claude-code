# Errors Reference

The decisions behind the core error rules in `SKILL.md`: which kind of error to create, where a
sentinel lives, how errors reach a client, where an error came from, how cleanup errors are kept,
and when to panic.

## Contents

- [Wrap With Context](#wrap-with-context)
- [Sentinel, Type or Plain](#sentinel-type-or-plain)
- [Sentinel Layering](#sentinel-layering)
- [Application Errors](#application-errors)
- [Error Origin Site](#error-origin-site)
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
   caller imports `database/sql` to ask "was it missing?". At the HTTP edge, store sentinels become
   [application errors](#application-errors) once, through one error map per module; handlers never
   switch on `store.ErrNotFound` with their own `http.Error` text.
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

---

## Application Errors

An error that reaches a client is an application error: a **kind**, a **public message** and the
**cause**. The zero kind is unknown, so an untranslated error can never pass as a 404.

```go
// package apperr

type Kind int

const (
    KindUnknown Kind = iota
    KindNotFound
    KindConflict
    KindInvalid
)

type Error struct {
    Kind   Kind
    Public string
    Err    error
}

func (e *Error) Error() string { return e.Public + ": " + e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func KindOf(err error) Kind {
    var e *Error
    if errors.As(err, &e) {
        return e.Kind
    }
    return KindUnknown
}
```

Callers branch on the kind. A freshly built value is a new pointer, so `errors.Is` against it is
always false:

```go
// Bad
if errors.Is(err, apperr.NotFound("task not found")) {

// Good
if apperr.KindOf(err) == apperr.KindNotFound {
```

Each module translates its store sentinels once, in one map at its HTTP edge, instead of every
handler switching on them:

```go
// package tasks

var errorMap = []struct {
    target error
    kind   apperr.Kind
    public string
}{
    {store.ErrNotFound, apperr.KindNotFound, "task not found"},
    {store.ErrDuplicateTitle, apperr.KindConflict, "a task with that title already exists"},
}

func translate(err error) error {
    for _, m := range errorMap {
        if errors.Is(err, m.target) {
            return &apperr.Error{Kind: m.kind, Public: m.public, Err: err}
        }
    }
    return err
}
```

One error writer serves every handler. It sends only the public message, never `err.Error()`,
which carries table names, SQL and driver detail. It logs the cause, with its
[origin site](#error-origin-site), only for an unknown kind, because a known kind is an expected
outcome and not a fault:

```go
var statusOf = map[apperr.Kind]int{
    apperr.KindNotFound: http.StatusNotFound,
    apperr.KindConflict: http.StatusConflict,
    apperr.KindInvalid:  http.StatusUnprocessableEntity,
}

func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
    var ae *apperr.Error
    if errors.As(err, &ae) && ae.Kind != apperr.KindUnknown {
        http.Error(w, ae.Public, statusOf[ae.Kind])
        return
    }
    logger.ErrorContext(r.Context(), "request failed", "err", err, "err_site", errsite.Of(err))
    http.Error(w, "internal error", http.StatusInternalServerError)
}
```

---

## Error Origin Site

A wrapped chain says what failed but not where. Record the file and line once, at the seam where a
foreign error enters our code: a driver call, an outbound HTTP request, a queue publish. Every
layer above adds only `<verb> <noun>` context, and the top-of-chain logger pulls the site out as
`err_site`.

```go
// package errsite

type siteError struct {
    err  error
    site string
}

func (e *siteError) Error() string { return e.err.Error() }
func (e *siteError) Unwrap() error { return e.err }

func Mark(err error) error {
    if err == nil {
        return nil
    }
    _, file, line, _ := runtime.Caller(1)
    site := filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))
    return &siteError{err: err, site: fmt.Sprintf("%s:%d", site, line)}
}

func Of(err error) string {
    var s *siteError
    if errors.As(err, &s) {
        return s.site
    }
    return ""
}
```

```go
// package store

if err != nil {
    return Task{}, fmt.Errorf("query task %d: %w", id, errsite.Mark(err))
}
```

Mark only at the seam. Marking at every layer repeats what the wrap chain already says, and
`errsite.Of` returns the outermost site, which is the least useful one. An expected condition the
seam translates into a sentinel (`sql.ErrNoRows` to `ErrNotFound`) needs no mark.

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
