---
name: swift-ios
description: >
  SwiftUI and iOS engineering standards: state and data flow, view structure, navigation,
  layout, modern and deprecated APIs, accessibility, animation, performance, Swift concurrency,
  SwiftData and Liquid Glass. TRIGGER — load BEFORE writing or editing Swift, not after: any
  .swift, Package.swift, .xcodeproj or .xcworkspace file is touched, SwiftUI, iOS, iPadOS, Xcode
  or Swift concurrency is named, or the working directory holds an Xcode project or Swift
  package and the task is implementation, refactor, review or test-writing. Plain requests in an
  iOS repo ("add a screen", "fix this view", "why does this list flicker", "make this async")
  count; the user will not say "SwiftUI best practices". Load it for test files too, alongside
  swift-testing, and even when another skill (tdd, simplify) is active. SKIP only when no Swift
  file is involved and Swift is not named.
license: MIT
metadata:
  sources: >
    Merged and rewritten from MIT-licensed skills: twostraws/SwiftUI-Agent-Skill@1eac100,
    AvdLee/SwiftUI-Agent-Skill@204dba7, twostraws/Swift-Concurrency-Agent-Skill@bee3f69,
    AvdLee/Swift-Concurrency-Agent-Skill@d577081, twostraws/SwiftData-Agent-Skill@8051f17,
    Dimillian/Skills@05ba982.
allowed-tools: Read, Edit, Write, Glob, Grep, Bash(xcodebuild:*), Bash(swift:*), Bash(xcrun:*), Bash(git:*), Agent
---

# Swift iOS

House rules for SwiftUI apps on iOS and iPadOS. The rules below apply to every change; the
references hold the detail. Read only the references the task touches.

## Before you write

1. **Read the project's settings.** Deployment target (`IPHONEOS_DEPLOYMENT_TARGET` in the
   `.pbxproj`, `platforms:` in `Package.swift`), Swift language mode, default actor isolation
   (`SWIFT_DEFAULT_ACTOR_ISOLATION`, or `.defaultIsolation(MainActor.self)` in `swiftSettings`)
   and strict-concurrency level. Advice that depends on any of these is wrong until you know
   them. Keep the existing target; don't raise it to use an API.
2. **Read the neighbouring code.** Its architecture, folder layout and naming win over this
   skill's defaults. If the codebase uses view models, add view models.
3. **Check dated facts** in `references/platform-facts.md` before calling an API speculative or
   new. That file is the only place this skill names OS, Xcode or Swift versions as current.

## Operating rules

- **Architecture: MV.** Views own UI state with `@State`; app logic lives in `@Observable`
  models and services injected through the environment or initialisers. Don't add a per-screen
  view model unless one of the cases in `references/views-and-state.md` § "When a view model is
  justified" applies, or the codebase already uses them.
- **`@MainActor` follows the project's isolation.** In a MainActor-default target, don't annotate.
  Otherwise give UI-state `@Observable` models `@MainActor` and leave non-UI services
  (networking, parsing, persistence) unisolated. Never add `@MainActor`, `@unchecked Sendable`,
  `nonisolated(unsafe)` or `@preconcurrency` just to silence a diagnostic; find the isolation
  bug (`references/concurrency.md`).
- **Deprecated APIs:** in code the task touches, replace hard-deprecated APIs and flag
  soft-deprecated ones (`deprecated: 100000.0`, so the build stays quiet) without rewriting them.
  Never introduce either in new code. Don't scan or flag views the task doesn't touch
  (`references/modern-api.md` has the hard/soft table).
- **Availability:** gate any API newer than the deployment target with `#available` and give it a
  working fallback. A runtime check doesn't help if the SDK lacks the symbol, so match the
  project's Xcode too.
- **Liquid Glass:** keep the glass stock components give you. Don't fight it with opaque toolbar
  or tab-bar backgrounds. Don't hand-roll `glassEffect` or custom glass unless the user asks
  (`references/liquid-glass.md`).
