---
name: to-seams
description: >
  Records an agreed-but-unlanded cross-repo interface as a seam: the file a consumer's agent gets
  pasted into its prompt as ground truth, plus the `[[seam]]` block command-centre reads for the
  retirement pointer. Use when two repos need to build against an interface neither has landed
  yet — "record a seam for X", "we agreed on this schema, seam it", "add a seam so Y can build
  against Z's new field". Halts for approval before writing either file.
disable-model-invocation: true
---

# Record a seam

A seam is a pasted-in agreement, not a parsed contract: the file goes into a consumer's prompt
byte-for-byte, so what you write here is what the agent reads as ground truth until the producer
lands the real thing (design §6).

## Process

1. **Find the workspace.** Same resolution `to-tickets` step 7 uses:

   ```bash
   ls "$(git rev-parse --show-toplevel)/../.claude/command-centre.toml" 2>/dev/null
   ```

   If it's absent, say so and stop — a seam with no `[[seam]]` config to read it is pointless.

2. **Ask for the four things a seam needs:**
   - **Name** — a short slug (`customer-tier-field`). Becomes the seam filename and the
     `[[seam]]` key.
   - **Producer repo** — the `[[repo]]` name in `command-centre.toml` that will land the real
     thing.
   - **Consumer repo(s)** — whichever repos are building against the agreement now. Not written
     to config — only `producer_repo` and `lands_at` are read by anything, the retirement pointer
     — this is so you know which tickets' next `to-tickets` run should list this seam by name in
     their `seams = [...]`.
   - **`lands_at`** — one or more paths, in the producer repo, the real thing will live at once it
     merges. Optional: if the landing path isn't settled yet, leave it empty and say so — an empty
     `lands_at` just means "no retirement", not an error.

3. **Get the agreed interface itself.** Ask the user to paste or dictate exactly what a consumer's
   agent should read — a schema, a type, a function signature, whatever the two sides agreed on.
   Paste it in verbatim: nothing here is parsed, marked up, or addressed by symbol (design §6). If
   the user hands you prose describing the agreement rather than the interface text itself, ask
   for the actual text instead of summarising it back to them.

4. **Show both writes, then halt for approval.** Before touching disk, show:
   - The seam file's exact planned content, in full.
   - The `[[seam]]` block exactly as it will be appended:
     ```toml
     [[seam]]
     name          = "<name>"
     producer_repo = "<producer repo>"
     lands_at      = ["<path>", …]
     ```
   Wait for explicit confirmation before writing either file (§10) — a seam silently written is a
   prompt silently changed under whoever cuts a task against it next.

5. **Write, in this order:** the seam file at `<workspace>/.claude/seams/<name>`, then append the
   `[[seam]]` block to `<workspace>/.claude/command-centre.toml`. Upsert on `name`: if a
   `[[seam]]` block for this name already exists, replace it in place rather than duplicating it,
   and say so.

## Finishing up

Tell the user the seam's name and which consumer tickets should list it in `seams = [...]` the
next time `to-tickets` runs for those repos.
