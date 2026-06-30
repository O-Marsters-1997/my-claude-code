# goquery API Reference

Package: `github.com/PuerkitoBio/goquery`

---

## Document Creation

```go
// From an io.Reader (use this for io.Reader/http.Response/os.File input)
doc, err := goquery.NewDocumentFromReader(r io.Reader) (*Document, error)

// From a pre-parsed html.Node
doc := goquery.NewDocumentFromNode(node *html.Node) *Document

// Clone an existing document
clone := goquery.CloneDocument(doc *Document) *Document
```

`NewDocumentFromReader` is the correct entry point for parsers — it decouples fetching from parsing.

---

## Selection — the core type

All traversal and extraction methods return or operate on `*Selection`.
Methods are chainable. An empty `Selection` (zero matches) never panics — it propagates silently,
so always check `sel.Length() > 0` before asserting results.

---

## Finding and Filtering

```go
// Find descendants matching a CSS selector
sel.Find(selector string) *Selection

// Filter current selection to elements matching selector
sel.Filter(selector string) *Selection

// Exclude elements matching selector
sel.Not(selector string) *Selection

// Keep only elements that have a descendant matching selector
sel.Has(selector string) *Selection

// Test if any element in the selection matches
sel.Is(selector string) bool
```

---

## Indexing

```go
sel.First() *Selection           // first matched element
sel.Last() *Selection            // last matched element
sel.Eq(index int) *Selection     // nth element (0-based; negative counts from end)
sel.Slice(start, end int) *Selection  // sub-slice
sel.Length() int                 // number of matched elements (also: Size())
sel.Get(index int) *html.Node    // raw html.Node at index
```

---

## Iteration

```go
// Iterate — i is 0-based index, s is the current element as *Selection
sel.Each(func(i int, s *goquery.Selection)) *Selection

// Iterate with early-exit — return false to stop
sel.EachWithBreak(func(i int, s *goquery.Selection) bool) *Selection

// Transform to []string
sel.Map(func(i int, s *goquery.Selection) string) []string

// Go 1.22+ range iterator
for i, s := range sel.EachIter() { ... }
```

---

## Text and HTML Extraction

```go
sel.Text() string                          // concatenated text of all matched nodes
sel.Html() (string, error)                 // inner HTML of the first matched node
sel.OuterHtml() (string, error)            // outer HTML (includes the element tag itself)
```

**Important:** `Text()` on a multi-element selection concatenates all text without separators.
Use `Each` + individual `Text()` calls when you need per-element values.

---

## Attributes

```go
sel.Attr(attrName string) (string, bool)          // value + exists flag
sel.AttrOr(attrName, defaultVal string) string    // value or default when absent

// Mutation (rarely needed in a parser):
sel.SetAttr(attrName, val string) *Selection
sel.RemoveAttr(attrName string) *Selection
```

---

## Traversal

```go
sel.Parent() *Selection              // immediate parent of each element
sel.Parents() *Selection             // all ancestors up to root
sel.ParentsUntil(selector) *Selection

sel.Children() *Selection            // direct children
sel.Contents() *Selection            // direct children including text nodes

sel.Siblings() *Selection
sel.Next() *Selection                // immediately following sibling
sel.NextAll() *Selection             // all following siblings
sel.NextUntil(selector) *Selection
sel.Prev() *Selection
sel.PrevAll() *Selection
sel.PrevUntil(selector) *Selection

sel.Closest(selector) *Selection     // nearest ancestor (or self) matching selector
```

---

## Set Operations

```go
sel.Add(selector string) *Selection          // add elements to the set
sel.AddSelection(sel2 *Selection) *Selection
sel.Union(sel2 *Selection) *Selection        // union of two selections
sel.Intersection(sel2 *Selection) *Selection // intersection
sel.End() *Selection                         // revert to previous selection (for chaining)
sel.AddBack() *Selection                     // add previous selection to current set
```

---

## Class Helpers

```go
sel.HasClass(class string) bool
sel.AddClass(classes string) *Selection
sel.RemoveClass(classes string) *Selection
sel.ToggleClass(classes string) *Selection
```

---

## Utility

```go
goquery.NodeName(sel *Selection) string  // tag name of the first node ("div", "a", etc.)
```

---

## Custom Matching (advanced)

Implement the `Matcher` interface to use custom logic instead of CSS selectors:

```go
type Matcher interface {
    Match(node *html.Node) bool
    MatchAll(node *html.Node) []*html.Node
    Filter(nodes []*html.Node) []*html.Node
}
```

Most traversal methods have a `*Matcher` variant (e.g., `FindMatcher`, `FilterMatcher`).
Use `goquery.Single(m)` to wrap a matcher that should stop after the first match — avoids
scanning the full tree.

---

## CSS Selector Reference

goquery uses the `cascadia` CSS selector library. Supported syntax:

| Pattern | Meaning |
|---------|---------|
| `div` | tag name |
| `.class` | class attribute |
| `#id` | ID attribute |
| `[attr]` | attribute presence |
| `[attr="val"]` | attribute equality |
| `[attr^="val"]` | starts with |
| `[attr$="val"]` | ends with |
| `[attr*="val"]` | contains |
| `a b` | descendant |
| `a > b` | direct child |
| `a + b` | adjacent sibling |
| `a ~ b` | general sibling |
| `:first-child` | first child of parent |
| `:last-child` | last child |
| `:nth-child(n)` | nth child (1-based) |
| `:not(selector)` | negation |
| `:has(selector)` | has descendant |
| `a, b` | union (matches either) |

**Not supported:** CSS4 pseudo-classes, `::before`/`::after`, computed style queries.
