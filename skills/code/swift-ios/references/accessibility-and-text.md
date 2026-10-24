# Accessibility and text

Read this when writing or reviewing anything users see or hear: text, fonts,
labels, colour, motion, localisation, theming, haptics or media.

## Contents

- Dynamic Type and fonts
- VoiceOver and Voice Control
- Tappable elements
- Grouping and custom controls
- Motion and colour
- Text: localised or verbatim
- Localisation
- Formatting values
- Design consistency and theming
- Haptics
- Media

## Dynamic Type and fonts

- Use text styles (`.body`, `.headline`, `.title2`). Never force a point size
  on primary content.
- Custom fonts scale through `Font.custom(_:size:relativeTo:)`. Plain
  `custom(_:size:)` scales relative to body.
- Scale non-text values (icon size, padding, spacing) with `@ScaledMetric`,
  adding `relativeTo:` to track a nearby text style. iOS 26+:
  `.font(.body.scaled(by:))` is also available.
- Read `@Environment(\.dynamicTypeSize)`, not `sizeCategory`. Switch to a
  roomier layout (often vertical) when `dynamicTypeSize.isAccessibilitySize`.
- Use `bold()` rather than `fontWeight(.bold)`; the system picks the right
  weight. Use other `fontWeight` values only with a design reason.
- Avoid `.caption2`; use `.caption` sparingly.
- Never hardcode text frame sizes; translations and scripts vary in length and height.

## VoiceOver and Voice Control

- Every icon button and menu gets a text title, even when only the icon shows:

  ```swift
  // Before
  Button(action: addUser) { Image(systemName: "plus") }
  // After
  Button("Add User", systemImage: "plus", action: addUser)
  ```

  Toolbars pick icon-only styling themselves. To force icon-only elsewhere,
  add `.labelStyle(.iconOnly)`, which keeps the title for VoiceOver. Same for
  `Menu("Options", systemImage: "ellipsis.circle") { }`.
- Flag images whose spoken name is meaningless (`Image(.banner2)`).
  Decorative asset: `Image(decorative:)`. Decorative SF Symbol:
  `.accessibilityHidden(true)`. Meaningful: `.accessibilityLabel(...)`.
- Give a button with a long or live-changing label `accessibilityInputLabels`
  with short spoken names (a ticker label "AAPL $271.68" gets "Apple").
- Use dedicated modifiers (`accessibilityLabel`, `accessibilityValue`,
  `accessibilityHint`, `accessibilityAddTraits`), never `accessibility(label:)`.
- Express state with traits: `.accessibilityAddTraits(isSelected ? [.isSelected, .isButton] : .isButton)`.
- `.disabled(true)` makes VoiceOver announce "dimmed"; use it for
  non-interactive states instead of only greying out.

## Tappable elements

- Use `Button` for anything tappable. Use `onTapGesture` only when you need tap
  location or count, and then add `.accessibilityAddTraits(.isButton)`.
- Every tap target is at least 44x44 points. Enforce it; grow the hit area
  with padding and `.contentShape(.rect)` when the visual is smaller.

## Grouping and custom controls

- `.accessibilityElement(children: .combine)` to read a row's labels as one
  element.
- `.accessibilityElement(children: .ignore)` plus one `accessibilityLabel` when
  the combined reading would be awkward.
- `.accessibilityElement(children: .contain)` with a label for a named group
  (a custom tab bar).
- Increment/decrement controls: `.accessibilityElement()`,
  `.accessibilityValue("Page 2 of 5")`, `.accessibilityAdjustableAction`.
- A custom view that behaves like a stock control declares it with
  `.accessibilityRepresentation { Toggle(label, isOn: $isOn) }`.
- Pair a separate label and control with `accessibilityLabeledPair(role:id:in:)`.
- When visual order differs from logical order, set
  `accessibilitySortPriority`. Keep VoiceOver order matching reading order
  across layout variants.

## Motion and colour

- When `@Environment(\.accessibilityReduceMotion)` is true, swap large motion
  (slides, zooms, parallax) for opacity changes.
- When colour carries meaning, honour `\.accessibilityDifferentiateWithoutColor`
  with an icon, pattern or stroke as well.
- Prefer hierarchical styles (`.secondary`, `.tertiary`) over manual opacity.
- Use SwiftUI `Color` or asset colours, not `UIColor`. Use `foregroundStyle`.
- Keep enough contrast between text and its background in light and dark mode.

## Text: localised or verbatim

- `Text("Save")` with a literal is a localisation key. This is the default.
- `Text(stringVariable)` is shown verbatim; no `verbatim:` needed.
- `Text(verbatim: "...")` only for a literal that must never be translated.
- Do not wrap literals in `String(localized:)`, `NSLocalizedString` or
  `LocalizedStringResource` when passing them to SwiftUI. That resolves early
  and ignores `\.locale` overrides.
- Wrapping a runtime `String` in `LocalizedStringKey(...)` does not localise it;
  Xcode cannot extract a key from a runtime value.
