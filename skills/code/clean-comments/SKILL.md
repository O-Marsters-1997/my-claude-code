---
name: clean-comments
model: haiku
description: Remove self-documenting comments from source files, leaving only comments that explain why, document non-obvious behaviour, or record external constraints. Use when user wants to clean comments, remove redundant comments, strip obvious comments, or run /clean-comments on staged files before committing.
---

# Clean comments

## The standard

`~/.claude/rules/comments.md` decides what earns a comment. It is the only
copy. Read it before editing a file and apply it as written.

Do not restate it here, and do not work from memory of it. A subagent running
this skill may not have it loaded, so read the file. If a case it does not
cover comes up, the fix belongs in `comments.md`, not in this skill.

Deleting a directive is the one unrecoverable mistake this skill can make.
`comments.md` permits them outright: `//go:build`, `// +build`, `//go:generate`,
`//nolint`, `// @ts-expect-error`, `// eslint-disable-*`, `# type: ignore`,
`# noqa`, shebangs. They are not comments to weigh. Leave them.

## Quick start

Run on staged files before a commit:
```
/clean-comments
```

Run on specific files:
```
/clean-comments src/foo.ts src/bar.go
```

## Workflow

1. Read `~/.claude/rules/comments.md`
2. If arguments are provided, treat them as file paths to clean
3. If no arguments, get staged files: `git diff --cached --name-only --diff-filter=ACM`
4. Filter to code files (`.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.rs`)
5. For each file, read it, remove the comments that fail the standard, write it back
6. Re-stage any modified files: `git add <file>`
7. Report what was removed and from which files; if nothing needed removing, say so

Comment-only edits. Never change code in the same pass.

## The commit hook

`~/.claude/hooks/check-staged-comments.sh` runs on `PreToolUse` for Bash and
blocks `git commit` when the staged diff adds comment lines. It ignores
directives, and reads `#` as a comment marker in Python only.

Trimming a long doc comment counts as adding lines, so a cleanup pass trips it.
When every remaining comment earns its place, say so and commit with:

```
CC_COMMENTS_REVIEWED=1 git commit ...
```
