# Modern API

Read this when writing new SwiftUI or Swift code, or when code you touch uses an
older API and you must decide whether to replace it.

## Contents

- How to act on a deprecation
- Views and modifiers
- Navigation, presentation and toolbars
- Text input, events and gestures
- Environment, colour and graphics
- Lists, scrolling and collections
- Toolbars (iOS 26+ / 27+)
- Swift language and Foundation idioms

## How to act on a deprecation

Status column key:

- **Hard**: the SDK names a real deprecation version and the compiler warns.
  In code the task touches, replace it.
- **Soft**: the SDK uses `deprecated: 100000.0`, so the build stays quiet. A
  clean build proves nothing. Flag it in code you touch; do not change it
  unasked. Never introduce it in new code.
- **Prefer**: not deprecated; the newer API is the house default for new code.
  Do not migrate existing working code for this alone.

Rules:

- Read the deployment target before choosing. If the replacement's minimum OS is
  above the target, gate it with `if #available` and keep the old call in the
  fallback branch. Never raise the target to make a rule fit.
- A toolchain requirement (Xcode N) is not a runtime gate. `#available` cannot
  make an older SDK see a new declaration.
- When unsure of a symbol's status, read its `@available` attribute in the SDK.
- Scope: only the views you edit. Do not scan, list or offer migrations for code
  outside the task.
- During feature or bug work, keep a soft-deprecated API that the edited view
  already uses. Migration changes behaviour (state resets, navigation) and
  belongs in its own change; offer it in one line afterwards.
- In a review, refactor or "modernise" task, report soft-deprecated APIs as
  improvements, not bugs.

## Views and modifiers

| Old | New | Status | Min OS |
|---|---|---|---|
| `foregroundColor(_:)` | `foregroundStyle(_:)` | Soft | 15 |
| `cornerRadius(_:)` | `clipShape(.rect(cornerRadius:))` | Soft | 15 |
| `accentColor(_:)` | `tint(_:)`; app-wide default via the asset catalog accent colour | Soft | 16 |
| `edgesIgnoringSafeArea(_:)` | `ignoresSafeArea(_:edges:)` | Soft | 14 |
| `animation(_:)` without `value:` | `animation(_:value:)` or `withAnimation` | Hard | 13 |
| `AnimatableModifier` | conform the modifier to `Animatable` | Soft | 27 SDK |
| Hand-written `animatableData` that only forwards properties | `@Animatable` (exclude with `@AnimatableIgnored`); keep manual code that clamps or normalises | Prefer | 26 |
| `overlay(V())` / `background(V())` / `mask(V())` view argument | builder closure: `.overlay { ... }` | Soft | 15 |
| Overlay used only to stroke a filled shape | chain `.fill(...)` then `.stroke(...)` | Prefer | 17 |
| `Font.system(size:weight:design:)` legacy forms, `Font.system(_:design:)` | current `system(_:weight:design:)` overloads | Soft | 27 SDK |
| `fontWeight(.bold)` | `bold()` | Prefer | 16 |
| Text built with `+` | interpolation: `Text("\(red)\(blue)")` | Hard | 26 SDK |
| `Image("avatar")` with generated asset symbols on | `Image(.avatar)` | Prefer | Xcode 15 toolchain |
| `UIScreen.main` / `.bounds` | proposed size, size classes, `@Environment(\.displayScale)` | Hard | 26 |
| `@ViewBuilder` on new builders | `@ContentBuilder` (same type; leave existing uses) | Prefer | toolchain only, see `platform-facts.md` |
| `ContentSizeCategory` / `\.sizeCategory` | `DynamicTypeSize` / `\.dynamicTypeSize` | Soft | 15 |
| `.accessibility(label:)` and siblings | `accessibilityLabel()`, `accessibilityValue()`, `accessibilityHint()`, `accessibilityAddTraits()`, `accessibilityHidden()` | Hard | 14 |

Text concatenation:

