#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_LINT_CHANGED_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
[ "$(jq -r '.stop_hook_active // false' <<<"$INPUT")" = true ] && exit 0
CWD=$(jq -r '.cwd // empty' <<<"$INPUT")
TOP=$(git -C "${CWD:-.}" rev-parse --show-toplevel 2>/dev/null) || exit 0
git -C "$TOP" rev-parse --verify -q HEAD >/dev/null || exit 0
TOP=$(cd "$TOP" && pwd -P)

CHANGED=$({
	git -C "$TOP" diff --name-only --diff-filter=d HEAD
	git -C "$TOP" ls-files --others --exclude-standard
} | grep -vE '(^|/)(vendor|node_modules)/' | sort -u)
REPORT=""

up() {
	local dir=$1
	shift
	while [ "$dir" != / ]; do
		for name in "$@"; do
			[ -e "$dir/$name" ] && { echo "$dir"; return; }
		done
		dir=$(dirname "$dir")
	done
	return 1
}

by_root() {
	local pattern=$1 f dir root
	shift
	grep -E "$pattern" <<<"$CHANGED" | while IFS= read -r f; do
		dir=$(dirname "$TOP/$f")
		root=$(up "$dir" "$@") || continue
		printf '%s\t%s\n' "$root" "$TOP/$f"
	done | sort
}

lint_go() {
	command -v golangci-lint >/dev/null || return
	local pairs root out pkgs
	pairs=$(by_root '\.go$' go.mod)
	while IFS= read -r root; do
		pkgs=()
		while IFS= read -r dir; do pkgs+=(".${dir#"$root"}/"); done < <(awk -F'\t' -v r="$root" '$1 == r { print $2 }' <<<"$pairs" | xargs -n1 dirname | sort -u)
		out=$(cd "$root" && golangci-lint run --allow-parallel-runners --new-from-rev=HEAD "${pkgs[@]}" 2>&1) || REPORT+="$out"$'\n'
	done < <(cut -f1 <<<"$pairs" | grep . | uniq)
}

lint_biome() {
	local pairs root bin out files
	pairs=$(by_root '\.(ts|tsx|js|jsx|mjs|cjs)$' biome.json biome.jsonc)
	while IFS= read -r root; do
		bin=$(up "$root" node_modules/.bin/biome) || continue
		files=()
		while IFS= read -r f; do files+=("$f"); done < <(awk -F'\t' -v r="$root" '$1 == r { print $2 }' <<<"$pairs")
		out=$(cd "$root" && "$bin/node_modules/.bin/biome" lint --no-errors-on-unmatched "${files[@]}" 2>&1) || REPORT+="$out"$'\n'
	done < <(cut -f1 <<<"$pairs" | grep . | uniq)
}

lint_go
lint_biome
[ -z "$REPORT" ] && exit 0
printf '%sFix these lint errors in the files changed this turn. Set CC_LINT_CHANGED_OFF=1 to skip this check.\n' "$REPORT" >&2
exit 2
