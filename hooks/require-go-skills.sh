#!/usr/bin/env bash
# PreToolUse gate: block an edit to a Go file until go-idiomatic has been
# loaded this session (and testing-policy, for a _test.go in a project that
# ships that skill). Fails open whenever it cannot tell.
command -v jq >/dev/null || exit 0
[ "${CC_GO_SKILLS_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
FILE=$(jq -r '.tool_input.file_path // empty' <<<"$INPUT")
case "$FILE" in
*.go) ;;
*) exit 0 ;;
esac

case "$FILE" in
*/vendor/*) exit 0 ;;
esac

DIR=$(dirname "$FILE")
ROOT=$DIR
while [ "$ROOT" != / ] && [ ! -f "$ROOT/go.mod" ]; do ROOT=$(dirname "$ROOT"); done
[ -f "$ROOT/go.mod" ] || exit 0

GENERATED='^// Code generated .* DO NOT EDIT\.$'
if [ -f "$FILE" ]; then
	head -n 20 "$FILE" | grep -qE "$GENERATED" && exit 0
elif jq -r '.tool_input.content // empty' <<<"$INPUT" | head -n 20 | grep -qE "$GENERATED"; then
	exit 0
fi

TRANSCRIPT=$(jq -r '.agent_transcript_path // empty' <<<"$INPUT")
if [ -z "$TRANSCRIPT" ]; then
	TRANSCRIPT=$(jq -r '.transcript_path // empty' <<<"$INPUT")
	AGENT_ID=$(jq -r '.agent_id // empty' <<<"$INPUT")
	[ -n "$AGENT_ID" ] && TRANSCRIPT="${TRANSCRIPT%.jsonl}/subagents/agent-$AGENT_ID.jsonl"
fi
[ -r "$TRANSCRIPT" ] || exit 0

loaded() {
	grep -E -e "\"skill\":\"$1\"" -e "<command-name>/?$1</command-name>" "$TRANSCRIPT" |
		jq -e --arg s "$1" '
			[
				(.message.content | arrays | any(.[]; .type == "tool_use" and .name == "Skill" and .input.skill == $s)),
				(.message.content | strings | (contains("<command-name>/" + $s + "</command-name>") or contains("<command-name>" + $s + "</command-name>")))
			] | any
		' 2>/dev/null | grep -q true
}

REQUIRED=(go-idiomatic)
case "$FILE" in
*_test.go)
	PROJECT=$(git -C "$DIR" rev-parse --show-toplevel 2>/dev/null || echo "${CLAUDE_PROJECT_DIR:-$ROOT}")
	[ -f "$PROJECT/.claude/skills/testing-policy/SKILL.md" ] && REQUIRED+=(testing-policy)
	;;
esac

MISSING=()
for skill in "${REQUIRED[@]}"; do
	loaded "$skill" || MISSING+=("$skill")
done
[ ${#MISSING[@]} -eq 0 ] && exit 0

printf 'Load %s with the Skill tool before editing %s, then retry this edit. Set CC_GO_SKILLS_OFF=1 to skip this check.\n' "${MISSING[*]}" "$FILE" >&2
exit 2
