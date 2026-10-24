#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/require-swift-skills.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
fails=0

HOME_IN="$TMP/home-in"
HOME_OUT="$TMP/home-out"
mkdir -p "$HOME_IN/.claude/skills/swift-ios" "$HOME_IN/.claude/skills/swift-testing" "$HOME_OUT"
touch "$HOME_IN/.claude/skills/swift-ios/SKILL.md" "$HOME_IN/.claude/skills/swift-testing/SKILL.md"

mkdir -p "$TMP/pkg/Sources/App" "$TMP/xc/App.xcodeproj" "$TMP/xc/App" "$TMP/ws/App.xcworkspace" "$TMP/noroot"
touch "$TMP/pkg/Package.swift"
for d in pkg xc ws; do git -C "$TMP/$d" init -q; done
printf 'import SwiftUI\n' >"$TMP/pkg/Sources/App/A.swift"
printf 'import SwiftUI\n' >"$TMP/xc/App/A.swift"
printf 'import SwiftUI\n' >"$TMP/ws/A.swift"
printf 'import SwiftUI\n' >"$TMP/noroot/A.swift"
printf '// swiftlint:disable all\n// Generated using SwiftGen\nenum L10n {}\n' >"$TMP/xc/App/Strings.swift"
printf '// Generated using Sourcery 2.0\nenum Mocks {}\n' >"$TMP/xc/App/Mocks.swift"
mkdir -p "$TMP/pkg/.build/checkouts/x" "$TMP/xc/DerivedData/x" "$TMP/xc/Pods/x" "$TMP/xc/Carthage/x" "$TMP/xc/SourcePackages/x"
for d in pkg/.build/checkouts/x xc/DerivedData/x xc/Pods/x xc/Carthage/x xc/SourcePackages/x; do
	printf 'import Foundation\n' >"$TMP/$d/X.swift"
done

skill_call() { printf '{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Skill","input":{"skill":"%s"}}]}}\n' "$1"; }
slash_cmd() { printf '{"message":{"role":"user","content":"<command-name>/%s</command-name>"}}\n' "$1"; }
mention() { printf '{"message":{"role":"user","content":"please remember swift-ios"}}\n'; }

T_NONE="$TMP/none.jsonl"
T_MENTION="$TMP/mention.jsonl"
T_SWIFT="$TMP/swift.jsonl"
T_SLASH="$TMP/slash.jsonl"
: >"$T_NONE"
mention >"$T_MENTION"
skill_call swift-ios >"$T_SWIFT"
slash_cmd swift-ios >"$T_SLASH"

payload() {
	jq -nc --arg f "$1" --arg t "$2" --arg tool "${3:-Edit}" \
		'{tool_name: $tool, tool_input: {file_path: $f, content: "import SwiftUI\n"}, transcript_path: $t, session_id: "s"}'
}

check() {
	local name=$1 want=$2 input=$3
	shift 3
	local err code
	err=$(env HOME="$HOME_IN" "$@" "$HOOK" <<<"$input" 2>&1 >/dev/null)
	code=$?
	if [ "$code" != "$want" ]; then
		printf 'FAIL %s: exit %s, want %s\n%s\n' "$name" "$code" "$want" "$err"
		fails=$((fails + 1))
	fi
}

T_PARENT="$TMP/parent.jsonl"
mkdir -p "$TMP/parent/subagents"
skill_call swift-ios >"$T_PARENT"
skill_call swift-ios >"$TMP/parent/subagents/agent-loaded.jsonl"
: >"$TMP/parent/subagents/agent-bare.jsonl"

agent_payload() {
	jq -nc --arg f "$1" --arg t "$2" --arg a "$3" \
		'{tool_name: "Edit", tool_input: {file_path: $f}, transcript_path: $t, agent_id: $a}'
}

A="$TMP/pkg/Sources/App/A.swift"

check "non-Swift file is ignored" 0 "$(payload "$TMP/pkg/README.md" "$T_NONE")" X=1
check "Swift file with no skill loaded blocks" 2 "$(payload "$A" "$T_NONE")" X=1
check "mentioning the skill by name does not count" 2 "$(payload "$A" "$T_MENTION")" X=1
check "Skill tool call unblocks" 0 "$(payload "$A" "$T_SWIFT")" X=1
check "slash command unblocks" 0 "$(payload "$A" "$T_SLASH")" X=1
check "Write is gated like Edit" 2 "$(payload "$A" "$T_NONE" Write)" X=1
check "Package.swift-only root is found" 2 "$(payload "$A" "$T_NONE")" X=1
check "xcodeproj-only root is found" 2 "$(payload "$TMP/xc/App/A.swift" "$T_NONE")" X=1
check "xcworkspace-only root is found" 2 "$(payload "$TMP/ws/A.swift" "$T_NONE")" X=1
check "no project root is skipped" 0 "$(payload "$TMP/noroot/A.swift" "$T_NONE")" X=1
check "SwiftGen output is skipped" 0 "$(payload "$TMP/xc/App/Strings.swift" "$T_NONE")" X=1
check "Sourcery output is skipped" 0 "$(payload "$TMP/xc/App/Mocks.swift" "$T_NONE")" X=1
for d in pkg/.build/checkouts/x xc/DerivedData/x xc/Pods/x xc/Carthage/x xc/SourcePackages/x; do
	check "$d is skipped" 0 "$(payload "$TMP/$d/X.swift" "$T_NONE")" X=1
