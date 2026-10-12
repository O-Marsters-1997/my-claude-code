# Go Patterns Reference

Design patterns and interface conventions for idiomatic Go.

## Contents

- [Core Principles](#core-principles)
- [Interface Design](#interface-design)
- [Anti-Patterns](#anti-patterns-to-avoid)

---

## Core Principles

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

### Define Types for Swappable Identifiers

Give an identifier its own type (`type RunID int64`) when it crosses a package boundary and
sits beside a sibling of the same underlying type, as in `Cancel(launchID, runID int64)`. The
compiler then rejects a swapped argument. Untyped literals (`Cancel(3, 4)`) and explicit
conversions (`RunID(launchID)`) still get through. An identifier with no same-typed sibling
stays a plain `string` or `int64`.

### Typed Value Sets

A fixed set of values is a string-backed type with named constants, defined once in the owning
package.

```go
type Status string

const (
	StatusPlaced    Status = "placed"
	StatusShipped   Status = "shipped"
	StatusDelivered Status = "delivered"
)

var statusLabels = map[Status]string{
	StatusPlaced:    "Placed",
	StatusShipped:   "On its way",
	StatusDelivered: "Delivered",
}

func ParseStatus(s string) (Status, error) {
	st := Status(s)
	if _, ok := statusLabels[st]; !ok {
		return "", fmt.Errorf("parse status %q: unknown value", s)
	}
	return st, nil
}
```

- The zero value is unknown, never a real member.
- Per-value data lives in a map keyed by the constant, which doubles as the membership set.
- Input becomes the type only through `ParseStatus`, which errors on an unknown value. No bare
  `Status(s)` casts, no fallback `default`.
- A `switch` on the type covers every constant and has no `default`.

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

## Anti-Patterns to Avoid

```go
// Bad: naked returns in long functions — reader can't tell what's returned
func process() (result int, err error) {
    // ... 50 lines ...
    return
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
