---
name: record-decision
description: Capture a design or architecture decision as an ADR in docs/adr/ and update CONTEXT.md if terminology is affected. Use when the user says "document this decision", "record this design choice", "write an ADR", "capture the decision to X", or after any implementation where a non-obvious trade-off was made. Also trigger after refactors or design changes where the *why* isn't visible in the code — e.g. "we moved X to Y", "we renamed params to inputs", "we decided all db functions take DTOs". Don't wait for the user to ask explicitly: if a decision just landed that meets the ADR criteria below, offer to record it.
---

Take the decision as stated — no interview, no grilling. Write the ADR immediately.

## Steps

1. **Scan for existing ADRs** — check `docs/adr/` for the highest existing number and increment. Create the directory if it doesn't exist.

2. **Write the ADR** — follow the format in [adr-format.md](../../shared/adr-format.md). One paragraph is enough.

3. **Update CONTEXT.md if relevant** — if the decision establishes or changes the canonical name for a concept, or resolves a naming ambiguity in the codebase, update the relevant entry. If no CONTEXT.md exists, only create one if the decision is fundamentally about naming/terminology.

## CONTEXT.md updates

Update only when the decision:
- Establishes or changes the canonical name for a concept (e.g. "params" → `CreateApplicationInput` in the `dto` package)
- Resolves a naming ambiguity in the codebase

Keep definitions to one sentence. List terms to avoid. Follow the existing file's format — don't restructure it.
