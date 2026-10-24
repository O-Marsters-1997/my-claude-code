# Performance

Read when writing or reviewing any SwiftUI view (part 1), and when diagnosing a measured slowdown or doing a requested deep performance review (part 2).

Contents
- Part 1. Routine rules — always apply
  - View bodies and initialisers
  - Identity and structure
  - Lists and rows
  - Frequently changing values
  - Async work and threading
  - Images
- Part 2. Deep restructuring — only on a diagnosed problem or explicit deep review
  - When this applies
  - Audit workflow
  - Narrowing dependencies
  - Splitting observation
  - Diffing tricks
  - Diagnosis tools

---

## Part 1. Routine rules — always apply

### View bodies and initialisers

- Treat `body` as called often. Keep it a pure description of state: no side effects, no sorting, filtering, grouping or formatting beyond the trivial.
- Keep `init` trivial. Move loading and setup into `.task`.
- Never create objects in `body` or `init` that live longer than one call: formatters, models, `NSCache`, `URLSession`.
- Format with `FormatStyle` in `Text` (`Text(date, format: .dateTime.day().month())`, `Text(price, format: .currency(code: code))`). If a legacy `Formatter` is unavoidable, hold it as a `static let`.
- Derive cheap values with a computed property or a `let`. Do not mirror derived values in separate `@State` (`itemCount` next to `items`).
- Do not use `@State` as a cache for expensive derived collections unless you also own the invalidation (an `onChange(of:)` on every input). Prefer computing it in the `@Observable` model when its inputs change.

### Identity and structure

- Vary modifier values, not views: `.opacity(isDisabled ? 0.4 : 1)`, not `if isDisabled { v.opacity(0.4) } else { v }`. Branching changes structural identity, resets state, recreates platform views and breaks animations.
- If a ternary's branches are different style types, wrap each in `AnyShapeStyle`. That is cheap; `AnyView` is not.
- Never add a `View.if(_:transform:)` helper. Report existing ones with the reason, but leave removal to a requested change; they usually have many callers.
- Avoid swapping whole top-level trees for small state changes. Keep a stable root; localise the condition to a section or modifier.
- Avoid `AnyView`. Use generics, `@ViewBuilder` parameters, `Group`, or an extracted view.
- Store built content, not escaping builder closures:

```swift
// Avoid
struct Card<Content: View>: View { let content: () -> Content; ... }
// Prefer
struct Card<Content: View>: View { @ViewBuilder let content: Content; ... }
```

- Extract a subview when part of a screen changes much more often than the rest, so only that part re-evaluates.
- Remove key-path `@Environment` and `@FocusedValue` properties the view never reads; the declaration alone subscribes it.

### Lists and rows

- Give `ForEach` stable, unique domain IDs. Never `id: \.self` on values that are not unique and stable, never `enumerated()` with `id: \.offset` on a collection that can reorder or change, never `UUID()` generated during render.
- Pass each row its element (`Row(item: item)`), not the whole store plus an index or ID; looking it up makes every row depend on the whole collection.
- Never construct a row's `@Observable` object inside `body` or the `ForEach` closure. Create once, retain in the parent or model.
- Each `ForEach` element must produce a constant number and kind of top-level views. No bare `if`/`switch` at the top of a row in a lazy container; wrap in a `VStack` (`Group` does not count). To omit elements, filter the collection before `ForEach`.
- Never filter or sort inline in the `ForEach`/`List` initialiser when it reruns often. Precompute in the model.
- Use `List` or `LazyVStack`/`LazyHStack`/`LazyVGrid` for long or unbounded content; flag eager stacks with many children. Do not build custom scroll containers where these suffice.
- Flatten nested `ForEach` where possible. No `AnyView` in rows.
- Give a `ScrollView` with an opaque, static, solid background `.scrollContentBackground(.visible)`.

### Frequently changing values

- Keep fast-changing values (scroll offset, geometry, drag location, timer ticks, animation progress) out of the environment; every write walks the descendant tree.
- Write state from hot paths only when a derived value changes:

```swift
.onScrollGeometryChange(for: Bool.self) { $0.contentOffset.y > 32 } action: { _, new in
    showTitle = new
}
```

  The same applies to `onPreferenceChange`, `onReceive` and gesture handlers: compare before assigning.
- Converting a raw measurement to a Bool inside `body` still subscribes the view to every raw change. Do the conversion where the value is produced.
- Use `scrollTransition` or `visualEffect` for scroll-driven visual effects (fade, scale, rotate); they skip `body`.

### Async work and threading

- Use `.task` (or `.task(id:)`) for async work, not `onAppear` plus `Task {}`; it cancels when the view goes away.
- Keep main-actor work short. Anything that can take tens of milliseconds (decoding, parsing, image processing, large sorts) runs off the main actor via async functions or a `nonisolated`/`@concurrent` function. No GCD.
- Closures SwiftUI may run off the main thread (`Shape.path(in:)`, `visualEffect`, `Layout` methods, `onGeometryChange` transform) must not touch main-actor state. Capture values: `.visualEffect { [pulse] content, _ in ... }`.

### Images

