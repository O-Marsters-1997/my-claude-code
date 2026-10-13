# Example plan

A made-up feature in a Go + TypeScript + Postgres repo, showing every kind of definition a plan
keeps and the density to aim for. The languages are this example's; write a real plan in its repo's
own languages.

````markdown
# Plan: Reading shelves

> Source: docs/prds/reading-shelves.md
> Label: project:reading-shelves
> Feature branch: feature/reading-shelves

Readers group books into named shelves and get a weekly nudge about the shelf they're reading
from. Shelves live in the `library` context; the nudge goes through `notify`.

Definitions are frozen once tickets are cut: changing one is a `contract-change` issue that blocks
dependent work, never a quiet edit.

## Phases

1. [Shelve books](#phase-1-shelve-books)
2. [Weekly nudge](#phase-2-weekly-nudge)

## Phase 1: Shelve books

Journeys: J1, J2, J3

**Glossary**

- **Shelf**: a reader's named, ordered group of Books; private unless shared.
- **Shelf visibility**: whether a Shelf's other readers can see it (`private`, `shared`).

**New**

```sql
-- schema.sql
CREATE TABLE shelves (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reader_id  UUID NOT NULL REFERENCES readers(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private'
        CONSTRAINT shelves_visibility_check CHECK (visibility IN ('private', 'shared')),
    UNIQUE (reader_id, name)
);
CREATE TABLE shelf_books (
    shelf_id UUID NOT NULL REFERENCES shelves(id) ON DELETE CASCADE,
    book_id  UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    position INT  NOT NULL,
    PRIMARY KEY (shelf_id, book_id)
);
```

```go
// internal/dto/shelf.go
type ShelfVisibility string

const (
	VisibilityPrivate ShelfVisibility = "private"
	VisibilityShared  ShelfVisibility = "shared"
)

type Shelf struct {
	ID         string
	Name       string
	Visibility ShelfVisibility
	BookIDs    []string
}

type CreateShelfInput struct {
	Name       string          `json:"name"`
	Visibility ShelfVisibility `json:"visibility"`
}
```

```ts
// frontend/src/types/shelf.ts
export const shelfVisibilitySchema = z.enum(["private", "shared"]);
export const shelfSchema = z.object({
	ID: z.string(),
	Name: z.string(),
	Visibility: shelfVisibilitySchema,
	BookIDs: z.array(z.string()),
});
```

**Modified**

```go
// internal/dto/book.go
type Book struct {
	ID       string
	Title    string
	ShelfIDs []string
}
```

## Phase 2: Weekly nudge

Journeys: J4

**Glossary**

- **Nudge day**: the weekday a reader gets their nudge; none means no nudges.

**New**

```go
// internal/services/library/ports.go
type Nudger interface {
	Nudge(ctx context.Context, readerID string, shelf dto.Shelf) error
}
```

**Modified**

```sql
-- schema.sql
CREATE TABLE readers (
    -- existing columns
    nudge_day SMALLINT CHECK (nudge_day BETWEEN 0 AND 6)
);
```

```go
// internal/dto/reader.go
type Reader struct {
	ID       string
	Name     string
	NudgeDay *time.Weekday
}
```

## Notes

- A shelf's books are only ever reordered whole; `position` is rewritten in one statement.
- `notify` owns delivery; `library` only decides who is due, through `Nudger`.
````
