# Example: a Go service's policy

A finished policy from a modular Go monolith (bounded-context services over Postgres, chi routes,
HTML and ATS scrapers, RabbitMQ consumers). Calibrate depth and tone against it; don't copy its
rows into a repo that lacks those layers. It predates the core rule and the Guards column.

````markdown
## Before writing a test

1. Name the row of the map below that the test fills. If no row fits, stop and ask.
2. Check the never-test list. If the code is on it, test it through its caller instead.
3. The test goes through an exported surface. Needing an unexported function is a design
   smell: test through the caller, or move the logic behind a surface that deserves a test.

A **unit** is the behaviour a package's exported surface offers, not a function or a file.

## The map

| Code | File | Size | Doubles | Asserts |
|---|---|---|---|---|
| Service | `service_test.go` | small | `<ctx>test.NewFakeStore()`; `<ctx>test` stubs for other contexts | returned `dto` + `apperr` kind |
| Store | `store/store_test.go` | medium (pgtest) | real Postgres; runs `<ctx>test.RunStoreContract` | only what the DB proves (below) |
| Store fake | none: covered by the contract suite | | | |
| Module facade | `facade_test.go` in `<ctx>_test` | small | `<ctx>.Build(Deps{…fakes})` | promises to callers, re-exported sentinels |
| Tx-scoped port (`JobsChanged(ctx, tx, …)`) | `store/store_test.go` | medium | real tx | effect lands on commit, vanishes on rollback |
| Module routes | `routes_test.go` in `<ctx>_test` | small | real `Routes` + middleware over `Build(fakes)` | `handlerstest` checks + one happy path per route |
| Hand-built `Handle` route (OAuth, cookies, export, ingest) | `routes_test.go` | small | same | its own decode/respond logic |
| `internal/handlers` generics | `generic_test.go` | small | none | apperr → status mapping, decode rules — once, for every route |
| HTML source adapter | `<source>_test.go` | small | `snapshots/*.html` | normalised golden JSON; `FuzzParse` never panics |
| ATS source adapter | `<source>_test.go` | small | `snapshots/*.json`/`.xml` | parsed fields |
| Queue consumer | handler func test | small | `httptest.NewServer`, stubs | ack / requeue / dead-letter, and what it sent |
| Cross-context journey | `internal/api/router_test.go` | component | real everything | one happy path per journey |

**Never tested directly:** unexported helpers, generated `store/sqlc/**`, `module.go` wiring,
CRUD-generic routes beyond the `handlerstest` checks, `cmd/*/main.go`, `dto` structs, anything
whose only behaviour is calling something else.

## Doubles

Fidelity order: real > fake > stub. No generated mocks, no testify.

- **Store the service owns → stateful fake.** A map plus the sentinels the real store returns,
  exported from `internal/services/<ctx>/<ctx>test`. It must pass `RunStoreContract`, the same
  suite the real store runs in `store_test.go`:

  ```go
  // internal/services/applications/applicationstest
  func RunStoreContract(t *testing.T, newStore func(t *testing.T) applications.Store) {
      t.Run("get missing returns ErrNotFound", func(t *testing.T) { … })
  }

  // store/store_test.go
  applicationstest.RunStoreContract(t, func(t *testing.T) applications.Store {
      return store.New(pgtest.New(t))
  })
  ```

  When a service test needs the fake to do something the real store can't (an error path),
  wrap the fake in a one-method override; don't add knobs to the fake.
- **Another context's facade → behaviour stub.** Named by what it does:
  `scoringtest.ScoresAs(80)`, `scoringtest.Fails(err)`. Prime the value, then assert what the
  code under test does with it. Never assert that a query was called.
- **Unmanaged edge (job boards, OAuth, email, LLM, outgoing messages)** → `httptest.NewServer`
  serving recorded fixtures, or a recording stub.

**Interaction assertions** (call counts, arguments) only for a command or billed call across an
unmanaged boundary: published messages, sent email, LLM calls and their questions, retries
against external HTTP. Asserting that one of our own packages was called is a change-detector:
rewrite it at the output or delete it.

## Store tests

- Beyond the contract suite, cover only what a fake can't: constraints and uniqueness,
  `ON CONFLICT`, joins onto other contexts' tables, rollback of tx-scoped ports, concurrent
  writes (N goroutines, exactly one wins). Sabotage once: remove the rollback or constraint and
  watch the test fail.

## Queue consumers

Call the consumer's handling function with a decoded message; don't start the broker. Every
consumer covers: a poison message (dead-lettered, not requeued forever), a transient failure
(requeued), and a redelivery of an already-processed message (no duplicate effect). Broker
delivery itself is tested once.

## Naming

- One `Test<Method>` per exported method; scenarios are `t.Run` subtests named by the behaviour
  in plain words: `"duplicate URL returns conflict"`, `"defaults status to saved"`.
- The file name says the row: `service_test.go`, `store_test.go`, `facade_test.go`,
  `routes_test.go`. A reader knows the size and doubles from the file name alone.
- When a component test catches a bug, add the lower-level test that should have caught it.
````
