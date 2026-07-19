#!/usr/bin/env bash
#
# Install the product workflow suite into the agent skills root.
#
# Why this exists: the `skills` CLI only discovers skills/<category>/<name>/SKILL.md — one level
# shallower than this directory — so every skill in here is invisible to `npx skills add`. Grouping
# the suite is worth more than CLI tracking, so we copy it in directly instead.
#
# The trade-off, and it is real: `npx skills check` and `npx skills update` will never see these
# skills again. Re-run this script after every change, or the installed copies go stale silently.
#
# Usage:  ./install.sh            # installs to ~/.agents/skills
#         SKILLS_DIR=... ./install.sh
set -euo pipefail

DEST="${SKILLS_DIR:-$HOME/.agents/skills}"
SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "$DEST"
status=0

for dir in "$SRC"/*/; do
  name=$(basename "$dir")
  [ -f "$dir/SKILL.md" ] || continue

  target="$DEST/$name"
  # Only ever clobber something that is itself a skill — never a stray directory that happens
  # to share a name.
  if [ -e "$target" ] && [ ! -f "$target/SKILL.md" ]; then
    echo "  SKIPPED $name — $target exists but is not a skill; remove it by hand"
    status=1
    continue
  fi

  rm -rf "$target"
  cp -R "$dir" "$target"

  # The check: a copy that didn't land is worse than one that never ran, because the stale
  # version stays in place and keeps triggering.
  if diff -r -q "$dir" "$target" >/dev/null 2>&1; then
    echo "  ok      $name"
  else
    echo "  FAILED  $name — copy does not match source"
    status=1
  fi
done

echo
echo "Installed to $DEST"
echo "Not tracked by 'npx skills check/update' — re-run this after any change."
exit $status
