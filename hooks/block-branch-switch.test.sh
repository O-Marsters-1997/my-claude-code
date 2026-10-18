#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/block-branch-switch.sh"
fails=0

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
MAIN="$TMP/main"
LINKED="$TMP/linked"
git init -q -b main "$MAIN"
git -C "$MAIN" -c user.name=t -c user.email=t@t commit -q --allow-empty -m init
git -C "$MAIN" branch feat
mkdir -p "$MAIN/src"
touch "$MAIN/src/a.go"
git -C "$MAIN" worktree add -q "$LINKED" -b other

check() {
	local name=$1 want=$2 cwd=$3 cmd=$4 got
	jq -nc --arg c "$cmd" --arg d "$cwd" '{tool_name: "Bash", cwd: $d, tool_input: {command: $c}}' | bash "$HOOK" >/dev/null 2>&1
	got=$?
	if [ "$got" -ne "$want" ]; then
		echo "FAIL: $name (want $want, got $got)"
		fails=$((fails + 1))
	fi
}

check "checkout -b in primary" 2 "$MAIN" "git checkout -b new"
check "switch -c in primary" 2 "$MAIN" "git switch -c new"
check "switch existing in primary" 2 "$MAIN" "git switch feat"
check "checkout existing branch in primary" 2 "$MAIN" "git checkout feat"
check "checkout previous branch in primary" 2 "$MAIN" "git checkout -"
check "checkout --detach in primary" 2 "$MAIN" "git checkout --detach"
check "gh pr checkout in primary" 2 "$MAIN" "gh pr checkout 12"
check "git -C primary from linked" 2 "$LINKED" "git -C $MAIN checkout -b new origin/main"
check "cd into primary then switch" 2 "$LINKED" "cd $MAIN && git switch feat"
check "chained after other commands" 2 "$MAIN" "git fetch && git checkout -b new"
check "env prefix" 2 "$MAIN" "GIT_TRACE=1 git switch feat"

check "checkout -- path in primary" 0 "$MAIN" "git checkout -- src/a.go"
check "checkout path without -- in primary" 0 "$MAIN" "git checkout src/a.go"
check "checkout ref -- path in primary" 0 "$MAIN" "git checkout feat -- src/a.go"
check "checkout -b in linked worktree" 0 "$LINKED" "git checkout -b new"
check "switch in linked worktree" 0 "$LINKED" "git switch feat"
check "git -C linked from primary" 0 "$MAIN" "git -C $LINKED switch -c new"
check "worktree add in primary" 0 "$MAIN" "git worktree add ../x -b new"
check "tp new in primary" 0 "$MAIN" "tp new feat-x --base origin/main"
check "branch listing" 0 "$MAIN" "git branch --show-current"
check "grep for the text" 0 "$MAIN" "grep -n 'git checkout -b' README.md"
check "outside a repo" 0 "$TMP" "git checkout -b new"

CC_BRANCH_SWITCH_GATE_OFF=1 bash "$HOOK" < <(jq -nc --arg d "$MAIN" '{tool_name: "Bash", cwd: $d, tool_input: {command: "git switch feat"}}') >/dev/null 2>&1 || { echo "FAIL: gate off allows"; fails=$((fails + 1)); }
echo 'not json' | bash "$HOOK" >/dev/null 2>&1 || { echo "FAIL: malformed input fails open"; fails=$((fails + 1)); }
jq -nc '{tool_name: "Edit", tool_input: {file_path: "a"}}' | bash "$HOOK" >/dev/null 2>&1 || { echo "FAIL: non-Bash tool ignored"; fails=$((fails + 1)); }

[ "$fails" -eq 0 ] && echo "all hook tests passed"
exit "$fails"
