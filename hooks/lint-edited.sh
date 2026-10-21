#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_LINT_EDITED_OFF:-}" = 1 ] && exit 0

FILE=$(jq -r '.tool_input.file_path // empty')
[ -f "$FILE" ] || exit 0
case "$FILE" in */vendor/* | */node_modules/*) exit 0 ;; esac

up() {
	local dir
	dir=$(dirname "$FILE")
	while [ "$dir" != / ]; do
		for name in "$@"; do
			[ -e "$dir/$name" ] && { echo "$dir"; return; }
		done
		dir=$(dirname "$dir")
	done
	return 1
}

case "$FILE" in
*.go)
	command -v golangci-lint >/dev/null || exit 0
	head -n 20 "$FILE" | grep -qE '^// Code generated .* DO NOT EDIT\.$' && exit 0
	ROOT=$(up go.mod) || exit 0
	git -C "$ROOT" rev-parse --verify -q HEAD >/dev/null || exit 0
	ROOT=$(cd "$ROOT" && pwd -P)
	DIR=$(cd "$(dirname "$FILE")" && pwd -P)
	OUT=$(cd "$ROOT" && golangci-lint run --allow-parallel-runners --new-from-rev=HEAD ".${DIR#"$ROOT"}/" 2>&1)
	;;
*.ts | *.tsx | *.js | *.jsx | *.mjs | *.cjs)
	ROOT=$(up biome.json biome.jsonc) || exit 0
	BIN=$(up node_modules/.bin/biome) || exit 0
	OUT=$(cd "$ROOT" && "$BIN/node_modules/.bin/biome" lint --no-errors-on-unmatched "$FILE" 2>&1)
	;;
*) exit 0 ;;
esac || {
	printf '%s\nFix these lint errors in %s. Set CC_LINT_EDITED_OFF=1 to skip this check.\n' "$OUT" "$FILE" >&2
	exit 2
}
