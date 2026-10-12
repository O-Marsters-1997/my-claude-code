# Plan format

`docs/plans/<feature>.md`, exactly this shape. `/to-tickets` copies each slice's blocks into its
tickets, so keep every block self-contained: a reader with only one slice's section must be able to
write that slice's code.

````markdown
# Plan: <Feature>

> Source: docs/prds/<feature>.md
> Label: project:<slug>
> Feature branch: <branch>

Signatures are frozen once tickets are cut: changing one is a `contract-change` issue that blocks
dependent work, never a quiet edit.

## Usage

### J1. <journey name>
```<lang>
<the caller's code for this journey>
```

## Shared

<Only if several slices need it before any can start; otherwise omit the section.>

new:
```<lang>
// path/to/file.go
type Thing struct { … }
func NewThing(…) (*Thing, error)
```

## Slices

### Slice 1: <PRD slice name: what is usable after it, in product terms> (J1, J2)

new:
```<lang>
// path/to/file.go
func (s *Service) DoThing(ctx context.Context, id ID) (Result, error)
```

- change: `path/to/existing.go:upsertCompany` — also writes `kind`
- route: `POST /things/{id}` → `Service.DoThing`

### Slice 2: … (J3)

<A PRD slice with no new or changed symbols still gets a heading, saying so.>

## Notes

Only what signatures cannot say: rationale, ordering and invariants, gotchas, conventions, auth
approach. About one screen.

### PRD drift

- <PRD decision> → <what the plan does instead, and why>. <Edited in PRD | Left in PRD>
````

Rules:

- Group a slice's `new:` signatures by file, one fenced block per file or one block with a
  `// path` line before each file's declarations. Signatures and type shapes only, no bodies.
- One `change:` bullet per existing symbol, one line each.
- Every new or changed symbol appears in exactly one place. Nothing references a symbol defined in a
  later slice.
- Omit `### PRD drift` when nothing drifted.
