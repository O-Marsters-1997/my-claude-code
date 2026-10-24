# Layout

Read this when building or reviewing screen layout, safe areas, lists, forms,
grids, scrolling, or behaviour as the window resizes or a device folds.

## Contents

- Adapt to the space offered
- Sizing and geometry
- Safe areas and bars
- Presentations at any size
- Foldables and reserved regions (iOS 27.1+)
- Lists and identity
- Forms and controls
- Scrolling
- Grids
- Overlays, loading and input bars
- Scroll-reveal surfaces
- Checking layout changes

Device facts (display shapes, cameras, hinge range) live in `platform-facts.md`.

## Adapt to the space offered

- Treat window size as changeable for the scene's whole life: iPad resizing,
  Split View, Stage Manager, sheets, folding phones.
- Size from the proposed size. Never use `UIScreen.main`, a cached screen size,
  device idiom or orientation to pick a layout. A portrait-only app can still
  get a regular-width window.
- Pick layout variants with `@Environment(\.horizontalSizeClass)` /
  `verticalSizeClass`. Test `== .regular` for the expanded layout; size classes
  are optional and `!= .compact` treats `nil` as spacious.
- Read size classes in the `View` or `ViewModifier` nearest the decision. Never
  cache them in an `App`, `Scene` or model; they go stale mid-resize. In a
  representable, read `context.environment.horizontalSizeClass`.
- Before replacing an idiom check, ask what it decides. Available space: use a
  size class. Genuine platform or capability: keep it, do not invent a mapping.
- Flag forced size-class overrides (`.environment(\.verticalSizeClass, .regular)`);
  they stop descendants adapting.
- Use size classes for what to show, the proposed size for how big to draw it.
  For breakpoint logic, observe `onChange(of: horizontalSizeClass)`, not every
  `geometry.size` step.
- Use `ViewThatFits` when a compact alternative should replace a layout that
  overflows (an `HStack` of actions falling back to a `VStack`).
- Use `AnyLayout` to switch stack direction when the children own `@State`; a
  plain `if` rebuilds them and drops state.
- Do not branch between containers that already adapt (`TabView`,
  `NavigationSplitView`). Swapping them mid-resize resets navigation state.
- Keep selections and navigation paths above any size-class branch.
- Give every screen that assumes its content fits (onboarding, purchase,
  account setup) a `ScrollView` or `ViewThatFits` fallback. Short, wide windows
  push bottom actions out of reach.
- Every feature must exist at every size. More width may show more of the
  hierarchy (list and detail together); it must not add or remove capability.
- Cap long-form text width with `.frame(maxWidth:)` at wide sizes.
- Build views that work anywhere: full screen, sheet, popover or embedded. A
  view never assumes it owns the screen.

## Sizing and geometry

- Prefer `containerRelativeFrame()`, `visualEffect()`, `onGeometryChange()` or
  a custom `Layout` over `GeometryReader`. `GeometryReader` is not deprecated;
  keep it where layout itself depends on measured size.

  ```swift
  // Before
  GeometryReader { geo in HeaderView().frame(height: geo.size.height * 0.2) }
  // After
  HeaderView().containerRelativeFrame(.vertical) { length, _ in length * 0.2 }
  ```

- Never nest `GeometryReader`s or chain preferences for one measurement.
- When a measurement drives state, reduce it to the smallest `Equatable` value
  (a `Bool` past a threshold) so the action fires rarely.
- No device-sized frames (`.frame(width: 390, height: 844)`) and no magic
  heights that only fit one screen.
- Fill width with `.frame(maxWidth: .infinity, alignment: .leading)`, not an
  `HStack` plus `Spacer`.
- Keep hierarchies flat; deep stack nesting multiplies layout passes.
- A custom view owns its static container (`HStack` of icon and title). The
  caller owns lazy or repeated containers (`LazyVStack` around `ForEach`).
- iOS 26+: use `ConcentricRectangle()` for shapes touching a screen or sheet
  edge so corners follow the container. Elsewhere `.rect(cornerRadius:)` is fine.
- Use `.leading` / `.trailing`, never `.left` / `.right`.
- Prefer flexible frames and system spacing. Hard-code padding or spacing only
  when the design asks.

## Safe areas and bars

- Delete padding that imitates a system inset (a fixed bottom gap for the home
  indicator). SwiftUI already respects the safe area and its size varies.
