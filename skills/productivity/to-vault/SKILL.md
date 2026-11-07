---
name: to-vault
description: >-
  File what the user learnt or decided into their Obsidian vault as atomic Notes, one idea each,
  linked into MOCs and logged. Use when the user types /to-vault, or says "save this to my
  vault", "put this in obsidian", "add this to my notes", "file this decision", "record this
  decision in the vault", or asks to keep takeaways from the session in their second brain. Also
  offer it right after record-decision writes an ADR. Works from any repo. Not for writing ADRs
  (record-decision), Claude's own memory, or ingesting a source into wiki/.
argument-hint: "[--inbox | what to capture; \"quoted text\" is kept verbatim]"
---

File what the user names into the vault as atomic Notes, then reply with their paths.

## Inputs

- **Vault**: `~/Documents/personal/my-obsidian-notes`. If it or its `CLAUDE.md` is missing, stop
  and say so; never create a vault.
- **Schema**: read the vault's `CLAUDE.md` every run, at least § Schema, § Ownership, § MOCs and
  § Log format. It is the only source of truth for frontmatter, MOC shape and log entries, so
  follow it over anything below if they differ.
- **Capture**: the user decides what goes in. The args name it, and the conversation supplies the
  detail. With no args, list the candidate learnings and decisions from the conversation in one
  short numbered list and ask which to file; write nothing until they answer. Never widen the
  Capture beyond what they named.
- **Exact wording**: text the user wrote inside double quotes in the args is theirs verbatim.
  Put it in the Note it belongs to unchanged, without the quote marks, as its own paragraph.
  Everything else you may summarise and reword.
- **`--inbox`**: opt-in, only when the user typed it. Skip Capture and Steps. Read § Triage in
  the vault's `CLAUDE.md`; if that section is missing, stop and say so, writing nothing.
  Otherwise follow it in full, including showing its plan and writing only after approval.
- **Project**: `basename "$(git rev-parse --show-toplevel)"`. Omit `project` when that fails (not
  in a repo) or when the toplevel is the vault itself.

## What you may write

Only these, as the vault's Ownership table says:

- `notes/`: create Notes, or append to existing ones. Never change existing lines, frontmatter
  included.
- `mocs/`: add list lines; create a project MOC when missing.
- `log.md`: append entries.

With `--inbox` only, you may also delete the `inbox/` files that were processed, after approval.

Never touch `raw/`, `wiki/` page text, `index.md`, or anything in the repo you were invoked from
(an ADR stays exactly as `record-decision` wrote it).

## Steps

1. **Split.** Break the Capture into atomic ideas: one claim, technique or decision per Note,
   each understandable on its own. A capture usually yields one to a few Notes, not one long
   one; never pad it out. Each quoted passage goes with the idea it is about; a quote that stands
   alone becomes its own Note.
2. **Classify each.** `decision` if it is a choice between options (an ADR written this session
   is a strong signal); otherwise `learning`.
3. **Look for existing Notes.** List `notes/` and read any whose title or body covers the same
   idea. If one does, append to it (step 5) instead of creating a duplicate.
4. **Pick topics.** List `mocs/` and choose the MOCs each idea belongs to; at least one per Note.
   With a `project`, always include `[[<Project> MOC]]` (Project is the repo name as written,
   e.g. `fleet MOC`). If that MOC doesn't exist, create it in the MOC shape from § MOCs with a
   one-line synthesis, and add `- [[<Project> MOC]] — <when to open it>` under `## Notes` in
   `mocs/System Design MOC.md`. Without a project and with no closer MOC, use System Design MOC.
5. **Write the Notes.**
   - New Note: `notes/<Title>.md`, title a short claim or name, no date. Frontmatter per § Schema
     with `type`, `origin: user`, `status: seedling`, `topics`, `project` (if any) and today's
     `date`.
   - Existing Note: append a `## <YYYY-MM-DD>` section at the end with the new material.
   - Keep it short and plain. Add no specifics the conversation didn't contain (names, values,
     example code); a filled-in guess reads later as something the user learnt. Link wiki concepts inline as `[[Page]]`, only for pages that exist
     in `wiki/`. Link Notes from the same run to each other where one builds on another.
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
6. **Link from MOCs.** In every MOC in a Note's `topics:`, add `- [[<Title>]] — <when to open it>`
   under `## Notes` unless it is already listed. The phrase is short and drawn from the Note's
   content, so a reader can skip the Note without opening it. Agents can't see Obsidian
   backlinks, so both directions must be written.
7. **Log.** Append one `## [YYYY-MM-DD] to-vault | <Title>` to `log.md` per Note written or
   appended to, per § Log format. A project MOC created on the way needs none of its own.

## Reply

One line per Note: its path, e.g.
`~/Documents/personal/my-obsidian-notes/notes/Leases beat locks for job claims.md`.
