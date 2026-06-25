---
name: frontend-best-practices
description: >
  Framework-agnostic frontend engineering best practices. Fire this skill whenever building,
  reviewing, or refactoring any frontend component, hook, state, form, or UI logic — even when
  no framework is named. It detects the project's framework and loads only the matching guidance.
  Use it proactively; clean component architecture, state, async, forms, accessibility,
  performance, and TypeScript discipline should be the default, not an afterthought.
  Trigger words: frontend, component, hook, jsx, tsx, props, react, useState, useEffect, solid,
  solidjs, signal, createSignal, store, accessibility, a11y, web performance, web vitals.
user-invocable: true
allowed-tools: Read Edit Write Glob Grep Bash Agent
---

# Frontend

Best practices for building UI that is correct, accessible, fast, and type-safe. The principles
below are framework-neutral — they hold whether you write React, SolidJS, or anything else. The
framework-specific *how* lives in references that you load on demand, because the single biggest
source of frontend bugs is applying one framework's mental model to another.

## Framework Selection — do this first

The shared principles tell you *what* to aim for. They do **not** tell you how to write reactive
code, because that diverges sharply by framework and getting it wrong is the most common cause of
"my UI won't update" and "this re-renders too much" bugs. So before writing or reviewing reactive
logic, detect the framework and load exactly one reference.

Detect in this order; stop at the first signal:

1. **Explicit user statement.** If the user names the framework ("this is a Solid project"), honor it.
2. **`package.json` dependencies** — the strongest signal:
   - `react` / `react-dom` (or `next`) → load `references/react.md`
   - `solid-js` / `@solidjs/*` (e.g. `@solidjs/router`, `solid-start`) → load `references/solidjs.md`
3. **The working file's extension + imports.** `.tsx`/`.jsx` with `import … from "react"` or hook
   calls like `useState` → React. `import { createSignal } from "solid-js"`, or JSX that uses
   `<For>`/`<Show>` → SolidJS. (Note: both use `.tsx`, so extension alone is not decisive — read the imports.)

```bash
# Quick check:
cat package.json | grep -E '"(react|react-dom|next|solid-js|@solidjs|solid-start)"'
```

Load **only** the matching reference — don't load both. If the signals genuinely conflict or
none are present (e.g. a fresh file in an empty repo), ask the user once which framework they're
using, then proceed.

**Why this matters (the core divergence):** React re-runs the whole component function on every
render and relies on dependency arrays + memoization to *prevent* unnecessary work and re-renders.
SolidJS runs each component function *once* and uses fine-grained signals to update only the exact
DOM nodes that changed — there is no re-render to optimize away. This single difference flips the
guidance on memoization, effects, derived state, list rendering, and props. The reference re-states
each shared topic in its framework's own model — read it there, don't infer it from the principles below.

To add a framework later: drop a new `references/<framework>.md` mirroring the taxonomy headings,
and add one detection rule above. Nothing in the shared body should need to change.

---

## Shared principles

These are framework-neutral. Each maps to a same-named, same-ordered heading in every reference,
where the framework-specific mechanics live.

### 1. Component architecture & composition

Favor composition over inheritance and over configuration-by-boolean-props. A component that
sprouts `isModal`, `isCompact`, `variant` flags is usually two or three components wearing a
trenchcoat — split it. Reach for compound components (a parent coordinating related children),
render props / slots, and polymorphic `as`-style components when you need flexible composition
rather than a growing prop matrix.

Give each component a single responsibility, and a clean, fully typed prop API that describes
*what it accepts*, not *how it's implemented*. Avoid prop drilling: passing a value through
several layers that don't use it is a smell — lift the state, share it via context, or restructure
so the consumer sits closer to the source. The test: a reader should understand a component's
contract from its props alone.

### 2. State & data flow

Distinguish **server state** (data fetched from an API — it has a canonical source elsewhere,
goes stale, needs caching and revalidation) from **client/local state** (UI concerns the client
owns: open/closed, selected tab, form-in-progress). Never park fetched server data in ad-hoc local
state and hand-sync it — that's how you get stale and inconsistent UIs. Use a server-state/query
layer for the former.

Colocate state with the component that uses it, and lift it only as far as the nearest common
ancestor that needs it — no higher. Use context for genuinely shared, low-frequency state (theme,
auth, locale), not as a global dumping ground. When state transitions get complex or
interdependent, model them with a reducer / state machine so invalid states are unrepresentable.
Prefer **derived** state — computed from existing state — over storing a copy you must keep in sync.

