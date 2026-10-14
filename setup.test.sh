#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")" && pwd)
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

git clone -q --local "$ROOT" "$TMP/repo"
git -C "$TMP/repo" switch -q main
cp "$ROOT/setup.sh" "$TMP/repo/setup.sh"
cp -R "$ROOT/codex" "$TMP/repo/codex"
cp "$ROOT/agents/implementation.md" "$TMP/repo/agents/implementation.md"
cp "$ROOT/skills/workflow/implement/SKILL.md" "$TMP/repo/skills/workflow/implement/SKILL.md"
cp "$ROOT/skills/code/code-review/SKILL.md" "$TMP/repo/skills/code/code-review/SKILL.md"
cp "$ROOT/hooks/codex-skill-context.sh" "$TMP/repo/hooks/codex-skill-context.sh"
mkdir -p "$TMP/home/.codex" "$TMP/bin"
cat > "$TMP/home/.codex/hooks.json" <<'JSON'
{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo existing"}]}]}}
JSON
cat > "$TMP/bin/go" <<'SH'
#!/usr/bin/env bash
exit 0
SH
chmod +x "$TMP/bin/go"

HOME="$TMP/home" CODEX_HOME="$TMP/home/.codex" PATH="$TMP/bin:$PATH" bash "$TMP/repo/setup.sh" >/dev/null
HOME="$TMP/home" CODEX_HOME="$TMP/home/.codex" PATH="$TMP/bin:$PATH" bash "$TMP/repo/setup.sh" >/dev/null

jq -e '[.hooks.PreToolUse[]?.hooks[]?.command | select(. == "echo existing")] | length == 1' "$TMP/home/.codex/hooks.json" >/dev/null
jq -e '[.hooks.SessionStart[]?.hooks[]?.command | select(contains("codex-skill-context.sh"))] | length == 1' "$TMP/home/.codex/hooks.json" >/dev/null
jq -e '[.hooks.PreToolUse[]?.hooks[]?.command | select(contains("rtk hook codex"))] | length == 1' "$TMP/home/.codex/hooks.json" >/dev/null
test "$(readlink "$TMP/home/.claude/agents/implementation.md")" = "$TMP/repo/agents/implementation.md"
test "$(readlink "$TMP/home/.claude/skills/implement")" = "$TMP/repo/skills/workflow/implement"
test "$(readlink "$TMP/home/.claude/skills/code-review")" = "$TMP/repo/skills/code/code-review"
test "$(readlink "$TMP/home/.config/opencode/agents/implementation.md")" = "$TMP/repo/agents/implementation.md"
test "$(readlink "$TMP/home/.config/opencode/skills/implement")" = "$TMP/repo/skills/workflow/implement"
test "$(readlink "$TMP/home/.config/opencode/skills/code-review")" = "$TMP/repo/skills/code/code-review"
rm "$TMP/home/.config/opencode/skills/implement"
mkdir "$TMP/home/.config/opencode/skills/implement"
printf '%s\n' 'local override' > "$TMP/home/.config/opencode/skills/implement/SKILL.md"
HOME="$TMP/home" CODEX_HOME="$TMP/home/.codex" PATH="$TMP/bin:$PATH" bash "$TMP/repo/setup.sh" >/dev/null
test "$(cat "$TMP/home/.config/opencode/skills/implement/SKILL.md")" = 'local override'
XDG_CONFIG_HOME="$TMP/xdg" HOME="$TMP/home" CODEX_HOME="$TMP/home/.codex" PATH="$TMP/bin:$PATH" bash "$TMP/repo/setup.sh" >/dev/null
test "$(readlink "$TMP/xdg/opencode/agents/implementation.md")" = "$TMP/repo/agents/implementation.md"
test "$(readlink "$TMP/xdg/opencode/skills/implement")" = "$TMP/repo/skills/workflow/implement"
echo 'setup: Codex hooks installed, user hook preserved, rerun is idempotent'