```swift
// Before
Text("Hello").foregroundStyle(.red) + Text("World").foregroundStyle(.blue)
// After
let red = Text("Hello").foregroundStyle(.red)
let blue = Text("World").foregroundStyle(.blue)
Text("\(red)\(blue)")
```

## Navigation, presentation and toolbars

| Old | New | Status | Min OS |
|---|---|---|---|
| `NavigationView` | `NavigationStack` / `NavigationSplitView` | Soft | 16 |
| `NavigationLink(destination:isActive:)`, `(tag:selection:)` | `NavigationLink(value:)` + `navigationDestination(for:)` | Hard | 16 |
| Eager `NavigationLink(destination:)` | closure destination or value-based link | Soft | 16 |
| `navigationBarTitle(_:)` | `navigationTitle(_:)` + `navigationBarTitleDisplayMode(_:)` | Soft | 14 |
| `navigationBarItems(...)` | `toolbar { ToolbarItem(...) }` | Soft | 14 |
| `.navigationBarLeading` / `.navigationBarTrailing` | semantic placement (`.cancellationAction`, `.confirmationAction`, `.primaryAction`), else `.topBarLeading` / `.topBarTrailing` | Soft | 17 |
| `navigationBarHidden(_:)`, `toolbar(.hidden, for:)` | `toolbarVisibility(_:for:)` | Soft | 18 |
| `toolbarBackground(.visible, for:)` (visibility overload) | `toolbarBackgroundVisibility(_:for:)`; keep `toolbarBackground` when it sets a style | Soft | 18 |
| `statusBarHidden(_:)` / `statusBar(hidden:)` | `toolbarVisibility(_:for: .statusBar)` | Soft | 27 |
| `ToolbarItem(..., showsByDefault:)` | `defaultCustomization(_:options:)` with `.hidden` | Soft | 27 SDK |
| `tabItem(_:)` | `Tab("Home", systemImage: "house") { ... }`; never mix with `Tab(role:)` | Soft | 18 |
| `TabView(selection:content:)` legacy builder | `TabContentBuilder`-based initialisers | Soft | 18 |
| `PresentationMode` / `\.presentationMode` | `\.dismiss`, `\.isPresented` | Soft | 15 |
| `actionSheet(...)` / `ActionSheet` | `confirmationDialog(_:isPresented:actions:message:)` | Soft | 15 |
| `alert(isPresented:content:)` / `Alert` | `alert(_:isPresented:actions:message:)` | Soft | 15 |
| `presentationBackground(_:)` on sheets | default Liquid Glass sheet material | Prefer | 26 |
| Custom toolbar background hacks | `scrollEdgeEffectStyle(_:for:)` | Prefer | 26 |
| `ContextMenu` / `contextMenu(_:)` | `contextMenu(menuItems:)` | Soft | 27 SDK |

Navigation structure, sheets and alerts live in another reference; this table
covers only the API swap.

## Text input, events and gestures

| Old | New | Status | Min OS |
|---|---|---|---|
| `onChange(of:perform:)` (one-parameter closure) | `onChange(of:) { }` or `{ old, new in }`; `initial: true` to fire on appear | Hard | 17 |
| `TextField` `onEditingChanged:` / `onCommit:` | `onSubmit` + `focused(_:)` | Soft | 15 |
| `autocapitalization(_:)` | `textInputAutocapitalization(_:)` (`.never` replaces `.none`) | Soft | 15 |
| `disableAutocorrection(_:)` | `autocorrectionDisabled(_:)` | Soft | 15 |
| `onTapGesture` for a tappable control | `Button`; keep `onTapGesture` only for location or count | Prefer | 13 |
| `onLongPressGesture(..., pressing:)` | `onLongPressGesture(minimumDuration:maximumDuration:perform:onPressingChanged:)` | Soft | 27 SDK |
| `MagnificationGesture` | `MagnifyGesture` (`value.magnification`) | Soft | 17 |
| `RotationGesture` | `RotateGesture` (`value.rotation`) | Soft | 17 |
| `UIImpactFeedbackGenerator` and siblings in views | `sensoryFeedback(_:trigger:)` | Prefer | 17 |
| `UIPasteboard.general` for user-initiated paste | `PasteButton` | Prefer | 16 |
| String type identifiers in paste/drop APIs | `UTType` overloads | Soft | 27 SDK |
| `dropDestination` with location / `isTargeted` | `dropDestination(for:isEnabled:action:)` with `DropSession` (behaviour change: action returns `Void`) | Soft | 26 |

