# Views and state

Read when writing, reviewing or refactoring a SwiftUI view: property wrappers, data flow, decomposition, modifiers, environment, async loading and previews.

Contents
- Architecture: MV by default
- Choosing a property wrapper
- `@State` rules
- `@Observable` models
- Bindings
- Identity and collections
- View decomposition
- Body, init and actions
- Modifiers and structural identity
- Environment
- Async loading state
- Previews
- Hygiene

## Architecture: MV by default

- Views express state and orchestrate; services and `@Observable` models own business logic.
- Reach for `@State`, `@Environment`, `@Query`, `.task`, `.task(id:)` and `onChange` before a view model.
- If the codebase already uses per-screen view models, follow that convention. Existing convention wins.
- Do not add a view model that only mirrors local state, wraps environment values, duplicates `@Query`/`@State`/bindings, exists because `body` is long, or holds one-off loading that fits in `.task` plus local state.
- A view model is justified only when at least one holds:
  1. the user asks for one;
  2. the codebase already standardises on view models for that feature;
  3. the screen needs a long-lived reference model with behaviour that does not fit in services;
  4. the feature bridges a non-SwiftUI API that needs a dedicated object;
  5. several views share presentation-specific state that is not better modelled as app-level environment data.
- When a view model exists: keep it small and non-optional, pass dependencies through the view's `init`, build it there with `_viewModel = State(initialValue: ...)`. No `bootstrapIfNeeded` or delayed setup.
- A long screen is split into subviews first, never "fixed" by adding a view model.
- Do not add a view model to make a simple view testable. Test services, models and transformations; cover views with previews and UI tests.
- With SwiftData, use `@Query` and `@Environment(\.modelContext)` in the view; do not mirror fetched data in a model.

## Choosing a property wrapper

| Situation | Use |
|---|---|
| View owns a value | `@State private var` |
| View owns an `@Observable` instance | `@State private var model = Model()` (never plain `let`; the instance would be recreated) |
| Read-only input from parent | `let` (react to changes with `.onChange`) |
| Child writes parent's value | `@Binding var` |
| Child needs bindings into a passed `@Observable` | `@Bindable var` |
| Shared app-level model or service | `.environment(model)` + `@Environment(Model.self) private var` |
| Framework value or action | `@Environment(\.dismiss) private var` |
| Per-scene UI state that must survive relaunch | `@SceneStorage` |
| Small user preference | `@AppStorage` in a view, never secrets |

- Prefer `@Observable` over `ObservableObject`. Use `ObservableObject`, `@Published`, `@StateObject`, `@ObservedObject` or `@EnvironmentObject` only when the deployment target is below iOS 17 or legacy code makes migration impractical.
- Legacy: `@StateObject` where the view creates the object, `@ObservedObject` where it receives it. Never create an object inline in `@ObservedObject`. In a custom init pass the expression straight to `StateObject(wrappedValue:)`, not via a local.
- A binding you react to must be `@Binding var x`, not `let x: Binding<T>`. A plain property is not tracked, so external changes are missed (often Release-only, and fatal in `UIViewRepresentable.updateUIView`).
- Do not use `@Binding` for a value the child only reads.
- Mark `@FocusState` `private`.

## `@State` rules

- Always `private`. Newer SDKs implement `@State` as a macro that evaluates its initial expression once only for private properties; a non-private one re-runs `SearchStore()` whenever the parent rebuilds. Flag existing non-private state; do not change it unasked, as callers may set it through the initializer.
- Never declare a changing parent-owned input as `@State`. The child keeps the first value forever. Use `let` or `@Binding`.
- Initialise each `@State` in exactly one place. If `init` seeds it, drop the declaration's default and assign once in `init`. Do not fix "used before being initialized" by reordering assignments.
- An init-seeded `@State` is a one-time seed. Later parent arguments do not replace it; use `let`/`@Binding` when they must.
- `@State` may cache an expensive non-observable object (`CIContext`, formatter) with no change tracking.
- Keep per-window UI state (selected tab, navigation path, displayed record) in scene-root `@State` or `@SceneStorage`, not in `@AppStorage`, singletons or shared models, so each scene navigates independently.

## `@Observable` models

- `@MainActor`: first read `SWIFT_DEFAULT_ACTOR_ISOLATION` in build settings or `.defaultIsolation(MainActor.self)` in `Package.swift`. MainActor-default project: no annotation. Otherwise UI-state models get `@MainActor`; non-UI models do not. Never add it as a blanket fix for a compiler error.
- Own with `@State`, pass with `let`/`@Bindable`/`@Environment`.
- Nested `@Observable` objects are tracked; nested `ObservableObject`s are not (pass the inner one directly).
- Give stored property types `Equatable` conformance (enums with associated values, structs, array elements). The generated setter then skips no-op writes.
- Do not use `@AppStorage` inside an `@Observable` class, even with `@ObservationIgnored`; read it in a view or wrap `UserDefaults` explicitly. Any other property wrapper inside an `@Observable` class needs `@ObservationIgnored`.
- Observation tracks stored properties, not fields:
  - a computed property depends on every stored property its getter reads;
  - reading `model.user.name` subscribes to all of `user` when `user` is a struct;
  - reading one element subscribes to the whole collection.
