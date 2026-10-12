# Logging Reference

Structured logging with `log/slog`, and keeping logs apart from what the program prints for its
user. When to log an error rather than return it is covered in `references/errors.md`.

## Contents

- [Inject the Logger](#inject-the-logger)
- [Messages and Keys](#messages-and-keys)
- [Levels](#levels)
- [Output Is Not Logging](#output-is-not-logging)
- [Logging in Tests](#logging-in-tests)
- [Lint Enforcement](#lint-enforcement)

---

## Inject the Logger

`main` builds the one logger and passes it down through constructors, like any other dependency.
Nothing below `main` calls `log.*`, `slog.Info` or `slog.Default()`: a global logger cannot be
silenced or inspected in a test, and it hides which code logs at all.

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
    if err := run(os.Args[1:], os.Stdout, logger); err != nil {
        logger.Error("exit", "err", err)
        os.Exit(1)
    }
}
```

- JSON for anything a machine collects; `slog.NewTextHandler` for a tool only a person reads in a
  terminal. Logs go to stderr.
- Read the level from a flag or environment variable, defaulting to `Info`.
- Scope a component's logger once, at construction: `logger.With("component", "scheduler")`.
  Don't repeat the same attribute on every call.
- When a logger is optional, the default is `slog.New(slog.DiscardHandler)`, never
  `slog.Default()`.

```go
type Scheduler struct {
    store  *store.Store
    logger *slog.Logger
}

func NewScheduler(st *store.Store, logger *slog.Logger) *Scheduler {
    return &Scheduler{store: st, logger: logger.With("component", "scheduler")}
}
```

Use the `*Context` variants (`InfoContext`, `ErrorContext`) whenever a `ctx` is in scope, so a
handler can attach request or trace IDs carried by the context.

---

## Messages and Keys

The message names the event and never changes; everything variable goes in attributes. Log
aggregators group by message, so a message with an ID in it becomes a new group for every ID, and
nothing can be filtered by that ID.

```go
// Bad: unstructured; neither the task nor the error can be filtered on
log.Printf("tick failed for task %d: %v", id, err)

// Bad: structured call, but the message still varies
logger.Error(fmt.Sprintf("tick failed for task %d", id), "err", err)

// Good
logger.ErrorContext(ctx, "tick", "task_id", id, "err", err)
```

Messages follow error-string style: lowercase, short, no punctuation. Keys:

- `snake_case`, as key-value pairs: `"task_id", id`.
- `err` for the error, always.
- `<noun>_id` for identifiers, `path` for a filesystem path, `duration` for a `time.Duration`.
- One key means one thing across the codebase. Search for the key the repo already uses before
  inventing a new one.

Never log secrets, tokens, credentials or personal data. Log the ID, not the email address.

---

## Levels

| Level   | Means                                                                                  |
| ------- | -------------------------------------------------------------------------------------- |
| `Debug` | Detail for diagnosing one problem. Off by default.                                     |
| `Info`  | The normal record: startup, configuration, a unit of work starting or finishing.       |
| `Warn`  | Something went wrong, but the operation succeeded or will recover on its own.          |
| `Error` | An operation failed, and this is the top of the chain: nobody else will report it.     |

Between `Warn` and `Error`, ask "did the operation succeed?". A retried request that then
succeeded, or a temp file that could not be removed, is `Warn`. A job that gave up is `Error`.

`Info` inside a tight loop is noise; that belongs at `Debug`, or as one summary line when the
loop ends.

---

## Output Is Not Logging

What the user asked a CLI for (results, a listening URL, a progress line) is the program's output,
not a log. It goes to an injected `io.Writer`, formatted for a person. Logs are diagnostics for
whoever runs the program, and go through the logger.

```go
type App struct {
    out    io.Writer
    logger *slog.Logger
}

func (a *App) Serve(ctx context.Context, srv *http.Server) error {
    if _, err := fmt.Fprintf(a.out, "serving http://%s\n", srv.Addr); err != nil {
        return fmt.Errorf("write output: %w", err)
    }
    a.logger.InfoContext(ctx, "serve", "addr", srv.Addr)
    return srv.ListenAndServe()
}
```

Inject the writer rather than calling `fmt.Println`, so a test can capture the output and assert on
it. A command whose only job is printing (a `version` or `list` subcommand) needs no logger.

---

## Logging in Tests

Pass a discard logger by default, so test output stays clean and nothing depends on log lines:

```go
func newTestScheduler(t *testing.T) *Scheduler {
    t.Helper()
    return NewScheduler(newTestStore(t), slog.New(slog.DiscardHandler))
}
```

Assert on a log line only when the log line is the behaviour, such as a cleanup failure that must
be reported. Capture it with a JSON handler on a buffer and decode it; never match the text:

```go
var buf bytes.Buffer
logger := slog.New(slog.NewJSONHandler(&buf, nil))
// ... exercise the code ...
var entry struct {
    Level string `json:"level"`
    Msg   string `json:"msg"`
    Path  string `json:"path"`
}
if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
    t.Fatalf("decode log line: %v", err)
}
```

---

## Lint Enforcement

Merge into the repo's golangci-lint (v2) config:

```yaml
linters:
  enable:
    - forbidigo
    - sloglint
  settings:
    forbidigo:
      forbid:
        - pattern: ^log\.(Print|Fatal|Panic)
          msg: inject a *slog.Logger; package log is for main only
    sloglint:
      kv-only: true
      no-global: all
      context: scope
      static-msg: true
      key-naming-case: snake
  exclusions:
    rules:
      - path: ^cmd/
        linters:
          - forbidigo
```

- `forbidigo`: bans package `log` outside `cmd/`.
- `sloglint`: bans the global logger, requires the `*Context` variant when a `ctx` is in scope,
  a constant message, and snake_case keys.
