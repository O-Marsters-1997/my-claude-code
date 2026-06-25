# React reference

Load this when the project uses `react` / `react-dom` (or Next.js). It mirrors the nine shared
taxonomy headings from `SKILL.md`, in order, covering only what is React-specific.

## React's rendering & reactivity model — read this first

React re-runs a component function **every time** it renders: on its own state change, on a parent
re-render, or on a context change it subscribes to. Each render produces a fresh set of locals,
closures, and a new virtual DOM tree that React diffs against the last to decide what real DOM to
touch. State lives outside the function (in the fiber), retrieved each render via hooks.

The consequences that drive everything below:

- **Everything in the function body runs again on every render** — including object/array/function
  literals, which get new identities each time. That fresh identity is what breaks `===` checks,
  effect dependency arrays, and memoized children.
- **You optimize by *preventing* re-renders and recomputation** (`memo`, `useMemo`, `useCallback`),
  the opposite of Solid, where components run once and there's nothing to prevent.
- **Effects and derived values are tied to dependency arrays** — you declare what a computation
  depends on, and React reruns it when those change. Get the deps wrong and you get stale closures
  or infinite loops.

Hold this model in mind; the rest is consequences of it.

### Rules of hooks

Call hooks only at the top level of a component or custom hook, never in conditions, loops, or after
an early return. React identifies hooks by call order, so a conditional hook desynchronizes state
between renders. Custom hooks (`useX`) are the unit of reusable stateful logic — extract shared
behavior into one rather than duplicating effects across components.

## 1. Component architecture & composition

Compose with `children` and render props; this is also React's main escape from prop drilling
besides context. Memoization (below) interacts with composition: passing `children` through a
component lets the parent re-render without forcing the children to, because their elements were
created higher up. Keep components as functions returning JSX; reserve `useImperativeHandle` +
`forwardRef` for the rare case a parent must call into a child (focus, scroll).

## 2. State & data flow

`useState` for independent values; `useReducer` when transitions are interdependent or complex.
Context provides shared state but **any context value change re-renders every consumer** — split
contexts by concern and memoize the provider's value, or readers will re-render on unrelated
changes. Derive during render (`const full = first + " " + last`) instead of mirroring state in
another `useState` + effect — a state-syncing effect is almost always a bug waiting to go stale.
For server state, use a query library (TanStack Query, RTK Query) rather than `useEffect` +
`useState` fetching.

## 3. Async & data fetching

`useEffect`-based fetching is race-prone: use the cleanup function to set an `ignore` flag or abort
the request so a stale response can't commit after a newer one. Prefer a query library that handles
caching, dedup, and race-safety for you. `<Suspense>` declaratively renders a fallback while a child
suspends, and an error boundary catches thrown errors — together they replace scattered `if (loading)`
/ `if (error)` branches. (Throwing-to-suspend is what query libraries and RSC integrate with.)

```tsx
type State<T> =
  | { status: "loading" } | { status: "error"; error: Error }
  | { status: "empty" }   | { status: "success"; data: T };

useEffect(() => {
  let ignore = false;                          // stale response can't win the race
  fetchUsers(query).then(d => {
    if (ignore) return;
    setState(d.length ? { status: "success", data: d } : { status: "empty" });
  });
  return () => { ignore = true; };
}, [query]);
```

## 4. Forms & validation

For anything beyond trivial, use React Hook Form (uncontrolled inputs via refs to minimize
re-renders) with a Zod resolver, or a controlled form with Zod on submit. Infer the form type from
the schema. Wire `aria-describedby` to the error element id and move focus to the first invalid field
on submit failure.

```tsx
const Schema = z.object({ email: z.string().email() });
type Values = z.infer<typeof Schema>;            // one source of truth: type follows schema
const { register, formState: { errors } } = useForm<Values>({ resolver: zodResolver(Schema) });
<input {...register("email")} aria-invalid={!!errors.email} aria-describedby="email-err" />
{errors.email && <p id="email-err">{errors.email.message}</p>}
```

## 5. Accessibility

Same semantic-HTML-first principles. React specifics: use `htmlFor` (not `for`) on labels and
`aria-*` attributes as written. Manage focus with refs + `useEffect` (focus on mount/open, restore on
unmount/close). `key` is about reconciliation, not a11y — but a wrong/index key can scramble focus and
state across list reorders, which reads as an accessibility bug.

## 6. Performance & rendering

This is where the React model bites hardest. A parent re-render re-renders all children by default.
Tools, in order of reach-for:

- **`React.memo`** — skip re-rendering a child when its props are shallow-equal. Only helps if you
  also stabilize the props you pass (see below).
- **`useMemo`** — cache an expensive computation, or stabilize an object/array identity passed to a
  memoized child or an effect dep.
- **`useCallback`** — stabilize a function identity for the same reasons (it's `useMemo` for functions).

Don't sprinkle these everywhere: each has a cost and most renders are cheap. Profile with the React
DevTools Profiler, find the component that re-renders expensively, then memoize the specific path.
A new object literal in props defeats a child's `memo`, so stabilize *before* reaching for `memo`.
`useTransition`/`useDeferredValue` keep the UI responsive during heavy updates. Lazy-load routes and
heavy components with `React.lazy` + `<Suspense>`. Virtualize long lists (TanStack Virtual). Always
give list items a stable, data-derived `key` — never the array index for reorderable lists.

## 7. TypeScript discipline

Type props with an explicit type/interface; prefer `type Props = {...}` and destructure in the
signature. Type children as `React.ReactNode`. For event handlers use the precise React event types
(`React.ChangeEvent<HTMLInputElement>`, `React.MouseEvent`). For polymorphic `as` components, use
`React.ComponentPropsWithoutRef<T>`. Type custom hooks' return values explicitly, and return tuples
`as const` (or objects) so call sites get correct types. Generic components are written
`function List<T>(props: { items: T[] }) {...}`. Avoid `React.FC` (it implies `children` and weakens
generics) — type props directly.

```tsx
const assertNever = (x: never): never => { throw new Error(String(x)); };
switch (state.status) {                         // State<T> union from §3
  case "loading": return <Spinner />;
  case "error":   return <Err msg={state.error.message} />;
  case "empty":   return <Empty />;
  case "success": return <List data={state.data} />;
  default:        return assertNever(state);    // compile error if a variant is added
}
```


## 8. Testing principles

React Testing Library is the standard, but per the shared body this skill doesn't teach a runner —
defer the TDD loop to **`tdd`** and browser/e2e to **`playwright`**. The React-relevant principle:
query by accessible role/text (what the user perceives), not by test-ids or component internals, and
mock fetch at the network boundary (MSW), not your own data hooks.

## 9. Project structure & tooling

Standard React conventions: `PascalCase` component files, custom hooks in `useX.ts`, co-locate
`Component.test.tsx`. ESLint `react-hooks` plugin enforces the rules-of-hooks and exhaustive-deps —
keep it on; the exhaustive-deps warning is usually right, and silencing it is a deliberate decision
to document, not the default.

## Meta-framework pointer

If the project uses **Next.js**, components are Server Components by default and only become Client
Components with `"use client"`. Server Components run on the server, can be `async`, and can't use
hooks or browser APIs; data mutations go through Server Actions. This is a substantial model on top of
React — when working in a Next.js app, get the App Router / Server Components / Server Actions details
from Next.js guidance rather than treating everything as a client component. Don't fold the
meta-framework into core React decisions.
