#!/bin/bash
INPUT=$(cat)
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path')

if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
  LAST_TOOL=$(tac "$TRANSCRIPT_PATH" \
    | jq -c 'select(.type == "assistant") | [.message.content[]? | select(.type == "tool_use") | .name] | select(length > 0)' \
    | head -1)

  if echo "$LAST_TOOL" | grep -q '"AskUserQuestion"'; then
    exit 0
  fi
fi

afplay "/Users/olly/Sounds/ElevenLabs_2026-03-25T20_32_07_Julian - Warm, Articulate and Engaging_pvc_sp96_s50_sb75_se0_b_m2.mp3" 2>/dev/null || true
