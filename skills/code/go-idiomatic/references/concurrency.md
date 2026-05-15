# Go Concurrency Reference

Goroutines are cheap but not free. Every one you spawn is a resource you must manage — it needs
a predictable exit, a clear owner, and proper error propagation. Default to sequential code and
add concurrency only when the problem genuinely calls for it.

## Contents

- [When to reach for concurrency](#when-to-reach-for-concurrency)
- [The five rules](#the-five-rules)
- [Choosing the right primitive](#choosing-the-right-primitive)
- [Sync primitives quick reference](#sync-primitives-quick-reference)
- [Coordination patterns](#coordination-patterns)
- [Cancellation and timeouts](#cancellation-and-timeouts)
- [Avoiding goroutine leaks](#avoiding-goroutine-leaks)
- [Testing concurrent code](#testing-concurrent-code)
- [Pre-spawn checklist](#pre-spawn-checklist)
- [Anti-patterns](#anti-patterns)

---

## When to reach for concurrency

Concurrency is justified when the work is genuinely independent. The common cases:

- **Independent I/O** — fetching multiple URLs, querying several services simultaneously.
- **CPU-bound parallelism** — divide a large computation across cores and merge results.
- **Natural producer/consumer flow** — a pipeline where one stage produces faster than the next
  consumes, and buffering would otherwise balloon memory.

If none of those apply, sequential code is simpler, easier to test, and less likely to leak.
Resist the urge to add goroutines because something "might be slow later."

---

## The five rules

These are not guidelines — violating any of them produces bugs that are hard to reproduce:

1. **Every goroutine has a known exit.** Context cancellation, a closed channel, or an explicit
   signal. Without one, it lives until the process dies.

2. **The sender owns the channel and is the only closer.** Closing from the receiver side panics
   when the sender writes after close. Closing twice also panics.

3. **Specify channel direction at function boundaries.** `chan<-` (send-only) and `<-chan`
   (receive-only) let the compiler enforce ownership. A function that only reads a channel should
   declare `<-chan T`; one that only writes should declare `chan<- T`.

4. **Default to unbuffered channels.** A buffer masks backpressure and hides the fact that
   consumers are keeping up. Buffer only when you have a measured reason: decoupling burst
   latency, or a known bounded producer where a buffered send avoids a goroutine just to hold
   the value.

5. **Always select on `ctx.Done()`.** A `select` that doesn't include `<-ctx.Done()` will block
   after the caller cancels — that's a guaranteed goroutine leak.

---

## Choosing the right primitive

| Scenario | Use | Why |
| --- | --- | --- |
| Pass data between goroutines | `chan T` | Transfers ownership explicitly |
| Coordinate goroutine lifecycle | `chan` + `context.Context` | Clean shutdown via `select` |
| Protect shared struct fields | `sync.Mutex` / `sync.RWMutex` | Simple critical section |
| Simple counter or flag | `sync/atomic` (`atomic.Int64`, `atomic.Bool`) | Lock-free, minimal overhead |
| Read-heavy concurrent map | `sync.Map` | Optimised for stable key sets; concurrent map read/write without it causes a hard crash |
| One-time initialisation | `sync.Once` | Guaranteed single execution regardless of concurrent callers |
| Deduplicate identical in-flight calls | `x/sync/singleflight` | Cache-stampede prevention |
| Fan-out with error propagation | `x/sync/errgroup` | Structured goroutine group with first-error cancellation |

---

## Sync primitives quick reference

| Primitive | Trap to avoid |
| --- | --- |
| `sync.Mutex` | Never hold across I/O — you block every other goroutine trying to acquire it |
| `sync.RWMutex` | Never upgrade an `RLock` to `Lock` — deadlock. Drop the read lock first |
| `sync/atomic` | Use typed atomics (`atomic.Int64`, not `atomic.AddInt64`) — Go 1.19+ |
| `sync.Map` | Don't use for write-heavy workloads; a `RWMutex`-guarded `map` is faster there |
| `sync.Pool` | Always `Reset()` the object before `Put()` — pooled objects are reused with their old state |
| `sync.Once` | Go 1.21+: prefer `sync.OnceFunc` / `sync.OnceValue` for one-shot functions that return a value |
| `sync.WaitGroup` | Call `Add` before the `go` statement, not inside the goroutine — `Wait` can return early otherwise |
| `x/sync/singleflight` | The first call's return value is shared with all deduplicated callers — do not mutate it |
| `x/sync/errgroup` | `g.Wait()` returns only the first error; other errors are silently dropped |

---

## Coordination patterns

### errgroup — fan-out with first-error cancellation

The standard tool for running N concurrent operations and collecting the first error:

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) ([][]byte, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([][]byte, len(urls))

    for i, url := range urls {
        g.Go(func() error {
            data, err := fetch(ctx, url)
            if err != nil {
                return err
            }
            results[i] = data
            return nil
        })
    }

    return results, g.Wait()
}
```

`errgroup.WithContext` cancels the derived context as soon as any goroutine returns an error —
the others should check `ctx.Done()` and exit early.

### Worker pool — bounded concurrency with errgroup

Prefer `errgroup.SetLimit(n)` over a hand-rolled worker pool. It's less code and handles the
`WaitGroup` and error propagation for you:

```go
func processAll(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // at most 10 goroutines in flight

    for _, item := range items {
        g.Go(func() error {
            return process(ctx, item)
        })
    }

    return g.Wait()
}
```

### Hand-rolled worker pool

Only reach for this when you need explicit control over job/result channels (e.g. streaming
results back as they complete, rather than collecting all at the end):

```go
func workerPool(ctx context.Context, jobs <-chan Job, numWorkers int) <-chan Result {
    results := make(chan Result, numWorkers)

    var wg sync.WaitGroup
    for range numWorkers {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                select {
                case results <- process(ctx, job):
                case <-ctx.Done():
                    return
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

The closing goroutine waits for all workers before closing `results`, so the consumer's
`for range results` exits cleanly.

### Pipeline — generator → stage → stage

Each stage receives an input channel and returns an output channel. The sender closes when done;
the receiver ranges until closed:

```go
func generate(ctx context.Context, values []int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, v := range values {
            select {
            case out <- v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

### Fan-out / fan-in

Distribute one channel to N workers, then merge N result channels back into one:

```go
func merge(ctx context.Context, channels ...<-chan Result) <-chan Result {
    out := make(chan Result)
    var wg sync.WaitGroup

    forward := func(ch <-chan Result) {
        defer wg.Done()
        for v := range ch {
            select {
            case out <- v:
            case <-ctx.Done():
                return
            }
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go forward(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### Graceful shutdown

Use `signal.NotifyContext` (Go 1.16+) — it wires OS signals to a context and cleans up
automatically on cancel:

```go
func run() error {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":8080", Handler: handler}

    go func() {
        <-ctx.Done()
        stop() // release signal resources early

        shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := srv.Shutdown(shutCtx); err != nil {
            log.Printf("shutdown error: %v", err)
        }
    }()

    if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
        return err
    }
    return nil
}
```

---

## Cancellation and timeouts

Context rules — violate any of these and you have either hidden coupling or a leak:

- **First argument, always.** `func Foo(ctx context.Context, ...)` — never store a context on a
  struct.
- **Derive at the edges.** HTTP handlers, background workers, and CLI entry points create child
  contexts with deadlines. Internal helpers receive and propagate, but never create fresh
  `context.Background()` roots mid-call-stack.
- **Always call `cancel()`.** Even when the operation succeeds, a deferred cancel releases the
  associated resources immediately rather than waiting for the parent to cancel.

```go
func fetchWithTimeout(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel() // critical even on success

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch %s: %w", url, err)
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}
```

---

## Avoiding goroutine leaks

### Blocked send

A goroutine that blocks forever on a channel send leaks if no receiver ever arrives:

```go
// Bad: blocks if caller discards the returned channel
func leakyWork() <-chan Result {
    ch := make(chan Result)
    go func() {
        ch <- doWork() // hangs forever if nothing receives
    }()
    return ch
}

// Good: select on ctx.Done() so the goroutine can exit
func safeWork(ctx context.Context) <-chan Result {
    ch := make(chan Result, 1)
    go func() {
        result := doWork()
        select {
        case ch <- result:
        case <-ctx.Done():
        }
    }()
    return ch
}
```

### `time.After` in a loop

Each `time.After` call allocates a timer that lives until it fires, regardless of whether you
ever read the channel. In a loop this accumulates:

```go
// Bad: timer per iteration, never GC'd until it fires
for {
    select {
    case msg := <-in:
        handle(msg)
    case <-time.After(5 * time.Second):
        log.Println("idle")
    }
}

// Good: one timer, reset on each iteration
timer := time.NewTimer(5 * time.Second)
defer timer.Stop()
for {
    select {
    case msg := <-in:
        if !timer.Stop() {
            <-timer.C
        }
        timer.Reset(5 * time.Second)
        handle(msg)
    case <-timer.C:
        log.Println("idle")
        timer.Reset(5 * time.Second)
    }
}
```

### Unbounded goroutine spawning

Spawning one goroutine per item in an unbounded loop will exhaust memory under load.
Use `errgroup.SetLimit(n)` — it blocks `g.Go` when the limit is reached, applying backpressure
automatically.

---

## Testing concurrent code

**`-race` is non-negotiable.** Add it to your CI test command permanently:

```bash
go test -race ./...
```

The race detector finds actual data races at runtime — it will not catch every race, but it
catches the ones that actually happen in your test runs. Running without `-race` in CI means
races ship silently.

**Goroutine leak detection with `goleak`:**

```go
import "go.uber.org/goleak"

func TestWorkerPool(t *testing.T) {
    defer goleak.VerifyNone(t)

    // test body
}
```

`goleak.VerifyNone` fails the test if any goroutines started during the test are still running
at the end. Use it on any test that touches goroutines.

**Avoid `time.Sleep` for synchronisation.** Tests that `sleep` to wait for a goroutine to finish
are flaky on slow machines and wasteful on fast ones. Use `sync.WaitGroup`, channels, or
`goleak` instead.

---

## Pre-spawn checklist

Before writing `go func()`:

- [ ] **How does it exit?** — context cancellation, channel close, or explicit signal.
- [ ] **Who can stop it?** — is a `context.Context` or done channel threaded in?
- [ ] **Who waits for it?** — `sync.WaitGroup`, `errgroup`, or a result channel the caller
  must drain.
- [ ] **Who owns the channels?** — the creating goroutine is the sender and the closer.
- [ ] **Should this be synchronous?** — if the caller always blocks waiting for the result,
  a direct function call is simpler and has no overhead.

---

## Anti-patterns

| Mistake | Fix |
| --- | --- |
| Fire-and-forget goroutine with no exit | Pass `context.Context`; use `WaitGroup` or `errgroup` to wait |
| Closing a channel from the receiver | Only the sender closes; receiver ranges until the channel is closed |
| `time.After` in a hot loop | Reuse `time.NewTimer` + `Reset` |
| `select` without `ctx.Done()` | Any `select` in a goroutine must include `<-ctx.Done()` |
| `wg.Add` inside the goroutine | Call `Add` before `go` — `Wait` may return early otherwise |
| Sending a pointer over a channel | Send a copy or an immutable value; pointers recreate shared memory |
| Holding a mutex across I/O or a slow call | Keep critical sections to in-memory operations only |
| Upgrading `RLock` to `Lock` | Drop the read lock before acquiring the write lock |
| Unbounded goroutine spawning | `errgroup.SetLimit(n)` or a semaphore channel |
| Storing `context.Context` on a struct | Context is per-request — pass it as the first function argument |
| Forgetting `defer cancel()` | Always defer `cancel()` immediately after `WithTimeout` / `WithCancel` |

---

## References

- [Effective Go — Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go blog: Pipelines and Cancellation](https://go.dev/blog/pipelines)
- [samber/cc-skills-golang — golang-concurrency](https://github.com/samber/cc-skills-golang/blob/main/skills/golang-concurrency/SKILL.md)
