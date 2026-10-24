# Swift concurrency

Read before writing, reviewing or fixing any `async`/`await`, actor, `Task`, `Sendable` or isolation code, or any strict-concurrency compiler error.

Contents

1. Before you change anything
2. Ground rules
3. Isolation and default actor isolation
4. `nonisolated(nonsending)` and `@concurrent`
5. Actors and reentrancy
6. Global and static state
7. Structured concurrency
8. Unstructured tasks and entry isolation
9. Cancellation
10. Capture and retain cycles
11. Async streams
12. Bridging callbacks, delegates and Combine
13. Replacing GCD and locks
14. Sendable
15. Diagnostics: error, cause, fix
16. Bug patterns and review hotspots
17. Newer APIs
18. Swift 6 migration summary

## 1. Before you change anything

Concurrency behaviour depends on per-module build settings. Read them first, for every target the code spans; neighbouring modules can differ.

- [ ] Toolchain and language mode: `swift-tools-version` and `swiftLanguageModes` in Package.swift, or `SWIFT_VERSION` in the Xcode project. Swift 5 mode turns most data-race errors into warnings.
- [ ] Default actor isolation: `SWIFT_DEFAULT_ACTOR_ISOLATION` in the Xcode project, or `.defaultIsolation(MainActor.self)` in a target's `swiftSettings`.
- [ ] Approachable concurrency / upcoming features: the Approachable Concurrency build setting, or `.enableUpcomingFeature("NonisolatedNonsendingByDefault")` (and `InferIsolatedConformances`, `InferSendableFromCaptures`, `DisableOutwardActorInference`, `GlobalActorIsolatedTypesUsability`).
- [ ] Strict concurrency level: `SWIFT_STRICT_CONCURRENCY` (minimal / targeted / complete), or `StrictConcurrency=` in `swiftSettings`.
- [ ] Deployment target, which gates newer runtime APIs (`Mutex`, `Task.immediate`, task naming).
- [ ] The exact compiler diagnostic and offending symbol, and whether the code is UI-bound or meant to run off the main actor.

Do not assume behaviour from a different module or from memory of an older Swift. Current version facts live in `platform-facts.md`.

## 2. Ground rules

- Use Swift concurrency, not GCD, in new code. Keep GCD only in low-level library code, framework interop, or synchronous hot paths where a queue or lock is genuinely the right tool.
- When an API has both an `async` and a completion-handler form, call the `async` one.
- Use `Task.sleep(for: .seconds(1))`, never `Task.sleep(nanoseconds:)`.
- Prefer structured concurrency (`async let`, task groups) over `Task {}`.
- Never add `@MainActor`, `@unchecked Sendable`, `nonisolated(unsafe)` or `@preconcurrency` as a blanket fix to silence an error. Name the isolation boundary being crossed and fix that.
- Make the smallest behaviour-preserving fix. Do not refactor surrounding code while fixing a diagnostic.
- Do not add third-party concurrency libraries without asking.
- Report only real problems in review; do not invent issues.
- Testing async code: see the swift-testing skill, once installed.

## 3. Isolation and default actor isolation

- In a MainActor-default module, unannotated declarations are already `@MainActor`. Do not add the annotation; a missing `@MainActor` is not a bug there.
- Otherwise put `@MainActor` on UI-state types (view models, `@Observable` models a view reads, UIKit controllers). Leave non-UI types (networking clients, parsers, persistence services) without it.
- MainActor-default does not move networking onto the main thread. Suspending I/O in another module runs where it always did; the caller just waits without blocking.
- `@MainActor` is inferred, so do not repeat it on:
  - subclasses of a `@MainActor` class;
  - extensions of a `@MainActor` type;
  - types conforming to a `@MainActor` protocol (SwiftUI `View` included); the inference covers the whole type;
  - storage behind an actor-isolated property wrapper.
