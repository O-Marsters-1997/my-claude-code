# Templates

## The layout head

The layout loads scripts by `src` only and carries htmx config in a meta tag, so nothing inline is
needed and a strict CSP stays possible.

```html
<meta name="htmx-config" content='{
  "allowEval": false,
  "includeIndicatorStyles": false,
  "responseHandling": [
    {"code": "204", "swap": false},
    {"code": "[23]..", "swap": true},
    {"code": "422", "swap": true, "error": false},
    {"code": "[45]..", "swap": false, "error": true},
    {"code": "...", "swap": false}
  ]
}'>
<script src="/assets/htmx.min.js" defer></script>
<script src="/assets/app.js" type="module"></script>
```

- `allowEval: false` makes htmx refuse `hx-on`, `js:` values in `hx-vals`, and trigger filters
  like `[event.key==='Enter']`. It enforces the no-inline rule for you.
- `includeIndicatorStyles: false` stops htmx injecting a `<style>` tag; ship the
  `.htmx-indicator` rule in the stylesheet instead.

## Fragments are `{{define}}` blocks the page also uses

```html
{{define "board"}}
<main id="main">
  <section id="board-rows" hx-get="/board/rows" hx-trigger="every 5s" hx-sync="this:drop">
    {{template "board-rows" .}}
  </section>
</main>
{{end}}

{{define "board-rows"}}
  {{range .Rows}}{{template "board-row" .}}{{end}}
{{end}}
```

The handler for `/board/rows` executes `board-rows`; the page executes `board`. One markup
source, so the fragment and the full page cannot disagree.

Name blocks after what they render (`board-rows`, `flash`, `nav-link`), not after the route.

## Repeated attribute bundles become a partial

When three or more elements share the same set of `hx-*` attributes, the set is a component.
`html/template` has no keyword arguments, so pass a small struct built by a `FuncMap` constructor:

```go
type NavLink struct {
	Href, Label string
	Current     bool
}

funcs := template.FuncMap{
	"navLink": func(href, label, current string) NavLink {
		return NavLink{Href: href, Label: label, Current: href == current}
	},
}
```

```html
{{define "nav-link"}}
<a href="{{.Href}}"
   hx-get="{{.Href}}" hx-select="#main" hx-target="#main" hx-swap="outerHTML" hx-push-url="true"
   {{if .Current}}aria-current="page"{{end}}>{{.Label}}</a>
{{end}}

{{template "nav-link" (navLink "/board" "Board" .Path)}}
```

The real `href` keeps the link working without JS and lets middle-click open a tab. Links built in
TypeScript (a graph, a canvas) emit the same attributes from one TS function, kept next to the
partial's name in the repo's AGENTS.md so the two stay paired.

## Forms

```html
<form action="/sessions/{{.ID}}/stop" method="post"
      hx-post="/sessions/{{.ID}}/stop" hx-target="#session" hx-swap="outerHTML"
      hx-disabled-elt="find button">
  <input type="hidden" name="csrf" value="{{.CSRF}}">
  <button type="submit">Stop</button>
</form>
```

- `action` + `method` first; htmx is an enhancement. With JS off, the POST gets a 303.
- `hx-disabled-elt` on every state-changing form, so a double-click cannot send twice.
- Validation errors come back as the re-rendered form (see `responses.md` § Errors); give the form
  a stable `id` the server can retarget to.

## Live regions

Put one persistent announcement region in the layout, outside anything that gets swapped:

```html
<div id="flash" role="status" aria-live="polite" aria-atomic="true"></div>
```

- Swap into it with `innerHTML`. Replacing the node itself (`outerHTML`) breaks the screen
  reader's subscription and nothing is announced.
- Use `role="alert"` (assertive) only for failures the user must act on now.
- A region that updates by polling gets `aria-live="polite"` only if each update is worth
  hearing; a ticking list that announces every five seconds is noise. Prefer announcing the
  change ("2 sessions finished") through `#flash`.

`aria-busy` on the region being replaced is set from the client module (see `client.md`), because
it must be cleared after the swap.

## Presentation-only branching

Templates may branch on presentation (`{{if .Rows}}…{{else}}empty state{{end}}`,
`{{if .Current}}`). They do not interpret domain values:

```html
{{if or (eq .Status "failed") (eq .Status "errored") (eq .Status "killed")}}  ← domain logic
<span class="badge {{.Badge.Class}}">{{.Badge.Label}}</span>                  ← presentation
```

Build the view model in Go, where it can be unit-tested and where adding a status is a compile
error rather than a missing branch.

## Nothing inline

No `<script>` blocks, `onclick=`, `onchange=`, `hx-on:*`, or `style=` that carries behaviour.
If a template seems to need one, the behaviour belongs in `client.md`'s module, keyed off a
`data-*` attribute the template emits.
