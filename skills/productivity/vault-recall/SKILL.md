---
name: vault-recall
description: >-
  Answer the user's conceptual or technical question using their Obsidian vault as well as general
  knowledge. Use when the user asks a question (not a task) whose topic appears in the vault MOC
  digest from session start. Not for task instructions, not mid-task, and not for writing to the
  vault (that is to-vault).
---

Answer the question from the vault plus general knowledge, in one reply. Read-only.

## Inputs

- **Vault**: `~/Documents/personal/obsidian/my-obsidian-notes`. If it is missing, stop and say so.
- **Question**: the user's question as asked. Do not read the vault's `CLAUDE.md`; this skill
  holds the procedure.

## Retrieval

The budget per firing is `index.md`, then at most 2 MOCs (`Home MOC` excluded), then at most 5
pages under `wiki/` and `notes/`. Reading a page's `summary:` counts as opening it.

1. Read `index.md` and pick the MOCs that fit the question.
2. Open at most 2 of them and pick pages from their links.
3. For each page, read its `summary:` first; read the body only if the summary fits.
4. If the links miss, grep `summary:`, `aliases:` and filenames, and nothing else, then open
   hits within the same budget.

Stop when the budget is spent, even if more pages look relevant.

## Answer

Give one merged answer. A claim taken from a page opened in this firing carries `[[Page]]`
inline; anything uncited is general knowledge. Cite only pages opened in this firing.

## Not covered

If no opened page bears on the question, add exactly one line, `Your vault doesn't cover this.`,
and answer from general knowledge with no cites.

## Conflicts

Where an opened note disagrees with the current view (a Bootstrap 4 note, say), add one line per
clash: `Your note [[X]] says Y; current view is Z because W.` Leave the note unchanged.

## Keep

When the user says "keep this" or "save that", invoke `to-vault` with the answer as the Capture.

## Never

- Write to the vault.
- Exceed the budget.
- Cite a page not opened in this firing.
