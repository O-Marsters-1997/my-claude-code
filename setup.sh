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
link statusline-command.sh
link RTK.md
link CLAUDE.md

# settings.json: symlink shared base, or merge with device-specific overrides if present
if [ -f "$REPO/settings.local.json" ]; then
  jq -s '.[0] * .[1]' "$REPO/settings.json" "$REPO/settings.local.json" > "$CLAUDE/settings.json"
  echo "settings.json: merged base + local override"
else
  link settings.json
  echo "settings.json: symlinked (no local override)"
fi

echo "Done. Install plugins manually if on a new machine (skill-creator, gopls-lsp)."
