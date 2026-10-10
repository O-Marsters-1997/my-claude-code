#!/usr/bin/env bash
command -v jq >/dev/null || exit 0
[ "${CC_INPLACE_GATE_OFF:-}" = 1 ] && exit 0

CMD=$(jq -r 'select(.tool_name == "Bash") | .tool_input.command // empty' 2>/dev/null)
[ -n "$CMD" ] || exit 0

SCRATCH='(^|[^[:alnum:]_])(/private)?/tmp/|/var/folders/|/scratchpad/|\$TMPDIR|\$\{TMPDIR'
CMDPOS='(^|&&|;|\|\||\||\$\(|\(|-exec|-execdir)[[:space:]]*(([A-Za-z_][A-Za-z_0-9]*=[^[:space:]]*|sudo|env|xargs|exec|command|time|rtk)[[:space:]]+)*([^[:space:]]*/)?'
SED_INPLACE="${CMDPOS}sed[[:space:]]+([^|&]*[[:space:]])?(-[a-zA-Z]*i[a-zA-Z]*|--in-place)([[:space:]=.]|$)"
PERL_INPLACE="${CMDPOS}perl[[:space:]]+([^|&]*[[:space:]])?-[a-zA-Z0-9]*i[a-zA-Z0-9.]*([[:space:]]|$)"
INTERP="${CMDPOS}(python3?|node|ruby|bun|deno)[[:space:]]"
FILE_WRITE='open\([^)]*,[[:space:]]*(mode[[:space:]]*=[[:space:]]*)?[bf]?["'\''][wax]\+?b?["'\'']|open\([^)]*["'\''](r\+|rb\+)["'\'']|writeFileSync|writeFile\(|appendFileSync|write_text|write_bytes|File\.write|Bun\.write'

deny() {
	printf 'Blocked: %s. Edit files with the Edit or Write tool so the comment and skill gates run. Redirects and scripts that write under /tmp or the scratchpad are fine. Set CC_INPLACE_GATE_OFF=1 to skip this check.\n' "$1" >&2
	exit 2
}

if grep -qE "$SED_INPLACE" <<<"$CMD"; then
	deny "sed -i rewrites a file in place"
fi
if grep -qE "$PERL_INPLACE" <<<"$CMD"; then
	deny "perl -i rewrites a file in place"
fi
if grep -qE "$INTERP" <<<"$CMD" && grep -qE "$FILE_WRITE" <<<"$CMD" && ! grep -qE "$SCRATCH" <<<"$CMD"; then
	deny "an inline interpreter script opens a repo file for writing"
fi
exit 0