## Environment, colour and graphics

| Old | New | Status | Min OS |
|---|---|---|---|
| Manual `EnvironmentKey` + computed property | `@Entry var x: T = default` (also `FocusedValues`, `Transaction`, `ContainerValues`) | Prefer | Xcode 16, back-deploys |
| `colorScheme(_:)` | `.environment(\.colorScheme, .light)`; `preferredColorScheme` changes the whole presentation, so it is not a drop-in | Soft | 13 |
| `Color.cgColor` | `color.resolve(in: environment).cgColor` | Soft | 17 |
| `Color(UIColor)` / `Color(CGColor)` unlabeled | `Color(uiColor:)`, `Color(cgColor:)` | Soft | 27 SDK |
| `UIColor` in SwiftUI code | `Color` or asset catalog colours | Prefer | 13 |
| `coordinateSpace(name:)` | `coordinateSpace(.named("x"))`; read with `frame(in: .named("x"))` | Soft | 17 |
| `CoordinateSpace` overloads | `CoordinateSpaceProtocol` overloads | Soft | 17 |
| Style initialisers taking `tint:` | `.tint(_:)` on the view | Soft | 27 SDK |
| `ObservableObject` / `@StateObject` / `@ObservedObject` | `@Observable` + `@State` / `@Bindable` | Prefer | 17 |

`@Entry` defaults must be equal across reads. Do not default to a new reference
instance, `.now` or `UUID()`; every read yields a new value and invalidates
readers. Use an optional entry with no default, or a shared `static let`.
Constant defaults (`true`, `1`, `.orange`) are fine.

Custom environment actions: pass a struct or `@Observable` object holding the
data plus a method (optionally `callAsFunction()`), not a stored closure.
Closures cannot be compared, so readers update on every change. System actions
(`\.dismiss`, `\.openURL`) need nothing.

If `ObservableObject` is genuinely needed (a Combine debouncer), add
`import Combine`; SwiftUI no longer re-exports it.

## Lists, scrolling and collections

| Old | New | Status | Min OS |
|---|---|---|---|
| `ScrollView(..., showsIndicators: false)` | `ScrollView { }.scrollIndicators(.hidden)` | Soft | 16 |
| `searchable(..., suggestions:)` | `searchable(...)` + `searchSuggestions { }` | Soft | 16 |
| `Section(header:footer:content:)` positional views | `Section { } header: { } footer: { }`; `Section("Title") { }` stays current | Soft | 15 |
| `GroupBox(label:content:)` | `GroupBox(content:label:)` | Soft | 27 SDK |
| `Picker(selection:label:content:)` | `Picker(selection:content:label:)` | Soft | 27 SDK |
| `Section.collapsible(_:)` | plain `Section` initialiser | Soft | 27 SDK |
| List style `alternatesRowBackgrounds:` argument | base style + `alternatingRowBackgrounds()` | Soft | 27 SDK |
| `ForEach(Array(items.enumerated()), ...)` | `ForEach(items.enumerated(), id: \.element.id)` | Prefer | 26 |

Keep the `Array(...)` wrapper when the target is below iOS 26.

## Toolbars (iOS 26+ / 27+)

Toolbar overflow and placement rules. Gate each with `#available` unless the
target already covers it.

