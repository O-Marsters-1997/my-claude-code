# Solid 2.0 differences

Read only when the repo is on `solid-js@2.x` or depends on `@solidjs/web`. As of September 2026 Solid
2.0 is a release candidate: the API is declared frozen but not yet `latest` on npm. This file lists
what changes relative to the 1.x guidance; when a detail matters, confirm it in the migration guide:
https://github.com/solidjs/solid/blob/next/documentation/solid-2.0/MIGRATION.md

A codemod exists: https://github.com/solidjs-community/solid-migration-assistant. Lint with
`eslint-plugin-solid`'s `v2` config — its `removed-api` rule flags 1.x APIs.

## Semantics that change behaviour, not just names

- **Batched by default.** Setters apply at the next microtask: after `setCount(1)`, `count()` still
  returns the old value until the flush. Call `flush()` when you need the DOM updated synchronously
  (e.g. before focusing a new element). `batch` is gone.
- **Writing a signal inside a reactive scope throws in dev** (memos, component bodies, effect compute).
  Derive instead; this makes the 1.x "don't sync state in effects" advice a hard error.
- **Top-level reactive reads in a component body warn**, which covers destructured props. Use `untrack`
  for a deliberate one-time read.
- **Effects are split:** `createEffect(() => source(), (value, prev) => { sideEffect; return cleanup })`.
  The first function tracks; the second runs untracked after the flush. `on()` and `createComputed`
  are gone; `createTrackedEffect` is the single-callback escape hatch.

## Renames and removals

| 1.x | 2.0 |
|---|---|
| `solid-js/web` | `@solidjs/web` (set `jsxImportSource` to it) |
| `solid-js/store` | merged into `solid-js` |
| `onMount` | `onSettled` (may return a cleanup) |
| `createResource`, `.loading` | async `createMemo(() => fetchX(id()))`; `isPending(() => x())` |
| `refetch` / `mutate` | `refresh(x)`; `createOptimistic` / `createOptimisticStore` |
| `<Suspense>` / `<SuspenseList>` | `<Loading>` / `<Reveal>` |
| `<ErrorBoundary>` | `<Errored>` |
| `<Index>` | `<For keyed={false}>` |
| `<Dynamic>` | `dynamic(() => Comp)` |
| `<Ctx.Provider value>` | `<Ctx value>`; missing provider throws, so `useX()` guards are unnecessary |
| `splitProps` / `mergeProps` | `omit(props, "a")` / `merge` |
| `produce`, path setters | draft setters `setStore(s => { s.x = 1 })`; `storePath(...)` for paths |
| `unwrap` | `snapshot` |
| `createSelector` | `createProjection` / derived `createStore(fn, seed)` |
| `createMutable` | removed |
| `classList` | `class={["a", { active: on() }]}` |
| `use:` directives | `ref={directive(opts)}`; `ref={[a, b]}` composes |
| `on:`, `oncapture:`, `attr:`, `bool:` | removed; use a ref + `addEventListener` for listener options |
| `useTransition` / `startTransition` | removed; transitions are built in |

`<Loading>` shows its fallback only on first load; later refreshes keep the stale content (pass
`on={dep}` to re-show it). `"use server"` lives in core (`@solidjs/web`) and SolidStart is superseded
by Solid 2's start mode — check which the repo uses before adding server code.

The router's `query`/`createAsync` have no confirmed 2.0 mapping yet; in core, async memos plus
`action`/`refresh` cover the same ground. Follow whatever the repo's router version exports.
