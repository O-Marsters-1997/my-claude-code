# Liquid Glass

Read when touching toolbars, tab bars, sheets, navigation chrome or anything glass-like, and before writing any `glassEffect` code.

Contents
- Default: let the system do it
- Stock components and what to leave alone
- When the user asks for custom glass
  - Availability and fallback
  - Applying glass
  - Grouping with GlassEffectContainer
  - Morphing and unions
  - Buttons
  - Review checklist

## Default: let the system do it

- Stock components (`NavigationStack` bars, `TabView`, toolbars, sheets, menus, search, standard controls) get Liquid Glass automatically when built with the iOS 26 SDK on iOS 26+. Keep it.
- Never fight system glass with opaque backgrounds:
  - no `.background(Color.white)` (or any solid colour) behind toolbars, tab bars or bar items;
  - no `.toolbarBackground(.visible, ...)` with a solid colour, or `.toolbarBackground(Color...)`;
  - no custom darkening or gradient views behind bars; they clash with the automatic scroll edge effect;
  - no `UINavigationBarAppearance`/`UITabBarAppearance` opaque configuration to restore an old look.
- Never hand-roll `glassEffect`, `GlassEffectContainer`, glass button styles, or blur-and-stroke "glass" imitations unless the user explicitly asks for custom glass. Do not convert existing UI to glass proactively.
- In code the task touches, remove the opaque overrides above and report them. Elsewhere, flag them.

## Stock components and what to leave alone

- Partial-height sheets get a glass background by default. Remove a custom `presentationBackground(_:)` unless the design needs it.
- Toolbar icons render monochrome. Use `.tint` on a toolbar item only to carry meaning (a primary call to action), not decoration.
- Let the system scroll edge effect keep bar content legible. Adjust it with `.scrollEdgeEffectStyle(_:for:)` (iOS 26+) rather than adding views behind the bar.
- Group and separate toolbar items with `ToolbarItemGroup` and `ToolbarSpacer` (iOS 26+), not custom backgrounds. Opt an item out of the shared glass with `.sharedBackgroundVisibility(.hidden)` (iOS 26+).
- Tab bar behaviours (`.tabBarMinimizeBehavior(_:)`, `.tabViewBottomAccessory`) are system features; prefer them to custom floating bars.
- For content that should extend under a sidebar or inspector, use `.backgroundExtensionEffect()` (iOS 26+).
- Gate every iOS 26+ modifier above with `#available` when the deployment target is lower.

## When the user asks for custom glass

Only when the user explicitly asks for glass on custom views. Use the native APIs below; never imitate glass with blur, materials and strokes on iOS 26+.

### Availability and fallback

- All glass APIs are iOS 26+. If the deployment target is lower, gate with `#available(iOS 26, *)` and fall back to a material in the same shape:

```swift
if #available(iOS 26, *) {
    content.glassEffect(.regular, in: .rect(cornerRadius: 16))
} else {
    content.background(.ultraThinMaterial, in: .rect(cornerRadius: 16))
}
```

- For repeated use, wrap that in one `@ViewBuilder` extension (for example `glassOrMaterial(in:)`); the `#available` branch is constant at runtime so it does not churn identity.
- Fallback for grouped glass controls: the stock control (`Picker(.segmented)`, `.buttonStyle(.bordered)`), not a hand-made glass look.

### Applying glass

- `.glassEffect(_:in:)` defaults to `.regular` in a capsule. Shapes: `.capsule`, `.circle`, `.rect(cornerRadius:)`, `.rect(cornerRadius:style:)`, or a custom `Shape`.
- `Glass` variants: `.regular` (default), `.clear` (for media-rich backgrounds), `.identity` (no glass, for toggling without changing the hierarchy). There is no `.prominent`; express emphasis with `.tint(color)`, adjusting the tint's opacity.
- Add `.interactive()` only to elements that respond to touch (buttons, tappable chips). Never on static content.
- Apply `.glassEffect` after layout and appearance modifiers (font, foreground, padding, frame) so the glass covers the final size:

```swift
Label("Edit", systemImage: "pencil")
    .font(.headline)
    .padding()
    .glassEffect(.regular.interactive(), in: .capsule)
```

- Toggle glass with the `isEnabled:` parameter or `.identity`, not by branching the view.
- Use glass sparingly: controls and floating surfaces over content, not every card or list row. Keep shapes, corner radii and tints consistent across a feature.
- Put glass on the navigation/control layer, not on content itself. Never stack glass on glass.

### Grouping with GlassEffectContainer

- Wrap any set of nearby glass elements in one `GlassEffectContainer`. Glass cannot sample other glass, so siblings outside a shared container render inconsistently; the container also renders them in one pass and enables blending and morphing.
- Set the container's `spacing` equal to the layout's spacing. Larger spacing makes shapes merge from further apart.

```swift
GlassEffectContainer(spacing: 16) {
    HStack(spacing: 16) {
        ForEach(tools) { tool in
            ToolButton(tool: tool)
                .glassEffect(.regular.interactive(), in: .circle)
        }
    }
}
```

- Do not nest containers without a reason.

### Morphing and unions

- To morph glass as views appear, disappear or swap: shared `@Namespace`, a `.glassEffectID(_:in:)` on each glass view, all inside one `GlassEffectContainer`, and the state change inside `withAnimation` (or `.animation(_:value:)` on the container).
- IDs must be stable model identifiers, not indices or `UUID()`.
- Choose the transition with `.glassEffectTransition(_:)` (`.matchedGeometry` default, `.materialize` for views without a counterpart).
- To render separate views as one glass shape, give them the same `.glassEffectUnion(id:namespace:)`. Use it for dynamically built groups or views that are not adjacent in one stack.
- Honour Reduce Motion: with `accessibilityReduceMotion` on, change state without the spring (or with a short fade).

### Buttons

- Prefer `.buttonStyle(.glass)` for secondary actions and `.buttonStyle(.glassProminent)` for the primary one over a manual `glassEffect` on a `Button`.
- Use manual `.glassEffect(.regular.interactive(), in: shape)` only when the button needs a custom shape or tint the styles cannot give.
- Keep hit targets at least 44x44 pt.

### Review checklist

- User asked for custom glass; otherwise none is hand-written.
- No opaque overrides on system bars, tabs or sheets.
- `#available(iOS 26, *)` with a material or stock fallback when the deployment target is lower.
- Grouped glass inside one `GlassEffectContainer`, spacing matching the layout.
- `.glassEffect` after layout and appearance modifiers.
- `.interactive()` only on interactive elements.
- Morphing uses `glassEffectID` with a shared `@Namespace` and an animated state change.
- No `.prominent`; emphasis via tint.
- Consistent shapes and tints; glass used sparingly.
