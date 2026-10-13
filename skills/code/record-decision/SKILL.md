---
name: record-decision
description: >-
  Capture a design or architecture decision as an ADR in docs/adr/ and update CONTEXT.md if terminology is affected. Use when the user says "document this decision", "record this design choice", "write an ADR", "capture the decision to X", or after any implementation where a non-obvious trade-off was made. Also trigger after refactors or design changes where the *why* isn't visible in the code: if a decision just landed that meets the ADR criteria below, offer to record it.
---

Take the decision as stated — no interview, no grilling. Write the ADR immediately. The one check is against the global skills below; it is not an interview about the decision's merits.

## Steps

1. **Scan for existing ADRs** — check `docs/adr/` for the highest existing number and increment. Create the directory if it doesn't exist.

2. **Check against the global skills** — compare the decision with `go-idiomatic`, `go-testing` and `postgres-best-practices` where they cover its area. Global skills outrank repo ADRs: a decision that conflicts with one is rewritten to follow the skill, never recorded as an exception. Tell the user which rule it conflicted with and how the ADR changed. When an existing ADR conflicts, amend that ADR rather than writing a new one beside it. If the user insists on the exception, the skill is what changes; point them to `skill-updater`.

3. **Write the ADR** — `docs/adr/NNNN-kebab-title.md`, matching the existing ADRs' format. With none to match, use a `# ADR NNNN — Title` heading and one paragraph covering the context, the decision and its consequences.

4. **Update CONTEXT.md if relevant** — if the decision establishes or changes the canonical name for a concept, or resolves a naming ambiguity in the codebase, update the relevant entry. If no CONTEXT.md exists, only create one if the decision is fundamentally about naming/terminology.

## CONTEXT.md updates

Update only when the decision:

- Establishes or changes the canonical name for a concept (e.g. "params" → `CreateApplicationInput` in the `dto` package)
- Resolves a naming ambiguity in the codebase

Keep definitions to one sentence. List terms to avoid. Follow the existing file's format — don't restructure it.
