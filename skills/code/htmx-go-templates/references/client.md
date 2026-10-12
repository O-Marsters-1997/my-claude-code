# Client: the TypeScript module

All client behaviour lives in one TS entry (`web/src/app.ts` or the repo's equivalent), built to
the `app.js` the layout loads. Templates opt elements in with `data-*` attributes; the module finds
them.

## Behaviours register through `htmx.onLoad` and return a disposer

`htmx.onLoad` runs for the initial document and for every swapped-in fragment, so one registration
covers first paint and later swaps. Scope queries to the element it hands you, never `document`,
or a fragment swap re-initialises the whole page.

```ts
type Dispose = () => void;
type Behaviour = (root: HTMLElement) => Dispose;

const disposers = new WeakMap<Element, Dispose[]>();

export function register(selector: string, behaviour: Behaviour): void {
  htmx.onLoad((loaded) => {
    const roots = loaded.matches(selector)
      ? [loaded, ...loaded.querySelectorAll<HTMLElement>(selector)]
      : [...loaded.querySelectorAll<HTMLElement>(selector)];
    for (const el of roots) {
      const list = disposers.get(el) ?? [];
      list.push(behaviour(el));
      disposers.set(el, list);
    }
  });
}

document.addEventListener("htmx:beforeCleanupElement", (e) => {
  const el = e.target as Element;
  disposers.get(el)?.forEach((d) => d());
  disposers.delete(el);
});
```

```ts
register("[data-collapsible]", (el) => {
  const toggle = () => el.toggleAttribute("data-collapsed");
  const button = el.querySelector("[data-collapse-toggle]");
  button?.addEventListener("click", toggle);
  return () => button?.removeEventListener("click", toggle);
});
```

Each behaviour is a plain function of an element, so it is testable with a DOM fixture and no
htmx at all.

## History

- `htmx:beforeHistorySave`: undo anything a third-party library did to the DOM (open menus,
  injected wrappers), or the snapshot restores a broken state.
- `htmx:historyRestore` and `htmx:pushedIntoHistory`: update things outside the swapped region
  (the active nav item, the document title, focus). Listen in the module, once.
- Restore from a cache miss re-requests the URL with `HX-History-Restore-Request: true`; the
  server returns the full page (see `responses.md`).

## Polling guards

`hx-trigger="every 5s"` plus `hx-sync="this:drop"` stops overlapping polls. The module adds the
user-first guard, since trigger filters are disabled by `allowEval: false`:

```ts
document.addEventListener("htmx:beforeRequest", (e) => {
  const { elt } = (e as CustomEvent<{ elt: HTMLElement }>).detail;
  if (!elt.hasAttribute("data-poll")) return;
  if (document.hidden || elt.contains(document.activeElement)) e.preventDefault();
});
```

- Cancelling `htmx:beforeRequest` skips that tick; the next interval tries again.
- A polling target never contains an input the user edits. If the design puts one there, split
  the region: the poll targets the list, the input sits outside it. `hx-preserve` with a stable
  `id` is the fallback when splitting is impossible.
- To stop polling for good (the job finished), the server answers `286`.
- When a user action and a poll hit the same region, give both `hx-sync="#region:replace"` so the
  user's request wins.

## Busy and error state

```ts
type TargetDetail = CustomEvent<{ target?: HTMLElement }>;

document.addEventListener("htmx:beforeSend", (e) => {
  (e as TargetDetail).detail.target?.setAttribute("aria-busy", "true");
});
document.addEventListener("htmx:afterRequest", (e) => {
  (e as TargetDetail).detail.target?.removeAttribute("aria-busy");
});
document.addEventListener("htmx:responseError", () => {
  const flash = document.getElementById("flash");
  if (flash) flash.textContent = "Something went wrong. Try again.";
});
document.addEventListener("htmx:sendError", () => {
  const flash = document.getElementById("flash");
  if (flash) flash.textContent = "Can't reach the server.";
});
```

`aria-busy` is set on `htmx:beforeSend`, not `htmx:beforeRequest`, so a poll the guard cancels
never leaves a region stuck busy. `htmx:responseError` fires only for statuses `responseHandling` marks as errors, so a 422 that
swaps never reaches it. In development, also listen for `htmx:targetError`: a mistyped
`hx-target` otherwise drops the response without a trace.
