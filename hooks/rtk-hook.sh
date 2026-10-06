#!/usr/bin/env bash
INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")
CWD=$(jq -r '.cwd // empty' <<<"$INPUT")

claude_code_refuses_rtk_git_here() {
	[[ "$CWD" == */.claude/worktrees/* && "$CMD" =~ (^|[[:space:]\;\&\|\(])git[[:space:]] ]]
}

claude_code_refuses_rtk_git_here && exit 0
exec rtk hook claude <<<"$INPUT"
