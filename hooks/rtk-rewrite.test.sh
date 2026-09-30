#!/usr/bin/env bash
# Run: hooks/rtk-rewrite.test.sh
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/rtk-rewrite.sh"
fails=0

payload() { jq -nc --arg c "$1" --arg d "$2" '{tool_name: "Bash", tool_input: {command: $c}, cwd: $d}'; }

check() {
	local name=$1 want=$2 cmd=$3 cwd=$4 out
	out=$(payload "$cmd" "$cwd" | "$HOOK" 2>/dev/null)
	if [[ $want == rewritten && -n $out ]] || [[ $want == plain && -z $out ]]; then
		echo "ok   $name"
	else
		echo "FAIL $name: want $want, got: $out"
		fails=$((fails + 1))
	fi
}

WT=/repo/.claude/worktrees/agent-x
check "git in main checkout is rewritten" rewritten "git status" /repo
check "git in worktree stays plain" plain "git status" "$WT"
check "git after cd in worktree stays plain" plain "cd sub && git diff" "$WT"
check "non-git in worktree is still rewritten" rewritten "ls -la" "$WT"
check "git push stays plain everywhere" plain "git push -u origin b" /repo

exit $((fails > 0))
