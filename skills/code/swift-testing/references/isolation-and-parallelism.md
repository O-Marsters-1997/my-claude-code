# Isolation and parallelism

## The model

Swift Testing runs tests in parallel and in random order, sync and async alike. That is a
feature: it exposes tests that depend on each other. A test that fails only sometimes, or only
in the full run, almost always shares state with another test.

```swift
enum SharedStore { static var counter = 0 }

@Test func incrementsCounter() {
    SharedStore.counter += 1
    #expect(SharedStore.counter >= 1)
}

@Test func expectsFreshCounter() {
    #expect(SharedStore.counter == 0)   // flaky: depends on what ran first
}
```

## Fix the state, not the scheduling

Find what the tests share and give each test its own copy:

| Shared thing                       | Per-test replacement                                                 |
| ---------------------------------- | -------------------------------------------------------------------- |
| `static var`, a singleton          | Inject the dependency; build a fresh one in the suite's `init()`      |
| `UserDefaults.standard`            | `UserDefaults(suiteName: "test-\(UUID())")`, removed at the end      |
| Files at a fixed path              | A directory under `FileManager.default.temporaryDirectory` named by a UUID |
| A SwiftData/Core Data store on disk | `ModelConfiguration(isStoredInMemoryOnly: true)` per test (`testing-app-code.md`) |
| `URLSession.shared`, real network  | A protocol double (`testing-app-code.md` § Services)                 |
| `Date()`, `Date.now`, randomness   | An injected clock or generator                                       |
| `@TaskLocal` configuration         | A scoping trait (`new-features.md`), which is parallel-safe          |

```swift
@Test func remembersLastUser() {
    let suite = "test-\(UUID().uuidString)"
    let defaults = UserDefaults(suiteName: suite)!
    defer { defaults.removePersistentDomain(forName: suite) }

    let settings = Settings(defaults: defaults)
    settings.lastUser = "taylor"
    #expect(settings.lastUser == "taylor")
}
```

Injecting with a default (`init(defaults: UserDefaults = .standard)`) keeps production call
sites unchanged.

## `.serialized`

What it does, exactly:

- On a parameterized test: runs that test's cases one at a time.
- On a non-parameterized test: **nothing**. Agents believe otherwise.
- On a suite: runs the tests and sub-suites inside it one at a time. Other suites still run in
  parallel alongside it.

```swift
@Suite(.serialized)
struct LegacyKeychainTests { … }
```

Use it only as a transition: while migrating a serial XCTest suite, or for a real shared
resource you can't isolate yet (the keychain, a hardware device). Leave a `.bug` or a dated
reason so it gets removed. It doesn't protect against tests in *other* suites that touch the same
state. That is one more reason to remove the shared state instead.

## Shared resources in integration tests

If tests hit a real database, file tree or service, choose one:

1. Isolate the backing state per test (a fresh store, a unique key prefix).
2. Use an in-memory substitute on the fast path and keep a few real integration tests.
3. Move the real-resource tests into their own tag and test plan, run serially.

## Don't

- Don't make everything `@MainActor` or `.serialized` to silence flakiness. It hides the bug.
- Don't reset a singleton at the start of each test. A parallel test can see it half-reset.
- Don't rely on execution order, ever.
