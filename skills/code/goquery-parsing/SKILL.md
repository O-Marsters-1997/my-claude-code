---
name: goquery-parsing
description: >
  Generate robust, maintainable Go HTML parsers using goquery. Use whenever given an HTML snapshot
  and asked to extract structured data from it — product listings, articles, job boards, tables,
  search results, any page where you need to pull out fields and return typed Go structs. Produces
  stable CSS selectors, idiomatic goquery code, and parsers that survive minor markup changes.
  Trigger phrases: "goquery", "parse html", "scrape this", "extract from html", "html parser",
  "web scraping", "parse this page", "extract data from html", "html snapshot", "parse the page",
  "write a scraper", "write a parser", "extract these fields". Also trigger when the user pastes
  raw HTML and asks for structured data out of it, even without naming goquery explicitly.
user-invocable: true
allowed-tools: Read Write Edit Bash(go:*) Bash(gofmt:*) Bash(goimports:*)
---

# goquery-parsing

Generate Go parsers from HTML snapshots. The goal is code that extracts the right data
reliably and keeps working when a page gets restyled or minor markup is shuffled.

For the full API method signatures, see `references/api.md`.

---

## Workflow

1. **Read the HTML** — scan the full snapshot before writing a single selector
2. **Name the target data** — define the Go struct that will hold the output
3. **Pick stable selectors** — see the hierarchy below
4. **Write the parser** — follow the code structure below
5. **Verify** — compile, test against the snapshot, check edge cases

---

## Selector Stability Hierarchy

A selector is only as durable as the thing it targets. From most to least stable:

| Tier | Examples | Why it lasts |
|------|----------|--------------|
| **1 — Semantic data attributes** | `[data-product-id]`, `[data-testid="price"]` | Set explicitly for machine consumption; survive redesigns |
| **2 — IDs** | `#main-content`, `#search-results` | Unique; high cost to change |
| **3 — ARIA / role attributes** | `[role="main"]`, `[aria-label="price"]` | Accessibility-driven; stable across CSS refactors |
| **4 — Semantic HTML elements** | `article`, `time[datetime]`, `address`, `main`, `nav` | Tag semantics don't change with styling |
| **5 — Meaningful class names** | `.product-title`, `.job-listing`, `.price-amount` | Component names outlast utility classes |
| **6 — Structural attribute combos** | `a[href^="/product/"]`, `img[alt]`, `input[name="q"]` | Content-based; not style-dependent |
| **7 — Tag + structure (last resort)** | `article h2`, `ul.results > li` | Fragile to layout changes; use only when tiers 1–6 yield nothing |

**Avoid:**
- Utility/positional classes: `mt-4`, `flex`, `col-md-6`, `text-sm` — they change constantly
- `:nth-child()` on variable-length lists — inserted ads or banners break the index
- Deep descendant chains (more than 3 levels) — brittle to layout refactors
- Chained traversals: `.Next().Next().Find()` — index arithmetic dressed as code

When the HTML offers nothing better than tier 7, note it with a comment explaining which
attribute to watch for if the page is ever cleaned up.

---

## Code Structure

### Define the output struct first

Name the struct after the entity, not the page. Fields should be the minimal set the caller
actually needs — leave optional/nullable fields as empty strings rather than pointers unless
presence vs absence genuinely matters to the caller.

```go
type Product struct {
    Name        string
    Price       string
    URL         string
    ImageURL    string
    Description string // empty when absent — caller checks len()
}
```

### Parse function signature

Accept `io.Reader` so the caller controls the source (file, HTTP response, string, test fixture).
Return a descriptive error rather than silently returning an empty slice.

```go
func ParseProducts(r io.Reader) ([]Product, error) {
    doc, err := goquery.NewDocumentFromReader(r)
    if err != nil {
        return nil, fmt.Errorf("parsing HTML: %w", err)
    }

    var products []Product
    doc.Find("[data-product-id]").Each(func(_ int, s *goquery.Selection) {
        products = append(products, extractProduct(s))
    })

    if len(products) == 0 {
        return nil, fmt.Errorf("no products found: selector may need updating")
    }
    return products, nil
}
```

### Extract helpers for repeated patterns

When the same field appears across multiple types, or when a single field needs >1 line to
extract, pull it into a small helper. `extractProduct` is a receiver-free function that takes
a `*goquery.Selection` and returns the struct:

```go
func extractProduct(s *goquery.Selection) Product {
    return Product{
        Name:     text(s, "[data-product-name]"),
        Price:    text(s, ".price-amount"),
        URL:      s.Find("a[href]").AttrOr("href", ""),
        ImageURL: s.Find("img").AttrOr("src", ""),
    }
}

// text finds the first matching element and returns trimmed text content.
func text(s *goquery.Selection, selector string) string {
    return strings.TrimSpace(s.Find(selector).First().Text())
}
```

