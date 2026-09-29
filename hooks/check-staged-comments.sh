#!/usr/bin/env bash
# Blocks `git commit` when the staged diff adds comment lines, prompting a
# clean-comments pass.
#
# The only way past is CC_COMMENTS_REVIEWED=1 on the commit itself. An earlier
# design let an unchanged retry through, which made the block a nudge the agent
# cleared by running the same command twice. The override keeps "I checked,
# they all earn their place" as a deliberate, greppable act instead.

INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")

[[ "$CMD" =~ (^|[^[:alnum:]_-])git[[:space:]]+(-C[[:space:]]+[^[:space:]]+[[:space:]]+)?commit([[:space:]]|$) ]] || exit 0
[[ "$CMD" =~ CC_COMMENTS_REVIEWED=1 ]] && exit 0

# The hook process's own cwd is not where the command runs: `git -C <path>` and
# a leading `cd <path> &&` both retarget it, and the session cwd is what git
# would see otherwise.
if [[ "$CMD" =~ git[[:space:]]+-C[[:space:]]+([^[:space:]]+) ]]; then
	DIR="${BASH_REMATCH[1]}"
elif [[ "$CMD" =~ ^[[:space:]]*cd[[:space:]]+([^[:space:]\&\;]+) ]]; then
	DIR="${BASH_REMATCH[1]}"
else
	DIR=$(jq -r '.cwd // empty' <<<"$INPUT")
fi
DIR="${DIR%\"}" && DIR="${DIR#\"}"
DIR="${DIR/#\~/$HOME}"
[ -d "$DIR" ] || DIR=.

# `#` opens a comment in Python only. Applied to the rest it reads Rust
# attributes (#[derive]) and JS private fields (#count) as comments.
SLASH_EXTS=('*.go' '*.ts' '*.tsx' '*.js' '*.jsx' '*.rs')
HASH_EXTS=('*.py')

STAGED=$(git -C "$DIR" diff --cached --name-only --diff-filter=ACM -- "${SLASH_EXTS[@]}" "${HASH_EXTS[@]}" 2>/dev/null)
[ -z "$STAGED" ] && exit 0

# Only load-bearing directives are skipped, matching comments.md: deleting one
# changes what the compiler, formatter or build does. Suppressions (@ts-ignore,
# eslint-disable, # type: ignore, //nolint) are debt that has to argue for
# itself there, so they stay in scope and get flagged like any other comment.
DIRECTIVES='^\+[[:space:]]*(//[[:space:]]*(go:|\+build|prettier-ignore)|/\*[[:space:]]*prettier-ignore|#!|#[[:space:]]*(fmt:|-\*-))'

# Line-leading only. Matching trailing comments would flag every staged URL in
# a string literal, so the gate misses `x = 1  # noise` by design.
ADDED=$( {
	grep -E '^\+[[:space:]]*(//|/\*)' <<<"$(git -C "$DIR" diff --cached --diff-filter=ACM -- "${SLASH_EXTS[@]}" 2>/dev/null)"
	grep -E '^\+[[:space:]]*#' <<<"$(git -C "$DIR" diff --cached --diff-filter=ACM -- "${HASH_EXTS[@]}" 2>/dev/null)"
} | grep -vE "$DIRECTIVES")
[ -z "$ADDED" ] && exit 0

printf 'Staged diff adds comment lines in:\n%s\n\nRun the clean-comments skill on these files, then commit again.\nThe standard is ~/.claude/rules/comments.md.\n\nIf every comment already earns its place, say so and commit with the override:\n  CC_COMMENTS_REVIEWED=1 git commit ...\n' "$STAGED" >&2
exit 2
