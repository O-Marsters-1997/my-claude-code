# Async and concurrency

## Contents

- Async tests
- Actors
- Actor isolation of tests
- Confirmations
- Code that spawns its own `Task`
- Callback APIs
- Streams and notifications
- Intermediate state (`isLoading`)
- Cancellation
- Deallocation
- No sleeps
- Data races
- XCTest async (legacy)

## Async tests

Mark the test `async` and `await` the work. No `Task {}` wrapper, no semaphores, no expectations.

```swift
@Test func userLoads() async throws {
    let user = try await UserService(client: .stub).load(id: "123")
    #expect(user.name == "Alice")
}
```

Keep synchronous logic in synchronous tests. Making a test `async` for no reason adds nothing.

## Actors

Read actor state with `await`, as production code does. Don't add `nonisolated` accessors or
test-only hooks to reach inside.

```swift
@Test func cachingReturnsSameImage() async throws {
    let cache = ImageCache(loader: .stub)
    let first = try await cache.image(for: url)
    let second = try await cache.image(for: url)
    #expect(first == second)
    #expect(await cache.loadCount == 1)
}
```

## Actor isolation of tests

Swift Testing runs tests on arbitrary executors. XCTest ran synchronous tests on the main
thread, so migrated tests can expose isolation bugs they used to hide.

- Put `@MainActor` on a test or suite when the code under test is main-actor isolated, such as a
  UI-state `@Observable` model. Not otherwise: it serialises work and hides concurrency bugs.
- `confirmation(isolation:)` and `withKnownIssue(isolation:)` run only their closure on an actor.
- The test target may have default MainActor isolation. Check build settings before adding or
  removing annotations.

## Confirmations

`confirmation()` checks that an event happens a set number of times. It replaces
`XCTestExpectation`.

```swift
@Test func saveNotifiesObserverOnce() async throws {
    let store = Store()
    try await confirmation { saved in
        store.onSave = { saved() }
        try await store.save(item)
    }
}
```

- The confirmed work must finish before the closure returns. `confirmation` does not wait.
  Await the work inside the closure.
- `expectedCount: 3` for an exact count, `expectedCount: 0` for "never happens".
- Ranges (Swift 6.1+): `expectedCount: 5...10`, `expectedCount: 1...`. `...10` is rejected.

## Code that spawns its own `Task`

```swift
func refresh() {
    Task { items = try await api.fetch() }
}
```

A test can't know when that finishes. A sleep makes it flaky. Fix the API:

1. Make it `async`: `func refresh() async throws`. The view calls it from `.task {}`.
2. If it must stay synchronous, return the task: `@discardableResult func refresh() -> Task<Void,
   Never>`. The test does `await model.refresh().value`.

Changing production code is a suggestion to the user unless the task already covers it. If you
can't change it and the type is `@Observable`, wait on observation instead (see
`testing-app-code.md` § Waiting for an observable change).

## Callback APIs

Don't modernise callback-based production code without asking. Bridge it in the test, and wait
for the callback before asserting.

```swift
@Test func loadReadings() async throws {
    let readings = try await withCheckedThrowingContinuation { continuation in
        loader.loadReadings { result in continuation.resume(with: result) }
    }
    #expect(readings.count >= 10)
}
```

Resume exactly once. A callback that can fire twice crashes the continuation, so use
`confirmation(expectedCount:)` for multi-fire callbacks.

## Streams and notifications

Start listening before you trigger the event, or the event can arrive first. An `AsyncStream`
buffers from creation, so get the iterator first:

```swift
@Test func publishesProgress() async throws {
    let uploader = Uploader(client: .stub)
    var events = uploader.progress.makeAsyncIterator()
    async let upload: Void = uploader.upload(file)
    #expect(await events.next() == .started)
    try await upload
}
```

A `Task` plus `Task.yield()` before posting a notification only makes the race less likely; it
doesn't remove it. Prefer an injected stream or callback you control.

## Intermediate state (`isLoading`)

Asserting `isLoading == true` straight after starting a task is a race. Make the dependency
suspend until the test lets it go:

```swift
@MainActor
@Test func showsLoadingWhileFetching() async throws {
    let (gate, release) = AsyncStream<Void>.makeStream()
    let model = FeedModel(api: .init(fetch: {
        for await _ in gate { break }
        return [.sample]
    }))

    let task = Task { await model.load() }
    await Task.yield()
    #expect(model.isLoading)

    release.yield()
    await task.value
    #expect(model.isLoading == false)
    #expect(model.items == [.sample])
}
```

`await Task.yield()` on the main actor lets `load()` reach its suspension point. If one yield
isn't enough, wait for the state change via observation instead of adding more yields.

If the project already depends on Point-Free's `swift-concurrency-extras`,
`withMainSerialExecutor` makes scheduling deterministic. It doesn't work with parallel tests, so
that suite needs `.serialized`. Don't add the package without asking.

## Cancellation

Test that the *production* code checks for cancellation. Cancel while it is suspended on
something the test controls, then expect `CancellationError`:

```swift
@Test func importStopsWhenCancelled() async throws {
    let (gate, _) = AsyncStream<Void>.makeStream()
    let importer = Importer(source: .init(next: {
        for await _ in gate {}
        try Task.checkCancellation()
        return nil
    }))

    let task = Task { try await importer.run() }
    task.cancel()

    await #expect(throws: CancellationError.self) {
        try await task.value
    }
}
```

## Deallocation

```swift
@Test func modelIsReleasedAfterStop() async {
    var model: SessionModel? = SessionModel()
    weak var weakModel = model
    await model?.stop()
    model = nil
    #expect(weakModel == nil)
}
```

If a long-running task keeps the object alive, that is the leak. Fix it, don't add a sleep.

## No sleeps

`Task.sleep`, `Thread.sleep` or fixed delays as synchronisation pass on a fast laptop and fail on
CI. Await the work, use a confirmation, or control the dependency. For code that itself waits
(debounce, retry backoff), inject a clock and advance it. Don't wait on real time.

## Data races

Turn on Thread Sanitizer in the test scheme (Edit Scheme → Test → Diagnostics), or run
`swift test --sanitize=thread`. It's slow, so a dedicated CI job is fine.

## XCTest async (legacy)

When you have to stay in XCTest (UI tests), use `await fulfillment(of: [exp], timeout: 5)`.
`wait(for:)` in an async test deadlocks. `setUp() async throws` and `tearDown() async throws`
are available.