- It is not inferred for closures passed to non-isolated functions unless the parameter is typed `@MainActor`.
- Some SwiftUI APIs run closures off the main actor (`Shape.path`, `Layout`, `visualEffect`, `onGeometryChange`) and require `Sendable` closures. Capture value copies there, not `self`.
- `MainActor.run {}` inside code that is already main-actor isolated is a no-op. If a function keeps hopping to main, make the function `@MainActor` instead.
- Use `MainActor.assumeIsolated {}` only in a callback that truly runs on main and the compiler cannot see it. A wrong assumption traps at runtime.
- `MainActor.assertIsolated()` checks isolation in debug builds only.
- Use an `isolated` parameter (`func f(on actor: isolated SomeActor)`) to run a function on whichever actor the caller passes.

## 4. `nonisolated(nonsending)` and `@concurrent`

Needs Swift 6.2+. With `NonisolatedNonsendingByDefault` on (part of approachable concurrency), a plain `nonisolated` async function runs on the caller's actor instead of hopping to the global pool. Without it, the old behaviour applies: it hops off.

- Plain `async` does not mean background. A CPU-heavy `async` helper called from the main actor blocks the main actor in this mode.
- Mark CPU-bound async work `@concurrent` to run it on the concurrent pool: parsing large payloads, image processing, compression, big transforms.
- Do not mark ordinary async I/O `@concurrent`; it already suspends.
- `@concurrent` needs a non-isolated context. Put it on a `nonisolated` type's method, a `static` method, or a free function:

```swift
nonisolated struct PhotoProcessor {
    @concurrent
    func process(_ data: Data) async -> Photo? { ... }
}

model.photo = await PhotoProcessor().process(data)
```

- Arguments to and results from a `@concurrent` function cross an isolation boundary, so they must be `Sendable` (or `sending`).
- Write `nonisolated(nonsending)` explicitly when a function must stay on the caller's actor regardless of the module's feature flags, typically to avoid sending a non-Sendable argument.
- When a function is `nonisolated` but the author expects it to run in the background, the missing piece is usually `@concurrent`.

## 5. Actors and reentrancy

- After every `await` inside an actor, assume all actor state may have changed. Never check, await, then act on the old check.
- Never force-unwrap actor state after an `await`.

```swift
// Before: two callers both download; `!` can crash
func video(for url: URL) async throws -> Video {
    if items[url] == nil { items[url] = try await download(url) }
    return items[url]!
}

// After
func video(for url: URL) async throws -> Video {
    if let cached = items[url] { return cached }
    let video = try await download(url)
    items[url] = video
    return video
}
```

- To deduplicate concurrent requests, store in-flight `Task` handles in a dictionary, await an existing one if present, and clear the entry on both success and failure.
- An actor adds a serial boundary: callers must `await`, and everything crossing must be `Sendable`. Do not introduce one where a value type, a `@MainActor` type or a `Mutex` would do. Flag actors that own little mutable state or mostly forward calls.
- Keep work inside actor methods short; long synchronous work on an actor blocks every other caller.
- A `deinit` on an actor-isolated class is not isolated by default and cannot touch isolated state. Use `isolated deinit` (Swift 6.2+) when teardown must touch it.

## 6. Global and static state

Every global or static `var`, and every static `let` of a non-Sendable type, needs an isolation story the compiler can check:

1. Make it an immutable `let` of a `Sendable` type. Best option when possible.
2. Isolate it to `@MainActor` when it belongs to UI code (`@MainActor static let shared = Library()`). Already implicit in MainActor-default modules.
3. Move mutable shared state into an actor, or behind a `Mutex`.
4. `nonisolated(unsafe)` only for genuinely immutable state the compiler cannot prove safe, such as C interop constants. Never for mutable state.

## 7. Structured concurrency

- `async let`: fixed number of independent operations, possibly different result types.
- `withTaskGroup` / `withThrowingTaskGroup`: dynamic number of operations of one type.
- `withDiscardingTaskGroup` / `withThrowingDiscardingTaskGroup`: child tasks with no result to collect; avoids accumulating results.
- Never spawn `Task {}` in a loop. Use a task group so cancellation propagates and errors surface.

