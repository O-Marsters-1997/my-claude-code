#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")" && pwd)
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

git clone -q --local "$ROOT" "$TMP/repo"
git -C "$TMP/repo" switch -q main
cp "$ROOT/setup.sh" "$TMP/repo/setup.sh"
cp -R "$ROOT/codex" "$TMP/repo/codex"
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
echo 'setup: Codex hooks installed, user hook preserved, rerun is idempotent'
