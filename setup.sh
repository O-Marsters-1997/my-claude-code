#!/usr/bin/env bash
# Wire my-claude-code into ~/.claude/
# Safe to re-run after pulling changes on any machine.
set -e

REPO="$(cd "$(dirname "$0")" && pwd)"
CLAUDE=~/.claude

BRANCH="$(git -C "$REPO" branch --show-current)"
if [ "$BRANCH" != main ] && [ "${SETUP_ALLOW_BRANCH:-}" != 1 ]; then
  echo "setup.sh: $REPO is on '${BRANCH:-detached HEAD}', not main. Running it here repoints ~/.claude at this checkout and regenerates settings.json from it." >&2
  echo "Run it from the main checkout, or set SETUP_ALLOW_BRANCH=1 to do this deliberately." >&2
  exit 1
fi
echo "setup.sh: wiring from $REPO ($BRANCH @ $(git -C "$REPO" rev-parse --short HEAD))"

mkdir -p "$CLAUDE"

link() { ln -sfn "$REPO/$1" "$CLAUDE/$1"; }

link hooks
link commands
link rules
link statusline-command.sh
link RTK.md
link CLAUDE.md

mkdir -p "$CLAUDE/bin"
ln -sfn "$REPO/bin/fleet-init" "$CLAUDE/bin/fleet-init"

ln -sfn "$REPO/githooks/post-merge" "$(git -C "$REPO" rev-parse --path-format=absolute --git-common-dir)/hooks/post-merge"

mkdir -p "$CLAUDE/bin"
(cd "$REPO/tools/reflect" && go build -ldflags "-X main.library=$REPO" -o "$CLAUDE/bin/reflect" ./cmd/reflect)
echo "reflect: built $CLAUDE/bin/reflect"

git config --global core.hooksPath "$REPO/githooks/global"

if [ "${1:-}" = "--reflect" ]; then
  mkdir -p "$CLAUDE/agents"
  link agents/reflect-reviewer.md
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