- `GeometryProxy.size` already excludes the safe area. Subtracting
  `safeAreaInsets` from it, or re-adding them as padding, counts them twice.
  Do not use `GeometryReader` only to read and reapply insets.
- Pick the modifier by content:
  - Bar content (custom control bar): `safeAreaBar(edge:)` on iOS 26+, which
    reserves space and gives scroll-edge treatment. Fallback: `safeAreaInset`.
  - Other pinned content (notices, input bars): `safeAreaInset(edge:)`.
  - A deliberate extra design margin: `safeAreaPadding`. Never use it as a
    stand-in for a bar or device inset.
- Move an interactive bar sitting in a `ZStack` or overlay over content into
  `safeAreaBar` / `safeAreaInset`. Leave full-bleed decoration where it is.
- Scope `ignoresSafeArea` to edges: `.ignoresSafeArea(edges: .top)` for content.
  Unscoped is for backgrounds, gradients, images and scrims only.
- iOS 26+: extend artwork behind sidebars, inspectors or vertical bars with
  `backgroundExtensionEffect()`, at most one per screen. `ignoresSafeArea` can
  hide part of the image under the bar.
- Use `TabView` and system toolbars. Custom bars do not relocate, overflow,
  clear cameras or pick up system glass.
- Keep the system glass on stock bars and controls. Do not paint opaque
  backgrounds behind them or hand-roll `glassEffect` unless asked.
- Give each tab and toolbar action a title and an icon; icon-only bars still
  need titles for overflow menus and VoiceOver.
- In `NavigationSplitView`, attach toolbars to the column whose content they act on.
- iOS 27+: on wide iPhone layouts, tabs become a sidebar only with
  `.tabViewStyle(.sidebarAdaptable)` plus `.defaultTabBarPlacement(.sidebar)`.

## Presentations at any size

- Use system sheets with `presentationDetents()`, popovers and alerts. A sheet
  built from a `ZStack`, offsets and a drag gesture does not adapt.
- Use `inspector(isPresented:content:)` for a side panel: a trailing column at
  regular width, a sheet at compact. Attach it around the `NavigationStack`,
  not on a pushed screen.
- iOS 27+: `.presentationPlacement(.leading | .trailing)` only when side
  placement is asked for or a centred sheet hides content that must stay
  visible. Sheets only. Keep the old presentation in the fallback.

## Foldables and reserved regions (iOS 27.1+)

Add none of these unless the task asks for the behaviour or testing shows a
real overlap. System containers already handle the fold and cameras. Gate with
`#available(iOS 27.1, *)`, put declarations naming new types in
`@available(iOS 27.1, *)` scope, and keep an equivalent older layout.

Choose by screen structure, in this order:

1. Rows that push screens (settings, mail, folders): `NavigationSplitView`.
   No fold code.
2. Cards or a feed in one `ScrollView`: reflow to two columns with the gutter
   over a vertical fold. Never split at a horizontal fold.
3. Two peer regions without navigation (media and controls): `ArrangementView`
   with `.split`, outside any scroll view.
4. Content plus a supplementary panel: keep the compact pattern (a bar that
   pushes the full view), or an `.inspector` around the stack.
5. Custom edge-to-edge chrome only: read `reservedRegions` directly.

Two-column reflow:

- Write a custom `Layout` reading each child's column from a `LayoutValueKey`;
  children with no column span full width. Switch via `AnyLayout` when cards
  own state.
- Keep reading and VoiceOver order equal to the single-column order. Keep
  accessibility Dynamic Type sizes in one column.
- Under MainActor default isolation, mark `LayoutValueKey` types and helpers
  called from `Layout` methods `nonisolated` if the compiler complains.

`ArrangementView`:

- `.arrangementViewStyle(.split)` for peers that must not overlap; `.overlay`
  where temporary overlap is fine. Restrict `axes(_:)` only when the content
  demands one direction.
- Never put it inside `List`, `ScrollView` or `NavigationSplitView`, or a
  `NavigationSplitView` inside it. Wrap it in a `NavigationStack` if it needs
  pushes. Each region scrolls itself.
- When the split cannot fit, only the primary shows. Give essential secondary
  content another route and use `layoutPriority()`.
- Prefer one arranged branch that varies its axis over switching to another
  layout; switching rebuilds both regions and drops state.
- Read `\.splitArrangementAxis` and `\.overlayArrangementZIndex` to adapt;
  neither tells you the other child is hidden.

Reserved regions:

- Read `proxy.reservedRegions(kind: .division)` for the fold and
  `.occlusion` for cameras, from a `GeometryReader` or `onGeometryChange()`
  proxy on the laid-out view.
- Filter on `isActive` before moving content. Pass `options: .includeInactive`
  when layout must not jump while folding; inactive frames can be zero-sized.
- Move the smallest coherent group into one area; never hardcode fold
  positions. Keep controls out of active camera frames.
- Keep `layoutDirectionBehavior` at its `.mirrors` default for UI content.
- For column layouts keep only vertical divisions (`height > width`); grids use
  an even column count whenever a division exists.

Vertical bars:

- Read `@Environment(\.toolbarVerticalEdge)` to align custom UI to the bar;
  never assume a fixed edge. It is `nil` where no vertical bar exists.
- Use `axisBehavior(.horizontalOnly)` sparingly; a kept top bar costs height.
- `toolbarVerticalBehavior(.disabled)` only for: a sheet with just Done/Close,
  a bottom-heavy full-width UI (dial pad, calculator), or a full-screen video
  player. Never propose adding it elsewhere; never toggle it at runtime.
- `toolbarVerticalCompressionBehavior(.prefersToolbarItems)` only when a
  focused workflow must keep its actions over tabs.

Hinge: `onHingeChange` is for effects and interactions, never layout.
`context.hinge` is optional; reset effect state when it is `nil` or its status ends.

Do not plan simultaneous content on both displays. For scene accessories, track
`onAvailabilityChange` and handle failed scene requests; never infer
availability from device pose.

## Lists and identity

- Use `List` for long repeated rows, feeds and settings. Use `ScrollView` +
  `LazyVStack` for custom or chat layouts. Do not nest `List` and a same-axis
  `ScrollView`.
- Give every `ForEach` (and data-driven `List`, `Picker`, `Table`,
  `OutlineGroup`) a stable, unique, cheap id:
  - never `.indices`, `\.offset` or `id: \.self` on a large struct;
  - never derive `id` from a mutable property (editing it resets the row);
  - never mint ids inside `body` (`items.map { Item(...) }`);
  - do not conform a type to `Identifiable` just for `ForEach`; pass `id:`.
- A class's default `ObjectIdentifier` id is unique only for its lifetime.
- Keep each row unary: one top-level view per element. Wrap branching content
  in a `VStack` (not `Group`), and filter out elements that should not be rows
  before `ForEach`. Launch with `-LogForEachSlowPath YES` to find offenders.
- No `AnyView` in rows.
- Do not filter or sort expensively inline in `ForEach`; compute in the model
  or cache. Cheap slices like `prefix(n)` are fine.
- `.enumerated()` is fine with `id: \.element.id`. Pass the sequence directly
  on iOS 26+; wrap in `Array(...)` below that.
- Custom list surface: `.scrollContentBackground(.hidden)` then `.background(...)`.
  Without the first, a background has no visible effect.
- Tune rows with `.listRowInsets`, `.listRowSeparator(.hidden)`,
  `.listRowBackground`; set `.environment(\.defaultMinListRowHeight, 1)` for
  dense custom rows.
- Use `.refreshable { await ... }` for pull to refresh.
- Use `ContentUnavailableView` for empty and error states. Under `searchable`,
  `ContentUnavailableView.search` fills in the query itself.
- Make whole rows tappable with a `Button` and `.contentShape(.rect)`.
- iOS 27+: drag reordering via `.reorderable()` on `ForEach` plus
  `.reorderContainer(for:)` on the container. Apply `ReorderDifference` by
  removing the moved items in source order, then inserting them at the
  destination. Rebuilding from a `Set` loses order.
- iOS 27+: `swipeActions` outside `List` need `.swipeActionsContainer()` on the
  scroll container. Inside `List` they need no gate.
- `Table` does not sort itself: bind `sortOrder` and re-sort in `onChange`. On
  compact width only the first column shows, so put combined info there.

## Forms and controls

- Use `Form` for settings and data entry, grouped in `Section`s. Very custom
  layouts belong in `ScrollView` + `VStack`, not `Form`.
- Add a `NavigationStack` only when the form has no navigation context (a sheet).
- Pick one background strategy: stock styling, or `.scrollContentBackground(.hidden)`
  with theme colours and `.listRowBackground`.
- Bind controls straight to state (`@State`, `@Binding`, `@Bindable`, `@AppStorage`).
- `Toggle` for booleans; `Picker` for discrete choices; `.segmented` only for
  2-4 options, menu or inline beyond.