done
check "opt-out env is honoured" 0 "$(payload "$A" "$T_NONE")" CC_SWIFT_SKILLS_OFF=1
check "uninstalled skill blocks even when loaded" 2 "$(payload "$A" "$T_SWIFT")" HOME="$HOME_OUT"
check "missing transcript fails open" 0 "$(payload "$A" "$TMP/absent.jsonl")" X=1
check "malformed input fails open" 0 "not json" X=1

check "subagent that loaded the skill unblocks" 0 "$(agent_payload "$A" "$T_PARENT" loaded)" X=1
check "subagent without the skill blocks even if the parent loaded it" 2 "$(agent_payload "$A" "$T_PARENT" bare)" X=1
check "subagent with no transcript fails open" 0 "$(agent_payload "$A" "$T_PARENT" absent)" X=1

new_write=$(jq -nc --arg f "$TMP/pkg/Sources/App/New.swift" --arg t "$T_NONE" \
	'{tool_name: "Write", tool_input: {file_path: $f, content: "// DO NOT EDIT\nenum X {}\n"}, transcript_path: $t}')
check "new generated file via Write is skipped" 0 "$new_write" X=1

T_BOTH="$TMP/both.jsonl"
{ skill_call swift-ios; skill_call swift-testing; } >"$T_BOTH"
mkdir -p "$TMP/pkg/Tests/AppTests" "$TMP/xc/AppTests" "$TMP/xc/AppUITests"
printf 'import Foundation\nimport Testing\n' >"$TMP/pkg/Sources/App/Imports.swift"

edit_only() {
	jq -nc --arg f "$1" --arg t "$2" '{tool_name: "Edit", tool_input: {file_path: $f}, transcript_path: $t}'
}
write_new() {
	jq -nc --arg f "$1" --arg t "$2" --arg c "$3" '{tool_name: "Write", tool_input: {file_path: $f, content: $c}, transcript_path: $t}'
}
check_names() {
	local name=$1 input=$2 err
	err=$(env HOME="$HOME_IN" "$HOOK" <<<"$input" 2>&1 >/dev/null)
	if [ $? != 2 ] || ! grep -q swift-testing <<<"$err"; then
		printf 'FAIL %s: want exit 2 naming swift-testing\n%s\n' "$name" "$err"
		fails=$((fails + 1))
	fi
}

check_names "SPM Tests/<X>Tests/ path needs swift-testing" "$(payload "$TMP/pkg/Tests/AppTests/Helpers.swift" "$T_SWIFT")"
check_names "Xcode <App>Tests/ path needs swift-testing" "$(payload "$TMP/xc/AppTests/Fixtures.swift" "$T_SWIFT")"
check_names "UITests/ path needs swift-testing" "$(payload "$TMP/xc/AppUITests/Launch.swift" "$T_SWIFT")"
check_names "filename ending Tests.swift needs swift-testing" "$(payload "$TMP/pkg/Sources/App/CartTests.swift" "$T_SWIFT")"
check_names "filename ending Test.swift needs swift-testing" "$(payload "$TMP/pkg/Sources/App/CartTest.swift" "$T_SWIFT")"
check_names "on-disk file importing Testing needs swift-testing" "$(edit_only "$TMP/pkg/Sources/App/Imports.swift" "$T_SWIFT")"
check_names "new Write importing XCTest needs swift-testing" "$(write_new "$TMP/pkg/Sources/App/New.swift" "$T_SWIFT" $'import XCTest\nfinal class X {}\n')"
check_names "test file with only swift-ios loaded names swift-testing" "$(payload "$TMP/pkg/Tests/AppTests/CartTests.swift" "$T_SWIFT")"
check "non-test Swift file does not need swift-testing" 0 "$(payload "$A" "$T_SWIFT")" X=1
check "test file with both skills loaded passes" 0 "$(payload "$TMP/pkg/Tests/AppTests/CartTests.swift" "$T_BOTH")" X=1

[ "$fails" -eq 0 ] && echo "all hook tests passed"
exit "$fails"
