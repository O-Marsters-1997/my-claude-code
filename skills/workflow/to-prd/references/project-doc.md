# The PRD

One doc describing one feature, written to `docs/prds/<feature>.md`. It states what is built and in
what order users get it, in product terms. Present tense, plain words, one solution argued: no option
matrices. Name the technical concepts a reader needs, since the scope boundary is often technical.

## The core

Every PRD has all three. A small change is only these.

### Problem

One line: who is stuck, and on what.

### Journeys

A journey is how one person gets one thing done. Number them J1, J2, ...

- **Title** — the job, in the user's words.
- **Surface** — CLI, API or UI.
- **Steps** — 2 to 5, each something the user does or sees.
- **Final line** — `→ <observable state>`: what is now true that you could check. A journey with no
  such line is not finished.

```markdown
### J1. Save a view of the thread list

Surface: UI

1. Filter the thread list by status and assignee.
2. Choose "Save view" and name it.
3. Open the views menu.

→ The named view is listed, and selecting it restores the same filters.
```

### Delivery slices

Each slice is a shippable step. It says what is usable after it, in product terms, and lists the
journey numbers it delivers.

```markdown
1. **Save and reopen** — a user can save a filtered list and return to it. Journeys: J1, J2
2. **Share a view** — a saved view can be sent to a teammate. Journeys: J3
```

Every journey appears in exactly one slice.

## Optional layer

For bigger features: why, goal, concepts, out of scope, open questions. Include only what a reader
needs; the format is the user's call (see SKILL.md step 6). Resolved open questions stay, with the
answer that settled them.
