#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/allow-reflect-slice.sh"
fails=0

expect() {
	local want="$1" cmd="$2" got=0
	jq -nc --arg c "$cmd" '{tool_name:"Bash",tool_input:{command:$c}}' | "$HOOK" 2>/dev/null || got=$?
	[ "$got" -eq "$want" ] && return
	echo "FAIL: '$cmd' (want exit $want, got $got)"
	fails=$((fails + 1))
}

SID=e3354183-f7ca-47a7-becc-89f00d21e443
expect 0 "~/.claude/bin/reflect slice $SID main 14 -C 20"
expect 0 "~/.claude/bin/reflect slice $SID a00553fcbf8529abd 14"
expect 0 "\$HOME/.claude/bin/reflect slice $SID main 7 -C 5"
expect 0 "/Users/x/.claude/bin/reflect slice $SID main 7"
expect 0 "~/.claude/bin/reflect replay $SID a00553fcbf8529abd 27"
expect 0 "~/.claude/bin/reflect replay $SID a00553fcbf8529abd 27 --cwd /Users/x/wt-1"
expect 0 "~/.claude/bin/reflect replay $SID main 27 --cwd /tmp/x --command 'cd /x && grep y'"
expect 2 "~/.claude/bin/reflect replay $SID main 27 --cwd /tmp/x;id"
expect 2 "~/.claude/bin/reflect replay $SID main 27 --command 'x'; id"
expect 2 "~/.claude/bin/reflect replay $SID main 27 --cwd \$(whoami)"
expect 2 "~/.claude/bin/reflect scan $SID"
expect 2 "~/.claude/bin/reflect slice $SID main L14"
expect 2 "~/.claude/bin/reflect slice $SID main 14; rm -rf /"
expect 2 "~/.claude/bin/reflect slice $SID main 14 && cat x"
expect 2 "~/.claude/bin/reflect slice $SID main 14 | head"
expect 2 "~/.claude/bin/reflect slice $SID \$(whoami) 14"
expect 2 "cat ~/.claude/projects/x/$SID.jsonl"
expect 2 "/tmp/evil;/.claude/bin/reflect slice $SID main 14"

[ "$fails" -eq 0 ] && echo "ok" || exit 1