- iOS 27+: use `ToolbarOverflowMenu` for overflow actions. The system already
  overflows items that do not fit, so a hand-made ellipsis `Menu` turns into a
  nested submenu.
- A single-purpose menu (filter) may stay a `Menu`. Give it a descriptive SF
  Symbol, not an ellipsis.
- iOS 27+: set `visibilityPriority(.high)` on frequent actions and status items,
  `.low` on expendable ones. Without priorities, overflow takes the trailing end.
- iOS 27+: use `.topBarPinnedTrailing` for at most one indispensable action.
- iOS 27+: consider `.toolbarMinimizationBehavior(.onScrollDown, for: .navigationBar)`
  on content-led screens. iOS 26+ tab bar equivalent: `.tabBarMinimizeBehavior(.onScrollDown)`.
- iOS 27+: apply `contentMarginsRemoved()` to a `ToolbarItem` holding custom
  imagery so the glass fits the image. Leave margins on ordinary buttons.
- Any OS: group related controls (previous/next) in one `ToolbarItemGroup`
  instead of separate items split by fixed spacers.
- `ForEach` may generate items directly in `toolbar()` on toolchains that
  support it (no runtime gate; see `platform-facts.md`).
- iOS 26.4+: toggle `ToolbarItem { ... }.hidden(!canExport)` instead of
  building the item conditionally. Check `platform-facts.md` for the minimum
  Xcode; the runtime gate alone is not enough.
- Give every tab and toolbar action both a title and an icon:
  `Button("Compose", systemImage: "square.and.pencil", action: compose)`.

## Swift language and Foundation idioms

Concurrency rules (GCD, `Task.sleep(for:)`, actors, `Sendable`, `MainActor.run`,
`Task.detached`) live in `concurrency.md`.

| Old | New |
|---|---|
| `replacingOccurrences(of:with:)` | `replacing(_:with:)` |
| `FileManager` directory lookups, `appendingPathComponent` | `URL.documentsDirectory`, `appending(path:)` |
| `String(format: "%.2f", v)` | `Text(v, format: .number.precision(.fractionLength(2)))` / `FormatStyle` |
| `DateFormatter` with `dateFormat` for display | `Text(date, format: .dateTime...)`; if a formatter is needed, `setLocalizedDateFormatFromTemplate("yMMMd")` (use `y`, not `yyyy`) |
| `DateFormatter` / manual parsing of ISO strings | `try Date(string, strategy: .iso8601)` |
| `items.joined(separator: ", ")` for display | `items.formatted()`, `.formatted(.list(type: .or))` |
| `"\(first) \(last)"` for a person's name | `PersonNameComponents` with its format style |
| `NSLocalizedString(...)` | `String(localized:)` |
| `filter { }.count` | `count(where:)` |
| `Date()` | `Date.now` |
| `CGFloat` in your own code | `Double`, unless an API requires `CGFloat` (optionals, `inout`, `[CGFloat]`) |
| `Circle()`, `BorderedProminentButtonStyle()` as arguments | static members: `.circle`, `.borderedProminent` |
| `if let value = value` | `if let value` |
| `contains()` / `localizedCaseInsensitiveContains()` on user search input | `localizedStandardContains()` |
| Closure-based API with an `async` twin | the `async` version |

Further rules:

- Omit `return` in single-expression bodies; use `if`/`switch` as expressions.

  ```swift
  var tileColor: Color { if isCorrect { .green } else { .red } }
  ```

- Avoid `!` and `try!`. Use `if let`, `guard let`, `??`, `try?` or `do/catch`.
  If failure truly cannot happen, `fatalError("reason")` beats a bare `!`.
- Repeated identical sort closures mean the type should conform to `Comparable`.
- Show errors from user actions; `print(error)` alone is a swallowed error.
- `import SwiftUI` already brings in UIKit types (`UIImage`); do not add
  `import UIKit` just for them.
- Keep fixed formats and `joined(separator:)` for storage and API exchange,
  where a stable format is the point.
