# Capability inventory

The inventory is the substrate for every idea. A row the code doesn't back is a defect. So is an
idea that re-proposes something the code already does. Every status is checkable by someone
re-running your search.

## 1. Pin and read intent

```bash
git rev-parse --short HEAD          # pin: the inventory describes this commit
git log --oneline -40               # where work is flowing
```

Read the domain glossary and decisions first: `CONTEXT.md` (repo root), `docs/adr/`,
`docs/agents/domain.md`, README, AGENTS.md/CLAUDE.md. Use their nouns for capability names
throughout. Note every promise they make, because docs that disagree with code are signals.

On a rerun with a stored inventory, run `git diff --stat <pinned>..HEAD`. Re-check only the rows
whose files changed, plus any new surface. Carry the rest forward unchanged.

## 2. List surfaces mechanically

Enumerate each surface class from the code, not from memory. Adapt the search to the stack. For a
large repo, fan out 2–3 `Explore` subagents in parallel (backend surfaces / frontend and email /
data and jobs), each returning raw rows with `file:line`.

| Surface | What to find |
|---|---|
| HTTP routes | router registrations, handler maps |
| Workers / jobs / schedules | queue consumers, cron, `schedule`/ticker loops, and which are commented out |
| UI routes | route files and pages, nav entries |
| Emails / notifications | templates, senders, triggers |
| CLI | `cmd/`, `bin/`, scripts, justfile/Makefile recipes |
| Tables | schema or migrations: tables, notable columns, enums |

## 3. Group into capabilities

Group the surfaces into 8–20 capabilities, each a thing a user can get done, named in the repo's
domain terms. "Feed refresh" is a capability. "worker/sources/" is a directory. Each row lists the
surfaces it spans. A capability with no UI, no email and no CLI is invisible to the user, so say so.

## 4. Mark status, with evidence

| Status | Meaning | Minimum evidence |
|---|---|---|
| COMPLETE | Works end to end and the result reaches the user | entry `file:line` and the surface that shows the result |
| PARTIAL | Exists but missing its second half: stored but not shown, one lifecycle step absent, manual where the rest is automatic | what exists (`file:line`) and the missing half (with the search) |
| ORPHAN | Built but unreachable or switched off: no consumer, disabled loop, UI wired to mocks | `file:line` of the built part, and the search showing no caller or the disabled line |
| GAP | A domain word or job step with nothing behind it | the search that came back empty |

**Every "missing", "no consumer" or GAP claim shows the search that proved it**, for example
`rg -n "posted_at" frontend/ internal/ → 0 hits`. If you didn't search, say `unverified`. Missing
evidence is not a flaw: unknown is not the same as broken.

## 5. Signals to check

Run what's cheap. Each hit attaches to a capability row.

- **Stored, never surfaced:** columns or enum values written by the backend but never read by the
  UI, API DTOs or emails (grep each notable column against the frontend and response types).
- **Endpoints with no consumer:** routes that nothing in the frontend, CLI or other services calls.
- **UI wired to mocks:** fixtures, `mock`, hard-coded arrays, or feature flags stuck at one value.
- **Disabled loops:** commented-out workers or schedules, `if false`, and env flags off by default.
- **TODO/FIXME with age:** `rg -n "TODO|FIXME|HACK" --glob '!*_test*'`, then `git blame -L` on a
  few. Old TODOs clustered in one capability mark it as half-finished.
- **Churn × size:** `git log --since=6.months --format= --name-only | sort | uniq -c | sort -rn | head -20`
  crossed with file length. High churn on a shallow surface is a deepening target.
- **Docs vs code:** an ADR, glossary entry or README claim with no enforcement point, or code that
  contradicts it.
- **Domain words with no first-class representation:** a term the glossary or UI copy uses
  repeatedly that has no table, type or route.

## 6. Overlay the job map

Place each capability on Ulwick's 8 steps of the user's core job: Define → Locate → Prepare →
Confirm → Execute → Monitor → Modify → Conclude. Phrase each step for this product (for a feed
reader, "Locate" means finding sources and "Conclude" means archiving or sharing what was read). Thin
steps are deepen candidates. **Empty steps are legitimate net-new slots**, recorded as GAP rows.

## Output

```markdown
_Inventory pinned at `abc1234` (2026-10-01)._

| # | Capability | Surfaces | Status | Evidence | Job step |
|---|---|---|---|---|---|
| C1 | Feed refresh | worker, `feed_state` | ORPHAN | loop commented out `cmd/worker/main.go:88` | Locate |
| C2 | Read history | `reads.opened_at` | PARTIAL | stored `schema.sql:140`; `rg opened_at web/src` → 0 hits | Monitor |
| C9 | Sharing | — | GAP | `rg -i "share" internal/ web/src` → 0 hits | Conclude |

**Signals:** 3–8 bullets, each tied to a row and its evidence.
**Job map:** step → rows (empty steps flagged).
```

Keep the table itself in the report body. Move raw search output to an appendix only if it is
long.
