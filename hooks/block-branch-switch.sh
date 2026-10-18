#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_BRANCH_SWITCH_GATE_OFF:-}" = 1 ] && exit 0

INPUT=$(cat)
CMD=$(jq -r 'select(.tool_name == "Bash") | .tool_input.command // empty' <<<"$INPUT" 2>/dev/null)
[ -n "$CMD" ] || exit 0
dir=$(jq -r '.cwd // empty' <<<"$INPUT" 2>/dev/null)
[ -n "$dir" ] || dir=$PWD

resolve() {
	local p=${1/#\~/$HOME}
	case $p in
	/*) printf '%s' "$p" ;;
	*) printf '%s/%s' "$2" "$p" ;;
	esac
}

is_primary() {
	local gd gc
	gd=$(git -C "$1" rev-parse --absolute-git-dir 2>/dev/null) || return 1
	gc=$(git -C "$1" rev-parse --path-format=absolute --git-common-dir 2>/dev/null) || return 1
	[ "$gd" = "$gc" ]
}

checkout_moves_head() {
	local target=$1 a
	shift
	for a in "$@"; do
		case $a in
		-b | -B | --orphan | --detach) return 0 ;;
		--) return 1 ;;
		esac
	done
	for a in "$@"; do
		case $a in
		-) return 0 ;;
		-*) ;;
		*) git -C "$target" rev-parse --verify --quiet "$a^{commit}" >/dev/null 2>&1 && return 0 || return 1 ;;
		esac
	done
	return 1
}

deny() {
	printf 'Blocked: %s moves HEAD in the primary checkout at %s, which parallel sessions share. Create the branch in its own worktree with `tp new <branch> --base <ref> --current` (see the treepad skill) and work there. The user can switch it themselves with `! git switch ...`, or set CC_BRANCH_SWITCH_GATE_OFF=1 to skip this check.\n' "$1" "$2" >&2
	exit 2
}

segments=${CMD//&&/$'\n'}
segments=${segments//||/$'\n'}
segments=${segments//;/$'\n'}
segments=${segments//|/$'\n'}

while IFS= read -r seg; do
	read -ra t <<<"${seg//[()]/ }"
	i=0
	while [ $i -lt ${#t[@]} ] && [[ ${t[$i]} =~ ^([A-Za-z_][A-Za-z_0-9]*=.*|sudo|env|command|rtk|exec)$ ]]; do
		i=$((i + 1))
	done
	[ $i -lt ${#t[@]} ] || continue
	prog=${t[$i]##*/}
	case $prog in
	cd)
		[ -n "${t[$((i + 1))]:-}" ] && dir=$(resolve "${t[$((i + 1))]}" "$dir")
		;;
	gh)
		[ "${t[$((i + 1))]:-}" = pr ] && [ "${t[$((i + 2))]:-}" = checkout ] &&
			is_primary "$dir" && deny "gh pr checkout" "$dir"
		;;
	git)
		target=$dir
		i=$((i + 1))
		while [ $i -lt ${#t[@]} ]; do
			case ${t[$i]} in
			-C) target=$(resolve "${t[$((i + 1))]:-}" "$dir") && i=$((i + 2)) ;;
			-c) i=$((i + 2)) ;;
			-*) i=$((i + 1)) ;;
			*) break ;;
			esac
		done
		case ${t[$i]:-} in
		switch)
			is_primary "$target" && deny "git switch" "$target"
			;;
		checkout)
			checkout_moves_head "$target" "${t[@]:$((i + 1))}" && is_primary "$target" &&
				deny "git checkout" "$target"
			;;
		esac
		;;
	esac
done <<<"$segments"
exit 0
