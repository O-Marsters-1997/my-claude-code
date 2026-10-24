# Navigation

Read when building or reviewing stacks, split views, tabs, sheets, alerts, toolbars, search or deep links.

Contents
- Stacks and typed routes
- Routers
- Split views
- Tabs
- Sheets and full-screen covers
- Alerts and confirmation dialogs
- Popovers and inspectors
- Toolbars
- Search
- Deep links

## Stacks and typed routes

- Use `NavigationStack` or `NavigationSplitView`. `NavigationView` is deprecated: replace it in code the task touches.
- Navigate by value: `NavigationLink("Profile", value: Route.profile)` plus `navigationDestination(for:)`. Replace `NavigationLink(destination:)` where it should become value-based.
- Never mix `navigationDestination(for:)` and `NavigationLink(destination:)` in one hierarchy.
- Register `navigationDestination(for:)` once per type per stack. Competing registrations in one stack are a bug; separate stacks (one per tab) may each register the same type.
- Routes are a `Hashable` enum carrying lightweight data (ids, small values), never view instances or heavy state.
- Programmatic navigation binds `NavigationStack(path:)` to `[Route]` (one route type) or `NavigationPath` (mixed types). Push with `append`, pop to root with `removeAll()`.
- Map routes to views in one place, e.g. a `withAppRouter()` modifier applied once per stack, rather than repeating the `switch` per screen.
- Reset paths when app context changes (logout, account switch).
- Keep the path in scene-root `@State` or a scene-owned router so each window navigates independently. Make the route `Codable` and persist the path with `@SceneStorage` when state restoration matters.

## Routers

- A router is an `@Observable` class owning `path` and, optionally, `presentedSheet`. It is app navigation state, not a view model. Apply the `@MainActor` rule from `views-and-state.md`.
- Give each tab its own router and its own `NavigationStack`. Share one path only if you want global history.
- Own the router with `@State` at the tab or scene root and inject it with `.environment(router)` so descendants navigate without prop-drilling.
- Bind to an environment router with a local `@Bindable var router = router` and `$router.path`, not `Binding(get:set:)`.
- Keep the router out of other `@Observable` models.
- Small apps with one stack and a couple of destinations do not need a router; `@State` path plus `navigationDestination` is enough.

## Split views

- Use `NavigationSplitView` for sidebar or list-detail layouts. It collapses to a stack in compact width.
- Drive it with selection: `List(items, selection: $selectedID)`, then resolve the detail from the id. Show `ContentUnavailableView` when nothing is selected.
- Respond to the space offered, not the device family. iPhones can present regular width, and the same scene can change between compact and regular. Keep selection consistent across the change: re-fill selection when it expands, clear regular-only selections on collapse so compact returns to the list.
- Put a `NavigationStack` in the detail column for deeper pushes.
- Configure with `columnVisibility:` (`.all`, `.doubleColumn`, `.detailOnly`), `preferredCompactColumn:`, `.navigationSplitViewStyle(.balanced | .prominentDetail)` and `.navigationSplitViewColumnWidth(min:ideal:)`. Avoid a `max:` that stops the system widening the sidebar.
- When hiding the sidebar would strand the user, use `columnVisibility` `.all` with `.toolbar(removing: .sidebarToggle)`.
- Only build a custom `HStack` column layout when `NavigationSplitView` cannot express the design (a column independent of selection). You then lose system collapse, glass and toolbar behaviour; say so.

## Tabs

- Use `TabView(selection:)` with `Tab("Home", systemImage: "house", value: .home)` bound to an enum, never an `Int` or `String`.
- Centralise tab identity in an `AppTab` enum (`CaseIterable`, `Identifiable`, stable `id`) that supplies label and content.
- No view models per tab; keep state local or in `@Observable` services.
- Group sidebar tabs with `TabSection` and `.tabPlacement(.sidebarOnly)`; pin one tab with `.tabPlacement(.pinned)`.
- `.tabViewStyle(.sidebarAdaptable)` morphs tab bar and sidebar on iPad. To ask whether a sidebar is available, read `@Environment(\.isTabViewSidebarAvailable)`, not size classes; when none is, keep nested content reachable another way.
- For a tab that triggers an action instead of selecting (compose), intercept in the setter of a computed selection binding and leave selection unchanged. This is the one place a custom binding is warranted; build it outside `body` (computed property or method).
- Tab content gets the stock tab bar and its Liquid Glass. Do not replace it with a custom bar or paint opaque backgrounds behind it.

## Sheets and full-screen covers

- Present optional data with `sheet(item:)`, not `sheet(isPresented:)` plus an `if let` inside.
- When the sheet view's only init parameter is the item, write `sheet(item: $item, content: EditView.init)`.
- Several sheets from one screen: one `Identifiable` enum, one `@State var sheet: Sheet?`, one `.sheet(item:)` with a `switch`. Not N Booleans and N modifiers. Never mix `sheet(isPresented:)` and `sheet(item:)` for the same concern.
- Mutually exclusive sheets can share an `id`; independent ones need distinct ids.
- App-wide sheets: store `presentedSheet` on the router and attach a single `withSheetDestinations(sheet: $router.presentedSheet)` at the root. Apply `.environment(router)` and other dependencies so the sheet content receives them; verify required environment values are visible inside the sheet.
- Sheets own their actions: read `@Environment(\.dismiss)` and save through an injected store, then `dismiss()`. Do not thread `onSave`/`onCancel` closures from the parent.
- Wrap sheet content in its own `NavigationStack` when it needs a title, toolbar or pushes. Use `.cancellationAction` and `.confirmationAction` placements for Cancel and Save.
- Keep sheet payloads light: ids or small models.
- `fullScreenCover` only for immersive or blocking flows; it has no swipe-to-dismiss.
- Zoom from a source: `.matchedTransitionSource(id:in:)` on the trigger and `.navigationTransition(.zoom(sourceID:in:))` on the destination (iOS 18+).

