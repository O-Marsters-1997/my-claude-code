#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
command -v claude >/dev/null || exit 0
[ "${CC_PONYTAIL_REVIEW_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
[ "$(jq -r '.stop_hook_active // false' <<<"$INPUT")" = true ] && exit 0
TRANSCRIPT=$(jq -r '.transcript_path // empty' <<<"$INPUT")
[ -f "$TRANSCRIPT" ] || exit 0
CWD=$(jq -r '.cwd // empty' <<<"$INPUT")
TOP=$(git -C "${CWD:-.}" rev-parse --show-toplevel 2>/dev/null) || exit 0
git -C "$TOP" rev-parse --verify -q HEAD >/dev/null || exit 0
TOP=$(cd "$TOP" && pwd -P)

EDITED=$(jq -rs '
	(map(.type == "user" and .isMeta != true and ((.message.content | type) == "string" or any(.message.content[]?; .type == "text"))) | rindex(true)) as $turn
	| .[($turn // -1) + 1:][]
	| select(.type == "assistant") | .message.content[]?
	| select(.type == "tool_use" and (.name | IN("Edit", "Write", "MultiEdit")))
	| .input.file_path // empty' "$TRANSCRIPT" 2>/dev/null | grep -E '\.(go|ts|tsx|js|jsx|py|rs|sql|sh)$' | sort -u)
[ -z "$EDITED" ] && exit 0

DIFF=""
while IFS= read -r f; do
	[ -f "$f" ] || continue
	f="$(cd "$(dirname "$f")" && pwd -P)/$(basename "$f")"
	case "$f" in "$TOP"/*) ;; *) continue ;; esac
	rel=${f#"$TOP"/}
	if git -C "$TOP" ls-files --error-unmatch -- "$rel" >/dev/null 2>&1; then
		DIFF+=$(git -C "$TOP" diff HEAD -- "$rel")$'\n'
	else
		DIFF+=$(git -C "$TOP" diff --no-index -- /dev/null "$rel")$'\n'
	fi
done <<<"$EDITED"
[ -z "${DIFF//[[:space:]]/}" ] && exit 0

PROMPT='You are the ponytail reviewer. Stdin holds the diff of the code files edited this turn. Read the code it touches as needed. Judge only lines this diff adds or changes, against these rules:
- delete: dead code, unused options, speculative features or scaffolding for later
- reuse: a helper that already exists in this repo (name its path)
- stdlib/native: the standard library, platform or an installed dependency already does it; a new dependency for a few lines
- yagni: an interface or abstraction with one implementation, config nobody sets, a factory for one product
- merge: near-copies that must change together
Every finding needs a concrete case and the smallest fix, preferring a fix that deletes code. Skip style, anything a linter catches, and code the diff did not change. A `ponytail:` or `shortcut:` comment naming its limit is a decision, not a finding.
If there are no findings, print exactly OK. Otherwise print only the findings, one per line, as: <file:line> - <finding> - <smallest fix>'

OUT=$(cd "$TOP" && claude -p "$PROMPT" --model sonnet --tools Read,Grep,Glob --settings '{"disableAllHooks":true}' <<<"$DIFF" 2>/dev/null) || exit 0
OUT=$(sed -e 's/[[:space:]]*$//' <<<"$OUT")
[ -z "$OUT" ] || [ "$OUT" = OK ] && exit 0
printf 'Ponytail review:\n%s\nApply the fixes, or say why a finding is wrong. Set CC_PONYTAIL_REVIEW_OFF=1 to skip this check.\n' "$OUT" >&2
exit 2
