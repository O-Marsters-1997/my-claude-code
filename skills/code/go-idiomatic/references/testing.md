# Go Testing Patterns Reference

Worked examples for the core rules in `SKILL.md` § Testing. The rules are the authority; this file
shows them applied. Which tests a change needs is the repo's policy, not this file's.

For the TDD workflow (red-green-refactor, when to write the test first), use the `tdd` skill.

## Contents

- [Unit Tests](#unit-tests)
- [Table-Driven Tests](#table-driven-tests)
- [Subtests and Parallelism](#subtests-and-parallelism)
- [Test Helpers and TestMain](#test-helpers-and-testmain)
- [Test Doubles: Fakes and Contract Suites](#test-doubles-fakes-and-contract-suites)
- [Databases](#databases)
- [HTTP Handler Testing](#http-handler-testing)
- [Asynchronous Code](#asynchronous-code)
- [Golden Files](#golden-files)
- [Fuzzing](#fuzzing)
- [Benchmarks](#benchmarks)
- [Lint Enforcement](#lint-enforcement)
- [Testing Commands](#testing-commands)

---

## Unit Tests

Tests live in the black-box `_test` package and call only the exported API. Compare with
`cmp.Diff` and report the diff as `(-want +got)`:

```go
package config_test

import (
    "testing"

    "github.com/google/go-cmp/cmp"

    "example.com/app/internal/config"
)

func TestParse(t *testing.T) {
    got, err := config.Parse(`{"host": "localhost", "port": 8080}`)
    if err != nil {
        t.Fatalf("Parse() error = %v", err)
    }
    want := &config.Config{Host: "localhost", Port: 8080}
    if diff := cmp.Diff(want, got); diff != "" {
        t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
    }
}
```

`t.Fatalf` on the error because nothing after it can run; `t.Errorf` on the comparison so every
failure in a test is reported.

Without go-cmp, compare scalars directly and structs field by field, using the same message shape:

```go
if got.Port != 8080 {
    t.Errorf("Parse().Port = %d, want %d", got.Port, 8080)
}
```

---

## Table-Driven Tests

Use a table when every case runs the same code and is checked the same way. Cases are a slice of
structs with a `name` field:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {name: "positive", a: 2, b: 3, want: 5},
        {name: "negative", a: -1, b: -2, want: -3},
        {name: "mixed signs", a: -1, b: 1, want: 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := calc.Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### Success and error cases are separate tests

A `wantErr bool` field forces an `if tt.wantErr` branch, so half the table is checked one way and
half another. Split them. The error table checks every case with `errors.Is`:

```go
func TestParse(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  *config.Config
    }{
        {name: "full", input: `{"host": "localhost", "port": 8080}`, want: &config.Config{Host: "localhost", Port: 8080}},
        {name: "empty object", input: `{}`, want: &config.Config{}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := config.Parse(tt.input)
            if err != nil {
                t.Fatalf("Parse(%q) error = %v", tt.input, err)
            }
            if diff := cmp.Diff(tt.want, got); diff != "" {
                t.Errorf("Parse(%q) mismatch (-want +got):\n%s", tt.input, diff)
            }
        })
    }
}

func TestParseErrors(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {name: "invalid JSON", input: `{invalid}`, wantErr: config.ErrSyntax},
        {name: "empty input", input: "", wantErr: config.ErrEmpty},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := config.Parse(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("Parse(%q) error = %v, want %v", tt.input, err, tt.wantErr)
            }
        })
    }
}
```

Never match errors on `err.Error()` text. If a caller needs to tell errors apart, the package
exports a sentinel or a type, and the test uses `errors.Is` or `errors.As`.

---

## Subtests and Parallelism

Subtests group related cases and can be run alone with `-run 'TestUser/Create'`.

Call `t.Parallel()` only when the test shares no state: no package globals, no `t.Setenv`, no
`t.Chdir`, no rows another test reads. Since Go 1.22 each loop iteration has its own `tt`, so the
old `tt := tt` copy is gone:

```go
func TestNormalise(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name, in, want string
    }{
        {name: "trims", in: "  a  ", want: "a"},
        {name: "lowercases", in: "ABC", want: "abc"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            if got := text.Normalise(tt.in); got != tt.want {
                t.Errorf("Normalise(%q) = %q, want %q", tt.in, got, tt.want)
            }
        })
    }
}
```

If the parent calls `t.Parallel()`, so do its subtests, and the other way round (`tparallel`
checks this).

---

## Test Helpers and TestMain

A helper takes `t`, calls `t.Helper()` so failures point at the caller, fails on setup errors and
registers its own cleanup. It builds things; the test does the asserting:

```go
func newTestServer(t *testing.T, store userstore.Store) *httptest.Server {
    t.Helper()
    srv := httptest.NewServer(api.NewRouter(store))
    t.Cleanup(srv.Close)
    return srv
}
```

Skip generic `assertEqual`/`assertNoError` helpers. They hide which value was compared, and the
failure message loses the `Func(in) = got, want want` context.

`t.TempDir()` gives a directory that is removed after the test. `t.Setenv` and `t.Chdir` restore
their state afterwards and make the test non-parallel.

`TestMain` is for expensive setup the whole package shares, such as starting one database
container. Don't use it for per-test setup:

```go
func TestMain(m *testing.M) {
    stop, err := startDatabase()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    code := m.Run()
    stop()
    os.Exit(code)
}
```

---

## Test Doubles: Fakes and Contract Suites

Fidelity order: **real, then fake, then stub, then mock.** Use the real implementation when it is
fast and deterministic. Otherwise use a fake: a working in-memory implementation with real state.
Stubs return canned answers for one call. Mocks that assert which calls were made are only for
state-changing calls across a system boundary, where the call is the behaviour (an email sent, a
card charged).

### The interface and a fake

```go
package userstore

var ErrNotFound = errors.New("user not found")

type Store interface {
    Get(ctx context.Context, id string) (User, error)
    Save(ctx context.Context, u User) error
}
```

```go
package memstore

type Store struct {
    mu    sync.Mutex
    users map[string]userstore.User
}

func New() *Store {
    return &Store{users: make(map[string]userstore.User)}
}

func (s *Store) Get(_ context.Context, id string) (userstore.User, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    u, ok := s.users[id]
    if !ok {
        return userstore.User{}, userstore.ErrNotFound
    }
    return u, nil
}

func (s *Store) Save(_ context.Context, u userstore.User) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.users[u.ID] = u
    return nil
}
```

### The contract suite keeps the fake honest

One exported function states the behaviour every `Store` must have. It lives in a small test-support
package and runs against both implementations, so the fake cannot drift from the real one:

```go
package storetest

func Run(t *testing.T, newStore func(t *testing.T) userstore.Store) {
    t.Run("get missing user", func(t *testing.T) {
        s := newStore(t)
        _, err := s.Get(context.Background(), "nobody")
        if !errors.Is(err, userstore.ErrNotFound) {
            t.Errorf("Get(nobody) error = %v, want %v", err, userstore.ErrNotFound)
        }
    })
    t.Run("save then get", func(t *testing.T) {
        s := newStore(t)
        want := userstore.User{ID: "123", Name: "Alice"}
        if err := s.Save(context.Background(), want); err != nil {
            t.Fatalf("Save() error = %v", err)
        }
        got, err := s.Get(context.Background(), "123")
        if err != nil {
            t.Fatalf("Get(123) error = %v", err)
        }
        if diff := cmp.Diff(want, got); diff != "" {
            t.Errorf("Get(123) mismatch (-want +got):\n%s", diff)
        }
    })
}
```

```go
func TestMemStore(t *testing.T) {
    storetest.Run(t, func(t *testing.T) userstore.Store { return memstore.New() })
}

func TestPostgresStore(t *testing.T) {
    storetest.Run(t, func(t *testing.T) userstore.Store { return pgstore.New(newTestDB(t)) })
}
```

### Using the fake

Tests of code that depends on the store seed the fake and assert on the result or the stored state,
not on which methods were called:

```go
func TestRenameUser(t *testing.T) {
    store := memstore.New()
    if err := store.Save(context.Background(), userstore.User{ID: "123", Name: "Alice"}); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    svc := users.NewService(store)

    if err := svc.Rename(context.Background(), "123", "Alicia"); err != nil {
        t.Fatalf("Rename() error = %v", err)
    }

    got, err := store.Get(context.Background(), "123")
    if err != nil {
        t.Fatalf("Get(123) error = %v", err)
    }
    if got.Name != "Alicia" {
        t.Errorf("Name after Rename(123, Alicia) = %q, want %q", got.Name, "Alicia")
    }
}
```

### Error paths use a one-method stub

Don't add failure switches to the fake; they make it behave in ways the real store never does.
Embed the interface and override the one method that should fail:

```go
type failingSave struct {
    userstore.Store
    err error
}

func (f failingSave) Save(context.Context, userstore.User) error { return f.err }

func TestRenameSaveFails(t *testing.T) {
    store := memstore.New()
    if err := store.Save(context.Background(), userstore.User{ID: "123", Name: "Alice"}); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    errDisk := errors.New("disk full")
    svc := users.NewService(failingSave{Store: store, err: errDisk})

    err := svc.Rename(context.Background(), "123", "Alicia")
    if !errors.Is(err, errDisk) {
        t.Errorf("Rename() error = %v, want %v", err, errDisk)
    }
}
```

---

## Databases

Test SQL against a real database of the same engine and version as production. Never mock
`database/sql` or the driver: a mocked driver proves the query string was sent, not that it works.
The real store is what the contract suite above runs against.

For the mechanics (containers, template databases, per-test isolation), use the repo's
DB-testing skill if it has one.

---

## HTTP Handler Testing

Test through the real router so routing, middleware and encoding are exercised with the handler.
Build it the same way `main` does, with fakes behind it:

```go
func TestGetUser(t *testing.T) {
    store := memstore.New()
    if err := store.Save(context.Background(), userstore.User{ID: "123", Name: "Alice"}); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    srv := newTestServer(t, store)

    resp, err := http.Get(srv.URL + "/users/123")
    if err != nil {
        t.Fatalf("GET /users/123 error = %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("GET /users/123 status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
    var got api.User
    if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
        t.Fatalf("decode body: %v", err)
    }
    want := api.User{ID: "123", Name: "Alice"}
    if diff := cmp.Diff(want, got); diff != "" {
        t.Errorf("GET /users/123 body mismatch (-want +got):\n%s", diff)
    }
}
```

`httptest.NewRecorder()` with `router.ServeHTTP(rec, req)` also goes through the router and skips
the network. Calling a handler function directly is only for a handler with no routing or
middleware behaviour worth testing.

Compare JSON by decoding it, never as a string: key order and whitespace are not behaviour. When
the expected body is raw JSON, decode both sides into `any` and diff those:

```go
func jsonDiff(t *testing.T, want, got []byte) string {
    t.Helper()
    var w, g any
    if err := json.Unmarshal(want, &w); err != nil {
        t.Fatalf("decode want: %v", err)
    }
    if err := json.Unmarshal(got, &g); err != nil {
        t.Fatalf("decode got: %v", err)
    }
    return cmp.Diff(w, g)
}
```

---

## Asynchronous Code

Never `time.Sleep` to wait for work to finish. Wait on a channel the code signals, or poll a
condition until a deadline:

```go
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
    t.Helper()
    deadline := time.Now().Add(timeout)
    for !cond() {
        if time.Now().After(deadline) {
            t.Fatalf("condition not met within %v", timeout)
        }
        time.Sleep(10 * time.Millisecond)
    }
}
```

`t.Fatal` and `t.FailNow` must run on the test's own goroutine. From a goroutine the test starts,
send the error back on a channel, or use `t.Error`, and let the test goroutine decide whether to
stop.

---

## Golden Files

Store expected output in `testdata/` and regenerate it with `-update`. Normalise volatile parts
(timestamps, IDs, absolute paths, line endings) before both writing and comparing, so the file only
changes when behaviour does. Every `-update` diff gets reviewed by a human before commit:

```go
var update = flag.Bool("update", false, "update golden files")

func TestRender(t *testing.T) {
    tests := []struct {
        name  string
        input report.Template
    }{
        {name: "simple", input: report.Template{Name: "test"}},
        {name: "with items", input: report.Template{Name: "test", Items: []string{"a", "b"}}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := normalise(report.Render(tt.input))
            golden := filepath.Join("testdata", tt.name+".golden")

            if *update {
                if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
                    t.Fatalf("write golden file: %v", err)
                }
            }
            want, err := os.ReadFile(golden)
            if err != nil {
                t.Fatalf("read golden file: %v", err)
            }
            if diff := cmp.Diff(string(want), got); diff != "" {
                t.Errorf("Render(%s) mismatch (-want +got):\n%s", tt.name, diff)
            }
        })
    }
}

var timestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)

func normalise(b []byte) string {
    s := strings.ReplaceAll(string(b), "\r\n", "\n")
    return timestamp.ReplaceAllString(s, "<TIME>")
}
```

---

## Fuzzing

Fuzz any parser of untrusted input. Seed with realistic cases and check properties that must hold
for every input, such as a round trip or no panic:

```go
func FuzzParse(f *testing.F) {
    f.Add(`{"host": "localhost", "port": 8080}`)
    f.Add(`{}`)
    f.Add(``)

    f.Fuzz(func(t *testing.T, input string) {
        cfg, err := config.Parse(input)
        if err != nil {
            return
        }
        again, err := config.Parse(cfg.String())
        if err != nil {
            t.Fatalf("Parse(%q) round trip error = %v", cfg.String(), err)
        }
        if diff := cmp.Diff(cfg, again); diff != "" {
            t.Errorf("round trip mismatch (-want +got):\n%s", diff)
        }
    })
}
```

Run with `go test -fuzz=FuzzParse -fuzztime=30s`. Failing inputs land in `testdata/fuzz/FuzzParse/`.
Commit that directory: plain `go test` replays it as regression cases.

---

## Benchmarks

Since Go 1.24, `b.Loop()` replaces the `b.N` loop and excludes setup before it from the timing:

```go
func BenchmarkSort(b *testing.B) {
    for _, size := range []int{100, 1_000, 10_000} {
        b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
            data := randomInts(size)
            for b.Loop() {
                slices.Sort(slices.Clone(data))
            }
        })
    }
}
```

Run with `go test -bench=BenchmarkSort -benchmem`, and compare runs with `benchstat`, not by eye.

---

## Lint Enforcement

The mechanical rules are cheaper to enforce with golangci-lint (v2 config) than to remember:

```yaml
linters:
  enable:
    - thelper
    - tparallel
    - errorlint
    - usetesting
    - forbidigo
  settings:
    forbidigo:
      forbid:
        - pattern: ^reflect\.DeepEqual$
          msg: use cmp.Diff from github.com/google/go-cmp
      analyze-types: true
  exclusions:
    rules:
      - path-except: _test\.go
        linters:
          - forbidigo
```

- `thelper`: helpers call `t.Helper()` and take `t` as the first argument.
- `tparallel`: `t.Parallel()` is used consistently between a test and its subtests.
- `errorlint`: errors are compared with `errors.Is`/`errors.As`, not `==` or type assertions.
- `usetesting`: `t.TempDir`, `t.Setenv`, `t.Context` instead of `os` equivalents.
- `forbidigo`: bans `reflect.DeepEqual` in tests.

`paralleltest` is left out on purpose: it demands `t.Parallel()` everywhere, and parallelism here is
opt-in.

---

## Testing Commands

```bash
go test ./...                                      # run all tests
go test -run 'TestUser/Create' ./...               # run one subtest
go test -race ./...                                # with race detector
go test -cover -coverprofile=coverage.out ./...    # with coverage
go tool cover -func=coverage.out                   # per-function coverage
go test -short ./...                               # skip long-running tests
go test -count=10 ./...                            # repeat to surface flakiness
go test -bench=. -benchmem ./...                   # run benchmarks
go test -fuzz=FuzzParse -fuzztime=30s ./pkg        # run one fuzzer
```