```swift
let results = try await withThrowingTaskGroup { group in
    for url in urls { group.addTask { try await fetch(url) } }
    var collected = [Data]()
    for try await data in group { collected.append(data) }
    return collected
}
```

- Groups start every child eagerly. To bound concurrency, add N children up front, then add one more each time a result arrives.
- When one child throws, the group cancels the rest. To keep partial results, catch inside each child and return a `Result`.
- Spell out `of:` when the child type is not trivially inferable (tuples, `Result`).

## 8. Unstructured tasks and entry isolation

- `Task {}` inherits the caller's actor isolation and priority. From a `@MainActor` context, or anywhere in a MainActor-default module, a bare `Task {}` starts on the main actor.
- A new `Task {}` is enqueued, not run inline: the caller continues first and the task body runs later.
- Choose entry isolation from the synchronous prefix (everything before the first `await`):
  - prefix touches main-actor state: keep the inherited main-actor start;
  - prefix does nothing main-bound (empty, or only a trivial line like `print`): start with `Task { @concurrent in ... }` and hop back with `await MainActor.run { ... }` only for the UI mutation.

```swift
// Prefix needs the main actor: keep inheritance
Task {
    isLoading = true
    await load()
}

// Prefix needs nothing on main: start off it
Task { @concurrent in
    let data = await service.fetch()
    await MainActor.run { self.apply(data) }
}
```

  A `@concurrent` task closure changes isolation, so any non-Sendable state from the enclosing actor may only be touched inside the `MainActor.run` hop.
- `Task.detached` drops isolation, priority and task-locals. It is rarely right; reach for `@concurrent` or a task group first.
- `Task {}` is a smell when:
  - created in SwiftUI `onAppear()`: use `.task()`, which cancels on disappear; use `.task(id:)` to restart on a value change;
  - the enclosing function could itself be `async`;
  - its thrown error is dropped. Handle errors inside the closure.
- Store any long-lived task handle, cancel the previous one before replacing it, and cancel it in `deinit`.

## 9. Cancellation

Cancellation is cooperative: `cancel()` only sets a flag.

- Cancelling a parent cancels its structured children and group members. `Task {}` and `Task.detached` must be cancelled through their stored handle.
- In loops and long CPU-bound work, call `try Task.checkCancellation()` (throwing contexts) or check `Task.isCancelled` (non-throwing, or when cleanup is needed) at points where exiting is safe. A loop with no `await` never sees cancellation otherwise.
- Treat `CancellationError` as a normal lifecycle event, not a failure. Catch it first and do nothing; no alert, no retry.

```swift
do { try await load() }
catch is CancellationError { }
catch { errorMessage = error.localizedDescription }
```

- Bridge to APIs that have their own cancel handle with `withTaskCancellationHandler`. `onCancel` runs immediately, possibly on another thread, while the body is suspended, so it must be thread-safe.
- `URLSession` async methods already respond to cancellation; do not wrap them.

## 10. Capture and retain cycles

- A `Task` closure captures like any escaping closure and keeps `self` alive until the task ends.
- Strong `self` is fine for short, finite tasks.
- For long-running or infinite tasks (polling loops, `for await` over an endless stream), stored on `self`: capture `[weak self]`, rebind it per iteration inside the loop, and drop the strong reference before the next suspension. Also cancel the stored task in `deinit`.
- `[weak self]` alone does not end a `for await` over a stream that never finishes; the task keeps waiting. Cancel the task or finish the stream.

## 11. Async streams

- Create streams with `AsyncStream.makeStream(of:bufferingPolicy:)` / `AsyncThrowingStream.makeStream(of:throwing:)`, not the closure initializer that smuggles the continuation out.
- Finish the continuation exactly once on every path, or the consumer's `for await` hangs.
- Set `continuation.onTermination` to stop the underlying producer when the consumer cancels or the stream ends.
- The default buffer is unbounded. For fast producers set `.bufferingNewest(n)` or `.bufferingOldest(n)`.
- `for await` exits on cancellation or when the stream finishes; no manual check inside the loop. Code after the loop still runs, so put cleanup there.
- An `AsyncStream` has one consumer. For several observers, broadcast through an `@Observable` model instead.

