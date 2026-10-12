#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_SQL_SKILLS_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
FILE=$(jq -r '.tool_input.file_path // empty' <<<"$INPUT")
case "$FILE" in
*/vendor/*) exit 0 ;;
*.sql) ;;
*) exit 0 ;;
esac

GENERATED='^-- (Code generated|Generated|@generated).*(DO NOT EDIT|generated)'
if [ -f "$FILE" ]; then
	head -n 20 "$FILE" | grep -qiE "$GENERATED" && exit 0
elif jq -r '.tool_input.content // empty' <<<"$INPUT" | head -n 20 | grep -qiE "$GENERATED"; then
	exit 0
fi

TRANSCRIPT=$(jq -r '.agent_transcript_path // empty' <<<"$INPUT")
if [ -z "$TRANSCRIPT" ]; then
	TRANSCRIPT=$(jq -r '.transcript_path // empty' <<<"$INPUT")
	AGENT_ID=$(jq -r '.agent_id // empty' <<<"$INPUT")
	[ -n "$AGENT_ID" ] && TRANSCRIPT="${TRANSCRIPT%.jsonl}/subagents/agent-$AGENT_ID.jsonl"
fi
[ -r "$TRANSCRIPT" ] || exit 0

SKILL=postgres-best-practices
if grep -E -e "\"skill\":\"$SKILL\"" -e "<command-name>/?$SKILL</command-name>" "$TRANSCRIPT" |
	jq -e --arg s "$SKILL" '
		[
			(.message.content | arrays | any(.[]; .type == "tool_use" and .name == "Skill" and .input.skill == $s)),
			(.message.content | strings | (contains("<command-name>/" + $s + "</command-name>") or contains("<command-name>" + $s + "</command-name>")))
		] | any
	' 2>/dev/null | grep -q true; then
	exit 0
fi

printf 'Load %s with the Skill tool before editing %s, then retry this edit. Set CC_SQL_SKILLS_OFF=1 to skip this check.\n' "$SKILL" "$FILE" >&2
exit 2
