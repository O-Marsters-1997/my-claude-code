#!/bin/bash
INPUT=$(cat)
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // empty')

if [ -n "$(echo "$INPUT" | jq -r '.agent_id // empty')" ]; then
  exit 0
fi

if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
  LAST_TOOL=$(tail -r "$TRANSCRIPT_PATH" \
    | jq -c 'select(.type == "assistant") | [.message.content[]? | select(.type == "tool_use") | .name] | select(length > 0)' \
    | head -1)

  if echo "$LAST_TOOL" | grep -q '"AskUserQuestion"'; then
    exit 0
  fi

  LAUNCHED=$(jq -r 'select(.toolUseResult.status? == "async_launched") | .toolUseResult.agentId // empty' "$TRANSCRIPT_PATH" 2>/dev/null | sort -u)
  FINISHED=$(grep -oE '<task-id>[a-z0-9]+</task-id>|\\?"senderTaskId\\?":\\?"[a-z0-9]+' "$TRANSCRIPT_PATH" | grep -oE '[a-z0-9]{12,}' | sort -u)
  PENDING=$(comm -23 <(echo "$LAUNCHED") <(echo "$FINISHED") | grep -c .)

  if [ "$PENDING" -gt 0 ]; then
    exit 0
  fi
fi

afplay "/Users/olly/Sounds/ElevenLabs_2026-03-25T20_32_07_Julian - Warm, Articulate and Engaging_pvc_sp96_s50_sb75_se0_b_m2.mp3" 2>/dev/null || true