- Use `AsyncImage` with phase handling (empty, success, failure). It honours HTTP cache headers; do not add a custom cache or dependency just for that. If images reload, check server cache headers first.
- Newer SDKs add `AsyncImage(request:)` and `.asyncImageURLSession(_:)` for per-request cache policy and a custom `URLSession`/`URLCache` (availability in `platform-facts.md`). Gate with `#available` and keep `AsyncImage(url:)` as the fallback.
- Never `UIImage(data:)` on the main thread for large images in lists or grids. Decode and downsample off the main actor with ImageIO, to the displayed size times `@Environment(\.displayScale)`:

```swift
func downsample(_ data: Data, maxPixel: CGFloat) -> UIImage? {
    let src = CGImageSourceCreateWithData(data as CFData,
        [kCGImageSourceShouldCache: false] as CFDictionary)
    let opts: [CFString: Any] = [
        kCGImageSourceCreateThumbnailFromImageAlways: true,
        kCGImageSourceThumbnailMaxPixelSize: maxPixel,
        kCGImageSourceCreateThumbnailWithTransform: true,
        kCGImageSourceShouldCacheImmediately: true,
    ]
    return src.flatMap { CGImageSourceCreateThumbnailAtIndex($0, 0, opts as CFDictionary) }
        .map(UIImage.init(cgImage:))
}
```

  Read `displayScale` in the view and pass it in; do not read global screen state. Do not multiply again if the caller already has pixels. Suggest this as an improvement; do not rewrite every `UIImage(data:)` unasked.
- `UIImage(named:)` caches in the system cache. For many single-use bundle images (galleries, sliders) use `UIImage(contentsOfFile:)`. Bound any processed-image cache with `NSCache.countLimit`.

---

## Part 2. Deep restructuring — only on a diagnosed problem or explicit deep review

### When this applies

Use this part only when (a) the user asks for a deep performance review, or (b) there is an established problem: measured hitches, hangs, janky scrolling, typing lag, or excessive body updates shown by `_logChanges` or Instruments. These changes trade simplicity for fewer updates. Without such a trigger, do not apply them and do not flag their absence.

Reading several properties of one `@Observable` object is not on its own a reason to split anything; Observation already tracks each property read.

### Audit workflow

1. Classify the symptom: slow render, janky scroll, CPU or battery, memory growth, hang (>100 ms on a discrete interaction), or overly broad updates.
2. Review code first. Triage in this order: broad invalidation and observation fan-out; unstable identity and list churn; main-thread work during render; image decode and resize; layout and animation cost.
3. If code does not explain it, ask the user for runtime evidence: profile a Release build on a real device with Instruments' SwiftUI template (SwiftUI timeline plus Time Profiler), one interaction per capture, with a note of what was happening. Debug builds and Simulator distort results.
4. Separate code-backed hypotheses from trace-backed findings; say which each is. Rank by impact, not by ease.
5. Apply targeted fixes from the sections below, then ask for the same capture again to confirm the change.

### Narrowing dependencies

- When a child needs up to about three fields of a struct, pass the fields, not the struct (`TrackTitle(title: track.title)`). Four or more: pass the struct. Does not apply to `@Observable` objects.
- Do not thread a very large value (nested collections, dozens of fields) through many layers; whole-value comparison runs on every parent update. Pass a smaller value or the owning `@Observable` object. Do not turn a struct into a class just to dodge comparison.
- Pass specific values instead of a broad config or context object, and instead of reading a whole model from the environment in every row.
- If an expensive `body` reads a property only to drive `onChange`, move the read and the `onChange` into a small `ViewModifier`.
- Isolate state-dependent subtrees from `GeometryReader`s and preference chains so layout readers do not re-run with them.

### Splitting observation

- If one fast-changing field inside a struct stored on an `@Observable` object invalidates readers of its siblings, lift that field to its own stored property on the object.
- If many views read a small slice of a large `@Observable` model and still over-update, extract that slice into a smaller `@Observable` object and pass it directly.
- Where rows change independently (favourite flags, per-row progress, visibility derived from scrolling), give each element its own `@Observable` model, created once and retained by the parent. Do not convert existing value models wholesale.
- For a measured cost from a computed property on an `@Observable` object, store the result and refresh it from each input's `didSet`. Keep every update path or the UI goes stale.
- Coarsen raw measurements in the model (`hasReachedEnd` rather than `offset`) so readers subscribe to the derived value only.

### Diffing tricks

- `.equatable()` with a custom `==` only when equality is cheaper than re-running the subtree and the inputs are value types. Update `==` whenever inputs change. Never a blanket fix.
- Views holding only plain value properties (no property wrappers) diff fastest. Wrapping an expensive stateful view in a thin plain-value parent can skip its diff when inputs are unchanged. Use only with evidence.

### Diagnosis tools

- `let _ = Self._logChanges()` (iOS 17+) or `Self._printChanges()` inside `body`, under `#if DEBUG`, shows which input triggered an update (`@self`, `@identity`, or a property name). Remove before commit.
- Launch argument `-LogForEachSlowPath YES` logs `ForEach`s that fall off the constant-row-count fast path.
- In Instruments' SwiftUI template: long view body updates, update groups with high counts, and the cause graph ("Show Causes") to find what triggers frequent updates; correlate with Time Profiler.
