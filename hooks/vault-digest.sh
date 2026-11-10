#!/usr/bin/env bash
export LC_ALL=C
dir="${VAULT_DIGEST_DIR:-$HOME/Documents/personal/my-obsidian-notes}/mocs"
names=""
for f in "$dir"/*\ MOC.md; do
	[ -f "$f" ] || continue
	n=$(basename "$f" " MOC.md")
	[ "$n" = Home ] && continue
	names="${names:+$names, }$n"
done
[ -n "$names" ] && printf 'Vault MOCs (vault-recall topics): %s\n' "$names"
exit 0