- So pass a row its element or the fields it reads, never the parent model plus an index.
- Narrowing dependencies further (flattened stored properties, per-item models, cached derived values in `didSet`) is a diagnosed-performance fix, not routine.

## Bindings

- Prefer bindings projected from `@State`, `@Binding` or `@Bindable`, including key-path and subscript bindings, over `Binding(get:set:)` in `body`. Closure bindings allocate per pass and cannot be compared.
- For a conversion, add a writable computed property and bind `$model.distanceInMiles`. For a parameterised lookup, add a labelled subscript and bind `$model[scoreFor: player]`. No marker-enum subscripts for argumentless projections.
- Need a binding to an environment model? Re-bind locally: `@Bindable var router = router`, then `$router.path`.
- Refactoring a custom binding preserves behaviour: a setter fires on every write including equal values; `onChange` fires on observed changes including programmatic ones. Leave it alone if the equivalent is unclear.
- Numeric input: `TextField("Score", value: $score, format: .number)` plus `.keyboardType(.numberPad)` or `.decimalPad`. The keyboard modifier alone is not enough.

## Identity and collections

- Make models `Identifiable` instead of `id: \.someProperty`.
- Identity comes from the element, never the position. No `ForEach(items.indices, id: \.self)` or `enumerated()` keyed by `\.offset` on mutable collections. Use `ForEach(items.enumerated(), id: \.element.id)` (iOS 26+) or `ForEach(Array(items.enumerated()), id: \.element.id)`.
- Create identity when the element is stored, not in `body`. Mapping to fresh values with `UUID()` inside `ForEach` recreates every row.
- Use a compact immutable key. `id: \.self` only for small unique values (Int, UUID, URL, short String), never a full struct. Never derive `id` from an editable field.
- Fixed ranges and constant collections are fine as they are.

## View decomposition

- Extract sections into dedicated `View` structs, not computed `some View` properties or `@ViewBuilder` methods. A struct is SwiftUI's invalidation boundary; a helper property is inlined into the parent and re-runs with it.
- Tolerate a handful of small private helpers that belong to the same concern and would fit in `body` if inlined. Otherwise extract.
- Flag a `body` longer than about a screen or with several logical sections. Split files over about 300 lines.
- Give extracted views narrow inputs: data, bindings, callbacks. Pass only the fields a value-type view reads (`avatarURL`, not `user`).
- Detail screens: one view type per section (`ProductHeader`, `ProductGallery`, ...), a thin parent composing them.
- Move a subview to its own file once it is reused or independently meaningful. One type per file is the default.
- Container views store `@ViewBuilder let content: Content`, not `let content: () -> Content`.
- Remove a `Group` wrapping a single view; keep one wrapping siblings, a `ForEach` or `if`/`else`. Inside a `List` row use a `VStack`, not `Group`, around branches.
- Avoid `AnyView`; use `@ViewBuilder` branches. Use `AnyShapeStyle` (not `AnyView`) when ternary style types do not unify, and only then.
- Use `overlay`/`background` to decorate a primary view; `ZStack` only for peers that jointly define layout.
- Repeated modifier stacks become a `ViewModifier` behind a `View` extension; reusable buttons become a `ButtonStyle` exposed via `extension ButtonStyle where Self == X { static var x: X { .init() } }`.
- Prefer `TextField(axis: .vertical)` with `lineLimit(5...)` over `TextEditor` unless full-screen editing is needed.
- Use `ImageRenderer`, not `UIGraphicsImageRenderer`, to render views to images.
- Order members: environment, `let`s, `@State`/stored, non-view computed vars, `init`, `body`, view helpers, functions. Follow a stronger local convention if one exists; never reorder unrelated code just to match.

## Body, init and actions

- `body` reads like layout. Put button actions in methods: `Button("Save", action: save)`, not `Button("Save") { save() }`.
- No business logic inline in `.task`, `.onAppear`, `.onChange` or `.refreshable`; call a small private method that delegates to a service or model.
- No filtering, sorting, mapping or formatter construction in `body`. Prepare data in the model or cache in `@State` updated from `.onChange`.
- Keep `init` constant-time: it re-runs on every parent body pass. No decoding, formatters, file I/O or large allocations.
- Use `Text(date, format: ...)` and other `FormatStyle`s instead of building formatters.
- Never start network work from `body`.
- Keep a stable root. Do not swap whole root trees with top-level `if`/`else`; put conditions inside sections or modifiers (`overlay`, `opacity`, `disabled`, `toolbar` content).
- When a dependency exists only to trigger a side effect in an expensive parent, consider moving it and its `.onChange` into a small `ViewModifier` to give it its own invalidation boundary. Skip this when the parent is trivial.
- Lists of non-trivial rows inside `ScrollView` use `LazyVStack`/`LazyVGrid`.
- `.redacted(reason: .placeholder)` on a container gives a loading skeleton; `.unredacted()` opts children out.

