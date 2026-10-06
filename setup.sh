#!/usr/bin/env bash
# Wire my-claude-code into ~/.claude/
# Safe to re-run after pulling changes on any machine.
set -e

REPO="$(cd "$(dirname "$0")" && pwd)"
CLAUDE=~/.claude

mkdir -p "$CLAUDE"

link() { ln -sfn "$REPO/$1" "$CLAUDE/$1"; }

link hooks
link commands
link rules
link statusline-command.sh
link RTK.md
link CLAUDE.md

# reflect: machine-wide command; `reflect on` inside a repo installs that repo's hooks
if [ "${1:-}" = "--reflect" ]; then
  mkdir -p "$CLAUDE/bin"
  (cd "$REPO/tools/reflect" && go build -ldflags "-X main.library=$REPO" -o "$CLAUDE/bin/reflect" ./cmd/reflect)
  echo "reflect: built $CLAUDE/bin/reflect"
fi

SETTINGS="$CLAUDE/settings.json"
USER_SETTINGS="$CLAUDE/settings.user.json"

if [ ! -f "$USER_SETTINGS" ]; then
  if [ -f "$SETTINGS" ]; then
    jq 'del(.permissions, .hooks, .statusLine, .enabledPlugins)' "$SETTINGS" > "$USER_SETTINGS"
  else
    echo '{}' > "$USER_SETTINGS"
  fi
  echo "settings.user.json: seeded"
fi

merged=$(jq -s '.[0] * .[1]' "$REPO/settings.shared.json" "$USER_SETTINGS")
if [ -f "$SETTINGS" ] && ! diff -q <(jq -S . "$SETTINGS") <(jq -S . <<<"$merged") >/dev/null; then
  cp -L "$SETTINGS" "$SETTINGS.bak"
  echo "settings.json: dropping edits not in settings.shared.json or settings.user.json (old copy: $SETTINGS.bak)"
  diff <(jq -S . "$SETTINGS.bak") <(jq -S . <<<"$merged") || true
fi
rm -f "$SETTINGS"
printf '%s\n' "$merged" > "$SETTINGS"
echo "settings.json: generated from settings.shared.json + settings.user.json"

echo "Done. Install plugins manually if on a new machine (skill-creator, gopls-lsp)."
