#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")/.." && pwd)/githooks/global/pre-commit"
fails=0
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

check() {
	if [ "$2" != "$3" ]; then
		echo "FAIL: $1 (want '$3', got '$2')"
		fails=$((fails + 1))
	fi
}

repo="$tmp/repo"
git init -q "$repo"
mkdir -p "$tmp/home/.claude/bin"
log="$tmp/log"

stub() {
	printf '#!/bin/sh\necho "reflect $*" >> "%s"\nexit %s\n' "$log" "$1" > "$tmp/home/.claude/bin/reflect"
	chmod +x "$tmp/home/.claude/bin/reflect"
}

repo_hook() {
	printf '#!/bin/sh\necho "repo-hook $*" >> "%s"\nexit %s\n' "$log" "$1" > "$repo/.git/hooks/pre-commit"
	chmod +x "$repo/.git/hooks/pre-commit"
}

run() {
	: > "$log"
	(cd "$repo" && HOME="$tmp/home" "$HOOK" 2> "$tmp/err")
	echo $?
}

stub 0
repo_hook 0
check "pass runs harvest then repo hook" "$(run; cat "$log")" "$(printf '0\nreflect learn harvest --block\nrepo-hook ')"

stub 1
check "harvest failure blocks and skips repo hook" "$(run; cat "$log")" "$(printf '1\nreflect learn harvest --block')"

rm "$tmp/home/.claude/bin/reflect"
check "missing binary skips harvest, runs repo hook" "$(run; cat "$log")" "$(printf '0\nrepo-hook ')"
grep -q "skipping" "$tmp/err" || { echo "FAIL: missing binary prints no warning"; fails=$((fails + 1)); }

rm "$repo/.git/hooks/pre-commit"
check "no repo hook passes" "$(run)" "0"

stub 0
repo_hook 3
check "repo hook failure propagates" "$(run)" "3"

[ "$fails" -eq 0 ] && echo "ok" || exit 1
