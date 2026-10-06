#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/rtk-hook.sh"
fails=0

bash_call() { jq -nc --arg c "$1" --arg d "$2" '{tool_name:"Bash",cwd:$d,tool_input:{command:$c}}'; }

expect() {
	local want="$1" name="$2" payload="$3"
	local out
	out=$("$HOOK" <<<"$payload" 2>/dev/null)
	case "$want" in
	plain) [ -z "$out" ] && return ;;
	rewritten) [[ "$out" == *"rtk git status"* ]] && return ;;
	esac
	echo "FAIL: $name (want $want, got: ${out:-<empty>})"
	fails=$((fails + 1))
}

WT=/r/repo/.claude/worktrees/agent-1
expect plain "git in worktree" "$(bash_call 'git status' "$WT")"
expect plain "chained git in worktree" "$(bash_call 'cd sub && git status' "$WT")"
expect plain "subshell git in worktree" "$(bash_call '(git status)' "$WT")"
expect rewritten "git in main checkout" "$(bash_call 'git status' /r/repo)"
expect rewritten "git in tp worktree" "$(bash_call 'git status' /r/repo-worktrees/cc-1)"

[ "$fails" -eq 0 ] && echo "ok" || exit 1
