# Client JavaScript

htmx exists so the server can own almost everything. Client JS is the exception, and each new
piece of it should have lost an argument first. Work down this ladder and stop at the first rung
that holds:

| Rung | Use when | Example |
|---|---|---|
| **1. Server** | The state matters after a reload, to another tab, or to another user | A collapsed rail as a saved preference: `hx-post="/prefs/rail"` returns the re-rendered rail |
| **2. HTML / CSS** | The behaviour is a toggle, disclosure or overlay the platform already has | `<details>`, `<dialog>`, the `popover` attribute, `:has()`, `:target` |
| **3. Island** | A widget holds real client state that HTML cannot express | A graph, a rich-text editor, drag-to-reorder, a chart with zoom |
| **4. htmx glue** | Wiring htmx itself: busy state, poll guards, error fallback, history | The fixed listener set at the end of this file |

There is no rung for "a small DOM script to add a feature". A feature that needs client state is
an island; a feature that doesn't belongs on rung 1 or 2. Scattered `querySelector` behaviours are
how inline-script drift comes back under a different name.

## Islands

An island is a component from the repo's frontend framework (Solid, Preact, React — whatever it
already uses) mounted into a server-rendered page. Inside the island, the framework's rules apply:
load `frontend-best-practices`. This file covers only the boundary.

### Mount and unmount with the swap lifecycle

htmx swaps can remove an island at any time, so mount through `htmx.onLoad` and unmount on
`htmx:beforeCleanupElement`. One registry handles every island:

```ts
type Unmount = () => void;
type Island = (root: HTMLElement, props: unknown) => Unmount;

const islands = new Map<string, Island>();
const mounted = new WeakMap<Element, Unmount>();

export function defineIsland(name: string, island: Island): void {
  islands.set(name, island);
}

htmx.onLoad((loaded) => {
  const roots = loaded.matches("[data-island]")
    ? [loaded, ...loaded.querySelectorAll<HTMLElement>("[data-island]")]
    : [...loaded.querySelectorAll<HTMLElement>("[data-island]")];
  for (const root of roots) {
    if (mounted.has(root)) continue;
    const island = islands.get(root.dataset.island ?? "");
    if (!island) continue;
    const props = JSON.parse(root.querySelector("script[type='application/json']")?.textContent ?? "null");
    mounted.set(root, island(root as HTMLElement, props));
  }
});

document.addEventListener("htmx:beforeCleanupElement", (e) => {
  const el = e.target as Element;
  mounted.get(el)?.();
  mounted.delete(el);
});
```

### Props come from the server's HTML

```html
<div data-island="task-graph" hx-preserve id="task-graph">
  <script type="application/json">{{.Graph}}</script>
</div>
```

- A `type="application/json"` script is data, not code: the browser never executes it, so the
  no-inline-script rule and a strict CSP both hold.
- Pass the Go value itself (`.Graph` is a struct). `html/template` JSON-encodes values in this
  context; a string you marshalled first would be emitted as a quoted JS string instead.
- The island never fetches JSON to draw something the server could render. If it needs fresh
  data, the server re-renders the island's container with new props.
- `hx-preserve` with a stable `id` keeps a live island (and its scroll, zoom, selection) intact
  when a surrounding region is swapped.

### Islands talk back through hypermedia

An island reports user intent the way any HTML control would, so the server stays the owner:

- **A form value.** An editor writes its content into a hidden `<input>` inside a normal form.
- **An event htmx listens for.** The island dispatches a `CustomEvent` on its root; an ancestor
  with `hx-trigger="reordered from:closest [data-island]"` sends the request.
- **`htmx.ajax`** only when neither fits, targeting a region so the response is still HTML.

### Not Alpine.js

Alpine is the most common htmx companion, and this skill leaves it out on purpose: it puts
behaviour into inline attributes (`x-data`, `@click`), and its standard build needs `unsafe-eval`
under a CSP. That trades away the testable-module rule. If a repo already uses Alpine, follow the
repo and say so; don't introduce it.

## htmx glue

One small module wires htmx itself. Its scope is this list; a feature does not get added here.

```ts
type HtmxDetail = CustomEvent<{ elt: HTMLElement; target?: HTMLElement }>;

document.addEventListener("htmx:beforeRequest", (e) => {
  const { elt } = (e as HtmxDetail).detail;
  if (!elt.hasAttribute("data-poll")) return;
  if (document.hidden || elt.contains(document.activeElement)) e.preventDefault();
});

document.addEventListener("htmx:beforeSend", (e) => {
  (e as HtmxDetail).detail.target?.setAttribute("aria-busy", "true");
});

document.addEventListener("htmx:afterRequest", (e) => {
  (e as HtmxDetail).detail.target?.removeAttribute("aria-busy");
});

document.addEventListener("htmx:responseError", () => announce("Something went wrong. Try again."));
document.addEventListener("htmx:sendError", () => announce("Can't reach the server."));

function announce(message: string): void {
  const flash = document.getElementById("flash");
  if (flash) flash.textContent = message;
}
```

- **Poll guard.** Cancelling `htmx:beforeRequest` skips that tick; the next interval tries
  again. Trigger filters would do this inline, but `allowEval: false` disables them.
- **Busy state** is set on `htmx:beforeSend`, not `htmx:beforeRequest`, so a cancelled poll
  never leaves a region stuck busy.
- **Error fallback** covers only responses the server sent no usable HTML for (5xx, network).
  `htmx:responseError` fires only for statuses `responseHandling` marks as errors, so a 422 that
  swaps never reaches it.
- **History.** If something outside the swapped region must follow navigation (the active nav
  item, the title), prefer the server rendering it (`hx-swap-oob`). Listen for
  `htmx:pushedIntoHistory` / `htmx:historyRestore` only when it can't.
- In development, listen for `htmx:targetError`: a mistyped `hx-target` otherwise drops the
  response without a trace.

## Polling

`hx-trigger="every 5s"` plus `hx-sync="this:drop"` stops overlapping polls; `data-poll` opts the
element into the guard above.

- A polling target never contains an input the user edits. Split the region: the poll targets
  the list, the input sits outside it. `hx-preserve` with a stable `id` is the fallback when
  splitting is impossible.
- To stop polling for good (the job finished), the server answers `286`.
- When a user action and a poll hit the same region, give both `hx-sync="#region:replace"` so the
  user's request wins.
