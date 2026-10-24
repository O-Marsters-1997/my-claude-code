# Platform facts

Read this when a rule depends on which OS, Xcode, Swift or device is current.

Last checked: 2026-10-24 against twostraws/SwiftUI-Agent-Skill@1eac100, AvdLee/SwiftUI-Agent-Skill@204dba7.
Update this file, and only this file, when a new OS, Xcode or device ships.

## Current releases

- iOS 27 is shipping and is Xcode's default deployment target for new apps.
- Xcode 27.1 and the 27.1 platform releases are available. Do not dismiss their
  APIs as speculative. AvdLee still labelled the 27.1 SDK beta when last
  checked; recheck semantics of 27.1-only APIs against the shipping SDK.
- Target Swift 6.4 or later with strict concurrency.
- Xcode 27.2 pairs with iOS 26.4+ for toolbar-item `.hidden()`; not in 27.1.

## Toolchain-only changes (Xcode 27, no deployment-target gate)

- `@ContentBuilder` is the new spelling of `@ViewBuilder`; both name one type.
- `@State` is a macro; it keeps its initial expression for one-time evaluation
  only when the property is `private`.
- `ForEach` can generate items directly inside `toolbar()`.
- The compiler diagnoses reference-type `@Entry` defaults and stored closures
  used as environment actions.
- View-argument `overlay`/`background`/`mask` can fail to type-check because
  `blendMode` becomes ambiguous; use the builder closure.
- Xcode 27.1: `.inspector` attached to a pushed screen breaks pushes and puts
  full-bleed content under the inspector. Attach it around the `NavigationStack`.

## iPhone Duo (folding iPhone)

- Two displays: a small outer display, broader and much shorter than other
  iPhones (near square), and a large inner display with a central fold.
- Identifies as a phone (`.phone` idiom). Outer display in portrait is compact
  width / regular height; inner display gives regular width and height,
  whatever orientations the app supports, and can still be landscape.
- Opening, closing, rotating, partial folding, Split View and pinned video all
  resize the same scene. No relaunch, no `scenePhase` change.
- With two windows open, closing the device keeps the most recent one active
  and backgrounds the other.
- The outer display's four corners differ, so no fixed radius matches it.
- The outer camera always reserves space; the inner FaceTime camera occludes
  content only while in use. Touches are suppressed over active camera regions.
- System bars can merge status, toolbar and tabs into one vertical bar on the
  outer display. Its side follows device orientation, not layout direction.
- The outer display cannot host app content while the device is open, except a
  camera app during capture.
- Hinge angle runs 180 (fully open) to 0 (closed), per-degree accuracy today;
  Apple calls this policy and may change it.
- Build toolchain decides the experience: Xcode 26 builds run in an
  iPhone-shaped compatibility area; Xcode 27.0 expands inner-display use;
  Xcode 27.1 enables the full display and vertical bars.
- Duo APIs (`ArrangementView`, `reservedRegions`, vertical-bar modifiers,
  `onHingeChange`) need the iOS 27.1 SDK and runtime.

## Testing a fold

- `simctl` cannot fold or rotate the Duo simulator; Device Hub needs manual use.
- RocketSim's CLI can: `rocketsim duo pose closed|book|open`, `rocketsim duo hinge`.
- Apple docs: developer.apple.com/iphone-duo, HIG "Designing for iPhone Duo".
