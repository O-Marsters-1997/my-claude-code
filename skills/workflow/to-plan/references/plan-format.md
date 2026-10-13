# Plan format

`docs/plans/<feature>.md`, exactly this shape. `/to-tickets` copies each phase into its tickets, so
keep every phase self-contained: a reader with only one phase's section, plus the phases before it,
must be able to build it. See [example-plan.md](example-plan.md) for the density to match.

````markdown
# Plan: <Feature>

> Source: docs/prds/<feature>.md
> Label: project:<slug>
> Feature branch: <branch>

<One or two sentences: what the feature adds to the domain and where it lives.>

Definitions are frozen once tickets are cut: changing one is a `contract-change` issue that blocks
dependent work, never a quiet edit.

## Phases

1. [<PRD slice name>](#phase-1-<anchor>)
2. [<PRD slice name>](#phase-2-<anchor>)

## Phase 1: <PRD slice name>

Journeys: J1, J2

**Glossary**

- **<Term>**: <what it is, what it relates to>.

**New**

```<lang>
// path/to/file
<type, enum, port interface or table definition>
```

**Modified**

```<lang>
// path/to/existing/file
<the definition's final shape, or the part that matters>
```

## Phase 2: …

<A phase with nothing new or modified still gets a heading and Journeys line, saying so.>

## Notes

<A few lines at most: invariants and rules the definitions cannot say (who may write what), and
module ownership where it isn't obvious.>

### PRD drift

- <PRD decision> → <what the plan does instead, and why>. <Edited in PRD | Left in PRD>
````

Rules:

- One fenced block per file, headed by a `// path` (or `-- path`) line. Definitions only, no bodies,
  no call sites, no migrations.
- Omit **Glossary**, **New** or **Modified** in a phase when it has none.
- Every definition appears in exactly one phase. Nothing references a definition from a later phase.
- Omit `### PRD drift` when nothing drifted.