### 3. Async & data fetching

Model every async operation as four explicit states: **loading, error, empty, and success**. The
empty state (a successful fetch that returned nothing) is the one people forget, and it's a
different UI from loading. Don't render `data.map(...)` against a maybe-undefined value.

Cache and dedupe requests so the same data isn't fetched five times across a screen. Guard against
**race conditions**: when inputs change faster than responses return (search-as-you-type), a stale
response can overwrite a fresh one — cancel or ignore out-of-order results. Use the framework's
suspense/async boundary primitives to keep loading and error handling declarative rather than
scattering `if (loading)` branches everywhere.

### 4. Forms & validation

Prefer controlled inputs so the component state is the single source of truth for field values.
Define one **schema** (Zod) and infer your TypeScript types from it rather than declaring types and
validators separately and letting them drift. Validate untrusted input at the boundary — client
validation is for UX, but never trust it for correctness or security; the server validates too.

Make errors accessible: associate each message with its field (`aria-describedby`), move focus to
the first error on failed submit, and announce validation results so screen-reader users aren't
stranded on a form that silently rejected them.

### 5. Accessibility

Reach for semantic HTML first — a real `<button>`, `<nav>`, `<label>`, `<dialog>` gives you
keyboard behavior, focus, and screen-reader semantics for free. Add ARIA only to fill genuine gaps
semantic HTML can't cover; incorrect ARIA is worse than none. Every interactive element must be
reachable and operable by keyboard, with visible focus and sensible focus management (trap focus in
modals, restore it on close). Meet WCAG 2.1 AA contrast (4.5:1 for normal text). Run automated
checks (axe, Lighthouse) as a floor — they catch maybe a third of issues, so tab through the UI
yourself too.

### 6. Performance & rendering

**Profile before optimizing.** Most perceived slowness is a few real bottlenecks, not pervasive
inefficiency — measure with the profiler and fix what's actually hot. Memoize *deliberately*, where
a measurement justifies it, not reflexively (how memoization works, and when it's even needed,
differs sharply by framework — see the reference). Code-split and lazy-load below-the-fold and
route-level chunks. Virtualize long lists so you render rows in the viewport, not thousands of
offscreen nodes. Analyze the bundle to catch accidental heavyweight imports, and optimize images
(right format, sizes, lazy loading). Keep an eye on Core Web Vitals (LCP, INP, CLS) — they're the
user-facing scoreboard.

### 7. TypeScript discipline

Turn on strict mode, plus `noUncheckedIndexedAccess` (array access is `T | undefined`) and
`exactOptionalPropertyTypes`. Model mutually-exclusive shapes as **discriminated unions** and handle
them exhaustively (a `never` default case fails the build when you add a variant). Narrow unknown
input with type guards. Prefer `satisfies` over `as` — it checks the value against the type without
widening or silencing errors, whereas `as` is an unchecked assertion. Use generics with constraints
for reusable components and hooks, and utility/mapped/conditional types to derive types rather than
restate them. Infer types from schemas (Zod) so runtime validation and compile-time types share one
source. Type component props, hook returns, and event handlers precisely — the signatures are your
first line of documentation.

### 8. Testing principles

Test **behavior and rendered output**, never internal implementation details — assert what the user
sees and can do, not which functions were called or what the internal state is. That way tests
survive refactors. Avoid shallow rendering; render the real thing. Mock the **network at the
boundary** (intercept HTTP) rather than stubbing your own modules, so tests exercise real code paths.

This skill does not re-teach a test runner. For the TDD workflow (red-green-refactor, writing the
test first), defer to the **`tdd`** skill. For end-to-end and browser testing, defer to the
**`playwright`** skill. Use those by name; don't reinvent their guidance here.

### 9. Project structure & tooling

Organize by **feature** (a folder per feature owning its components, hooks, and tests) rather than
by type (`components/`, `hooks/`, `utils/` mega-folders that scatter one feature across the tree).
Co-locate tests and helpers with the code they serve. Prefer named exports (greppable, refactor-safe,
no rename-on-import drift) and consistent file/identifier naming.

Baseline tooling: Vite for dev/build, pnpm for installs, ESLint (flat config) + Prettier for lint
and format. Run typecheck and lint before every commit (a pre-commit hook or CI gate) so broken
types never land on the main branch.

---

## How to apply this skill

When building or refactoring: read the relevant shared principle above, then load the matching
framework reference for the concrete mechanics. When reviewing, walk the nine topics as a checklist
against the diff — but only flag what's actually wrong, with the *why*, not a rule recital.
