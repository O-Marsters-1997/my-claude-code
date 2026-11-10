#!/usr/bin/env bash
set -u
HOOK="$(cd "$(dirname "$0")" && pwd)/vault-digest.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
fails=0

check() {
	local name=$1 want=$2 dir=$3 out code
	out=$(VAULT_DIGEST_DIR="$dir" "$HOOK" 2>&1)
	code=$?
	if [ "$code" != 0 ] || [ "$out" != "$want" ]; then
		printf 'FAIL %s: exit %s, got %q, want %q\n' "$name" "$code" "$out" "$want"
		fails=$((fails + 1))
	fi
}

mkdir -p "$TMP/v/mocs"
touch "$TMP/v/mocs/Home MOC.md" "$TMP/v/mocs/CSS MOC.md" "$TMP/v/mocs/CSS Layout MOC.md" "$TMP/v/mocs/Web Development MOC.md" "$TMP/v/mocs/notes.md"
check "lists MOCs without Home or suffix" 'Vault MOCs (vault-recall topics): CSS Layout, CSS, Web Development' "$TMP/v"

mkdir -p "$TMP/empty/mocs"
check "no MOCs prints nothing" '' "$TMP/empty"
check "missing vault prints nothing" '' "$TMP/absent"

[ "$fails" = 0 ] && echo "vault-digest: all passed"
exit "$fails"
