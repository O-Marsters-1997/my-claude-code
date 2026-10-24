# Testing app code

`swift-ios` builds apps as MV: views stay thin, and logic lives in `@Observable` models and in
services. So the unit-test surface is the models and services. Views are not unit-tested.

## Contents

- What to test where
- `@Observable` models
- Waiting for an observable change
- Services and their doubles
- SwiftData
- When XCUITest is right

## What to test where

| Code                                       | Test with                                         |
| ------------------------------------------ | ------------------------------------------------- |
| `@Observable` model: state, actions, derived values | Swift Testing, real model, faked services |
| Service (API client, repository, store)    | Swift Testing, against a double of what *it* talks to |
| SwiftData models, queries, relationships   | Swift Testing, in-memory `ModelContainer`          |
| Pure functions, formatters, parsers        | Swift Testing, parameterized                      |
| Flows across screens, navigation, system alerts | XCUITest (XCTest)                            |
| Timing and memory (`XCTMetric`)            | XCTest                                            |
| A SwiftUI `View`'s `body`                  | Nothing. Move the logic it holds into a model     |

When the logic you need to test lives in a view (`@State` plus closures in `body`), suggest
moving it into an `@Observable` model. Make that change only if the user agrees or the task
already includes it.

## `@Observable` models

An `@Observable` class is directly testable: create it, call its methods, read its properties.
It needs no protocol wrapper and no view.

```swift
@MainActor
struct CartModelTests {
    let model = CartModel(pricing: .fixed(rate: 0.2))

    @Test func addingItemUpdatesTotal() {
        model.add(.coffee, quantity: 2)
        #expect(model.items.count == 1)
        #expect(model.total == 7.20)
    }

    @Test func removingLastItemEmptiesCart() {
        model.add(.coffee)
        model.remove(.coffee)
        #expect(model.isEmpty)
    }
}
```

- If the model is `@MainActor` (or the module defaults to MainActor), the suite or test is too.
  A non-UI model doesn't need it.
- Assert on observable state, the thing the view renders, not on which service methods were
  called.
- Inject services through `init`, with production defaults, so tests can pass doubles.

## Waiting for an observable change

Prefer an `async` method you can `await` (`async-and-concurrency.md` § Code that spawns its own
`Task`). When the model starts its own work and you can't change it, wait for the property to
change:

```swift
@MainActor
@Test func searchPublishesResults() async {
    let model = SearchModel(index: .sample)

    await withCheckedContinuation { continuation in
        _ = withObservationTracking {
            model.results
        } onChange: {
            continuation.resume()
        }
        model.startSearch("swift")
    }

    #expect(model.results.isEmpty == false)
}
```

`onChange` fires once, *before* the new value is stored. Assert after the `await`, never inside
`onChange`. If the property changes more than once (cleared, then filled), wait for the final
value.

With Swift 6.2 and an iOS 26 / macOS 26 deployment target, `Observations` gives an async sequence
of values instead:

```swift
for await state in Observations({ model.state }) where state == .loaded { break }
```

Put a `.timeLimit` on such tests so a value that never arrives fails the test rather than hanging
it.

## Services and their doubles

Services hide `URLSession`, persistence, the clock and other system edges. Make the dependency
explicit, then pass a double in tests:

```swift
protocol HTTPClient: Sendable {
    func data(for request: URLRequest) async throws -> (Data, URLResponse)
}

extension URLSession: HTTPClient {}

struct ProfileService {
    var http: any HTTPClient = URLSession.shared
}
```

- Prefer a small, stateful fake that behaves like the real thing (stores what was saved, returns
  it, throws the same errors) to a mock that counts calls.
- To force an error path, use a one-off stub that throws, not failure switches on the fake.
- A closure-based client (`struct APIClient { var fetch: @Sendable () async throws -> [Item] }`)
  works too. Follow what the codebase already does.
- To test the real client's request building and decoding, stub at the `URLProtocol` level on an
  `.ephemeral` session configuration. The handler is usually a `static var`, which is shared
  state, so key responses by URL or put those tests in one `.serialized` suite.
- Never hit the live network in unit tests.

## SwiftData

Give each test its own in-memory container. One built in the suite's `init()` is per-test,
because each test gets a fresh suite instance.

```swift
@MainActor
struct LibraryTests {
    let container: ModelContainer
    var context: ModelContext { container.mainContext }

    init() throws {
        container = try ModelContainer(
            for: Book.self, Author.self,
            configurations: ModelConfiguration(isStoredInMemoryOnly: true)
        )
    }

    @Test func deletingAuthorCascadesToBooks() throws {
        let author = Author(name: "Le Guin")
        author.books = [Book(title: "The Dispossessed")]
        context.insert(author)
        try context.save()

        context.delete(author)
        try context.save()

        #expect(try context.fetchCount(FetchDescriptor<Book>()) == 0)
    }
}
```

- List every model type in the schema the test touches, including relationship targets.
- `mainContext` is main-actor bound, so the suite is `@MainActor`. For code using a
  `@ModelActor`, create the actor from the same container and `await` its methods.
- Test the queries the app runs: build the same `FetchDescriptor` with `#Predicate` and sort the
  app uses, ideally by calling the code that builds it.
- Never share a container through a `static` or a singleton. Never point tests at the on-disk
  default store.
- Models and `ModelContext` don't cross actors in tests any more than in the app. Pass
  `PersistentIdentifier`s.
- A model can wrap a `ModelContext` behind a repository service. Then test the model against a
  fake repository, and test the repository against an in-memory container.

## When XCUITest is right

Swift Testing doesn't drive UI. Use XCTest's `XCUIApplication` for:

- a handful of critical end-to-end flows (sign in, purchase, onboarding);
- navigation and deep links that cross screens;
- system integration: permission alerts, share sheets, widgets, App Intents;
- checking accessibility identifiers and labels are reachable.

Keep UI tests few: they are slow and the most brittle tests you have. Logic that a UI test
would cover belongs in a model test.

- Find elements by `accessibilityIdentifier`, not by visible text that changes with copy or
  locale.
- Wait with `element.waitForExistence(timeout:)`, never `sleep`.
- Launch with arguments or environment (`app.launchArguments = ["-ui-testing"]`) that the app
  reads to swap in stub services and an in-memory store.
- Put UI tests in a `*UITests` target. Don't import `Testing` there.
