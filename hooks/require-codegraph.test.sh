#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/require-codegraph.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
fails=0

mkdir -p "$TMP/indexed/.codegraph" "$TMP/plain" "$TMP/t/none/subagents" "$TMP/t/mcp/subagents"
touch "$TMP/indexed/.codegraph/codegraph.db"
for d in indexed plain; do git -C "$TMP/$d" init -q; done

result() { printf '{"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"%s","is_error":%s}]}}\n' "$1" "$2"; }
mcp_call() { printf '{"message":{"role":"assistant","content":[{"type":"tool_use","id":"m1","name":"mcp__codegraph__codegraph_explore","input":{"query":"x"}}]}}\n'; result m1 false; }
shell_call() { printf '{"message":{"role":"assistant","content":[{"type":"tool_use","id":"b1","name":"Bash","input":{"command":"codegraph explore \\"Loop Run\\""}}]}}\n'; result b1 false; }
denied_call() { printf '{"message":{"role":"assistant","content":[{"type":"tool_use","id":"d1","name":"Bash","input":{"command":"codegraph explore \\"Loop Run\\" | head"}}]}}\n'; result d1 true; }
mention() { printf '{"message":{"role":"user","content":"use codegraph explore please, mcp__codegraph__ too"}}\n'; }

T_NONE="$TMP/t/none.jsonl"
T_MENTION="$TMP/t/mention.jsonl"
T_MCP="$TMP/t/mcp.jsonl"
T_SHELL="$TMP/t/shell.jsonl"
T_DENIED="$TMP/t/denied.jsonl"
: >"$T_NONE"
mention >"$T_MENTION"
mcp_call >"$T_MCP"
shell_call >"$T_SHELL"
denied_call >"$T_DENIED"
mcp_call >"$TMP/t/none/subagents/agent-a1.jsonl"
: >"$TMP/t/mcp/subagents/agent-a2.jsonl"

payload() {
	jq -nc --arg tool "$1" --argjson in "$2" --arg cwd "$3" --arg t "$4" --arg a "${5:-}" \
		'{tool_name: $tool, tool_input: $in, cwd: $cwd, transcript_path: $t, session_id: "s"}
		+ (if $a == "" then {} else {agent_id: $a} end)'
}

check() {
	local name=$1 want=$2 input=$3
	shift 3
	local err code
	err=$(env "$@" "$HOOK" <<<"$input" 2>&1 >/dev/null)
	code=$?
	if [ "$code" != "$want" ]; then
		printf 'FAIL %s: exit %s, want %s\n%s\n' "$name" "$code" "$want" "$err"
		fails=$((fails + 1))
	fi
}

I="$TMP/indexed"
check "grep blocked before query" 2 "$(payload Grep '{"pattern":"func Run"}' "$I" "$T_NONE")"
check "glob blocked before query" 2 "$(payload Glob '{"pattern":"**/*.go"}' "$I" "$T_NONE")"
check "bash rg blocked" 2 "$(payload Bash '{"command":"rg -n \"func Run\" internal"}' "$I" "$T_NONE")"
check "bash grep after cd blocked" 2 "$(payload Bash '{"command":"cd internal && grep -rn Spawn ."}' "$I" "$T_NONE")"
check "mention is not a query" 2 "$(payload Grep '{"pattern":"Spawn"}' "$I" "$T_MENTION")"
check "grep allowed after mcp query" 0 "$(payload Grep '{"pattern":"Spawn"}' "$I" "$T_MCP")"
check "grep allowed after shell query" 0 "$(payload Grep '{"pattern":"Spawn"}' "$I" "$T_SHELL")"
check "denied query does not unlock" 2 "$(payload Grep '{"pattern":"Spawn"}' "$I" "$T_DENIED")"
T_LATE="$TMP/t/late.jsonl"
: >"$T_LATE"
(sleep 0.3 && mcp_call >>"$T_LATE") &
check "query flushed after the hook starts unlocks" 0 "$(payload Grep '{"pattern":"Spawn"}' "$I" "$T_LATE")"
check "piped grep passes" 0 "$(payload Bash '{"command":"go test ./... | grep FAIL"}' "$I" "$T_NONE")"
check "other bash passes" 0 "$(payload Bash '{"command":"go build ./..."}' "$I" "$T_NONE")"
check "markdown grep passes" 0 "$(payload Grep '{"pattern":"Repo","glob":"*.md"}' "$I" "$T_NONE")"
check "template grep passes" 0 "$(payload Grep '{"pattern":"hx-get","path":"internal/cc/board.tmpl"}' "$I" "$T_NONE")"
check "docs dir grep passes" 0 "$(payload Grep '{"pattern":"Track","path":"docs"}' "$I" "$T_NONE")"
check "bash grep on css passes" 0 "$(payload Bash '{"command":"grep -n pill web/app.css"}' "$I" "$T_NONE")"
check "unindexed repo passes" 0 "$(payload Grep '{"pattern":"x"}' "$TMP/plain" "$T_NONE")"
check "non-repo passes" 0 "$(payload Grep '{"pattern":"x"}' "$TMP" "$T_NONE")"
check "missing transcript fails open" 0 "$(payload Grep '{"pattern":"x"}' "$I" "$TMP/t/absent.jsonl")"
check "off switch" 0 "$(payload Grep '{"pattern":"x"}' "$I" "$T_NONE")" CC_CODEGRAPH_GATE_OFF=1
check "read never gated" 0 "$(payload Read '{"file_path":"x.go"}' "$I" "$T_NONE")"
check "subagent with own query passes" 0 "$(payload Grep '{"pattern":"x"}' "$I" "$T_NONE" a1)"
check "subagent without own query is blocked" 2 "$(payload Grep '{"pattern":"x"}' "$I" "$T_MCP" a2)"

[ "$fails" = 0 ] && echo ok || exit 1
