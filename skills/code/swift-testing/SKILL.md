---
name: swift-testing
description: >
  How Swift tests are written in these repos: Swift Testing (`@Test`, `#expect`, `#require`),
  suites, traits and tags, parameterized tests, async and actor-isolated tests, confirmations,
  testing `@Observable` models and SwiftData, and migrating from XCTest. TRIGGER — load BEFORE
  writing, editing or reviewing any Swift test file: a path under a `*Tests/` or `*UITests/`
  directory, a file ending `Tests.swift`/`Test.swift`, or one that imports `Testing` or
  `XCTest`. Also when the task says `@Test`, `#expect`, XCTestCase, XCTest migration, test plan,
  flaky test or "write a test" in a Swift or Xcode project. Plain requests count: "add tests for
  this model", "fix the failing test", or a feature task whose last step is testing it. Load
  alongside swift-ios, not instead of it. SKIP when no Swift test code is written or judged.
license: MIT
metadata:
  sources: >
    Merged from twostraws/Swift-Testing-Agent-Skill (Paul Hudson, MIT), AvdLee/Swift-Testing-Agent-Skill
    (Antoine van der Lee, MIT), and the testing references of twostraws/Swift-Concurrency-Agent-Skill and
    AvdLee/Swift-Concurrency-Agent-Skill (MIT).
---

# Swift Testing

How a Swift test is written. What to test is the repo's `testing-policy` when it ships one. For
the red-green-refactor loop, use `tdd`; for the code under test, `swift-ios`.

Swift Testing ships new APIs several times a year, and most training data and blog posts are
XCTest. Your instincts reskin XCTest. The rules below exist to stop that. Trust the installed
toolchain over memory and over Apple's docs, which lag behind it.

## Before you write

1. Read the toolchain: `swift --version`, `xcodebuild -version`, or `swift-tools-version` in
   `Package.swift`. Gate anything in `references/new-features.md` on it.
2. Check the test target's default actor isolation (`SWIFT_DEFAULT_ACTOR_ISOLATION` in build
   settings, `.defaultIsolation(MainActor.self)` in `Package.swift`). It decides whether tests
   already run on the main actor.
3. See what the target already uses. New tests are Swift Testing. Existing XCTest stays as it
   is unless the task is a migration.

## Your defaults that are wrong here

| You reach for                                       | Do this instead                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------- |
| `class FooTests: XCTestCase`, `XCTAssert*`          | A `struct` with `@Test` methods, `#expect`                              |
| `@Suite` on every test type                         | Nothing. `@Suite` only to give a display name or attach traits          |
| `setUp()` / `tearDown()`                            | `init()` (async/throws allowed); `deinit` on a class; a scoping trait for async teardown |
| `func testUserLogsOut()`                            | `func userLogsOut()`; the `test` prefix means nothing                   |
| `#expect(!isLoggedIn)`                              | `#expect(isLoggedIn == false)`; `!` hides the value in the failure     |
| `users.first!`, `XCTUnwrap`                         | `let first = try #require(users.first)`                                 |
| `#expect(throws: Error.self)`                       | The specific error or case: `#expect(throws: LoginError.locked)`        |
| Copy-pasted tests that differ by one input, a `for` loop in a test | `@Test(arguments:)`                                       |
| `Task.sleep` to wait for work                       | `await` the work; `confirmation()`; return the `Task` and await it      |
| `XCTestExpectation` + `wait(for:)`                  | `await confirmation { confirm in … }`, or a continuation               |
| `.serialized` to fix a flaky test                   | Remove the shared state. `.serialized` only as a dated transition       |
| `.timeLimit(.seconds(10))`                          | `.timeLimit(.minutes(1))`; minutes is the only unit                     |
| `@available` on a suite                             | `@available` on each test function                                      |
| `@MainActor` on every test                          | Only where the code under test is main-actor isolated                   |
| Commenting a test out                               | `.disabled("reason", .bug(...))`, or `withKnownIssue` to keep it running |
| Unit-testing a SwiftUI view                         | Test the `@Observable` model or service behind it                       |

## Rules

1. XCTest stays only for UI automation (`XCUIApplication`), performance metrics (`XCTMetric`,
   `measure`) and Objective-C tests. Everything else new is Swift Testing.
2. `import Testing` only in test targets, never in app, library or executable targets.
3. Suites are `struct`s unless they need `deinit` or reference identity. Each test gets a fresh
   instance, so stored properties set in `init()` are per-test state, not shared.
4. `#expect` is the default. `try #require` when the rest of the test is meaningless if it fails:
   preconditions and optional unwraps. Never `!` inside either macro.
5. Throwing code: `#expect(throws: SpecificError.case)`, or capture the returned error and
   assert on it. `#expect(throws: Never.self)` for "does not throw".
6. Parameterize when tests share logic and differ in data. Pair inputs with expected outputs as
   an array of tuples, not `zip` of two arrays. Two argument collections form a Cartesian
   product. Expected values are literals, never derived from the input.
7. Every test runs in parallel and in random order. Shared mutable state (`static var`,
   singletons, `UserDefaults.standard`, the file system) is the bug. Inject it per test.
8. Async: make the test `async` and `await` the real work. Code that spawns a `Task` it never
   returns can't be tested reliably: make it `async` or return the `Task`. Wrap callback APIs in
   `withCheckedContinuation`. No sleeps as synchronisation.
9. Test behaviour through the public or `@testable` internal API. Don't add `nonisolated`
   accessors or test-only hooks to production code to reach into an actor.
10. Unit tests never touch the network, the real clock, or persistent storage. Use a protocol
    double for `URLSession`, an injected clock and an in-memory `ModelContainer`.
11. New dependencies (Swift Numerics, swift-concurrency-extras, snapshot libraries) need the
    user's permission first.
12. Mirror the production folder layout in the test target for new tests. Don't reorganise
    existing tests without asking.

## Reviews

When asked to review tests, list findings by file. For each: line, the rule broken, a short
before/after. Skip clean files, and end with a prioritised summary. Report real problems only.
When asked to write or fix tests, make the change; don't produce a report.

## Where to look

| If you're…                                                          | Load                                     |
| ------------------------------------------------------------------- | ---------------------------------------- |
| Structuring suites, `init`/`deinit`, nested suites, `@available`     | `references/fundamentals.md`             |
| Choosing `#expect`/`#require`, testing throws, `Issue.record`, `withKnownIssue` | `references/expectations.md`  |
| Turning repeated tests into `arguments:`, pairing inputs             | `references/parameterized.md`            |
| Disabling, gating, tagging, linking bugs, test plans                 | `references/traits-and-tags.md`          |
| Testing async code, callbacks, actors, `confirmation()`, cancellation | `references/async-and-concurrency.md`   |
| Fixing a flaky or order-dependent test, deciding on `.serialized`    | `references/isolation-and-parallelism.md` |
| Testing an `@Observable` model, a service, SwiftData, or picking XCUITest | `references/testing-app-code.md`    |
| Converting XCTest                                                    | `references/migrating-from-xctest.md`    |
| Using raw identifiers, exit tests, attachments, scoping traits, range confirmations | `references/new-features.md` |

## See also

| Skill            | When                                                    |
| ---------------- | ------------------------------------------------------- |
| `swift-ios`      | The code under test: SwiftUI, models, concurrency       |
| `tdd`            | Writing the test first, one behaviour at a time         |
| `testing-policy` | What to test in this repo (per repo, when present)      |