- Build sentences with interpolation, never `+` or fragments glued together.
  Word order differs between languages.

  ```swift
  // Before
  Text("Folder: " + folderName)
  // After
  Text("Folder: \(folderName)")
  ```

- For counts in English, French, German, Italian, Portuguese and Spanish, use
  automatic grammar agreement: `Text("^[\(people) person](inflect: true)")`.

## Localisation

- Add strings through the project's existing mechanism. Keep `.strings` /
  `.stringsdict` projects as they are; moving to a String Catalog needs an
  explicit request. A catalog must already exist; Xcode does not create one.
- With `Localizable.xcstrings` and generated string symbols on, add new keys as
  symbols with `extractionState` `"manual"` and use them (`Text(.libraryEmpty)`).
  Otherwise follow the project's key style (opaque keys or natural strings).
- Offer translations of new keys in every language the project ships.
- Route groups of strings to a table with `tableName:`.
- Type new properties, parameters and enum payloads carrying UI text as
  `LocalizedStringResource` (models included). It resolves at display time in
  the active locale. Leave existing `String` or `LocalizedStringKey` storage
  alone outside the task.

  ```swift
  enum Category {
      case mains, desserts
      var name: LocalizedStringResource {
          switch self { case .mains: "Mains"; case .desserts: "Desserts" }
      }
  }
  ```

- Below iOS 26, `LabeledContent` and plain `TextField(_:text:)` do not take a
  `LocalizedStringResource` title; pass `Text(resource)` as the label instead.
- In a framework or Swift package, pass the bundle:
  `Text("Reading list", bundle: #bundle)`. Without it lookup uses the main
  bundle and fails silently. Prefer `#bundle` over `Bundle.module`.
- Outside views, use `String(localized:)`, never `NSLocalizedString`
  (interpolated keys cannot be extracted from it).
- Add `comment:` for ambiguous words ("Edit", "Record") and describe each
  interpolated placeholder by position. Keep one home per string: call site or
  catalog.
- Bake case into the string ("SECTION HEADER") instead of `.textCase` or
  runtime uppercasing, so translators control it. Show user-entered text as typed.
  If a transform is unavoidable, use `localizedUppercase`.
- Read `@Environment(\.locale)`, `\.calendar` and `\.timeZone` in views, not
  their `.current` forms, so previews and per-view overrides work.

## Formatting values

- Use `Text(value, format:)` or `.formatted()`: `.number`, `.percent`,
  `.currency(code:)`, `.dateTime.month().day().year()`. Fields choose what
  shows; the locale chooses order.
- Never `String(format:)`, `specifier: "%.2f"` or a hand-set `dateFormat` for
  display. If a `DateFormatter` is unavoidable, use
  `setLocalizedDateFormatFromTemplate`.
- Lists for display: `items.formatted()`, not `joined(separator:)`.
- People's names: `PersonNameComponents` with its format style.
- User search: `localizedStandardContains()`.

## Design consistency and theming

- Put shared fonts, spacing, padding, corner radii and animation timings in one
  constants enum so screens stay uniform.
- Add an `@Observable` theme object, injected at the root with `.environment`
  and read with `@Environment(Theme.self)`, only when the theme changes at
  runtime (user-selectable). Apply `@MainActor` per the project's isolation
  setting. Persist user choices if needed.
- Name theme colours semantically (`primaryBackground`, `label`); do not
  duplicate system colours or scatter raw `Color` literals in views.
- Theme complements `\.colorScheme`; it does not replace it. Test light and dark.
- Use `Label` for icon-plus-text, not an `HStack`.
- Use `LabeledContent` for title-value rows, with a custom `LabeledContentStyle`
  when several screens need one look.
- `RoundedRectangle` is already `.continuous`; do not pass the style.
- Use `ContentUnavailableView` for empty or missing data, not a custom design.
- Keep stock component glass; never cover it with opaque fills.

## Haptics

- Use `sensoryFeedback(_:trigger:)` on the view that owns the state change, not
  UIKit feedback generators fired in actions.

  ```swift
  Button("Favorite", systemImage: isFavorite ? "heart.fill" : "heart") {
      isFavorite.toggle()
  }
  .sensoryFeedback(.selection, trigger: isFavorite)
  ```

- Use the condition overload to fire only on specific transitions:
  `.sensoryFeedback(.success, trigger: phase) { old, new in new == .done }`.
- Fire sparingly: distinct feedback for selection, success and error; never
  several in quick succession.
- If the app offers a haptics setting, gate on it.

## Media

- Load remote images with `AsyncImage` (or the project's image loader), showing
  a placeholder until loaded. Use resized previews in rows; load full size only
  in the viewer.
- Open the viewer through one shared entry point (one `sheet(item:)` at the
  root driven by an `@Observable` state object), so it never presents twice.
- Make media tappable with a `Button`, and label it for VoiceOver.
