# Go Patterns Reference

Design patterns, error handling, concurrency, and interface conventions for idiomatic Go.

## Contents

- [Core Principles](#core-principles)
- [Error Handling](#error-handling-patterns)
- [Interface Design](#interface-design)
- [Memory and Performance](#memory-and-performance)
- [Go Tooling](#go-tooling-integration)
- [Quick Reference](#quick-reference-go-idioms)
- [Anti-Patterns](#anti-patterns-to-avoid)

---

## Core Principles

### Simplicity and Clarity

Go favors simplicity over cleverness. Code should be obvious to read.

```go
// Good: clear and direct
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return user, nil
}

// Bad: overly clever
func GetUser(id string) (*User, error) {
    return func() (*User, error) {
        if u, e := db.FindUser(id); e == nil {
            return u, nil
        } else {
            return nil, e
        }
    }()
}
```

### Make the Zero Value Useful

Design types so their zero value is immediately usable without explicit initialisation.

```go
// Good: zero value is useful
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}

// Bad: requires initialisation before use
type BadCounter struct {
    counts map[string]int // nil map panics on write
}
```

### Accept Interfaces, Return Structs

Functions should accept interface parameters and return concrete types. Returning an interface
hides useful type information from the caller without benefit.

```go
// Good
func ProcessData(r io.Reader) (*Result, error) {
    data, err := io.ReadAll(r)
    if err != nil {
        return nil, err
    }
    return &Result{Data: data}, nil
}
```

---

## Error Handling Patterns

### Wrap Errors with Context

Each layer of the call stack adds the context it knows about. Use `%w` to preserve the original
error for `errors.Is` / `errors.As` checks:

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

### Custom Error Types and Sentinels

```go
// Sentinel errors for common cases
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// Structured error with context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}
```

### Check Errors with `errors.Is` and `errors.As`

```go
func HandleError(err error) {
    if errors.Is(err, sql.ErrNoRows) {
        log.Println("no records found")
        return
    }

    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("validation error on %s: %s", validationErr.Field, validationErr.Message)
        return
    }

    log.Printf("unexpected error: %v", err)
}
```

### Never Ignore Errors

```go
// Bad
result, _ := doSomething()

// Good
result, err := doSomething()
if err != nil {
    return err
}

// Acceptable: when error truly doesn't matter (rare — document why)
_ = writer.Close() // best-effort cleanup
```

---

## Concurrency

Goroutines, channels, errgroup, worker pools, leak avoidance, and graceful shutdown are covered
in `references/concurrency.md`. Load that reference whenever the work involves `go`, `chan`,
`select`, `sync.*`, or `context` cancellation.

---

## Interface Design

### Keep Interfaces Small

The standard library's one-method interfaces (`io.Reader`, `io.Writer`, `io.Closer`) are the
model. Compose them when you need the combination:

```go
type ReadWriteCloser interface {
    io.Reader
    io.Writer
    io.Closer
}
```

### Define Interfaces in the Consumer Package

Define an interface where it is *used*, not where the concrete type lives. The concrete type
doesn't need to know it satisfies the interface:

```go
// In the service package (the consumer)
type UserStore interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

type Service struct {
    store UserStore
}
```

### Optional Behaviour with Type Assertions

```go
type Flusher interface {
    Flush() error
}

func WriteAndFlush(w io.Writer, data []byte) error {
    if _, err := w.Write(data); err != nil {
        return err
    }
    if f, ok := w.(Flusher); ok {
        return f.Flush()
    }
    return nil
}
```

---

## Memory and Performance

### Preallocate Slices When Size Is Known

```go
// Bad: multiple reallocations
func processItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// Good: single allocation
func processItems(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}
```

### Use `sync.Pool` for Frequent Allocations

```go
var bufferPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func ProcessRequest(data []byte) []byte {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    return buf.Bytes()
}
```

### Avoid String Concatenation in Loops

```go
// Bad: creates many short-lived strings
func join(parts []string) string {
    var result string
    for _, p := range parts {
        result += p + ","
    }
    return result
}

// Best: standard library handles this
func join(parts []string) string {
    return strings.Join(parts, ",")
}

// When you need more control: strings.Builder
func join(parts []string) string {
    var sb strings.Builder
    for i, p := range parts {
        if i > 0 {
            sb.WriteByte(',')
        }
        sb.WriteString(p)
    }
    return sb.String()
}
```

---

## Go Tooling Integration

```bash
# Build and run
go build ./...
go run ./cmd/myapp

# Testing
go test ./...
go test -race ./...
go test -cover ./...

# Static analysis
go vet ./...
staticcheck ./...
golangci-lint run

# Module management
go mod tidy
go mod verify

# Formatting
gofmt -w .
goimports -w .
```

---

## Quick Reference: Go Idioms

| Idiom | Description |
|-------|-------------|
| Accept interfaces, return structs | Functions accept interfaces, return concrete types |
| Errors are values | Treat errors as first-class values, not exceptions |
| Don't communicate by sharing memory | Use channels for coordination between goroutines |
| Make the zero value useful | Types should work without explicit initialisation |
| A little copying is better than a little dependency | Avoid unnecessary external dependencies |
| Clear is better than clever | Prioritise readability over cleverness |
| Return early | Handle errors first; keep the happy path unindented |

---

## Anti-Patterns to Avoid

```go
// Bad: naked returns in long functions — reader can't tell what's returned
func process() (result int, err error) {
    // ... 50 lines ...
    return
}

// Bad: panic for control flow — only panic for truly unrecoverable programmer errors
func GetUser(id string) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err)
    }
    return user
}

// Bad: context stored in a struct — context is per-request, not per-object
type Request struct {
    ctx context.Context
    ID  string
}

// Good: context as first parameter
func ProcessRequest(ctx context.Context, id string) error {
    // ...
}

// Bad: mixed value and pointer receivers on the same type
type Counter struct{ n int }
func (c Counter) Value() int  { return c.n }  // value
func (c *Counter) Inc()       { c.n++ }       // pointer
// Pick one and be consistent across all methods on the type
```

**Remember**: Go code should be boring in the best way — predictable, consistent, and easy to
understand. When in doubt, keep it simple.
