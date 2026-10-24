# Animation

Read when adding, changing or reviewing animations, transitions, matched/zoom transitions, or custom `Animatable` types.

Contents
- Implicit vs explicit
- Placement and scope
- Timing
- Disabling and transactions
- Transitions
- Identity
- Sequencing: completion, phase, keyframe
- Custom animatable types
- Matched geometry and zoom transitions
- Reduce Motion
- Performance

## Implicit vs explicit

- Always pass a value: `.animation(.spring, value: isExpanded)`. Replace the value-less `.animation(_:)` (deprecated; it animates every change in the subtree).
- Use `withAnimation { ... }` for event-driven changes (taps, gestures, async results landing).
- Use `.animation(_:value:)` when one specific value change should animate one specific region.
- Implicit animations later in the tree override an explicit `withAnimation` for the views they cover.

## Placement and scope

- Put `.animation` after the modifiers it should animate; modifiers below it are not covered.
- Attach it to the smallest subtree that changes, not the root container.
- Animate some properties and not others with the scoped form (iOS 17+):

```swift
Rectangle()
    .foregroundStyle(isOn ? .blue : .red)          // not animated
    .animation(.spring) { $0.scaleEffect(isOn ? 1.2 : 1) }
```

  On older targets, stack `.animation(.spring, value:)` then `.animation(nil, value:)` after the part to exclude.

## Timing

- Default to springs (`.spring`, `.smooth`, `.snappy`, `.bouncy`) for UI. Use `.easeInOut` for plain appearance changes, `.linear` only for progress or continuous loops.
- Keep tap feedback short (well under half a second). A one-second ease on a button press feels broken.
- Adjust with `.speed()`, `.delay()`, `.repeatCount(_:autoreverses:)`, `.repeatForever(autoreverses:)`.

## Disabling and transactions

- Remove animation from a subtree with `.transaction { $0.animation = nil }`. Never fake it with `.linear(duration: 0)`.
- Stop a parent's implicit animations overriding yours with `.transaction { $0.disablesAnimations = true }`.
- `withAnimation(x)` is shorthand for `withTransaction(Transaction(animation: x))`.
- `.transaction { }` without `value:` fires on every change, like the deprecated value-less `.animation`. Prefer `.transaction(value:) { }`.
- To vary animation by cause (user vs server), define a `TransactionKey` (iOS 17+), set it in `withTransaction`, and read it in `.transaction`.

## Transitions

- Transitions animate insertion and removal; property animations interpolate a view that exists before and after. Pick the one that matches what happens.
- A transition needs an animation context outside the conditional: `withAnimation` around the state change, or `.animation(_:value:)` on the parent container. An `.animation` inside the `if` is removed with the view and the removal does not animate.

```swift
// Wrong: removal never animates
if show { Detail().transition(.slide).animation(.spring, value: show) }

// Right
VStack { if show { Detail().transition(.slide) } }
    .animation(.spring, value: show)
```

- Built-ins: `.opacity` (default), `.scale`, `.slide`, `.move(edge:)`, `.offset`, `.push(from:)`, `.blurReplace`. Combine with `.combined(with:)`.
- Use `.asymmetric(insertion:removal:)` when entry and exit should differ.
- Write custom transitions as a `Transition` conformer (iOS 17+) switching on `phase.isIdentity`. Use `AnyTransition.modifier(active:identity:)` only on older targets.
- Never imitate a transition with `.opacity(show ? 1 : 0)` inside the `if`; it cannot animate removal.
- For text and number changes, use `.contentTransition(.numericText(value:))` or `.contentTransition(.interpolate)` rather than a transition.

## Identity

- Different `if`/`else` branches are different views: switching runs a transition, not an interpolation. To interpolate, keep one view and vary its modifiers.
- Changing `.id(x)` destroys and recreates the view, also a transition. Use it deliberately (resetting state, replaying an entrance), never as a way to "refresh".

## Sequencing: completion, phase, keyframe

- Chain steps with the completion closure (iOS 17+), never `DispatchQueue.asyncAfter` or `Task.sleep` between `withAnimation` calls:

