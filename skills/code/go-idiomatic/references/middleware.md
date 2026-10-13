# HTTP Middleware Reference

Request checks that every handler needs belong in one middleware, not repeated in handlers. A check
written per handler is missing from the next handler someone adds.

## Contents

- [Content-Type](#content-type)
- [CSRF](#csrf)

---

## Content-Type

One middleware requires `Content-Type: application/json` on every request with a body and answers
`415 Unsupported Media Type` otherwise. Handlers never check it.

A browser sends a cross-origin `text/plain` or form POST without a preflight, so a JSON endpoint
that decodes whatever arrives accepts requests no page of yours made.

```go
func requireJSON(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.ContentLength != 0 {
            mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
            if err != nil || mt != "application/json" {
                http.Error(w, "content type must be application/json", http.StatusUnsupportedMediaType)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

`ContentLength` is `-1` for a chunked body, so the `!= 0` test covers it. Routes that take a file
upload or a form are registered outside this middleware, not exempted inside it.

---

## CSRF

- **Session cookies are `SameSite=Lax` by default**, with `HttpOnly` and `Secure`. Lax stops a
  cross-site POST from carrying the cookie while keeping top-level navigation logged in.
- **A middleware checks `Origin` against an allowlist on unsafe methods** (anything but `GET`,
  `HEAD`, `OPTIONS`). SameSite treats every subdomain of the site as same-site; the `Origin` check
  does not. Reject a missing `Origin` on an unsafe method unless the request carries no cookies.
- **`SameSite=None` only together with a CSRF token.** Without SameSite, the token is the only
  defence, so a cookie that must cross sites (an embedded widget, a separate API domain) needs one.

```go
func checkOrigin(allowed map[string]bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            switch r.Method {
            case http.MethodGet, http.MethodHead, http.MethodOptions:
                next.ServeHTTP(w, r)
                return
            }
            origin := r.Header.Get("Origin")
            if origin == "" && len(r.Cookies()) == 0 {
                next.ServeHTTP(w, r)
                return
            }
            if !allowed[origin] {
                http.Error(w, "forbidden origin", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

On Go 1.25+, prefer `http.CrossOriginProtection`: it makes the same check from `Sec-Fetch-Site`,
falls back to `Origin` against `Host`, and takes the allowlist through `AddTrustedOrigin`.