## Alerts and confirmation dialogs

- Use the action-based `alert(_:isPresented:actions:message:)`; never the old `Alert`-returning `alert(item:)`.
- When presentation concerns one optional value, use the item overloads `alert(_:item:actions:message:)` or `confirmationDialog(_:item:actions:message:)`. They pass the unwrapped value and reset to `nil` on dismissal, with no `Identifiable` requirement. They need a recent SDK but back-deploy to iOS 15 (iOS 16 with a `LocalizedStringResource` title); check the toolchain, no `#available` needed.
- An alert with only a dismissing "OK" button needs no actions: `.alert("Saved", isPresented: $showSaved) { }`.
- Attach `confirmationDialog` to the control that triggers it so the presentation animates from the right source.
- Exception: a dialog attached to a button inside `ToolbarOverflowMenu` never appears. Attach it outside the menu and drive it with state.
- Use `role: .destructive` and `role: .cancel` on buttons, not custom colours.

## Popovers and inspectors

- `popover` adapts to a sheet in compact width; add `.presentationCompactAdaptation(.popover)` only when it must stay a popover.
- `inspector(isPresented:)` gives a trailing column in regular width and a sheet in compact. Size with `.inspectorColumnWidth(min:ideal:max:)`. Toggle it from a toolbar button.

## Toolbars

- Attach `.toolbar` to content inside a system `NavigationStack`, `NavigationSplitView` or `TabView`. System containers supply Liquid Glass, overflow and adaptive layout; a hand-built bar gets none of it. Do not set opaque toolbar backgrounds that fight the glass.
- Use semantic placements (`.cancellationAction`, `.confirmationAction`, `.primaryAction`, `.topBarTrailing`, `.bottomBar`) instead of positioning by physical edge.
- Give every action a title and icon, `Button("Share", systemImage: "square.and.arrow.up", action: share)`, even when only the symbol shows. The title drives accessibility and overflow.
- Group toolbar content with `ToolbarSpacer(.fixed | .flexible)` (iOS 26+). Opt an item out of the shared glass background with `.sharedBackgroundVisibility(.hidden)` (iOS 26+). Show counts with `.badge(_:)`.
- Place system items with `DefaultToolbarItem(kind: .search | .sidebarToggle, placement:)` (iOS 26+).
- User-customisable toolbars use `.toolbar(id:)` with a stable, meaning-based id on every `ToolbarItem`. Prefer `defaultCustomization(_:options:)` over the soft-deprecated `showsByDefault:` (flag, do not change unasked).
- Dynamic toolbar content: `ForEach` with stable element identity.
- Title menus: `ToolbarTitleMenu { ... }` for filtering or context switching only. Keep it short, grouped with `Divider`, and never put destructive actions in it.
- Title plus subtitle: prefer `navigationSubtitle(_:)` where available; otherwise a `.principal` item with headline and secondary caption, with `.navigationBarTitleDisplayMode(.inline)`.
- A custom bar above scroll content (filter pills): `safeAreaBar(edge: .top)` on iOS 26+; fall back to `safeAreaInset(edge: .top, spacing: 0)`. Do not stack several top insets.
- Newer overflow and adaptive-bar APIs (`visibilityPriority`, `ToolbarOverflowMenu`, `.topBarPinnedTrailing`, vertical-bar `axisBehavior`, minimisation behaviours) need iOS 27+; check each API's availability against the deployment target. Wrap the whole toolbar content in one `if #available` with the plain toolbar as the fallback:

```swift
.toolbar {
    if #available(iOS 27, *) {
        ToolbarItem(placement: .topBarPinnedTrailing) { ShareButton() }
        ToolbarOverflowMenu { ExportButton() }
    } else {
        ToolbarItem { ShareButton() }
    }
}
```

- Leave adaptive toolbar behaviour on `.automatic`; override axis or vertical behaviour only for a specific measured problem.
- Hide the status bar with `statusBarHidden(_:)` as the fallback for `toolbarVisibility(.hidden, for: .statusBar)` where that is unavailable.

## Search

- Bind `searchable(text:)` to local `@State`. Add `.searchScopes($scope) { ... }` for modes.
- Run queries in `.task(id: query)`: debounce with `try? await Task.sleep(for: .milliseconds(250))`, check `Task.isCancelled`, clear results for an empty query, treat `CancellationError` as normal.
- Show progress while loading and `ContentUnavailableView.search` for no results.
- Keep the system search field. Use `.searchToolbarBehavior(.minimize)` (iOS 26+) for a compact search button; a dedicated search tab uses `Tab(role: .search)`.
- Never block the main actor while fetching.

## Deep links

- One entry point. Handle `.onOpenURL` once at the root and route through the router (`handle(url:) -> OpenURLAction.Result`).
- Validate before trusting: parse into a `Route` and ignore or fall back on anything unrecognised. Never assume a URL is internal.
- Intercept in-app links by installing `.environment(\.openURL, OpenURLAction { router.handle(url: $0) })`. Return `.handled` when routed, otherwise `.systemAction`.
- Deep-linking selects the right tab, resets or builds that tab's path, then appends routes. Do not push onto whatever stack happens to be visible.
- Resolve remote lookups in a `Task` without blocking UI; show loading in the destination.
