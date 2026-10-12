# Logging Reference

The decisions behind the core logging rules in `SKILL.md`: wiring the logger, naming keys,
choosing a level, and keeping logs apart from what the program prints for its user.

---

## Wiring

`main` builds the one logger and passes it down through constructors:

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
    if err := run(os.Args[1:], os.Stdout, logger); err != nil {
        logger.Error("exit", "err", err)
        os.Exit(1)
    }
}
```

- JSON for anything a machine collects; `slog.NewTextHandler` for a tool only a person reads.
  Logs go to stderr.
- Scope a component once, at construction: `logger.With("component", "scheduler")`.
- An optional logger defaults to `slog.New(slog.DiscardHandler)`, never `slog.Default()`: a
  global logger cannot be silenced or inspected in a test.
- Use the `*Context` variants whenever a `ctx` is in scope, so request or trace IDs on the context
  reach the log.
- Tests pass `slog.New(slog.DiscardHandler)`. Assert on a log line only when it is the behaviour,
  by decoding a JSON handler's output.

---

## Messages and Keys

Log aggregators group by message. A message with an ID in it is a new group per ID, and nothing can
be filtered by that ID:

```go
// Bad
log.Printf("tick failed for task %d: %v", id, err)
logger.Error(fmt.Sprintf("tick failed for task %d", id), "err", err)

// Good
logger.ErrorContext(ctx, "tick", "task_id", id, "err", err)
```

- Messages follow error-string style: lowercase, short, no punctuation.
- Keys are `snake_case` key-value pairs: `err` for the error, `<noun>_id` for identifiers, `path`,
  `duration`.
- One key means one thing across the codebase. Search for the key the repo already uses before
  inventing one.
- Never log secrets, tokens or personal data. Log the ID, not the email address.

---

## Levels

| Level   | Means                                                                              |
| ------- | ---------------------------------------------------------------------------------- |
| `Debug` | Detail for diagnosing one problem. Off by default.                                 |
| `Info`  | The normal record: startup, configuration, a unit of work starting or finishing.   |
| `Warn`  | Something went wrong, but the operation succeeded or will recover on its own.      |
| `Error` | An operation failed, and this is the top of the chain: nobody else will report it. |

Between `Warn` and `Error`, ask "did the operation succeed?". `Info` inside a tight loop belongs at
`Debug`, or as one summary line when the loop ends.

---

## Output Is Not Logging

What the user asked a CLI for (results, a listening URL, progress) is output, written to an
injected `io.Writer` so a test can capture it. Logs are diagnostics for whoever runs the program.

```go
func (a *App) Serve(ctx context.Context, srv *http.Server) error {
    if _, err := fmt.Fprintf(a.out, "serving http://%s\n", srv.Addr); err != nil {
        return fmt.Errorf("write output: %w", err)
    }
    a.logger.InfoContext(ctx, "serve", "addr", srv.Addr)
    return srv.ListenAndServe()
}
```

A command whose only job is printing (`version`, `list`) needs no logger.