```swift
withAnimation { scale = 2 } completion: {
    withAnimation { scale = 1 }
}
```

- For a completion that must fire on every change of a value, use `.transaction(value:) { $0.addAnimationCompletion { ... } }`. Without `value:` it fires once.
- Use `.phaseAnimator(phases, trigger:)` (iOS 17+) for multi-step sequences that return to the start (shake, pulse, bounce). Model phases as a `CaseIterable` enum when there are more than two or three. Omit `trigger` only for a deliberate infinite loop. Vary per-phase timing with the `animation:` closure.
- Use `.keyframeAnimator(initialValue:trigger:content:keyframes:)` (iOS 17+) when properties need independent timing. Each `KeyframeTrack` animates one key path and tracks run in parallel. Keyframe kinds: `CubicKeyframe` (smooth), `LinearKeyframe`, `SpringKeyframe`, `MoveKeyframe` (jump).
- Use `KeyframeTimeline` to sample keyframe values in tests or outside SwiftUI.

## Custom animatable types

- A custom `Shape` or `ViewModifier` whose properties should interpolate must expose them via `animatableData`. Missing it fails silently: the value jumps to the end.
- iOS 26+: mark the type `@Animatable`; it synthesises `Animatable` and `animatableData`. Mark properties that must not interpolate (Bool, Int, enums, IDs) `@AnimatableIgnored`.
- Animated `ViewModifier`: use `@Animatable` (iOS 26+) or conform to `Animatable` directly. Replace `AnimatableModifier` (deprecated).
- Write `animatableData` by hand only when the setter needs logic (clamping, normalising) or the target is below iOS 26. Use `AnimatableValues` on iOS 26+, nested `AnimatablePair` below.

```swift
@Animatable
struct Wedge: Shape {
    var startAngle: Angle
    var endAngle: Angle
    @AnimatableIgnored var clockwise: Bool
    func path(in rect: CGRect) -> Path { ... }
}
```

## Matched geometry and zoom transitions

- Cross-screen (push, sheet, full-screen cover) on iOS 18+: mark the source with `.matchedTransitionSource(id:in:)` and the destination with `.navigationTransition(.zoom(sourceID:in:))`. Gate with `#available`; fall back to the default transition.

```swift
@Namespace private var ns
ThumbnailView(item: item)
    .matchedTransitionSource(id: item.id, in: ns)
// destination
DetailView(item: item)
    .navigationTransition(.zoom(sourceID: item.id, in: ns))
```

- Within one hierarchy, use `.matchedGeometryEffect(id:in:)` and change state inside `withAnimation`. Exactly one view per ID should be the source (`isSource: true`, the default) at a time.
- Keep IDs stable model IDs. Never `UUID()` or an index that shifts.
- Avoid wildly different sizes or shapes between source and destination unless the design asks for it.

## Reduce Motion

- Read `@Environment(\.accessibilityReduceMotion)`. When true, replace large movement (slides, zooms, parallax, bounces, spinning, repeating loops) with an opacity fade or no animation. Keep subtle feedback that carries meaning.

```swift
@Environment(\.accessibilityReduceMotion) private var reduceMotion

.transition(reduceMotion ? .opacity : .move(edge: .bottom))
.animation(reduceMotion ? nil : .spring, value: isExpanded)
```

- Stop `.repeatForever` and phase-animator loops when Reduce Motion is on.
- System components (navigation pushes, sheets, zoom transitions) already respect the setting; do not override them.

## Performance

- Prefer animating transforms and opacity (`scaleEffect`, `offset`, `rotationEffect`, `opacity`) over layout (`frame`, `padding`), which relayouts every frame.
- Never call `withAnimation` on every scroll or gesture tick. Animate only when a derived value crosses a threshold:

```swift
.onScrollGeometryChange(for: Bool.self) { $0.contentOffset.y > 50 } action: { _, show in
    withAnimation(.easeOut(duration: 0.2)) { showTitle = show }
}
```

- Debug a misbehaving animation by temporarily slowing it (`.speed(0.2)`) rather than adding logging inside `animatableData`.
