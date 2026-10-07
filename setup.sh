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

ln -sfn "$REPO/githooks/post-merge" "$(git -C "$REPO" rev-parse --path-format=absolute --git-common-dir)/hooks/post-merge"

if [ "${1:-}" = "--reflect" ]; then
  mkdir -p "$CLAUDE/bin"
  (cd "$REPO/tools/reflect" && go build -ldflags "-X main.library=$REPO" -o "$CLAUDE/bin/reflect" ./cmd/reflect)
  echo "reflect: built $CLAUDE/bin/reflect"
fi

SETTINGS="$CLAUDE/settings.json"
USER_SETTINGS="$CLAUDE/settings.user.json"

if [ ! -f "$USER_SETTINGS" ]; then
  if [ -f "$SETTINGS" ]; then
    jq 'del(.permissions, .hooks, .statusLine, .enabledPlugins, .autoMode)' "$SETTINGS" > "$USER_SETTINGS"
  else
    echo '{}' > "$USER_SETTINGS"
  fi
  echo "settings.user.json: seeded"
fi

merged=$(jq -s '.[0] * .[1]' "$REPO/settings.shared.json" "$USER_SETTINGS")
if [ -f "$SETTINGS" ] && ! diff -q <(jq -S . "$SETTINGS") <(jq -S . <<<"$merged") >/dev/null; then
  cp -L "$SETTINGS" "$SETTINGS.bak"
  echo "settings.json: changes below (< old, > new); old copy at $SETTINGS.bak. Move any < lines you want to keep into settings.shared.json or settings.user.json"
  diff <(jq -S . "$SETTINGS.bak") <(jq -S . <<<"$merged") || true
fi
rm -f "$SETTINGS"
printf '%s\n' "$merged" > "$SETTINGS"
echo "settings.json: generated from settings.shared.json + settings.user.json"

echo "Done. Install plugins manually if on a new machine (skill-creator, gopls-lsp)."