- **Performance:** apply the routine rules in `references/performance.md` always. Restructure
  for performance (splitting observation, narrowing dependencies) only for a diagnosed problem
  or a requested deep performance review.
- Prefer SwiftUI to UIKit unless bridging is the only way. Don't add a third-party dependency
  without asking.
- One primary type per file. Group folders by feature.

## Correctness checklist

Treat a violation as a bug, in your own code and in code you review:

- [ ] `@State` and `@FocusState` are `private`; parent-owned inputs are not copied into `@State`
- [ ] `@Observable` over `ObservableObject` for new code; `@State` owns it, `@Bindable` for
      injected models that need bindings
- [ ] `@Binding` only where the child writes parent state
- [ ] No `Binding(get:set:)` in `body`; bind directly and react with `onChange`
- [ ] `ForEach` identity is stable: never `.indices` or `\.self` on mutable values; one view
      per element
- [ ] Tappable things are `Button`s with a text label (use `Button("Add", systemImage:)` for
      icon buttons), not `onTapGesture`
- [ ] `.animation(_:value:)`, never without `value:`; new code uses `foregroundStyle`, not
      `foregroundColor`
- [ ] Environment `@Entry` defaults are stable values, never `Model()`, `Date()` or `UUID()`
- [ ] No GCD in new code; `Task.sleep(for:)`, not nanoseconds
- [ ] `ModelContext` and `@Model` instances never cross actors; pass `PersistentIdentifier`
- [ ] Version-specific APIs are gated with `#available` and have fallbacks
- [ ] Previews use in-memory, self-contained data; no network or live store

## Where to look

| If you're working on… | Read |
|---|---|
| Property wrappers, data flow, `@Observable` models, environment, view decomposition, modifier order, async loading state, previews | `references/views-and-state.md` |
| `NavigationStack`/`NavigationSplitView`, routes, sheets, alerts, dialogs, tabs, toolbars, search, deep links | `references/navigation.md` |
| Stacks, grids, lists, forms, scroll views, safe areas, size classes, iPad multitasking, foldables | `references/layout.md` |
| Whether an API is deprecated, its replacement, modern Swift idioms | `references/modern-api.md` |
| VoiceOver, Dynamic Type, Reduce Motion, localisation, text formatting, HIG, theming, haptics | `references/accessibility-and-text.md` |
| Implicit/explicit animation, transitions, matched geometry, phase and keyframe animation | `references/animation.md` |
| Slow lists, janky scrolling, excess body evaluations, image decoding, a perf review | `references/performance.md` |
| `async`/`await`, actors, `Task`, cancellation, `Sendable` errors, Swift 6 migration, bridging callbacks | `references/concurrency.md` |
| `@Model`, `@Query`, predicates, relationships, indexes, CloudKit sync, `ModelActor` | `references/swiftdata.md` |
| Toolbars, tab bars or custom controls on iOS 26+, any `glassEffect` request | `references/liquid-glass.md` |
| Current OS, Xcode and Swift versions, new devices | `references/platform-facts.md` |
| Writing or reviewing tests | the `swift-testing` skill, if installed, alongside this one |

Not covered, so don't assume this skill has an answer: macOS scenes, windows and views; Swift
Charts; WebKit; document-based apps; styled text editing; focus management; Instruments trace
recording and analysis; Core Data.

## Feature work vs review

**During feature work or a bug fix**, make the change and stop. Don't rewrite unrelated code.
If the code you edited uses soft-deprecated APIs, offer the migration in one line at the end.

**When asked to review**, report only real problems, grouped by file:

1. The file and line(s).
2. The rule broken, named in one line (e.g. "Use `foregroundStyle()` instead of `foregroundColor()`").
3. A short before/after fix.

Skip files with no findings. End with a prioritised summary, most impactful first, each item
tagged with its area and severity, e.g. **Accessibility (high):** the add button has no label.
For a whole-app review, pick focus areas first and work through them one at a time.
