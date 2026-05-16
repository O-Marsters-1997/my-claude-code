# Go Third-Party HTTP Client

Only build a full client if HTTP calls are shared across multiple callsites or you need middleware
(retry, rate-limiting, logging) applied automatically. For a single endpoint:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
req.Header.Set("Authorization", "Bearer "+token)
resp, err := client.Do(req)
```

---

## Constructor with Functional Options

```go
type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Option func(*Client)

func WithBaseURL(u string) Option           { return func(c *Client) { c.baseURL = u } }
func WithHTTPClient(hc *http.Client) Option { return func(c *Client) { c.httpClient = hc } }
func WithTimeout(d time.Duration) Option    { return func(c *Client) { c.httpClient.Timeout = d } }

func NewXxxClient(token string, opts ...Option) *Client {
	c := &Client{
		baseURL: "https://api.example.com",
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	// Auth wraps after options so WithHTTPClient can replace the base transport first.
	c.httpClient.Transport = &bearerTransport{
		token: token,
		inner: c.httpClient.Transport,
	}
	return c
}
```

## Auth at the Transport Layer

```go
type bearerTransport struct {
	token string
	inner http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone before mutating: the RoundTripper contract forbids modifying the original
	// request — it may be retried by an outer middleware layer.
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)
	return t.inner.RoundTrip(clone)
}
```

## Context-Aware Requests

Use `NewRequestWithContext` — `ctx` at construction can't be forgotten or missed:

```go
req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
```

Not this — `req` is context-free between lines and `WithContext` is easy to omit:

```go
req, _ = http.NewRequest(http.MethodGet, url, nil)
req = req.WithContext(ctx)
```

## Transport Chain Composition

Stack transports outermost-first; each layer only knows its immediate inner:

```go
var rt http.RoundTripper = http.DefaultTransport
rt = &loggingTransport{inner: rt}
rt = &bearerTransport{token: tok, inner: rt}

client := &http.Client{Transport: rt, Timeout: 30 * time.Second}
```

New middleware slots in as another layer — no restructuring needed.

## What NOT To Do

- **Don't use `http.DefaultClient`** — shared process-wide, no timeout.
- **Don't set auth per-request** in client methods — a missed callsite is a silent auth failure.
- **Don't ignore `ctx`** — pass it to `NewRequestWithContext`; callers cannot cancel what they can't reach.
