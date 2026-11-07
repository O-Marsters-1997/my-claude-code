---
name: to-vault
description: >-
  File what the user learnt or decided in this conversation into their Obsidian vault as a Note,
  linked into its MOCs and logged. Use when the user types /to-vault, or says "save this to my
  vault", "put this in obsidian", "add this to my notes", "remember this learning", "file this
  decision", "record this decision in the vault", or asks to keep a takeaway from the session in
  their second brain. Also offer it right after record-decision writes an ADR. Works from any repo.
  Not for writing ADRs (record-decision), Claude's own memory, or ingesting a source into wiki/.
argument-hint: "[what to capture, optional]"
---

File one Capture from this conversation into the vault as a Note, then reply with its path.

## Inputs

- **Vault**: `~/Documents/personal/my-obsidian-notes`. If it or its `CLAUDE.md` is missing, stop
  and say so; never create a vault.
- **Schema**: read the vault's `CLAUDE.md` every run, at least § Schema, § Ownership, § MOCs and
  § Log format. It is the only source of truth for frontmatter, MOC shape and log entries, so
  follow it over anything below if they differ.
- **Capture**: the free-text args if given, otherwise what the user learnt or decided in this
  conversation. If the conversation holds several, take the one most recent and most clearly
  theirs; if none stands out, ask which one in a single question.
- **Project**: `basename "$(git rev-parse --show-toplevel)"`. Omit `project` when that fails (not
  in a repo) or when the toplevel is the vault itself.

## What you may write

Only these, as the vault's Ownership table says:

- `notes/`: create a Note, or append to one. Never change existing lines, frontmatter included.
- `mocs/`: add list lines; create a project MOC when missing.
- `log.md`: append one entry.

Never touch `raw/`, `wiki/` page text, `index.md`, or anything in the repo you were invoked from
(the ADR stays exactly as `record-decision` wrote it).

## Steps

1. **Classify.** `decision` if the Capture is a choice between options (an ADR written this
   session is a strong signal); otherwise `learning`.
2. **Look for an existing Note.** List `notes/` and read any whose title or body covers the same
   idea. If one does, append to it (step 4) instead of creating a duplicate.
3. **Pick topics.** List `mocs/` and choose the MOCs the Capture belongs to; there must be at least
   one. With a `project`, always include `[[<Project> MOC]]` (Project is the repo name as written,
   e.g. `fleet MOC`). If that MOC doesn't exist, create it in the MOC shape from § MOCs with a
   one-line synthesis, and add `- [[<Project> MOC]]` under `## Notes` in
   `mocs/System Design MOC.md`. Without a project and with no closer MOC, use System Design MOC.
4. **Write the Note.**
   - New Note: `notes/<Title>.md`, title a short claim or name, no date. Frontmatter per § Schema
     with `type`, `origin: user`, `status: seedling`, `topics`, `project` (if any) and today's
     `date`.
   - Existing Note: append a `## <YYYY-MM-DD>` section at the end with the new material.
   - Write in the user's words: keep their phrasing and framing, trimmed, not Claude's
     explanation of it. The vault is for what they learnt, and a Claude-voiced summary is the
     reference-library failure the vault exists to avoid.
   - Link wiki concepts inline as `[[Page]]`, only for pages that exist in `wiki/`.
   - A `decision` body is exactly:

     ```markdown
     ## What
     ## Why
     ## Alternatives
     ## ADR
     ```

     `## Why` says why the chosen option wins; `## Alternatives` lists each rejected option with
     the reason it lost. `## ADR` holds a link to the ADR: its GitHub URL when the repo has a GitHub
     `origin` (`https://github.com/<owner>/<repo>/blob/<default-branch>/docs/adr/<file>`, branch
     from `git symbolic-ref --short refs/remotes/origin/HEAD`, else `main`), else its
     repo-relative path. Drop `## ADR` for a decision with no ADR (a non-code decision); its full
     reasoning then lives in the other sections.
5. **Link from MOCs.** In every MOC in `topics:`, add `- [[<Title>]]` under `## Notes` unless it is
   already listed. Agents can't see Obsidian backlinks, so both directions must be written.
6. **Log.** Append `## [YYYY-MM-DD] to-vault | <Title>` to `log.md`, per § Log format. One entry
   per run, titled with the Note; a project MOC created on the way needs none of its own.

## Reply

One line: the Note's path, e.g. `~/Documents/personal/my-obsidian-notes/notes/Leases beat locks for job claims.md`.
