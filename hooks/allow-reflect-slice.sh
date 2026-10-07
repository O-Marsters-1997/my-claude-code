#!/usr/bin/env bash
CMD=$(jq -r '.tool_input.command // empty')
BIN='(~|\$HOME|/[^ ;&|`$()]*)/\.claude/bin/reflect'
[[ "$CMD" =~ ^${BIN}\ slice\ [0-9a-f-]+\ [A-Za-z0-9_-]+\ [0-9]+(\ -C\ [0-9]+)?$ ]] && exit 0
REPLAY="replay\\ [0-9a-f-]+\\ [A-Za-z0-9_-]+\\ [0-9]+(\\ --cwd\\ /[A-Za-z0-9_./-]*)?(\\ --command\\ '[^']*')?"
[[ "$CMD" =~ ^${BIN}\ ${REPLAY}$ ]] && exit 0
echo "Only '~/.claude/bin/reflect slice <sid> <agent> <line> [-C n]' and '... replay <sid> <agent> <line> [--cwd dir] [--command '"'"'cmd'"'"']' may run here, as one plain command." >&2
exit 2
