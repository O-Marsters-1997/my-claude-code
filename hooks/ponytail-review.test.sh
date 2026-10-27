#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/ponytail-review.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
fails=0

mkdir -p "$TMP/bin"
cat >"$TMP/bin/claude" <<'EOF'
#!/usr/bin/env bash
cat >"$FAKE_DIR/stdin"
printf '%s' "${FAKE_OUT:-OK}"
exit "${FAKE_CODE:-0}"
EOF
chmod +x "$TMP/bin/claude"

R="$TMP/repo"
mkdir -p "$R"
printf 'package a\n' >"$R/a.go"
printf 'notes\n' >"$R/notes.md"
git -C "$R" init -q && git -C "$R" add -A && git -C "$R" -c user.name=t -c user.email=t@t commit -qm init
R=$(cd "$R" && pwd -P)
printf 'package a\n\nfunc Dead() {}\n' >"$R/a.go"
printf 'package a\n\nfunc New() {}\n' >"$R/new.go"
printf 'more notes\n' >"$R/notes.md"

prompt() { jq -nc --arg t "$1" '{type: "user", message: {content: $t}}'; }
meta() { jq -nc '{type: "user", isMeta: true, message: {content: [{type: "text", text: "Stop hook feedback"}]}}'; }
edit() { jq -nc --arg p "$1" '{type: "assistant", message: {content: [{type: "tool_use", name: "Edit", input: {file_path: $p}}]}}'; }
result() { jq -nc '{type: "user", message: {content: [{type: "tool_result", content: "ok"}]}}'; }

check() {
	local name=$1 want=$2 called=$3 active=$4
	shift 4
	local err code
	rm -f "$TMP/stdin"
	err=$(jq -nc --arg c "$R" --arg t "$TMP/transcript.jsonl" --argjson a "$active" \
		'{hook_event_name: "Stop", cwd: $c, transcript_path: $t, stop_hook_active: $a}' |
		env PATH="$TMP/bin:$PATH" FAKE_DIR="$TMP" "$@" "$HOOK" 2>&1 >/dev/null)
	code=$?
	if [ "$code" != "$want" ]; then
		printf 'FAIL %s: exit %s, want %s\n%s\n' "$name" "$code" "$want" "$err"
		fails=$((fails + 1))
	fi
	if { [ "$called" = yes ] && [ ! -f "$TMP/stdin" ]; } || { [ "$called" = no ] && [ -f "$TMP/stdin" ]; }; then
		printf 'FAIL %s: reviewer called=%s, want %s\n' "$name" "$([ -f "$TMP/stdin" ] && echo yes || echo no)" "$called"
		fails=$((fails + 1))
	fi
}

{ prompt "hi"; edit "$R/a.go"; result; prompt "just a question"; } >"$TMP/transcript.jsonl"
check "no edit this turn skips the reviewer" 0 no false X=1

{ prompt "fix"; edit "$R/notes.md"; result; } >"$TMP/transcript.jsonl"
check "non-code edit skips the reviewer" 0 no false X=1

{ prompt "fix"; edit "$TMP/elsewhere.go"; result; } >"$TMP/transcript.jsonl"
touch "$TMP/elsewhere.go"
check "edit outside the repo skips the reviewer" 0 no false X=1

{ prompt "fix"; edit "$R/a.go"; result; meta; edit "$R/new.go"; result; } >"$TMP/transcript.jsonl"
check "clean review passes" 0 yes false X=1
grep -q 'func Dead' "$TMP/stdin" && grep -q 'func New' "$TMP/stdin" || {
	echo "FAIL reviewer stdin lacks the tracked and untracked diffs (meta message reset the turn?)"
	fails=$((fails + 1))
}
check "findings block" 2 yes false FAKE_OUT='a.go:3 - dead - delete it'
check "reviewer failure passes" 0 yes false FAKE_CODE=1
check "second stop in a turn passes" 0 no true FAKE_OUT='x'
check "opt-out env skips" 0 no false CC_PONYTAIL_REVIEW_OFF=1

[ "$fails" = 0 ] && echo "ponytail-review: all passed"
exit "$fails"
