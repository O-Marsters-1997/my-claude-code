# Migrating from XCTest

## Whether to migrate

- Don't rewrite XCTest to Swift Testing unless asked. New tests in a mixed target are Swift
  Testing.
- Keep XCTest for UI automation (`XCUIApplication`), performance (`XCTMetric`, `measure`) and
  Objective-C tests.
- Both frameworks coexist in one target, and in one file during migration (`import XCTest` and
  `import Testing`). Don't mix their assertions in one test: depending on the toolchain, an
  `XCTAssert` inside an `@Test` may not fail it.

## Order

Migrate incrementally, one file or suite per reviewable change:

1. **Assertions and structure, mechanically.** Keep type and method names. `final class
   FooTests: XCTestCase` becomes `struct FooTests`, `func testBar()` becomes `@Test func bar()`,
   `XCTAssert*` becomes `#expect`/`#require`, `setUp` becomes `init`.
2. **Suites.** Nest or split types where it helps. Add `@Suite` only for a name or traits.
3. **Parameterize** methods that differ only in input.
4. **`#require`** for preconditions at the top of tests.
5. **Traits** in place of XCTest conventions: `.disabled` / `.enabled(if:)` instead of
   `XCTSkip`, `.timeLimit`, `.tags`, `.bug`.

## Mapping

| XCTest                                  | Swift Testing                                         |
| --------------------------------------- | ----------------------------------------------------- |
| `XCTAssertEqual(a, b)`                  | `#expect(a == b)`                                     |
| `XCTAssertNotEqual(a, b)`               | `#expect(a != b)`                                     |
| `XCTAssertTrue(x)`                      | `#expect(x)`                                          |
| `XCTAssertFalse(x)`                     | `#expect(x == false)`; not `#expect(!x)`              |
| `XCTAssertNil(x)` / `XCTAssertNotNil(x)` | `#expect(x == nil)` / `#expect(x != nil)`            |
| `XCTAssertLessThan(a, b)` etc.          | `#expect(a < b)`                                      |
| `XCTAssertIdentical(a, b)`              | `#expect(a === b)`                                    |
| `try XCTUnwrap(x)`                      | `try #require(x)`                                     |
| `XCTAssertThrowsError(try f())`         | `#expect(throws: SpecificError.self) { try f() }`     |
| `XCTAssertNoThrow(try f())`             | `#expect(throws: Never.self) { try f() }`             |
| `XCTFail("msg")`                        | `Issue.record("msg")`                                 |
| `XCTAssertEqual(a, b, accuracy: e)`     | `#expect(abs(a - b) < e)`, or Swift Numerics if already a dependency |
| `continueAfterFailure = false`          | `#require` on the lines that must stop the test       |
| `throw XCTSkip("…")`                    | `.enabled(if:)` / `.disabled("…")` trait              |
| `XCTExpectFailure`                      | `withKnownIssue`                                      |
| `setUp()` / `setUpWithError()`          | `init()` / `init() throws` / `init() async throws`    |
| `tearDown()`                            | `deinit` on a `final class` suite, or a scoping trait for async |
| `expectation` + `fulfill` + `wait`      | `await` directly, or `confirmation()`                 |
| `addTeardownBlock`                      | `defer` in the test, or a scoping trait               |
| `XCTAttachment`                         | `Attachment.record` (Swift 6.2+, see `new-features.md`) |

## Before and after

```swift
// Before
final class PriceTests: XCTestCase {
    var calculator: PriceCalculator!

    override func setUp() {
        calculator = PriceCalculator(taxRate: 0.2)
    }

    func testDiscountedTotal() throws {
        let order = try XCTUnwrap(Order.sample)
        XCTAssertEqual(calculator.total(order, discount: 5), 15)
    }
}

// After
struct PriceTests {
    let calculator = PriceCalculator(taxRate: 0.2)

    @Test func discountedTotal() throws {
        let order = try #require(Order.sample)
        #expect(calculator.total(order, discount: 5) == 15)
    }
}
```

## Behaviour that changes

- **Parallelism.** XCTest ran a class's methods serially; Swift Testing runs everything in
  parallel and in random order. Migrated tests that share state start failing. Fix the state
  (`isolation-and-parallelism.md`). `.serialized` on the suite is acceptable as a dated stopgap.
- **Main thread.** XCTest ran synchronous tests on the main thread. Swift Testing doesn't. Add
  `@MainActor` only where the code under test needs it, not to every migrated test.
- **Instances.** Both create a fresh instance per test, so `init` setup keeps XCTest's
  per-test semantics.
