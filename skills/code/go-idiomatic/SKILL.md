---
name: go-idiomatic
description: >
  Go engineering standards for this repo — idiomatic style, package layout, error handling,
  concurrency, testing and design patterns.
  TRIGGER — load BEFORE writing or editing Go, not after: any .go file or go.mod/go.sum/go.work
  is being created or edited; the prompt names Go, Golang, goroutine, channel, interface, struct,
  `go test`, `go build`, golangci-lint, gofmt, table-driven test, or benchmark; or the working
  directory contains go.mod and the task is implementation, refactor, review or test-writing.
  Trigger on plain feature requests inside a Go repo — "add an endpoint", "extract a helper",
  "where should this live", "reduce this duplication", "write a test for this" — the user will
  not say "idiomatic"; the .go file is the trigger. Do not wait to be asked for style guidance,
  and load it even when another skill (tdd, simplify) is already active.
  SKIP only when no Go file is involved and Go is not named.
allowed-tools: Read, Edit, Write, Glob, Grep, Bash(go:*), Bash(golangci-lint:*), Bash(git:*), Agent
---

# Go Idiomatic

Guidance for writing Go that is simple, self-documenting, easy to test, and follows established
best practices. The goal is code a future reader can understand and modify without needing to ask
questions.

## Where to Look

Route to the right reference before reading further:

| If you're working on…                                                          | Load this                        |
| ------------------------------------------------------------------------------ | -------------------------------- |
| Goroutines, channels, sync primitives, errgroup, leaks, races                  | `references/concurrency.md`      |
| Error handling, interface design, anti-patterns                                | `references/patterns.md`         |
| Writing or reviewing tests, benchmarks, fuzzing, HTTP tests                    | `references/testing.md`          |
| Optional fields or return values — when to use `Option[T]`                     | `references/optional.md`         |
| Pointer vs value receivers, mutation vs optionality                             | `references/value-vs-pointer.md` |
| Scaffolding a Go HTTP client for a third-party API, writing a `NewXxxClient`, or attaching auth to outbound HTTP calls | `references/http-client.md` |

---

## Lint First

Let tooling handle mechanical rules so human judgment is reserved for the things linters can't catch.
Before touching code:

```bash
gofmt -w .            # canonical formatting — non-negotiable
goimports -w .        # manages import blocks
go vet ./...          # catches common correctness mistakes
golangci-lint run     # aggregated linter suite
```

Configure the linter and let it run — don't write rules here for anything it already enforces.

---

## Code Organisation

### Where code lives

The standard layout for an application (not a library):

```text
myproject/
├── cmd/myapp/main.go   # entry point only — wire dependencies here, no business logic
├── internal/           # all application code; Go prevents external import
│   ├── handler/
│   ├── service/
│   └── repository/
└── testdata/           # fixtures and golden files
```

`internal/` is the right home for nearly everything in an application. `pkg/` only belongs in
library modules — packages explicitly designed to be imported by other Go modules outside the repo.
Don't add a `pkg/` layer to an application just because it looks tidy; it signals a public API that
doesn't exist.

### When to extract a shared package

Repetition is normal; premature abstraction is worse than duplication. Extract shared logic on the
**third** repetition across meaningfully different packages — not the second. Two callsites might
be coincidence; three signal genuine multi-purpose value.

Before extracting, ask: is this helper tightly coupled to one domain? If yes, it belongs inside
`internal/` next to its primary caller. If it's genuinely domain-agnostic (string manipulation,
ID generation, time utilities), it earns its own package inside `internal/`.

Name by what the package _does_, not what it _is_:

```
internal/
├── stringx/     # string normalisation, truncation, slugs
├── idgen/       # UUID / snowflake ID generation
└── timeutil/    # time formatting, rounding, duration helpers
```

### A note on `utils`

`utils` is a legitimate package name — but only when the package is genuinely cross-cutting:
useful across many packages in the project and with **zero dependencies on any other internal
package**. `utils` functions cannot import anything else from elsewhere in `internal/`, functions that require these imports must be moved elsewhere.

`helpers` and `common` are not equivalent to `utils`. They're vague catch-alls with no signal about
purpose. If you reach for one, pause — the package probably has a real name waiting to be found.
If the code truly is general-purpose utility material that could live in any Go project, `utils` is
fine; otherwise, name it by what it does.

### Avoid package-level state

Global `var db *sql.DB` initialised in `init()` is invisible coupling that makes tests fragile.
Pass dependencies explicitly through constructors:

```go
type Service struct {
    db     *sql.DB
    logger *slog.Logger
}

func NewService(db *sql.DB, logger *slog.Logger) *Service {
    return &Service{db: db, logger: logger}
}
```

---

## Function Design

### Keep argument lists short

Functions with more than 3–4 arguments are hard to call correctly and easy to misread. Group
related arguments into a struct. For constructors or functions that grow over time, the
**functional options** pattern avoids a brittle positional argument list:

```go
type Server struct {
    addr    string
    timeout time.Duration
    logger  *slog.Logger
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func WithLogger(l *slog.Logger) Option {
    return func(s *Server) { s.logger = l }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second,
        logger:  slog.Default(),
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

Adding a new option is backward-compatible; adding a new positional argument is not.

### Declare complex arguments before the call

Inline constructor calls with multiple arguments are hard to read. Extract into named locals first —
the call site becomes self-explanatory and test cases become easier to construct:

```go
// Hard to read
client.Send(NewMessage("alice", "bob", time.Now().Add(-24*time.Hour), Priority(3)))

