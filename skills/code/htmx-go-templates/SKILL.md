---
name: htmx-go-templates
description: >
  htmx on Go html/template: how a handler picks page, fragment or redirect, how templates
  carry hx-* attributes, where client JS lives, and how errors and polling surface.
  TRIGGER — load BEFORE writing, not after: a .tmpl or .gohtml file is created or edited;
  an hx-* attribute or HX-* header is added or changed; a Go handler calls ExecuteTemplate
  or reads HX-Request; the prompt names htmx, hx-get, hx-swap, a fragment, partial, swap
  target, polling or an htmx event. Trigger on plain requests in a Go + htmx repo — "add a
  page", "make this form inline", "refresh this every 5s", "why doesn't the error show".
  Load alongside go-idiomatic. SKIP for templ components, JSON APIs and React or Solid UIs.
allowed-tools: Read, Edit, Write, Glob, Grep, Bash(go:*), Bash(git:*)
---

# htmx on Go html/template

The server owns state; a template only presents it; htmx moves HTML between the two. Most drift
in these apps comes from that split being decided ad hoc, per handler and per template: seven ways
to detect an htmx request, the same four attributes hand-copied across files, errors that vanish.
The rules below make each of those decisions once.

This skill holds the general rules. The repo's AGENTS.md holds what only the repo knows: its
swap-target ids, the names of its shared partials, and the htmx version it vendors. Read it first.

## Where to look

| If you're working on… | Load |
|---|---|
| A handler: choosing page / fragment / redirect, htmx headers, `Vary`, 303, error responses | `references/responses.md` |
| A template: partials, repeated attribute bundles, forms, live regions, conditionals | `references/templates.md` |
| Client behaviour: the TS module, `htmx.onLoad`, history events, polling guards, `aria-busy` | `references/client.md` |

## The rules

1. **htmx headers go through one helper set, never hand-typed strings.**
   A typo in `"HX-Request"` or a check for `""` instead of `"true"` fails silently; one package
   of constants and predicates makes the wrong spelling unrepresentable. If the repo has no such
   set yet, introduce it the first time a handler you touch needs an htmx header — `Vary` included.
2. **Every route declares its response shape — page, fragment or redirect — in one place.**
   A handler that sniffs headers inline grows a new branch per feature. When one URL can return
   either a page or a fragment, set `Vary: HX-Request`, or a cache or the back button can serve
   the bare fragment as a page. A history-restore request (`HX-History-Restore-Request: true`)
   always gets the full page.
3. **A non-htmx form POST answers 303 to a GET; an htmx one answers `HX-Location` or a fragment.**
   303 stops a refresh from resubmitting. A 303 to an htmx request is followed by the XHR and the
   whole redirected page gets swapped into the target, which is never what you want.
4. **Forms work without JavaScript.** Real `action` and `method`, a real submit button; htmx
   attributes layer on top. The non-JS path is also the path the 303 rule serves.
5. **A repeated attribute bundle becomes a named `{{define}}` partial.** Four attributes copied to
   seventeen links drift the moment one copy changes. One partial, one place to change it.
   (`codebase-design` has the vocabulary for why: one deep module beats many shallow copies.)
6. **Polling never fights the user.** `hx-sync` on every polling element, a guard that skips the
   poll while the user is focused or typing inside the target, and a target that never contains
   an input the user may be editing.
7. **Errors reach the user.** htmx 2.x does not swap 4xx/5xx bodies by default, so a refused
   action looks like nothing happened. Fix it in two halves: a `responseHandling` entry that swaps
   the status you use (typically 422), *and* `HX-Retarget` / `HX-Reswap` from the server so the
   error lands in a known region. That region is a persistent `aria-live` container.
8. **In-flight state is visible.** `hx-disabled-elt` on submit controls so a double-click cannot
   double-submit; `hx-indicator` or `aria-busy` on the region being replaced.
9. **No inline script in templates — no `<script>` blocks, no `onclick`, no `hx-on`.** Behaviour
   lives in a TypeScript module that registers through `htmx.onLoad` and returns a disposer. Inline
   handlers are untestable, duplicate across fragments, and re-run unpredictably on swaps. htmx
   config goes in `<meta name="htmx-config">`, so the layout needs only `<script src>` tags; set
   `"allowEval": false` there and htmx itself refuses `hx-on` and `js:` values.
10. **Templates branch on presentation only.** `{{if eq .Status "failed"}}` chains are domain
    logic in the wrong layer. Compute a view-model field in Go (`.Badge`, `.CanRetry`) and let the
    template read it.

## Choosing a response shape

```
request ──► HX-History-Restore-Request: true? ──yes──► page
               │ no
               ▼
            HX-Request: true? ──no──► POST? ──yes──► 303 to GET (or re-render page with 422 on invalid)
               │ yes                       └─no──► page
               ▼
            route declares: fragment ──► ExecuteTemplate(w, "<block>", vm)
                            page+select ──► page (template uses hx-select)
                            redirect ──► HX-Location (or HX-Redirect for a full reload)
```

`references/responses.md` has the helper set and a handler written both ways.

## htmx version

These rules target **htmx 2.x**. Check the version the repo vendors before relying on defaults.
If it is on 4.x, re-check rules 7 and 9 against the
[4.x migration notes](https://four.htmx.org/docs/whats-new-in-htmx-4):

- error responses swap by default, with per-status control through `hx-status`;
- attribute inheritance is explicit, so bundles that relied on inheriting `hx-target` break;
- event names changed, so listeners in the TS module need renaming.

## What not to do

- Don't reach for the `response-targets` extension to route errors. It moves a server decision
  into markup and adds a second mechanism next to `HX-Retarget`.
- Don't swap an `aria-live` region with `outerHTML`. Screen readers track the original node; a
  replaced live region announces nothing. Swap its `innerHTML`.
- Don't render into the `ResponseWriter` directly. Execute into a buffer first, so a template
  error becomes a 500 rather than a half-written 200.
- Don't add `hx-boost` to fix a navigation problem. Declare the nav-link partial instead; boost
  changes every link and form under it at once.
- Don't restate the htmx attribute reference here or in the repo; link
  [htmx.org/reference](https://htmx.org/reference/).
