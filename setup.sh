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

# settings.json: symlink shared base, or merge with device-specific overrides / opt-in reflect hooks
WITH_REFLECT=false
[ "${1:-}" = "--reflect" ] && WITH_REFLECT=true

if $WITH_REFLECT; then
  mkdir -p "$CLAUDE/bin"
  (cd "$REPO/tools/reflect" && go build -ldflags "-X main.library=$REPO" -o "$CLAUDE/bin/reflect" ./cmd/reflect)
  echo "reflect: built $CLAUDE/bin/reflect"
fi

if [ -f "$REPO/settings.local.json" ] || $WITH_REFLECT; then
  tmp="$(mktemp)"
  if [ -f "$REPO/settings.local.json" ]; then
    jq -s '.[0] * .[1]' "$REPO/settings.json" "$REPO/settings.local.json" > "$tmp"
  else
    cp "$REPO/settings.json" "$tmp"
  fi
  if $WITH_REFLECT; then
    jq --slurpfile r "$REPO/tools/reflect/hooks.json" \
      'reduce ($r[0].hooks | to_entries[]) as $e (.; .hooks[$e.key] = ((.hooks[$e.key] // []) + $e.value))' \
      "$tmp" > "$tmp.merged" && mv "$tmp.merged" "$tmp"
  fi
  mv -f "$tmp" "$CLAUDE/settings.json"
  echo "settings.json: merged (local override: $([ -f "$REPO/settings.local.json" ] && echo yes || echo no), reflect: $WITH_REFLECT)"
else
  link settings.json
  echo "settings.json: symlinked (no local override)"
fi

echo "Done. Install plugins manually if on a new machine (skill-creator, gopls-lsp)."