## Modifiers and structural identity

- Modifier order is semantic: each wraps what came before.
  - `padding` before `background` or `border` to colour the padded area.
  - `frame` before `background` to size the background.
  - `contentShape` before a gesture to widen its hit area.
  - `.compositingGroup()` before `.clipShape()` on layered views, else corners fringe.
- Do not write or add an `.if(condition) { $0... }` extension. Its branches are different types, so toggling resets state and breaks animation. Flag existing ones as identity risks; do not rewrite them during unrelated work.
- When two states describe the same view, vary the modifier's value: `.opacity(isVisible ? 1 : 0)`, `.foregroundStyle(isOn ? .red : .primary)`.
- Use real `if` only when content is genuinely different or should leave layout, hit testing and accessibility. Opacity keeps it in all three.
- Use `foregroundStyle`, not `foregroundColor`.
- `.animation(_:value:)` always takes a value; the value-less form is deprecated. Chain animations with `withAnimation { } completion: { }`, not delays.
- Prefer the `@Animatable` macro (iOS 26+) to hand-written `animatableData`; mark non-animating properties `@AnimatableIgnored`. No `AnimatableModifier`.
- Use `Button` for anything tappable, not `onTapGesture`.

## Environment

- Use `@Environment` for app-level services, shared clients, theme and configuration that many descendants need. Use initializer injection for feature-local dependencies; do not move one into the environment to save passing an argument.
- Keep mutable feature state out of the environment unless it is deliberately shared across the app.
- Inject the app's shared models and services in one place near the root: the `App`'s `WindowGroup` or a single `withAppDependencies()` modifier. The `App` owns them with `@State`. Never create shared models inline at reader sites.
- React to account or client switches with `.task(id: currentClient.id)` in that root wiring. Keep it light; long work belongs in services.
- Install one `modelContainer` at the root so tabs and sheets share a store.
- Declare `@Environment` in the smallest view that reads it. Delete unused key-path declarations; they still subscribe.
- Define custom values with `@Entry`. Flag hand-written `EnvironmentKey` boilerplate; do not rewrite it unasked.
- Never store closures in custom environment or focused values. Store a struct with `callAsFunction`, or share an `@Observable`. Framework actions (`dismiss`, `openURL`, `refresh`) are fine.
- Keep `@Entry` defaults stable. No `Model()`, `Date()` or `UUID()` inline; reference a `static let`, or use an optional with `nil` when absence means something.
- Never push per-frame values (scroll offset, drag position, geometry, timers) through the environment. Use `scrollTransition`/`visualEffect`, or a model exposing a coarse value (`isWide`).
- Below iOS 17, `@EnvironmentObject` is the equivalent; use it otherwise only where the project already standardises on it.

## Async loading state

- `.task` for load-on-appear; `.task(id:)` to restart when an input (query, selection, id) changes. Both cancel automatically.
- Model loading as an enum held in `@State`: `idle`, `loading`, `loaded(T)`, `failed(Error)`. One source of truth, not parallel flags.
- Treat cancellation as normal: `catch is CancellationError { return }`, check `Task.isCancelled` in long flows, never show it as an error.
- Debounce typed input inside `.task(id:)` with `try? await Task.sleep(for: .milliseconds(250))`, then `guard !Task.isCancelled`. Skip empty queries.

```swift
@State private var state: LoadState<Item> = .idle
@Environment(ItemClient.self) private var client

var body: some View {
    content.task(id: id) { await load() }
}

private func load() async {
    state = .loading
    do { state = .loaded(try await client.fetch(id: id)) }
    catch is CancellationError { return }
    catch { state = .failed(error) }
}
```

- Do background work in services; publish results back to main-actor UI state.
- Move the flow into a service or model when it spans screens, must survive dismissal, or gains retry, caching or offline policy.
- A small service can be a struct of async closures with `static func live()` and `static func mock(...)`; keep URL building and decoding in it and UI state out of it. Inject the store that uses it via `.environment`, not a singleton.

## Previews

- Use `#Preview`, not `PreviewProvider`, in new code.
- One named preview per meaningful state: loaded, empty, loading, error.
- Previews never touch the network, real databases, disk or global singletons. Use deterministic fixtures (`static let sample`, `static var preview` on the model) and mock services.
- Inject every environment dependency the view reads; a missing `@Environment(Model.self)` crashes the preview.
- Wrap destinations in `NavigationStack` so titles and toolbars render.
- Interactive state: `@Previewable @State var isOn = false` (iOS 18+). Below that, a private wrapper view.
- Do not make a dependency optional to stop a preview crashing; fix the wiring.
- Preview helpers touching main-actor models need `@MainActor`.

## Hygiene

- No secrets in the repository. No credentials in `@AppStorage`; use the Keychain.
- Extract button actions and logic so core behaviour has unit tests; UI tests only where unit tests cannot reach.
- If SwiftLint is configured, leave it with no warnings.
- Debug re-renders with `let _ = Self._printChanges()` and remove it before committing.
- "Unable to type-check this expression in reasonable time": extract subviews, split long modifier chains and inline closures.
