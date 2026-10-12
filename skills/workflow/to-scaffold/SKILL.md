---
name: to-scaffold
description: >
  Turns one feature's PRD into compiling stubs and a contract: usage example, types and
  signatures, `not implemented` bodies, a draft PR into the feature branch, docs/plans/<feature>.md
  and the project:<slug> label. Use for "scaffold this feature", "stub out the modules", "write
  the contract". One feature per run. Slash-only.
disable-model-invocation: true
---

# To Scaffold

```
FEATURE — one feature per run    to-prd → to-scaffold → to-tickets → fleet
                                          ^^^^^^^^^^^ you are here
```

Put everything code can express (types, signatures, routes, migrations, integration interfaces)
in compiling stubs. The contract orders the work and holds what code cannot say.

Judge every boundary with the `codebase-design` vocabulary.

## Process

### 1. Locate the input

Read `docs/prds/<feature>.md`. If the user names a feature, use that file. If more than one PRD
exists and the target is not obvious, ask which one before starting; never scaffold several
features in one run. If no PRD exists, write the core inline (below) and carry on, rather than sending the user to `/to-prd`.

`<feature>` is the PRD's filename without `.md`; `<slug>` below is the same string.

**No PRD: write the core.** Ask a short interview, one question at a time and skipping what the
conversation already answers: the problem in one line, the journeys (each with a surface, 2 to 5
steps and a `→ <observable state>` final line), then the delivery slices. Write it to
`docs/prds/<feature>.md` (kebab-case, creating `docs/prds/` if missing) in `to-prd`'s core
format (its `references/project-doc.md`): `### Problem`, numbered `### J1.` journeys, and numbered
delivery slices each naming what is usable after it and its `Journeys: J…`. Every journey sits in
exactly one slice. Show it to the user, confirm, then continue with step 2. Skip the optional layer.

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
The slices are the PRD's delivery slices, in its order, each named as the PRD names it with its
journey numbers; split one only if the user agrees, keeping the parts under its name.
For each slice show what is usable after it in product terms. Ask whether the granularity and
boundaries feel right. Iterate until approved. Slices are vertical: each is demoable on its own
and cuts through every layer.

### 6. Open the scaffold PR

The feature branch is the one the user names, else the branch checked out when the run started.
Commit the stubs and usage example on a branch off it, push, and open a draft PR
whose base is the feature branch (never `main`):

```bash
gh pr create --draft --base <feature-branch> --title "Scaffold: <Feature>" --body "<summary of modules and the usage example>"
```

Record the PR number. Merging it freezes the signatures.

### 7. Create the feature label

The label is `project:<slug>` and goes on the contract's `Label:` line.

```bash
gh label list --search "project:<slug>" --json name -q '.[].name' | grep -qx "project:<slug>" || \
  gh label create "project:<slug>" -d "<Feature> (docs/plans/<slug>.md)" -c "<any unused hex>"
```

### 8. Write the contract

Create `docs/plans/` if missing and write `docs/plans/<feature>.md` in exactly this format:

```markdown
# Contract: <Feature>

> Source: docs/prds/<feature>.md
> Label: project:<slug>
> Scaffold PR: #<n>

The work is implementing the stubs. Signatures are frozen: changing one is a `contract-change` issue that blocks dependent work, never a quiet edit.

## Implementation order

### Slice 1: <PRD slice name: what is usable after it, in product terms> (J1, J2)
- `path/to/file.go:Symbol`: one line on what it does

## Notes

Only what code cannot say: rationale, ordering and invariants, gotchas, conventions, auth approach. About one screen.
```

Rules:

- Every slice heading matches a PRD delivery slice and lists that slice's journey numbers, in the
  PRD's order. A PRD slice with no stubs still gets a heading, saying so.
- Every stub appears under exactly one slice. Check by listing the stubs from the diff against the
  contract; none missing, none invented.
- One bullet per symbol, one line each. No signatures in the contract: the code is the source.
- Notes stay to about one screen. If something is expressible in code, move it into the stubs.

Commit the contract on the scaffold branch and push it, so it ships with the PR. Mention the label
and PR URL in the closing message so the user knows what `/to-tickets` will file under.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
