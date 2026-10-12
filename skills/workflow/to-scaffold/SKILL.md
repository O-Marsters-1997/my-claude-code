---
name: to-scaffold
description: >
  Turns one feature's PRD into compiling stubs and a contract: writes a usage example first,
  derives types and signatures from it, commits module boundaries with `not implemented` bodies,
  opens them as a draft PR into the feature branch, then writes docs/plans/<feature>.md and
  creates the project:<slug> label. Use for code features after /to-prd and before /to-tickets:
  "scaffold this feature", "stub out the modules", "write the contract". Covers exactly one
  feature per run. Slash-only.
disable-model-invocation: true
---

# PRD to Scaffold

```
FEATURE — one feature per run    to-prd → to-scaffold → to-tickets → fleet
                                          ^^^^^^^^^^^ you are here
```

Replaces the prose plan for code features. Anything code can express (types, signatures, routes,
migrations, integration interfaces) goes in compiling stubs, not in a document. The only document
is the contract, which orders the work and holds what code cannot say.

Use the `codebase-design` vocabulary (deep modules, seams, interface vs implementation) to judge
every boundary you draw.

## Process

### 1. Locate the input

Read `docs/prds/<feature>.md`. If the user names a feature, use that file. If more than one PRD
exists and the target is not obvious, ask which one before starting; never scaffold several
features in one run. If no PRD exists, stop and point the user to `/to-prd`.

`<feature>` is the PRD's filename without `.md`; `<slug>` below is the same string.

### 2. Explore the codebase

Read the existing architecture, patterns, test layout and build command. Note the command that
proves the stubs compile (for example `go build ./...`, `tsc --noEmit`, `cargo check`).

### 3. Write the usage example first

Before any type, write the code a caller of this feature would write, for each user journey in
the PRD. Put it where the repo keeps examples or tests (an `example_test.go`, a `examples/`
file, a test file) so it compiles against the stubs.

The example is the design: every type, function and route you introduce must be demanded by it.
If the example reads awkwardly, change the interface now, while it costs nothing.

### 4. Derive the stubs

From the example, derive the minimum types and signatures it needs, then place them in modules.
Judge each boundary with `codebase-design`:

- Prefer deep modules: a small, stable interface over substantial behaviour.
- Put a seam only where behaviour must vary or be tested in isolation.
- Name the modules that warrant isolated tests.

Write every stub with a body that fails loudly in the language's idiom (`panic("not implemented")`,
`unimplemented!()`, `throw new Error("not implemented")`). Include routes, migrations and
integration interfaces as stubs or declarations too. Stubs carry doc comments only where the
repo's comment rules permit; do not narrate behaviour in comments, the contract holds that.

Run the build command from step 2. The stubs must compile and the usage example must build
against them. Fix until it does. Do not implement behaviour.

### 5. Quiz the user

Present the module boundaries, the usage example, and the proposed slices as a numbered list.
For each slice show what is usable after it in product terms, and the PRD journeys it covers.
Ask whether the granularity and boundaries feel right. Iterate until approved. Slices are
vertical: each is demoable on its own and cuts through every layer.

### 6. Open the scaffold PR

Commit the stubs and usage example on a branch off the feature branch, push, and open a draft PR
whose base is the feature branch (never `main`):

```bash
gh pr create --draft --base <feature-branch> --title "Scaffold: <Feature>" --body "<summary of modules and the usage example>"
```

Record the PR number. Merging it freezes the signatures.

### 7. Write the contract

Create `docs/plans/` if missing and write `docs/plans/<feature>.md` in exactly this format:

```markdown
# Contract: <Feature>

> Source: docs/prds/<feature>.md
> Label: project:<slug>
> Scaffold PR: #<n>

The work is implementing the stubs. Signatures are frozen: changing one is a `contract-change` issue that blocks dependent work, never a quiet edit.

## Implementation order

### Slice 1: <what is usable after it, in product terms> (journeys 1, 2)
- `path/to/file.go:Symbol`: one line on what it does

## Notes

Only what code cannot say: rationale, ordering and invariants, gotchas, conventions, auth approach. About one screen.
```

Rules:

- Every stub appears under exactly one slice. Check by listing the stubs from the diff against the
  contract; none missing, none invented.
- One bullet per symbol, one line each. No signatures in the contract: the code is the source.
- Notes stay to about one screen. If something is expressible in code, move it into the stubs.

### 8. Create the feature label

`to-tickets` files tickets under this label and `/fleet` selects by it. The label is `project:<slug>`
and is the `Label:` line in the contract header.

```bash
gh label list --search "project:<slug>" | grep -q "^project:<slug>\b" || \
  gh label create "project:<slug>" -d "<Feature> (docs/plans/<slug>.md)" -c "<any unused hex>"
```

Commit the contract (on the scaffold branch, so it ships with the PR) and mention the label and
PR URL in the closing message so the user knows what `/to-tickets` will file under.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
