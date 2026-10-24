# Fundamentals

## Building blocks

- `@Test` marks a test. It can be a global function or a method on a type.
- Any type containing `@Test` methods is a suite. `@Suite` is not needed to make it one.
- Add `@Suite` only to give the suite a display name or attach traits:
  `@Suite("Checkout", .tags(.payments))`. Agents add a bare `@Suite` everywhere; don't.
- Suites can be `struct`, `class`, `actor` or `enum`. Prefer `struct`: value semantics, and no
  temptation to share state.
- No `test` prefix. `userCanLogOut()`, not `testUserCanLogOut()`.
- A test that reaches no `#expect`/`#require` and doesn't throw passes.

```swift
import Testing
@testable import FoodTruck

struct MenuTests {
    @Test("Default menu has no duplicates")
    func uniqueItems() {
        let items = Menu.default.items
        #expect(Set(items).count == items.count)
    }
}
```

## Setup and teardown

Swift Testing makes a fresh suite instance for every test, so `init()` is per-test setup.

```swift
struct PlayerTests {
    let sut: Player

    init() {
        sut = Player(name: "Natsuki Subaru")
    }

    @Test func nameIsCorrect() {
        #expect(sut.name == "Natsuki Subaru")
    }
}
```

- `init` can be `async`, `throws`, or both.
- Teardown goes in `deinit`, so use a `final class` or `actor` suite when you need it. `deinit` is
  synchronous. For async teardown use a scoping trait (`new-features.md` § Test scoping traits).
- A suite with instance tests needs an initializer callable with zero arguments. Give stored
  properties defaults, or write `init()`, or default every parameter of a custom `init`.
- If you can't meet that, make the tests `static` or global functions.

## Nested suites

Nest types to group by feature. Group by behaviour, not just by the class under test.

```swift
struct CheckoutTests {
    struct Taxes {
        @Test func roundsToTwoDigits() {
            #expect(Checkout.total(subtotal: 10.00, taxRate: 0.0825) == 10.83)
        }
    }

    struct Discounts {
        @Test func promoCodeAppliesFixedAmount() {
            #expect(Checkout.total(subtotal: 20.00, discount: .fixed(5)) == 15.00)
        }
    }
}
```

Traits on a suite (tags, time limits, `.serialized`) apply to everything inside it.

## Availability

`@available` works on test functions but not on suites. If every test in a suite needs iOS 26,
annotate each test.

```swift
struct PushTests {
    @available(iOS 26, *)
    @Test func supportsNewPushFormat() { … }
}
```

Prefer `@available` on the test to an `if #available` inside the body, so the report shows the
test as skipped rather than passed.

## Layout

- New tests mirror production: `Extensions/URLSession+Decodable.swift` is tested by
  `Extensions/URLSession+DecodableTests.swift` in the test target.
- Fixtures go in a `Fixtures` folder, or several beside the tests that use them.
- Don't reorganise existing tests to match without asking.

## What a good test looks like

Fast, isolated, repeatable, self-verifying. For each function, cover the happy path, the
boundaries and invalid input, plus concurrency where it applies. One behaviour per test;
several `#expect`s are fine when they check that one behaviour.
