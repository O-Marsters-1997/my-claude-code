#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/require-swift-skills.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
fails=0

HOME_OK="$TMP/home"
mkdir -p "$HOME_OK/.claude/skills/swift-testing" "$HOME_OK/.claude/skills/swift-ios"
touch "$HOME_OK/.claude/skills/swift-testing/SKILL.md" "$HOME_OK/.claude/skills/swift-ios/SKILL.md"

APP="$TMP/app"
mkdir -p "$APP/Sources/App" "$APP/Tests/AppTests" "$APP/AppTests" "$APP/AppUITests" "$APP/.build/checkouts/x"
git -C "$APP" init -q
printf 'struct A {}\n' >"$APP/Sources/App/A.swift"
printf 'struct A {}\n' >"$APP/Sources/App/Helpers.swift"
printf 'import Foundation\nimport Testing\n' >"$APP/Sources/App/Imports.swift"
printf 'struct A {}\n' >"$APP/.build/checkouts/x/XTests.swift"

skill_call() { printf '{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Skill","input":{"skill":"%s"}}]}}\n' "$1"; }
slash_cmd() { printf '{"message":{"role":"user","content":"<command-name>/%s</command-name>"}}\n' "$1"; }
mention() { printf '{"message":{"role":"user","content":"please remember swift-testing"}}\n'; }

T_NONE="$TMP/none.jsonl"
T_MENTION="$TMP/mention.jsonl"
T_IOS="$TMP/ios.jsonl"
T_TESTING="$TMP/testing.jsonl"
T_SLASH="$TMP/slash.jsonl"
T_BOTH="$TMP/both.jsonl"
: >"$T_NONE"
mention >"$T_MENTION"
skill_call swift-ios >"$T_IOS"
skill_call swift-testing >"$T_TESTING"
slash_cmd swift-testing >"$T_SLASH"
{ skill_call swift-ios; skill_call swift-testing; } >"$T_BOTH"

payload() {
	jq -nc --arg f "$1" --arg t "$2" --arg tool "${3:-Edit}" --arg c "${4:-}" \
		'{tool_name: $tool, tool_input: ({file_path: $f} + (if $c == "" then {} else {content: $c} end)), transcript_path: $t, session_id: "s"}'
}

check() {
	local name=$1 want=$2 input=$3 grep_for=$4
	shift 4
	local err code
	err=$(env HOME="$HOME_OK" "$@" "$HOOK" <<<"$input" 2>&1 >/dev/null)
	code=$?
	if [ "$code" != "$want" ]; then
		printf 'FAIL %s: exit %s, want %s\n%s\n' "$name" "$code" "$want" "$err"
		fails=$((fails + 1))
	elif [ -n "$grep_for" ] && ! grep -q -- "$grep_for" <<<"$err"; then
		printf 'FAIL %s: stderr lacks %s\n%s\n' "$name" "$grep_for" "$err"
		fails=$((fails + 1))
	fi
}

T_PARENT="$TMP/parent.jsonl"
mkdir -p "$TMP/parent/subagents"
skill_call swift-testing >"$T_PARENT"
skill_call swift-testing >"$TMP/parent/subagents/agent-loaded.jsonl"
: >"$TMP/parent/subagents/agent-bare.jsonl"

agent_payload() {
	jq -nc --arg f "$1" --arg t "$2" --arg a "$3" \
		'{tool_name: "Edit", tool_input: {file_path: $f}, transcript_path: $t, agent_id: $a}'
}

check "non-Swift file is ignored" 0 "$(payload "$APP/README.md" "$T_NONE")" ""
check "non-test Swift file does not need swift-testing" 0 "$(payload "$APP/Sources/App/A.swift" "$T_NONE")" ""
check "non-test Swift file with swift-ios loaded passes" 0 "$(payload "$APP/Sources/App/A.swift" "$T_IOS")" ""

check "SPM Tests/<X>Tests/ path needs swift-testing" 2 "$(payload "$APP/Tests/AppTests/Helpers.swift" "$T_NONE")" swift-testing
check "Xcode <App>Tests/ path needs swift-testing" 2 "$(payload "$APP/AppTests/Fixtures.swift" "$T_NONE")" swift-testing
check "UITests/ path needs swift-testing" 2 "$(payload "$APP/AppUITests/Launch.swift" "$T_NONE")" swift-testing
check "filename ending Tests.swift needs swift-testing" 2 "$(payload "$APP/Sources/App/CartTests.swift" "$T_NONE")" swift-testing
check "filename ending Test.swift needs swift-testing" 2 "$(payload "$APP/Sources/App/CartTest.swift" "$T_NONE")" swift-testing
check "on-disk file importing Testing needs swift-testing" 2 "$(payload "$APP/Sources/App/Imports.swift" "$T_NONE")" swift-testing
check "new Write importing XCTest needs swift-testing" 2 "$(payload "$APP/Sources/App/New.swift" "$T_NONE" Write $'import XCTest\nfinal class X {}\n')" swift-testing
check "Write content without a test import passes" 0 "$(payload "$APP/Sources/App/Plain.swift" "$T_NONE" Write $'import Foundation\n')" ""

check "test file with only swift-ios loaded blocks and names swift-testing" 2 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_IOS")" swift-testing
check "test file with both skills loaded passes" 0 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_BOTH")" ""
check "Skill tool call unblocks" 0 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_TESTING")" ""
check "slash command unblocks" 0 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_SLASH")" ""
check "mentioning the skill by name does not count" 2 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_MENTION")" ""
check "uninstalled skill names the install command" 2 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_NONE")" "npx skills add" HOME="$TMP/nohome"

check ".build is skipped" 0 "$(payload "$APP/.build/checkouts/x/XTests.swift" "$T_NONE")" ""
check "opt-out env is honoured" 0 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$T_NONE")" "" CC_SWIFT_SKILLS_OFF=1
check "missing transcript fails open" 0 "$(payload "$APP/Tests/AppTests/CartTests.swift" "$TMP/absent.jsonl")" ""
check "malformed input fails open" 0 "not json" ""

check "subagent that loaded the skill unblocks" 0 "$(agent_payload "$APP/Tests/AppTests/CartTests.swift" "$T_PARENT" loaded)" ""
check "subagent without the skill blocks even if the parent loaded it" 2 "$(agent_payload "$APP/Tests/AppTests/CartTests.swift" "$T_PARENT" bare)" ""

[ "$fails" -eq 0 ] && echo "all hook tests passed"
exit "$fails"
