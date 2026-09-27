# SolidJS reference

Load this when the project depends on `solid-js` / `@solidjs/*`. It mirrors the nine shared taxonomy
headings from `SKILL.md`, in order, covering only what is Solid-specific. Deeper material lives in
`references/solidjs/` — open a file only when the task touches its topic:

| File | Read when |
|---|---|
| `solidjs/stores.md` | nested/array state, store setters, `produce`/`reconcile`, selection |
| `solidjs/async.md` | fetching, mutations, `createResource`, router `query`/`createAsync`/`action`, SolidStart, SSR |
| `solidjs/solid-2.md` | the repo is on Solid 2 (see version check) |

## Version check — do this first

Solid 1.x is the stable line; Solid 2.0 is at release candidate with a heavily changed API. Most docs,
tutorials and training data describe 1.x, so the risk runs one way: writing 1.x code in a 2.0 repo.

```bash
grep -E '"(solid-js|@solidjs/web)"' package.json
```

`solid-js` at `^1.x` → this file is correct as written. `solid-js` at `2.x` (including `2.0.0-rc`/`beta`)
or `@solidjs/web` present → also read `solidjs/solid-2.md` before writing code; several APIs below are
renamed or gone there.

## Solid's model — read this first

A Solid component function runs **once**. It is setup, not render. There is no virtual DOM and no
re-render: JSX expressions, `createMemo` and `createEffect` are *tracking scopes*, and only they re-run
when a signal they read changes, updating exactly the DOM nodes involved.

Every rule below follows from one question: **is this signal read inside a tracking scope?** A read
in the component body happens once and freezes. A read inside JSX, a memo, an effect, or a function
called from one of those stays live. Most "it won't update" bugs are a read in the wrong place, not a
missing dependency — Solid has no dependency arrays.

Choose derived state in this order, because each step adds cost and a way to go wrong:

1. **Derived function** — `const total = () => price() * qty()`. The default. Re-evaluates per read.
2. **`createMemo`** — when the derivation is expensive or read in many places; caches, notifies only on change.
3. **`createEffect`** — only to push state *out* of the reactive system (DOM APIs, storage, third-party
   widgets, logging). Never to compute one signal from another: `createEffect(() => setB(a() * 2))`
   causes extra passes and loops, and Solid 2 throws on it.

## Check every Solid diff for these

These are the mistakes a model with React habits makes. `eslint-plugin-solid` catches most of them;
the rule name is in brackets.

1. **Destructured props or stores** freeze at setup. Read `props.x` at point of use, or
   `splitProps`/`mergeProps`. Never `{...defaults, ...props}`. [`no-destructure`, `reactivity`]
2. **Early return in a component** (`if (!props.user) return null`) is evaluated once. Use
   `<Show>`/`<Switch>`. [`components-return-once`]
3. **Signal read into a local** (`const n = count()` in the body) is a snapshot. Wrap it in a function.
4. **Dependency arrays** — `createEffect(fn, [dep])` does nothing useful; the second argument is an
   initial value. Use `on(dep, fn)` for explicit deps. [`no-react-deps`]
5. **`.map` in JSX** recreates every row when the array changes. Use `<For>`/`<Index>`. [`prefer-for`]
6. **Work after `await`** runs without an owner: `useContext`, `onCleanup` and effects there attach to
   nothing. Capture `getOwner()` first and wrap with `runWithOwner`. See `solidjs/async.md`.
7. **React DOM props**: `class` not `className`, `for` not `htmlFor`, `style` keys are kebab-case
   (`{"font-size": "12px"}`) and numbers get no `px`. [`no-react-specific-props`, `style-prop`]
8. **Event handlers bind once.** `onClick={props.onClick}` captures the handler at setup; write
   `onClick={e => props.onClick?.(e)}` when it can change.
9. **In-place store mutation** (`todo.title = x`) bypasses the setter and notifies nobody. Write through
   `setStore(...)`. See `solidjs/stores.md`.
10. **`innerHTML`** with untrusted content is an XSS sink. Sanitise or render as text. [`no-innerhtml`]

## 1. Component architecture & composition

Composition works as in any JSX framework; the difference is that setup code runs once per instance.
Use `children(() => props.children)` when you need to inspect or reuse children — reading
`props.children` twice creates the DOM twice. Choose among components at runtime with `<Dynamic
component={...}>`. Forward refs by accepting `props.ref` and passing it to the element.

## 2. State & data flow

Local state is a signal; nested or collection state is a store (see `solidjs/stores.md`). Use
`splitProps(props, ["a"])` to separate props and `mergeProps({ size: "md" }, props)` for defaults —
both keep reactivity. Context (`createContext` + `<Ctx.Provider value>`) only updates readers when the
value *contains* signals or a store; a plain object is a one-time snapshot. The idiom is to provide
`[state, actions]` built from a store, behind a `useX()` that throws when the provider is missing.
Module-scope signals are fine in a client-only SPA but leak between requests under SSR.

