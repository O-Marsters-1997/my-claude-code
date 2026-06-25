# SolidJS reference

Load this when the project depends on `solid-js` / `@solidjs/*` / `solid-start`. It mirrors the nine
shared taxonomy headings from `SKILL.md`, in order, covering only what is Solid-specific.

## Solid's rendering & reactivity model — read this first

A Solid component function runs **exactly once**, at creation. It is *setup*, not *render*. There is
no re-render and no virtual DOM. Reactivity is fine-grained: a **signal** is a value with a getter
and setter, and Solid tracks which DOM nodes and computations *read* the getter. When you set the
signal, Solid re-runs only those exact subscribers and updates only those exact DOM nodes.

The consequences that drive everything below — and they are the opposite of React's:

- **Signals are functions.** Read with `count()`, not `count`. Reading inside JSX or an effect is
  what subscribes that location to updates. Read the wrong way (destructure, cache in a local) and
  you sever the subscription — the UI goes stale. *This is the #1 Solid bug.*
- **Components run once, so there is nothing to memoize for re-render avoidance.** `useMemo`/`memo`
  have no equivalent purpose. `createMemo` exists only to cache an *expensive derived computation* or
  to share one computation among many subscribers — not to prevent re-renders (there are none).
- **Props are live reactive getters, not a plain snapshot.** Destructuring props at the top of a
  component reads them once at setup and freezes them — they'll never update. Keep props intact and
  access `props.foo` at point of use, or split them reactively (below).

Hold this model; most "it won't update" bugs are a broken subscription, not a missing dependency.

### Primitives

- `createSignal(initial)` → `[get, set]`. `const [count, setCount] = createSignal(0); count(); setCount(c => c + 1)`.
- `createMemo(fn)` → a cached derived getter; recomputes when its tracked deps change, shared by all readers.
- `createEffect(fn)` → runs after render and re-runs when tracked signals change. For side effects only — don't set signals you also read in it without care.
- `createResource(source, fetcher)` → async data as a signal with `.loading` and `.error`; re-fetches when `source` changes.

## 1. Component architecture & composition

Composition works the same conceptually (children, slots) but the once-only model means setup code
in the parent doesn't re-run per child update. `children(() => props.children)` helper resolves and
memoizes children when you need to inspect them. No prop drilling concerns differ; context (below)
is the escape hatch.

## 2. State & data flow

