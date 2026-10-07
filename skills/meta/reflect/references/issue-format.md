# Filing an issue

One issue per approved item, in the repo that owns the fix ([routing.md](routing.md)).

## Repo

- Local: run `gh` from the current repo.
- Library: take the `library:` path from `~/.claude/bin/reflect status`, get its remote with
  `git -C <library> remote get-url origin`, and pass `-R <owner>/<repo>` to every `gh` call.

Make sure the `reflect` and `status:ready` labels exist in that repo, creating whichever is
missing (`gh label create reflect -c "#5319e7"`, `gh label create status:ready -c "#0e8a16"`).

## Dedupe

List open reflect issues and look for any of the item's fingerprints in their bodies:

```bash
gh issue list [-R <owner>/<repo>] --label reflect --state open --limit 200 --json number,url,body
```

If one carries a matching `reflect-fp:` line, comment on it instead of filing:
`Seen again in session <sid8>: <n>x <category>. <one-line evidence>`.

## Body

Title: imperative, naming the outcome (`Point the explore agent at docs/CODEMAPS before grepping`).

```markdown
## Context

<1–3 sentences: what went wrong, how often, what it cost. Name the session id.>

## Where to look

- `path/to/file` — <what to change there>

## Acceptance criteria

- [ ] <observable outcome>

## Evidence

- <agent> L<n> (`<tool_use_id>`): <one line>

reflect-fp: <fp> <fp> …
from: <originating repo name>
```

The `from:` line appears only on library issues. Include a proposed diff under "Where to look"
only when it is a few lines and you have read the target file.

```bash
gh issue create [-R <owner>/<repo>] --title "<title>" --body-file <tmpfile> --label reflect --label status:ready
```
