#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/block-bash-inplace-edits.sh"
fails=0

check() {
	local name=$1 want=$2 cmd=$3 got
	jq -nc --arg c "$cmd" '{tool_name: "Bash", tool_input: {command: $c}}' | bash "$HOOK" >/dev/null 2>&1
	got=$?
	if [ "$got" -ne "$want" ]; then
		echo "FAIL: $name (want $want, got $got)"
		fails=$((fails + 1))
	fi
}

PYEDIT=$'python3 - <<\'EOF\'\np = "api/x.go"\ns = open(p).read()\nopen(p, \'w\').write(s.replace("a", "b"))\nEOF'
PYSCRATCH=$'python3 - <<\'EOF\'\nopen("/private/tmp/claude-501/scratchpad/out.txt", "w").write("x")\nEOF'
NODEEDIT='node -e "require(\"fs\").writeFileSync(\"src/a.ts\", \"x\")"'

check "python heredoc open w on repo file" 2 "$PYEDIT"
check "python heredoc writing to scratchpad" 0 "$PYSCRATCH"
check "node writeFileSync on repo file" 2 "$NODEEDIT"
check "python read-only script" 0 $'python3 - <<\'EOF\'\nprint(open("a.go").read())\nEOF'
check "sed -i with empty suffix" 2 "sed -i '' 's/a/b/' file.go"
check "sed -i bare" 2 "sed -i 's/a/b/' file.go"
check "sed -i.bak" 2 "sed -i.bak 's/a/b/' file.go"
check "sed --in-place" 2 "sed --in-place 's/a/b/' file.go"
check "sed -n print" 0 "sed -n 1,40p file.go"
check "sed -E to stdout" 0 "sed -E 's/a/b/' file.go"
check "sed -i after a pipe-less chain" 2 "cd x && sed -i 's/a/b/' f"
check "perl -0pi" 2 "perl -0pi -e 's/a/b/' file_test.go"
check "perl -pi" 2 "perl -pi -e 's/a/b/' f"
check "perl -i.bak" 2 "perl -i.bak -pe 's/a/b/' f"
check "perl -e without -i" 0 "perl -e 'print 1'"
check "gofmt -w" 0 "gofmt -w ."
check "sqlc generate" 0 "sqlc generate"
check "bun run check --write" 0 "bun run check --write"
check "redirect to tmp" 0 "echo hi > /tmp/x.txt"
check "ls" 0 "ls -la"
check "grep for sed -i text" 0 "grep -n 'sed -i' README.md"

echo 'not json' | bash "$HOOK" >/dev/null 2>&1 || { echo "FAIL: malformed input fails open"; fails=$((fails + 1)); }
jq -nc '{tool_name: "Edit", tool_input: {file_path: "a"}}' | bash "$HOOK" >/dev/null 2>&1 || { echo "FAIL: non-Bash tool ignored"; fails=$((fails + 1)); }

[ "$fails" -eq 0 ] && echo "all hook tests passed"
exit "$fails"