```tsx
// ✗ destructuring freezes props at setup
function Hi({ name }: Props) { return <p>{name}</p>; }
// ✓
function Hi(props: Props) { return <p>{props.name}</p>; }
```

## 3. Async & data fetching

In a `@solidjs/router` or SolidStart app, fetch with `query()` + `createAsync` and mutate with
`action` — that is the cache/dedupe/revalidate layer the shared principle asks for. Without the router,
use `createResource`. Both suspend inside `<Suspense>`; wrap with `<ErrorBoundary>`, and model empty
explicitly. Details, race-safety and SSR in `solidjs/async.md`.

```tsx
<ErrorBoundary fallback={(err, reset) => <Retry onClick={reset} />}>
  <Suspense fallback={<Spinner />}>
    <Show when={users()?.length} fallback={<Empty />}>
      <For each={users()}>{u => <li>{u.name}</li>}</For>
    </Show>
  </Suspense>
</ErrorBoundary>
```

## 4. Forms & validation

Controlled inputs cost nothing extra in Solid, so keep them controlled: `value={email()}` +
`onInput={e => setEmail(e.currentTarget.value)}`. Derive validation from a Zod schema with a function
or memo, and only show errors after the field is touched or on submit — deriving it eagerly flags
every field as invalid on first paint. In a router app, `<form action={myAction} method="post">` plus
`useSubmission(myAction)` gives pending/error state without hand-rolled signals.

## 5. Accessibility

Semantic HTML first, as in the shared body; Solid uses native attribute names (`for`, `class`,
`aria-*`). Manage focus with a ref: `let input!: HTMLInputElement; <input ref={input} />`, then
`input.focus()` in `onMount` or a submit handler. Refs are assigned during render, so they are only
safe to use from `onMount`, effects, or handlers.

## 6. Performance & rendering

There are no re-renders to prevent, so React-style memoisation work is unnecessary. The levers:

- **Control flow over expressions.** `<Show>` and `<For>` keep branches and rows alive; `.map` and
  heavy ternaries recreate them. `<Show when={user()}>{u => u().name}</Show>` — without `keyed`, the
  callback gets an accessor, which also narrows the type.
- **`<For>` vs `<Index>`.** `<For>` keys by item identity (item is a value, index an accessor) — use it
  for objects that reorder. `<Index>` keys by position (item is an accessor) — use it for primitives
  and fixed-length lists such as form rows.
- **Stores:** `reconcile` server data so only changed leaves notify; `createSelector` for
  "is selected" across a large list.
- `lazy(() => import(...))` + `<Suspense>` for code splitting; `@tanstack/solid-virtual` for long lists.
  `batch` is rarely needed — handlers, effects and store setters already batch.

## 7. TypeScript discipline

Accept `props: Props` (never destructure in the signature). Component types: `Component<P>`,
`ParentComponent<P>` (adds optional `children`), `VoidComponent<P>` (forbids children),
`FlowComponent<P, C>` (required/callback children). Values: `JSX.Element`, `Accessor<T>`, `Setter<T>`.
Pass-through props: `JSX.HTMLAttributes<HTMLButtonElement>` with `splitProps`. `createSignal<T>()`
with no initial value is `T | undefined`. Type refs as `let el!: HTMLElement`. Store setters are
`SetStoreFunction<T>`.

## 8. Testing principles

Defer the TDD loop to **`tdd`** and e2e to **`playwright`**. Solid specifics for
`@solidjs/testing-library` with Vitest:

- `render(() => <Counter />)` takes a **function**, not an element — passing `<Counter />` runs the
  component outside a root.
- There is no `rerender`; drive changes by setting signals or through user events.
- `renderHook(useThing)` for custom primitives; `testEffect(done => createEffect(...))` for effects
  instead of polling. Test bare primitives inside `createRoot(dispose => ...)`.

## 9. Project structure & tooling

`vite-plugin-solid` (it also configures Vitest). Add `eslint-plugin-solid` if missing, with its
`recommended` or `typescript` flat config — on Solid 2 use its `v2` config. Treat its `reactivity`
warnings as real broken subscriptions, not noise. Check `tsconfig.json` has `"jsx": "preserve"` and
`"jsxImportSource": "solid-js"`.

## Meta-framework pointer

Routing is `@solidjs/router`; full-stack (SSR, `"use server"` functions, file routing) is SolidStart.
Both change how data is loaded and where code runs — read `solidjs/async.md` before touching data or
server code in either.
