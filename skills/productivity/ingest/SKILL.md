---
name: ingest
description: >-
  Ingest one external source (an article, course lesson, doc, PDF or URL) into the user's
  Obsidian vault: save the original to raw/, then build wiki/ pages per the vault's § Ingest.
  Use when the user types /ingest, or says "ingest this", "add this source to my wiki", "put
  this article/lesson in the vault as a wiki page". Works from any repo. Routing rule: someone
  else's words go to wiki/ (this skill); the user's own learnings and decisions go to notes/
  (/to-vault).
argument-hint: "<file path or URL>"
---

Ingest one source into the vault's `wiki/`, then reply with the paths written.

## Inputs

- **Vault**: `~/Documents/personal/my-obsidian-notes`. If it or its `CLAUDE.md` is missing, stop
  and say so; never create a vault.
- **Schema**: read the vault's `CLAUDE.md` every run, at least § Schema, § Ownership, § MOCs,
  § Wiki, § Index format, § Log format and § Ingest. § Ingest is the procedure; follow it over
  anything below if they differ.
- **Source**: exactly one file path or URL from the args. Given several, ingest the first and list
  the rest for later runs. Given none, ask for one and write nothing.
- **Routing**: if the source is the user's own words (their typed notes, a Notion page of their
  notes, a takeaway from this conversation), stop and point them to `/to-vault`. Material written
  by someone else, including a lesson Claude wrote from cited sources, is a source.

## Steps

1. **Load.** Read the source: a local file with Read, a URL with WebFetch. If it cannot be read,
   stop and say so, writing nothing.
2. **Already in?** Grep `log.md` for `ingest | <title>` and list `raw/` for the same title. On a
   hit, say so and ask whether to re-ingest; write nothing until the user answers.
3. **Discuss.** Do § Ingest steps 1–2: read `index.md` and the summary grep, tell the user the key
   takeaways and which wiki pages they would create or touch, then stop. Write no file, `raw/`
   included, until the user replies.
4. **Save the original** to `raw/<Title> (YYYY-MM-DD).md`, never overwriting an existing file:
   - Markdown or plain text: copy it unchanged.
   - HTML or a web page: convert to markdown, keeping every sentence, number, table and quote
     word for word; drop only markup, scripts and navigation. Add `Source: <url or path>` as the
     first line.
   - PDF or image: copy the file to `raw/assets/` and write the `.md` as a transcription that
     links it.
5. **Ingest.** Do § Ingest steps 3–6 against that `raw/` file: source page, concept/component/
   pattern pages, `index.md` and MOC links, `log.md`. Respect § Ownership: under `notes/` add only
   the links § Ingest allows, and never edit a `raw/` file after step 4.

## Reply

One line per file written or changed, as a full path, e.g.
`~/Documents/personal/my-obsidian-notes/wiki/Capacity estimation.md`, then any `conflict` entries
logged.