// Readable
sender := "alice"
recipient := "bob"
sentAt := time.Now().Add(-24 * time.Hour)
priority := Priority(3)
client.Send(NewMessage(sender, recipient, sentAt, priority))
```

### Return early

Validate and guard at the top. Handle the happy path at the bottom, unindented. Every extra nesting
level is cognitive overhead the reader has to track:

```go
func Process(input string) (Result, error) {
    if input == "" {
        return Result{}, ErrEmptyInput
    }
    if len(input) > MaxLen {
        return Result{}, ErrTooLong
    }
    return compute(input), nil
}
```

---

## Control Flow

### Eliminate unnecessary `else`

When a branch returns, the `else` block is dead weight — remove it:

```go
// Bad
if err != nil {
    return err
} else {
    process(data)
}

// Good
if err != nil {
    return err
}
process(data)
```

### Prefer `switch` over `if`/`else if` chains

A `switch` makes intent clearer and is easier to extend with a new case:

```go
// Bad
if status == "active" {
    handleActive()
} else if status == "pending" {
    handlePending()
} else {
    handleUnknown()
}

// Good
switch status {
case "active":
    handleActive()
case "pending":
    handlePending()
default:
    handleUnknown()
}
```

### Use `range` for iteration

Always prefer `range` over an index-based loop unless the index is integral to the logic:

```go
for _, item := range items {
    process(item)
}
```

---

## Value vs Pointer

The quick rule: use pointers for **mutation** and measurably **large structs**; use values for
small, immutable types. Never use `*T` to signal "this field might not be set" — that's what
`Option[T]` is for.

Load `references/value-vs-pointer.md` for the full decision guide: receiver consistency, when
"large" is actually large, interface satisfaction edge cases, and the nil-vs-optional problem.

---

## Self-Documenting Code

Well-named identifiers are the primary documentation. A name that explains what something does
makes a comment redundant — and redundant comments rot as the code evolves.

- Name functions by what they return or do: `ParseConfig`, `MustConnect`, `IsExpired`.
- Name booleans as predicates: `isActive`, `hasExpired`, `canRetry`.
- Avoid abbreviations that aren't Go conventions (`ctx`, `err`, `id`, `r`, `w` are fine; `b` for
  a business object is not).

`~/.claude/rules/comments.md` decides what earns a comment. It is the authority and this
skill does not restate or soften it. Read it before editing a file and apply it as written,
including its line cap and its exported-identifiers-only rule for doc comments. `internal/`
is not public API.

Never explain _what_ the code does — the code already does that. Comments that say "increment
counter" next to `count++`, or "returns error if nil" above `if err != nil`, add noise and train
readers to skip comments entirely.

When reviewing Go code for comment quality, use the `clean-comments` skill. Do not instruct it
to match the file's existing comment density.

---

## Testing

This section is the authority on how a Go test is written. Which tests a change needs, and how
much coverage is enough, is the repo's policy, not this skill's. For the TDD loop, use the `tdd`
skill.

Core rules:

1. Test behaviour through the exported API. Default to the black-box `package foo_test`; use
   `package foo` only when the unit is unexported by design, and `export_test.go` as a last resort.
2. Assert on outputs and state, not on which calls were made. A test that breaks when the code is
   refactored without a behaviour change is a change-detector: delete it.
3. Pick the most faithful double: real, then fake, then stub, then mock. Mock only state-changing
   calls across a system boundary (sending email, charging a card).
4. A fake is stateful and honest. One contract suite runs against both the fake and the real
   implementation. Force error paths with a one-method stub, not failure knobs on the fake.
5. Test SQL against a real database, never a mocked driver. Use the repo's DB-testing skill for
   the mechanics if it has one.
6. Compare with `cmp.Diff(want, got)` from `github.com/google/go-cmp` and report `(-want +got)`.
   Without go-cmp, compare scalars with `==` and structs field by field. Never `reflect.DeepEqual`.
7. Failure messages read `Func(in) = got, want want`. Match errors with `errors.Is`/`errors.As`,
   never by string.
8. Use a table only when every case is checked the same way: a slice of structs with a `name`
   field, run with `t.Run(tt.name, ...)`. An `if tt.wantErr` branch means split the test.
9. `t.Error` to keep going, `t.Fatal` when the rest of the test can't run. Never call `t.Fatal`
   from a goroutine the test didn't start on.
10. Helpers take `t`, call `t.Helper()`, fail on setup errors and register `t.Cleanup`. They set
    up; they don't assert. `TestMain` only for expensive setup shared by the whole package.
11. No `time.Sleep`. Poll a condition with a deadline or wait on a channel.
12. `t.Parallel()` only for tests that share no state: no globals, env, working dir or shared rows.
13. Compare JSON by decoding it, never as bytes. Normalise golden output before comparing, and a
    human reviews every `-update` diff.
14. Fuzz parsers of untrusted input, and commit the `testdata/fuzz/` corpus.
15. Test HTTP handlers through the real router and middleware.

Load `references/testing.md` for worked examples of each kind of test (unit, fake plus contract
suite, HTTP, golden, fuzz, benchmark) and a golangci-lint snippet that enforces the mechanical rules.

---

## See Also

| Skill                           | When to reach for it                                                        |
| ------------------------------- | --------------------------------------------------------------------------- |
| `tdd`                           | TDD cycle, red-green-refactor, writing the test first                       |
| `clean-comments`                | Audit and remove explanatory comments across Go code                        |
| `improve-codebase-architecture` | Identify structural improvements, shallow modules                           |
| `simplify`                      | Review changed code for quality, reuse, and efficiency after implementation |
