---
name: file-issue
description: >
  Files ONE ad-hoc GitHub issue for this repo and prints its URL. Use when the
  user says "file an issue", "open a ticket for this", "make an issue so we can /implement it
  later", "log this as a ticket", "park this in the tracker", or when a conversation surfaces a
  concrete follow-up (a bug spotted mid-task, a refactor out of scope, a TODO worth tracking) and
  the user wants it captured rather than done now. Also the instant capture for small fixes:
  "inbox: …", "inbox this", "add to the inbox" files a bare `inbox` issue for /sweep to batch
  later. The singular sibling of /to-tickets: no plan, no slicing, no feature label. For a bug
  that needs root-cause investigation first, use triage-issue; for breaking a whole feature into
  tickets, use /to-tickets.
---

# File issue

Two modes. **Inbox** when the user says "inbox" (or "add to the inbox"): skip to
[Inbox mode](#inbox-mode). Otherwise the full ticket below.

One ticket, filed fast, picked up later by `/implement`. Speed matters more than polish: the user
checks the result via the URL, so draft from what you already know and file without a confirm step.

## 1. Draft

Take the content from the user's ask and the conversation. Only explore the code if you can't yet
name where the work lives, and then just enough to name it: a couple of greps, not a review.

Title: imperative, names the outcome (`Retry stuck scrape runs after timeout`).

Body, exactly these sections, kept short:

```markdown
## Context

<1–3 sentences: what's wrong or wanted, and why. Link the PR/issue/conversation it came from if one exists.>

## Where to look

- `path/to/file.go` — `FuncName`: <what's relevant here>
- `path/to/other.ts` — <what's relevant here>

## Acceptance criteria

- [ ] <observable outcome, checkable without reading the diff>
- [ ] <…>
```

"Where to look" is the point of the ticket: it saves the implementing agent its exploration. Name
files and symbols, not directories. Acceptance criteria are outcomes, usually 1–4; a ticket
needing more is probably a feature for /to-tickets.

## 2. File and report

```bash
gh issue create --title "<title>" --body-file <tmpfile> --label "status:ready"
```

If `status:ready` doesn't exist, create it (`gh label create status:ready -c "#0e8a16"`) and retry.
No feature label: ad-hoc tickets stay out of `/fleet dispatch` and are run with
`/implement <number>`.

Finish with the URL on its own line, then `/implement <number>` as the command that picks it
up. Nothing else.

## Inbox mode

Capture only, for /sweep to explore, cluster and turn into tickets later. No exploration, no
acceptance criteria, no `status:*` label: an inbox issue is not ready to implement.

```bash
gh issue create --title "<title>" --body "<one line: what you noticed and where, if known>" --label inbox
```

If `inbox` doesn't exist, create it (`gh label create inbox -c "#c5def5"`) and retry. Several
items in one ask become one issue each. Finish with the URLs, one per line, then `/sweep` as the
command that batches them.
