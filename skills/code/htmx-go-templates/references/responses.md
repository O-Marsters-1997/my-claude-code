# Responses: the handler side

## One helper set for htmx headers

Put the header names and predicates in one small package (`internal/web/hx` or wherever the repo
keeps web helpers). Handlers call these and never spell a header.

```go
package hx

import "net/http"

const (
	HeaderRequest        = "HX-Request"
	HeaderBoosted        = "HX-Boosted"
	HeaderHistoryRestore = "HX-History-Restore-Request"
	HeaderCurrentURL     = "HX-Current-URL"
	HeaderTarget         = "HX-Target"
	HeaderLocation       = "HX-Location"
	HeaderRedirect       = "HX-Redirect"
	HeaderRetarget       = "HX-Retarget"
	HeaderReswap         = "HX-Reswap"
	HeaderTrigger        = "HX-Trigger"
)

func IsRequest(r *http.Request) bool {
	return r.Header.Get(HeaderRequest) == "true"
}

func IsHistoryRestore(r *http.Request) bool {
	return r.Header.Get(HeaderHistoryRestore) == "true"
}

func WantsFragment(r *http.Request) bool {
	return IsRequest(r) && !IsHistoryRestore(r)
}

func Retarget(w http.ResponseWriter, selector, swap string) {
	w.Header().Set(HeaderRetarget, selector)
	w.Header().Set(HeaderReswap, swap)
}
```

htmx sends `HX-Request: true`, never an empty value, so a check for `!= ""` passes forged or
stale headers and a check for `== ""` inverts the logic. One predicate ends that argument.

Avoid inventing parallel signals (`?part=rows`, `?from=rail`) to pick a shape. If a route really
serves two fragments, give them two routes, or key on `HX-Target` through the helper.

## Declare the shape per route

Pick one of three shapes for each route and make it visible where the route is registered, not
buried in the handler body.

| Shape | Use when | htmx request gets | Non-htmx request gets |
|---|---|---|---|
| **page** | Navigation; the template uses `hx-select` to cut the region out | full page | full page |
| **fragment** | A region refreshes on its own (a list, a row, a counter) | the `{{define}}` block | full page containing that block |
| **redirect** | A state change that ends somewhere else | `HX-Location` | 303 |

A fragment route always has a full-page fallback, because history restore, a refresh and a
shared link all arrive without `HX-Request`. Make the fragment a `{{define}}` that the page
template also calls, so both paths render the same markup:

```go
type shape int

const (
	page shape = iota
	fragment
)

func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, sh shape, block string, vm any) {
	w.Header().Add("Vary", hx.HeaderRequest)
	name := "page"
	if sh == fragment && hx.WantsFragment(r) {
		name = block
	}
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, vm); err != nil {
		s.serverError(w, r, err)
		return
	}
	w.WriteHeader(status)
	buf.WriteTo(w)
}
```

Rendering into a buffer first means a template error becomes a clean 500 instead of a
half-written 200, and every header is set before the status line goes out.

`Vary: HX-Request` belongs on every route whose body depends on that header. Without it a shared
cache, or the browser's own cache on back/forward, can hand a bare fragment to a full-page load.

## Redirects

```go
func (s *Server) seeOther(w http.ResponseWriter, r *http.Request, url string) {
	if hx.IsRequest(r) {
		w.Header().Set(hx.HeaderLocation, url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
```

- **303, not 302**, for a non-htmx POST: 303 tells the browser to follow with a GET, so refresh
  never resubmits.
- **`HX-Location`** for htmx: htmx performs a client-side navigation to the URL, honouring
  `hx-target`/`hx-select` defaults. Use `HX-Redirect` only when a full reload is wanted (login,
  logout, a layout change).
- Never send a 3xx to an htmx request expecting a redirect. The XHR follows it and swaps the
  destination page into the target.

## Errors

htmx 2.x's default `responseHandling` swaps 2xx/3xx and drops 4xx/5xx with only an
`htmx:responseError` event. Two halves are needed:

**1. Client config** (in the layout's meta tag; see `templates.md`): swap the status you use for
recoverable errors. `responseHandling` replaces the default list, so restate the defaults.

```json
[
  {"code": "204", "swap": false},
  {"code": "[23]..", "swap": true},
  {"code": "422", "swap": true, "error": false},
  {"code": "[45]..", "swap": false, "error": true},
  {"code": "...", "swap": false}
]
```

**2. Server headers** choose where the error goes:

```go
func (s *Server) refuse(w http.ResponseWriter, r *http.Request, vm pageVM) {
	if hx.IsRequest(r) {
		hx.Retarget(w, "#flash", "innerHTML")
	}
	s.render(w, r, http.StatusUnprocessableEntity, fragment, "flash", vm)
}
```

- **422** for "understood, refused" (validation, a disallowed transition). The body is HTML meant
  for the user.
- **Form validation** retargets to the form's own `id` with `outerHTML`, re-rendering it with
  field errors.
- **Action refusals** retarget to the page's persistent flash region with `innerHTML`.
- **5xx** stays unswapped. Never put a stack trace or framework error page into the DOM; the
  client module turns `htmx:responseError` into a generic message in the same flash region.

Set headers before `WriteHeader`; headers written after it are dropped silently. Routing every
response through `render` keeps that ordering in one place.