### Single-item parsers

When extracting one record from a page (e.g., an article detail), `Find().First()` returns
a zero-length selection if nothing matched — check with `.Length()`:

```go
func ParseArticle(r io.Reader) (Article, error) {
    doc, err := goquery.NewDocumentFromReader(r)
    if err != nil {
        return Article{}, fmt.Errorf("parsing HTML: %w", err)
    }

    titleSel := doc.Find("h1[data-article-title], article > h1").First()
    if titleSel.Length() == 0 {
        return Article{}, fmt.Errorf("article title not found")
    }

    return Article{
        Title:  strings.TrimSpace(titleSel.Text()),
        Author: text(doc.Selection, "[rel=author]"),
        Date:   doc.Find("time[datetime]").First().AttrOr("datetime", ""),
        Body:   extractBody(doc),
    }, nil
}
```

---

## Key Extraction Patterns

### Text content

Always `strings.TrimSpace()` — HTML text nodes typically contain leading/trailing whitespace
and newlines from indentation.

```go
title := strings.TrimSpace(s.Find("h1").Text())
```

For multi-paragraph body text, collect paragraphs individually:

```go
var paragraphs []string
s.Find("p").Each(func(_ int, p *goquery.Selection) {
    if t := strings.TrimSpace(p.Text()); t != "" {
        paragraphs = append(paragraphs, t)
    }
})
body := strings.Join(paragraphs, "\n\n")
```

### Attributes

`Attr` returns `(string, bool)` — use the bool when presence matters:

```go
href, exists := s.Attr("href")
if !exists {
    return // skip elements without a link
}
```

Use `AttrOr` when absence is acceptable and you have a sensible default:

```go
imgURL := s.Find("img").AttrOr("src", "")
```

### Dates and times

Prefer `datetime` attribute on `<time>` elements — it's machine-readable and consistent:

```go
date := doc.Find("time[datetime]").AttrOr("datetime", "")
```

Fall back to visible text only when `datetime` is absent.

### Links

Resolve relative URLs after extraction — goquery returns raw `href` values:

```go
href := s.Find("a").AttrOr("href", "")
if href != "" && !strings.HasPrefix(href, "http") {
    href = baseURL + href
}
```

### Tables

Row-by-row extraction; skip the header row:

```go
doc.Find("table tbody tr").Each(func(i int, row *goquery.Selection) {
    cells := row.Find("td")
    if cells.Length() < 3 {
        return // malformed row — skip
    }
    records = append(records, Record{
        Name:  strings.TrimSpace(cells.Eq(0).Text()),
        Value: strings.TrimSpace(cells.Eq(1).Text()),
        Date:  strings.TrimSpace(cells.Eq(2).Text()),
    })
})
```

---

## Fallback Selectors

When a primary selector is uncertain, chain alternatives with a comma — goquery tries all:

```go
doc.Find("[data-price], .price-amount, .product-price, span.price")
```

Or try selectors in priority order and take the first that matches:

```go
priceSel := doc.Find("[data-price]")
if priceSel.Length() == 0 {
    priceSel = doc.Find(".price-amount")
}
price := strings.TrimSpace(priceSel.First().Text())
```

---

## Self-Check (minimal test)

Leave a runnable check in a `_test.go` file that reads the HTML snapshot and asserts the key
fields are non-empty. The test should fail if a future page change breaks the selectors:

```go
func TestParseProducts(t *testing.T) {
    f, err := os.Open("testdata/products.html")
    if err != nil {
        t.Fatal(err)
    }
    defer f.Close()

    got, err := ParseProducts(f)
    if err != nil {
        t.Fatal(err)
    }
    if len(got) == 0 {
        t.Fatal("expected products, got none")
    }
    first := got[0]
    if first.Name == "" {
        t.Error("product name is empty")
    }
    if first.Price == "" {
        t.Error("product price is empty")
    }
}
```

Save the HTML snapshot to `testdata/` so the test is self-contained.

---

## Import Block

```go
import (
    "fmt"
    "io"
    "strings"

    "github.com/PuerkitoBio/goquery"
)
```

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| `.Text()` on a multi-element selection | `.First().Text()` or `.Eq(i).Text()` — `.Text()` on many nodes concatenates all text |
| Ignoring `Attr` bool | Use `AttrOr` or check the bool — missing attrs return `""` regardless |
| Selector targets a container, not the value | Scope with `.Find()` inside `Each` rather than selecting text containers globally |
| Selecting by presentational class | Find a `data-*` or semantic anchor; note the fragility if none exists |
| Returning empty slice on zero results | Return an error — silent empty is harder to debug than an explicit message |
| Fetching URLs inside the parser | Parser takes `io.Reader`, caller fetches — keeps the parser testable without HTTP |