- Wrap a `Slider` in `LabeledContent` inside a `Form`, and show its value with
  a format style, never `String(format:)`.
- Never hide control labels; use `.labelsHidden()` only when visible context remains.
- Reflect locked settings with `.disabled(...)`.
- Use `.scrollDismissesKeyboard(.interactively)` (or `.immediately`) on
  input-heavy screens, and `@FocusState` for focus.

## Scrolling

- Choose the API by target:
  - iOS 18+: `scrollPosition(_:)` with `ScrollPosition` to scroll by id,
    offset or edge; `onScrollGeometryChange(for:of:action:)` to observe.
  - iOS 17+: `scrollPosition(id:)` with an optional id binding.
  - Earlier, or proxy jumps: `ScrollViewReader` + `scrollTo(id, anchor:)`.
- Add `.scrollTargetLayout()` to the stack holding identified views.
- Extract the smallest value from scroll geometry. Return a `Bool` past a
  threshold rather than raw `contentOffset`, which changes every frame.

  ```swift
  .onScrollGeometryChange(for: Bool.self) { geo in
      geo.contentOffset.y + geo.contentInsets.top > 50
  } action: { _, isPast in showHeader = !isPast }
  ```

- Scroll-driven visuals (fade, parallax): `visualEffect` reading
  `geometry.frame(in: .scrollView)`, not state updates.
- Paging: `.scrollTargetBehavior(.paging)` with `.containerRelativeFrame(.horizontal)`
  pages. Snap: `.viewAligned` plus `.contentMargins`.
- Hide indicators with `.scrollIndicators(.hidden)`.
- Use lazy stacks for large or unbounded content; plain stacks for small fixed content.
- Never nest scroll views on the same axis.
- Animate programmatic scrolls with `withAnimation` when the jump should be seen.

## Grids

- `LazyVGrid` for large collections; `.adaptive(minimum:)` columns to scale
  across widths, several `.flexible` columns for a fixed count.
- Square cells: use `.aspectRatio(1, contentMode: .fit)` on a container and
  fill it, not a `GeometryReader` per cell.

  ```swift
  Color.clear
      .aspectRatio(1, contentMode: .fit)
      .overlay { thumbnail.resizable().scaledToFill() }
      .clipShape(.rect(cornerRadius: 6))
  ```

- Keep gutters small and consistent. Do not nest grids without cause; keep
  per-cell overlays light.
- Inside a two-column reflow, nested grids may need fewer columns.

## Overlays, loading and input bars

- Transient UI (toasts, banners): `.overlay(alignment: .top)` so layout does
  not move. Show one at a time; queue or replace. Do not block interaction.
- Auto-dismiss with structured concurrency, not GCD:

  ```swift
  .task(id: toast) {
      try? await Task.sleep(for: .seconds(2))
      withAnimation { toast = nil }
  }
  ```

- Loading: keep the real layout with placeholder data and
  `.redacted(reason: .placeholder)`, 3-6 rows. Then content, or
  `ContentUnavailableView` if empty. Use `ProgressView` only for short global
  work; one per section at most. Keep frames stable while redacted.
- Bottom input bars (chat, composer): `.safeAreaInset(edge: .bottom)` on the
  scroll view, never inside the scroll content. Pair with
  `.scrollDismissesKeyboard(.interactively)`; make send reachable from the
  return key.

## Scroll-reveal surfaces

For a detail screen with a primary surface that reveals secondary content on
scroll:

- Use one vertical `ScrollView` with `.scrollTargetBehavior(.paging)`: a
  viewport-height primary section, then the secondary section.
- Derive one normalised `progress` from the scroll offset divided by the
  measured secondary height (clamp to 0...1, guard zero). Drive offset,
  opacity, blur, scale and toolbar state from it alone.
- Do not keep parallel booleans (`isExpanded`); derive them from `progress` or
  `onScrollTargetVisibilityChange`.
- One property, one animation source.
- A control that moves between sections is one overlay interpolated between
  two anchors (`anchorPreference` + `overlayPreferenceValue`), not two copies.
- Disable scrolling while a conflicting gesture (zoom, crop) is active, and
  disable hit testing on overlays that have faded out.

## Checking layout changes

Check each screen, sheet and popover at the extremes: narrowest, widest,
shortest, tallest, roughly square, Split View, and right-to-left. For foldable
work, test both displays and all rotations; see `platform-facts.md` for
simulator tooling. Say which cases you could not test and ask the user to
cover them.
