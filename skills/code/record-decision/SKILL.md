---
name: record-decision
description: Capture a design or architecture decision as an ADR in docs/adr/ and update CONTEXT.md if terminology is affected. Use when the user says "document this decision", "record this design choice", "write an ADR", "capture the decision to X", or after any implementation where a non-obvious trade-off was made. Also trigger after refactors or design changes where the *why* isn't visible in the code — e.g. "we moved X to Y", "we renamed params to inputs", "we decided all db functions take DTOs". Don't wait for the user to ask explicitly: if a decision just landed that meets the ADR criteria below, offer to record it.
---

Take the decision as stated — no interview, no grilling. Write the ADR immediately.

## Steps

1. **Scan for existing ADRs** — check `docs/adr/` for the highest existing number and increment. Create the directory if it doesn't exist.

2. **Write the ADR** — follow the format below. One paragraph is enough.

3. **Update CONTEXT.md if relevant** — if the decision establishes or changes the canonical name for a concept, or resolves a naming ambiguity in the codebase, update the relevant entry. If no CONTEXT.md exists, only create one if the decision is fundamentally about naming/terminology.

## ADR format

Files live in `docs/adr/` as `NNNN-slug.md` (e.g. `0003-dtos-for-db-inputs.md`).

```md
# {Short title}

{1–3 sentences: context, what was decided, why.}
```

One paragraph is the goal. Add these sections only when they genuinely add value — most ADRs won't need them:

- **Considered Options** — when the rejected alternatives are worth remembering (someone will suggest them again)
- **Consequences** — when non-obvious downstream effects need calling out

## ADR criteria

Before writing, ask: **would a reader encountering the code cold reach the same conclusion without this ADR?** If yes, skip it — the code is already the documentation.

All three must hold before writing unsolicited. If the user asks for an ADR, write it regardless.

1. **Hard to reverse** — changing your mind later has real cost
2. **Surprising without context** — a future reader would wonder "why did they do it this way?" The code does not lead them to the answer.
3. **The result of a real trade-off** — there were genuine alternatives and one was chosen for specific reasons

Decisions that qualify tend to be *invisible* in the code: deliberate non-use of a pattern or library, ownership/boundary agreements not enforced by types, external constraints (SLAs, compliance, partner contracts), rejected alternatives that will be proposed again. Decisions that don't qualify: anything where the code clearly shows the pattern and a reader would just follow it.

## CONTEXT.md updates

Update only when the decision:
- Establishes or changes the canonical name for a concept (e.g. "params" → `CreateApplicationInput` in the `dto` package)
- Resolves a naming ambiguity in the codebase

Keep definitions to one sentence. List terms to avoid. Follow the existing file's format — don't restructure it.
