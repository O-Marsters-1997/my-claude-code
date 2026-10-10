#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_CODEGRAPH_GATE_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
TOOL=$(jq -r '.tool_name // empty' <<<"$INPUT")
CWD=$(jq -r '.cwd // empty' <<<"$INPUT")

reads_ref() {
	local words tok have_pattern=0
	read -ra words <<<"$(sed -E 's/^.*git[[:space:]]+grep//; s/[;&|].*//' <<<"$1")"
	for tok in "${words[@]}"; do
		tok=${tok//[\'\"]/}
		[ "$tok" = -- ] && return 1
		[ "$tok" = -e ] && have_pattern=1
		[[ $tok == -* ]] && continue
		if [ "$have_pattern" = 0 ]; then
			have_pattern=1
			continue
		fi
		git -C "${CWD:-.}" rev-parse --verify -q "$tok^{commit}" >/dev/null 2>&1 && return 0
	done
	return 1
}
NON_CODE='\.(md|mdx|txt|tmpl|html|css|json|jsonl|toml|ya?ml|sql|lock|log|env|sh)\b|(^|[/ "'\''])(docs|ideas|\.claude|\.codegraph|node_modules|testdata)(/|\b)'

case "$TOOL" in
Grep) TARGET=$(jq -r '[.tool_input.path, .tool_input.glob, .tool_input.type] | map(select(. != null)) | join(" ")' <<<"$INPUT") ;;
Glob) TARGET=$(jq -r '[.tool_input.path, .tool_input.pattern] | map(select(. != null)) | join(" ")' <<<"$INPUT") ;;
Bash)
	TARGET=$(jq -r '.tool_input.command // empty' <<<"$INPUT")
	grep -qE '(^|&&|;|\|\||\$\()[[:space:]]*(rtk[[:space:]]+)?(rg|grep|egrep|ag|find|git[[:space:]]+grep)([[:space:]]|$)' <<<"$TARGET" || exit 0
	grep -qE 'git[[:space:]]+grep' <<<"$TARGET" && reads_ref "$TARGET" && exit 0
	;;
*) exit 0 ;;
esac
grep -qE "$NON_CODE" <<<"$TARGET" && exit 0

DIR=$(jq -r '.tool_input.path // empty' <<<"$INPUT")
[ -d "$DIR" ] || DIR=$(dirname "${DIR:-.}")
[ -d "$DIR" ] && [ "$DIR" != . ] || DIR=$CWD
ROOT=$(git -C "$DIR" rev-parse --show-toplevel 2>/dev/null) || exit 0
[ -f "$ROOT/.codegraph/codegraph.db" ] || exit 0
[ "$(git -C "$ROOT" rev-parse --absolute-git-dir)" = "$(git -C "$ROOT" rev-parse --path-format=absolute --git-common-dir)" ] || exit 0

TRANSCRIPT=$(jq -r '.agent_transcript_path // empty' <<<"$INPUT")
if [ -z "$TRANSCRIPT" ]; then
	TRANSCRIPT=$(jq -r '.transcript_path // empty' <<<"$INPUT")
	AGENT_ID=$(jq -r '.agent_id // empty' <<<"$INPUT")
	[ -n "$AGENT_ID" ] && TRANSCRIPT="${TRANSCRIPT%.jsonl}/subagents/agent-$AGENT_ID.jsonl"
fi
[ -r "$TRANSCRIPT" ] || exit 0

queried() {
	local ids
	ids=$(grep -E -e '"name":"mcp__codegraph__' -e 'codegraph (explore|node|query|callers|callees|impact|context)' "$TRANSCRIPT" |
		jq -r '
			.message.content | arrays | .[] |
			select(.type == "tool_use" and (
				(.name | startswith("mcp__codegraph__")) or
				(.name == "Bash" and (.input.command | test("\\bcodegraph (explore|node|query|callers|callees|impact|context)\\b")))
			)) | .id
		' 2>/dev/null)
	[ -n "$ids" ] && grep -F "$ids" "$TRANSCRIPT" |
		jq -e --arg ids "$ids" '
			($ids | split("\n")) as $ids |
			.message.content | arrays | any(.[];
				.type == "tool_result" and .is_error != true and (.tool_use_id as $id | $ids | index($id) != null))
		' 2>/dev/null | grep -q true
}

# Claude Code appends to the transcript a few hundred ms after the event it records.
for _ in 1 2 3 4 5; do
	queried && exit 0
	sleep 0.25
done

printf 'This repo has a CodeGraph index. Query it once before searching code: call mcp__codegraph__codegraph_explore (load it via ToolSearch if deferred), or run `codegraph explore "<symbols or question>"` in Bash if the MCP tool is not available. After one query, Grep/Glob/grep are unlocked for the rest of this agent. Searches of docs, templates, CSS and config files are never gated. CC_CODEGRAPH_GATE_OFF=1 skips this check only when set in the environment before Claude Code launches; setting it inside the command has no effect.\n' >&2
exit 2
