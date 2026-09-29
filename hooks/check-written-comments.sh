#!/usr/bin/env bash
# PreToolUse gate for Edit, MultiEdit and Write: rejects an edit that adds
# comment lines, so ~/.claude/rules/comments.md holds at write time.
#
# Line-leading comments only. Matching trailing comments would flag every URL
# or `//` inside a string literal or regex, so `x = 1 // noise` passes by
# design. Comments carrying a link are skipped: comments.md permits an outside
# constraint and asks for the issue or RFC to be linked.
INPUT=$(cat)
FILE=$(jq -r '.tool_input.file_path // empty' <<<"$INPUT")
[ -n "$FILE" ] || exit 0

case "$FILE" in
*/node_modules/* | */vendor/* | */.git/* | */sqlc/* | */snapshots/* | *.gen.* | *.min.* | *.pb.go | *_gen.go) exit 0 ;;
esac

NAME=$(basename "$FILE")
case "$NAME" in
*.go | *.ts | *.tsx | *.js | *.jsx | *.mjs | *.cjs | *.rs | *.java | *.kt | *.swift | *.c | *.h | *.cc | *.cpp | *.hpp | *.cs | *.scala | *.dart)
	PATTERN='^[[:space:]]*(//|/\*|\*([[:space:]]|/|$)|\{/\*)'
	;;
*.css | *.scss | *.less)
	PATTERN='^[[:space:]]*(/\*|\*([[:space:]]|/|$)|//)'
	;;
*.py | *.rb | *.sh | *.bash | *.zsh | *.yaml | *.yml | *.toml | *.tf | Dockerfile | Dockerfile.* | Makefile | justfile | Justfile | *.just)
	PATTERN='^[[:space:]]*#'
	;;
*.sql | *.lua)
	PATTERN='^[[:space:]]*--'
	;;
*.html | *.vue | *.svelte | *.xml)
	PATTERN='^[[:space:]]*(<!--|//)'
	;;
*) exit 0 ;;
esac

# Directives are skipped when deleting the line changes what the compiler,
# formatter, build or codegen does, matching comments.md. Suppressions
# (@ts-ignore, eslint-disable, # type: ignore, //nolint) are debt and stay in
# scope. `-- name:` is a sqlc query directive.
DIRECTIVES='^[[:space:]]*((//|#|--)[[:space:]]*(go:|\+build|prettier-ignore|fmt:|-\*-|syntax=|escape=|name:)|/\*[[:space:]]*prettier-ignore|<!--[[:space:]]*prettier-ignore|#!|//[[:space:]]*@vitest-environment|//[[:space:]]*@jsx)'
NOTICES='(Copyright|SPDX-License-Identifier|[Ll]icen[sc]e|https?://)'

case $(jq -r '.tool_name' <<<"$INPUT") in
Edit)
	NEW=$(jq -r '.tool_input.new_string // empty' <<<"$INPUT")
	OLD=$(jq -r '.tool_input.old_string // empty' <<<"$INPUT")
	;;
MultiEdit)
	NEW=$(jq -r '.tool_input.edits[]?.new_string' <<<"$INPUT")
	OLD=$(jq -r '.tool_input.edits[]?.old_string' <<<"$INPUT")
	;;
Write)
	NEW=$(jq -r '.tool_input.content // empty' <<<"$INPUT")
	DIR=$(dirname "$FILE")
	REL=$(git -C "$DIR" ls-files --full-name -- "$FILE" 2>/dev/null)
	OLD=$([ -n "$REL" ] && git -C "$DIR" show "HEAD:$REL" 2>/dev/null)
	;;
*) exit 0 ;;
esac

if head -n 5 <<<"$NEW" | grep -qE 'Code generated|DO NOT EDIT|@generated|auto-generated'; then
	exit 0
fi

ADDED=$(grep -vxFf <(printf '%s\n' "$OLD" | grep -v '^[[:space:]]*$') <<<"$NEW" |
	grep -E "$PATTERN" | grep -vE "$DIRECTIVES" | grep -vE "$NOTICES")
[ -z "$ADDED" ] && exit 0

printf 'This edit added comment lines to %s:\n%s\n\nApply ~/.claude/rules/comments.md now: delete each one unless it is a doc comment on an exported identifier, a load-bearing directive, or names a constraint from outside the repo (link the issue or RFC). If the code needs the comment to be understood, rename or refactor the code instead.\n' "$FILE" "$ADDED" >&2
exit 2
