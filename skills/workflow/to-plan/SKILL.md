---
name: to-plan
description: >
  Turns one feature's PRD into a plan: a usage example per journey, then the types, signatures
  and changes to existing symbols each slice needs, written as markdown in docs/plans/<feature>.md
  with the project:<slug> label. No code, branch or PR; the user's review of the file is the gate.
  Use for "plan this feature", "write the contract", "what are the signatures". One feature per
  run. Slash-only.
disable-model-invocation: true
---

# To Plan

```
FEATURE — one feature per run    to-prd → to-plan → to-tickets → fleet
                                          ^^^^^^^ you are here
```

Write the technical design as markdown signatures, not code. The user reads types and interfaces to
check that the product maps onto the domain; writing them as compiling stubs for a whole feature
costs far more and adds nothing implementers can't do when they turn each signature into code.

Judge every boundary with the `codebase-design` vocabulary.

## Process

### 1. Locate the input

Read `docs/prds/<feature>.md`. If the user names a feature, use that file. If more than one PRD
exists and the target is not obvious, ask which one before starting; never plan several features in
one run. If no PRD exists, write the core inline (below) and carry on, rather than sending the user
to `/to-prd`.

`<feature>` is the PRD's filename without `.md`; `<slug>` below is the same string.

**No PRD: write the core.** Ask a short interview, one question at a time and skipping what the
conversation already answers: the problem in one line, the journeys (each with a surface, 2 to 5
steps and a `→ <observable state>` final line), then the delivery slices. Write it to
`docs/prds/<feature>.md` (kebab-case, creating `docs/prds/` if missing) in `to-prd`'s core format
(its `references/project-doc.md`): `### Problem`, numbered `### J1.` journeys, and numbered delivery
slices each naming what is usable after it and its `Journeys: J…`. Every journey sits in exactly one
slice. Show it to the user, confirm, then continue with step 2. Skip the optional layer.

### 2. Explore the codebase

Read the existing architecture, patterns, naming and test layout, and the symbols the feature will
have to change. Write nothing outside `docs/` in this skill: no stubs, no branch, no build.

**PRD drift.** When exploration shows a PRD decision won't hold (a cost model, an ownership rule, a
place the logic can't live), list each one and ask the user: edit the PRD now, or record it under
`PRD drift` in the plan's Notes. Never leave a changed decision only in the plan without asking,
or the PRD goes stale while still looking authoritative.

### 3. Write the usage example first

For each journey, write the code a caller would write, as one fenced block in the repo's language
under the journey's number. The example is the design: every signature in the plan must be used by
some example, and any the examples never use is cut. If an example reads awkwardly, change the
interface now, while it costs nothing.

### 4. Derive the signatures per slice

From the examples, derive the minimum types, functions, routes, migrations and integration
interfaces, and place each in the slice that first needs it. Judge each boundary with
`codebase-design`: prefer deep modules, put a seam only where behaviour must vary or be tested in
isolation, and name the modules that warrant isolated tests.

- **Nothing points forward.** Whatever a slice needs is defined in that slice or an earlier one. If
  slice 3 needs a type and slice 1 also uses it, it goes in slice 1. Tickets are built in parallel
  from these blocks with no compiled stubs, so a forward reference is a slice that cannot start.
- **Shared, only if needed.** A `## Shared` section above the slices holds what several slices need
  before any of them can start (a migration, a core type). Leave it out when the first slice can
  carry it.
- **New vs changed.** New symbols go in a fenced `new:` block with full signatures. Edits to existing
  symbols are `change:` bullets naming the symbol and what changes, since their code already exists.

### 5. Quiz the user

Present the module boundaries, the usage examples and the slices as a numbered list. The slices are
the PRD's delivery slices, in its order, each named as the PRD names it with its journey numbers;
split one only if the user agrees, keeping the parts under its name. For each slice show what is
usable after it in product terms. Ask whether the granularity and boundaries feel right. Iterate
until approved. Slices are vertical: each is demoable on its own and cuts through every layer.

### 6. Settle the feature branch and label

The feature branch is the one the user names, else the branch checked out when the run started.
If that is the default branch, propose `feature/<slug>` and ask; tickets never build on the default
branch. The plan records the name only; this skill creates no branch.

```bash
gh label list --search "project:<slug>" --json name -q '.[].name' | grep -qx "project:<slug>" || \
  gh label create "project:<slug>" -d "<Feature> (docs/plans/<slug>.md)" -c "<any unused hex>"
```

### 7. Write the plan

Write `docs/plans/<feature>.md` in the format in [references/plan-format.md](references/plan-format.md).
If the file already exists, show what would change and ask before overwriting: it may hold an
approved plan that tickets were already cut from.
Check coverage: every new or changed symbol appears in exactly one place (Shared or one slice), and
every slice heading matches a PRD delivery slice. Run `git check-ignore -q docs/plans/<feature>.md`
and say in the closing message whether the plan is local-only or will be committed.

Close by naming the plan path, the label and the feature branch, and ask the user to review the file:
their approval is the gate `/to-tickets` waits on.

When finished, ask: 'Would you like to log feedback? (yes/no)'. If yes, invoke skill-feedback-collector passing this skill's name and path.
