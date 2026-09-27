# Solid async, router data and SSR (1.x)

Pick the data primitive by what the app already has:

| App has | Read with | Write with |
|---|---|---|
| `@solidjs/router` or SolidStart | `query()` + `createAsync` | `action` + `useSubmission` |
| neither | `createResource` | plain handlers + `mutate`/`refetch` |

Don't add the router just for data, and don't hand-roll caching in a router app — `query` already
dedupes and revalidates.

## Router / SolidStart

```ts
const getUser = query(async (id: string) => {
  "use server";                              // SolidStart only: runs on the server as RPC
  return db.user.find(id);
}, "user");

const updateUser = action(async (form: FormData) => {
  "use server";
  await db.user.update(/* ... */);
  return redirect("/users");                 // or reload({ revalidate: getUser.keyFor(id) })
}, "updateUser");

export const route = { preload: ({ params }) => getUser(params.id) } satisfies RouteDefinition;

export default function UserPage(props: RouteSectionProps) {
  const user = createAsync(() => getUser(props.params.id));
  const saving = useSubmission(updateUser);
  return (
    <form action={updateUser} method="post">
      <Show when={user()}>{u => <input name="name" value={u().name} />}</Show>
      <button disabled={saving.pending}>Save</button>
      <Show when={saving.error}><p role="alert">{saving.error.message}</p></Show>
    </form>
  );
}
```

- `query(fn, name)` caches by name + serialised arguments and dedupes concurrent calls. It replaced the
  older `cache()` export — code or tutorials using `cache` are outdated.
- `createAsync` only *reads*; the caching comes from `query`.
- A route's `preload` starts the fetch in parallel with loading the route's code.
- A successful `action` revalidates active queries by default; return `redirect()`/`reload()`/`json()`
  with `revalidate` keys to narrow that. `useSubmission` exposes `pending`, `error`, `result` and
  `input` (the latter is how to render optimistic UI).
- `"use server"` functions must be async, must not capture client-side variables, and must not touch
  `window`/`document`. They are network endpoints: validate their input.

## createResource

```ts
const [user, { mutate, refetch }] = createResource(userId, fetchUser);
```

- The fetcher re-runs when the source changes and is skipped while it is `undefined`/`null`/`false`.
  A stale response from an earlier source is dropped, so search-as-you-type is race-safe.
- `user()` suspends inside `<Suspense>`; `user.latest` returns the last value without suspending (keep
  showing old results while refetching). Other fields: `loading`, `error`, `state`
  (`unresolved | pending | ready | refreshing | errored`).
- `mutate(v)` sets the value locally (optimistic update); `refetch()` re-runs the fetcher.
- `useTransition()` → `[pending, start]` keeps current UI up during a refetch instead of flashing the
  Suspense fallback. Router navigation already runs in a transition.

## Ownership after await

Only synchronous code sees the current owner. After the first `await` (or inside `setTimeout`),
`useContext` returns undefined and `onCleanup`/`createEffect` attach to nothing and leak.

```ts
const owner = getOwner();
const data = await load();
runWithOwner(owner, () => { const ctx = useContext(Ctx); /* ... */ });
```

## Error boundaries

`<ErrorBoundary>` catches errors thrown while rendering and in effects — including a resource that
errors. It does **not** catch errors in event handlers or in async code after `await`; handle those
where they occur and put the result in state.

## SSR and hydration

- Server and client must render the same markup. `Date.now()`, `Math.random()`, locale formatting and
  `window` checks during render cause hydration mismatches — move them into `onMount`.
- `isServer` from `solid-js/web` is a build-time constant; branches on it are tree-shaken.
- Browser-only components: `clientOnly(() => import("./Map"))` from `@solidjs/start`.
- Module-scope signals and stores are shared across every request on the server. Put per-user state in
  context created inside the app.