Local reactive state is a signal. **Props are reactive getters — never destructure them.** To pull
specific props while keeping reactivity, use `splitProps(props, ["a", "b"])`; to supply defaults use
`mergeProps({ size: "md" }, props)`. For nested/object state use a **store** (`createStore`) — it
gives fine-grained reactivity *per property*, so mutating one field updates only its subscribers.
Mutate stores with the setter's path syntax or `produce` (Immer-style draft); use `reconcile` to
diff new data into an existing store (key for #2 below — see list rendering), and `unwrap` to get the
raw non-reactive object. Derive with `createMemo` rather than storing a synced copy. Context via
`createContext` + `useContext` shares state without the re-render cost React's context has, because
only the readers of the changed signal update.

```tsx
// Props: ✗ destructuring freezes them at setup — never updates
function Hi({ name }: Props) { return <p>{name}</p>; }
// ✓ keep props intact; split reactively when you need locals
function Hi(props: Props) {
  const [local] = splitProps(props, ["name"]);
  return <p>{local.name}</p>;
}

// Store update: ✗ in-place mutation + same array ref — fine-grained reactivity never fires
todo.title = title; setTodos([...todos]);
// ✓ write through the setter path — notifies just that field's subscribers
setTodos(i => i.id === id, "title", title);
```

## 3. Async & data fetching

Use `createResource` — it models loading/error/success natively (`resource.loading`, `resource.error`)
and re-fetches when its source signal changes, which also gives you race-safety (Solid tracks the
latest). Wrap in `<Suspense fallback={...}>` to show a fallback while resources load, and
`<ErrorBoundary fallback={...}>` to catch errors — these are first-class control-flow components, not
add-ons. For the empty state, check the resolved value explicitly (`<Show when={data()?.length} fallback={<Empty/>}>`).

```tsx
const [users] = createResource(query, fetchUsers);   // re-fetches when query() changes
<Suspense fallback={<Spinner />}>
  <Show when={users()?.length} fallback={<Empty />}>  {/* empty is its own state */}
    <For each={users()}>{u => <li>{u.name}</li>}</For>
  </Show>
</Suspense>
```

## 4. Forms & validation

Bind inputs to signals (`value={name()}` / `onInput={e => setName(e.currentTarget.value)}`). Define a
Zod schema and infer the type; validate on submit (and optionally per-field via a memo). Solid has no
re-render cost for controlled inputs, so the React motive for uncontrolled forms doesn't apply — keep
them controlled. Wire `aria-describedby` to the error node and focus the first invalid field with a
ref on submit failure.

```tsx
const Schema = z.object({ email: z.string().email() });
type Values = z.infer<typeof Schema>;            // one source of truth: type follows schema
const [email, setEmail] = createSignal("");
const result = () => Schema.safeParse({ email: email() });   // derived validation
<input value={email()} onInput={e => setEmail(e.currentTarget.value)}
  aria-invalid={!result().success} aria-describedby="email-err" />
<Show when={!result().success}><p id="email-err">{/* result().error message */}</p></Show>
```

## 5. Accessibility

Semantic-HTML-first is unchanged. Solid uses native DOM attribute names: `for` (not `htmlFor`),
`class` (not `className`), and standard `aria-*`. Manage focus with `ref` + `onMount`. Because there's
no re-render, focus isn't disturbed by state updates the way it can be in React — but list reordering
still matters: see `For` vs `Index` below.

## 6. Performance & rendering

Solid is fast by default — fine-grained updates mean no wasted component re-runs, so most React-style
performance work is simply unnecessary. **Don't reach for memoization to prevent re-renders; there
are none.** Use `createMemo` only when a derived computation is genuinely expensive or shared across
many subscribers. The real performance levers here:

- **`<For>` vs `<Index>`**: `<For each={items()}>` keys by item *reference/identity* and moves DOM
  nodes on reorder — use it for lists of objects. `<Index>` keys by *index* and is for primitives or
  fixed-position lists. Using the wrong one causes both perf and correctness bugs.
  ```tsx
  <For each={todos()}>{t => <Todo item={t} />}</For>   // objects: keyed by identity
  <Index each={labels()}>{l => <span>{l()}</span>}</Index> // primitives: l is a signal
  ```
- **Don't break the reactive graph** by reading signals into plain locals — that's the most common
  "perf" issue and it's really a correctness one.
- `batch(() => {...})` coalesces multiple signal writes into one update; `untrack(() => sig())` reads
  without subscribing. Lazy-load with `lazy(() => import(...))` + `<Suspense>`. Virtualize long lists
  (`@tanstack/solid-virtual`). Core Web Vitals apply equally.

## 7. TypeScript discipline

Signals are typed by inference or `createSignal<Type>(initial)`. Type component props with
`type Props = { ... }` and **don't destructure in the signature** (it breaks reactivity) — accept
`props: Props` and read `props.x`. Children type is `JSX.Element`. `splitProps`/`mergeProps` preserve
types. Event handlers: use Solid's `JSX.EventHandler` types or annotate `e.currentTarget`. Stores are
typed by their initial value; `SetStoreFunction<T>` types the setter. Everything else from the shared
TypeScript section (strict flags, discriminated unions, `satisfies`, schema-inferred types) applies
unchanged.

## 8. Testing principles

Per the shared body, this skill doesn't teach a runner — defer the TDD loop to **`tdd`** and
browser/e2e to **`playwright`**. Solid-relevant principle: test rendered output and behavior via
`@solidjs/testing-library` (render, query by accessible role/text), not signal internals, and mock
the network at the boundary.

## 9. Project structure & tooling

Standard Solid/Vite conventions: `vite-plugin-solid`, `PascalCase` components, co-located tests.
`eslint-plugin-solid` catches the high-value Solid mistakes the model is most likely to make —
reactivity violations like destructured props and signals read without calling — so keep it on; its
warnings usually point at a real broken subscription.

## Meta-framework pointer

For routing use **`@solidjs/router`**; for full-stack (SSR, server functions, file routing) use
**SolidStart**. SolidStart adds server/client boundaries and `"use server"` functions on top of Solid
core — when working in a SolidStart app, get those specifics from SolidStart guidance rather than
treating everything as client-side. Don't fold the meta-framework into core Solid decisions.