## 12. Bridging callbacks, delegates and Combine

- Leave existing, working completion-handler APIs in place unless asked to modernise; add an `async` wrapper beside them.
- Use the SDK's own async overload when one exists instead of wrapping.
- Wrap single-shot callbacks with `withCheckedContinuation` / `withCheckedThrowingContinuation`. Resume exactly once on every path: zero resumes hang the caller forever, two trap. Audit early returns, thrown errors, and callbacks that may never fire (deallocated owner, silent timeout).
- Keep the checked variants in production. Use `withUnsafe…Continuation` only after profiling shows the check matters.
- If two callbacks could both resume, restructure so only one path reaches the continuation.
- Wrap multi-value delegates with `AsyncStream`: store the continuation, yield from delegate methods, stop the delegate source in `onTermination`.
- Combine: `sink` becomes `for await`; `map`/`filter` exist on `AsyncSequence`; `PassthroughSubject` becomes `makeStream`; consume `publisher.values` directly instead of rewrapping. Avoid Combine in new code.

## 13. Replacing GCD and locks

| Legacy | Replacement |
|---|---|
| `DispatchQueue.main.async { … }` | Make the function or type `@MainActor`; callers `await` it |
| `DispatchQueue.global().async` (one-off CPU work) | `@concurrent` async function |
| `DispatchQueue.global()` fan-out | Task group |
| Serial queue guarding mutable state | `actor`, when the API can become async |
| Lock around state in a synchronous API | `Mutex` (Synchronization framework), which keeps checked `Sendable` |
| `asyncAfter` | `try await Task.sleep(for:)` inside a task |

- Keep an API synchronous with a `Mutex` rather than making it an actor just to serialise access. Choose an actor only when the API should become actor-isolated.
- A type guarded by `NSLock` / `os_unfair_lock` usually ends up `@unchecked Sendable`; prefer `Mutex`.

## 14. Sendable

- Structs and enums whose stored members are all `Sendable` are implicitly `Sendable` when internal. Declare the conformance explicitly on `public` types.
- A class can be `Sendable` only if it is `final`, has only immutable `let` properties of `Sendable` types, and has no superclass other than `NSObject`.
- Actors and global-actor-isolated classes are already `Sendable`.
- For mutable shared state choose, in order: convert to a value type, an actor, `@MainActor` if UI-owned, a `Mutex`.
- Region-based isolation lets a non-Sendable value cross a boundary when the sender never uses it again. Check whether the compiler already accepts the code before adding conformances.
- Mark a parameter or result `sending` when ownership is transferred and the sender stops using it.
- `@unchecked Sendable` is a promise you have verified thread safety by hand. It is legitimate only for types with real internal synchronisation. A class with unsynchronised `var`s marked `@unchecked Sendable` still races; fix the type instead.
- In closures that cross boundaries, capture the specific values needed (`let id = model.id; Task { use(id) }`), not a whole non-Sendable object.

## 15. Diagnostics: error, cause, fix

| Diagnostic | Cause | Fix, in order |
|---|---|---|
| Sending 'x' risks causing data races | Non-Sendable value crosses an isolation boundary while still reachable from the sender | Stop using it after the send (region isolation); keep the work on one actor; enable or write `nonisolated(nonsending)` so no hop happens; mark the parameter `sending`; make the type genuinely `Sendable` |
| Capture of 'x' with non-sendable type in a `@Sendable` closure | `Task`, `Task.detached` or `addTask` closure captures non-Sendable state | Capture only the `Sendable` pieces; make the type `Sendable` if it truly is; keep the work on the caller's actor |
| Static property 'x' is not concurrency-safe | Unprotected global or static | See section 6. Check whether a sibling module that compiles cleanly is MainActor-default; that is a settings difference, not a code fix |
| Main actor-isolated 'x' cannot be used from a nonisolated context | Non-isolated code touches main-actor state | If the caller is UI code, isolate it to `@MainActor`; else `await` the access, or move the state off the main actor |
| Expression is 'async' but is not marked with 'await' | The call crosses into another isolation domain | Add `await`; if the caller cannot become async, see section 8 before adding `Task {}` |
| Conformance of 'X' to 'Y' crosses into main actor-isolated code | `@MainActor` type satisfies a non-isolated protocol | If the protocol is only used on main: `extension X: @MainActor Y`. If the type's isolation is incidental: remove it. If the requirement touches no isolated state: make that member `nonisolated` |
| Main actor-isolated conformance of 'X' to 'Y' cannot be used in nonisolated context | Isolated conformance used off its actor | Move the use site onto that actor, or drop the conformance's isolation |
| Isolated conformance cannot satisfy conformance requirement for a 'Sendable' type parameter (`SendableMetatype`) | `X.self` with an isolated conformance passed to a generic requiring `SendableMetatype` | Make the conformance non-isolated (requirements `nonisolated`), or stop passing the metatype across boundaries |
| `Thread.current` unavailable from asynchronous contexts | Debugging by thread | Reason about isolation; use `assertIsolated()` or the debugger |

