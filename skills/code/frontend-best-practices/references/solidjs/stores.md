# Solid stores (1.x)

Use a signal for primitives and values you replace whole. Use a store for nested objects and arrays
that change piece by piece: `createStore` wraps the value in a proxy with one signal per property,
created lazily on first tracked read, so updating `todos[3].title` notifies only what read that title.

## Reading

Read like a plain object — `store.user.name`, no call. The proxy is read-only: assigning to it
(`store.user.name = "x"`) is a dev warning in 1.x and notifies nobody. Destructuring or spreading a
store reads every key once and loses reactivity, exactly like props.

## Writing — the path setter

```ts
setStore("user", "name", "Ada");                         // one leaf
setStore("user", { name: "Ada" });                        // objects shallow-merge; no spread needed
setStore("todos", t => t.id === id, "done", d => !d);     // filter, then update from previous
setStore("todos", [0, 2], "done", true);                  // several indices
setStore("todos", { from: 0, to: 9 }, "done", false);     // a range
setStore("todos", todos => [...todos, newTodo]);          // arrays are replaced, not merged
```

The path form is preferred because it names exactly which leaf changed. For several fields at once use
`produce`, which gives an Immer-style draft:

```ts
setStore("todos", t => t.id === id, produce(t => { t.title = title; t.done = false; }));
```

## Server data

`reconcile(next, { key: "id" })` diffs a fresh immutable payload into the existing store so only the
changed leaves notify and `<For>` keeps rows whose identity is stable. Use it whenever a refetch
replaces a list — replacing the array wholesale remounts every row.

```ts
setStore("todos", reconcile(await fetchTodos(), { key: "id" }));
```

`unwrap(store)` returns the underlying plain object for third-party libraries that must not see the
proxy.

## Selection in large lists

`createSelector(selectedId)` returns `isSelected(id)`, which notifies only the row that gained and the
row that lost selection instead of every row:

```tsx
const isSelected = createSelector(selectedId);
<For each={rows()}>{r => <Row active={isSelected(r.id)} />}</For>
```

## Avoid

- `createMutable` / `modifyMutable`: a writable proxy that removes the read/write split stores are
  built on. Solid 2 removes them.
- Keeping a signal and a store in sync by hand — derive one from the other.