`@preconcurrency import` is acceptable only for a dependency you cannot change that has not adopted concurrency annotations; note the module and remove it once the dependency updates.

## 16. Bug patterns and review hotspots

Grep for these and inspect each:

- Actor methods that read state, `await`, then write; `!` on actor state after `await`.
- `Task {` inside a `for` loop.
- `Task {` in `onAppear`.
- `Task { try await … }` with no `do`/`catch`: the error vanishes.
- `catch` blocks that alert or retry without excluding `CancellationError`.
- `Task.detached`.
- `withCheckedContinuation` / `withCheckedThrowingContinuation`: resume exactly once on every path.
- `AsyncStream {` closure initializer; missing `finish()` or `onTermination`; unbounded buffer on a fast producer.
- `@unchecked Sendable`, `nonisolated(unsafe)`, `@preconcurrency`: is there real synchronisation or an external reason?
- `MainActor.run` in already-main code.
- CPU-heavy work in a `@MainActor` method or a plain `async` function in a nonsending-by-default module.
- `DispatchQueue`, `asyncAfter`, `Task.sleep(nanoseconds:)`.
- Stored `Task` properties never cancelled.

## 17. Newer APIs

Check the toolchain and deployment target before using these.

- Isolated conformances (Swift 6.2+): `class User: @MainActor Equatable`. Use when the conformance is only needed on that actor; the compiler rejects off-actor use.
- `Task.immediate` (Swift 6.2+): runs synchronously up to the first suspension when already on the target executor. Use only when that immediate start matters; otherwise plain `Task {}`. Groups gain `addImmediateTask` / `addImmediateTaskUnlessCancelled`.
- Task names (Swift 6.2+): `Task(name:)`, `group.addTask(name:)`, read via `Task.name`. Debugging aid only.
- Priority escalation (Swift 6.2+): `withTaskPriorityEscalationHandler`, `task.escalatePriority(to:)`. The runtime escalates automatically; manual use is rare.
- `isolated deinit` (Swift 6.2+): see section 5.

## 18. Swift 6 migration summary

Strategy:

- Migrate one isolated module or feature at a time, smallest dependency-free piece first.
- Update dependencies to concurrency-aware versions before fixing your own code.
- Change as little as possible; do not refactor during migration.
- Do not `@MainActor` everything to make errors go away.
- New code and new packages start in Swift 6 mode.

Order:

1. Add `async` alternatives beside completion-handler APIs.
2. App targets: set default actor isolation to MainActor (Swift 6.2+).
3. Raise strict concurrency minimal → targeted → complete, fixing everything at each level.
4. Add `Sendable` to value types that will cross boundaries.
5. Enable upcoming features one by one, using migrate mode (Xcode build setting set to "Migrate", or `swift package migrate --to-feature <Name>`), rather than flipping Approachable Concurrency on an existing project at once.
6. Switch the language mode to Swift 6.

Loop for each change: build, fix one category of diagnostic, rebuild clean, run tests, then move on. Treat any new warning a fix produces as a fresh diagnosis.
