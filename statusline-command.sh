#!/bin/sh
input=$(cat)

cwd=$(echo "$input" | jq -r '.workspace.current_dir // .cwd // ""')
dir=$(basename "$cwd")
model=$(echo "$input" | jq -r '.model.display_name // .model.id // ""')

branch=$(git -C "$cwd" --no-optional-locks symbolic-ref --short HEAD 2>/dev/null)

ctx_used=$(echo "$input" | jq -r '.rate_limits.five_hour.used_percentage // empty')
ctx_reset=$(echo "$input" | jq -r '.rate_limits.five_hour.reset_at // empty')
week_used=$(echo "$input" | jq -r '.rate_limits.seven_day.used_percentage // empty')

parts=""

# Directory (green, distinct from plan mode's yellow/orange)
parts=$(printf '\033[32m%s\033[0m' "$dir")

# Git branch (blue + red)
if [ -n "$branch" ]; then
  parts="$parts $(printf '\033[1;34mgit:(\033[0;31m%s\033[1;34m)\033[0m' "$branch")"
fi

# Model
if [ -n "$model" ]; then
  parts="$parts $(printf '\033[2m%s\033[0m' "$model")"
fi

# Session context window usage + time remaining
if [ -n "$ctx_used" ]; then
  ctx_int=$(printf '%.0f' "$ctx_used")

  # Compute time remaining in the 5-hour window if reset_at is available
  time_left=""
  if [ -n "$ctx_reset" ]; then
    now_epoch=$(date +%s)
    reset_epoch=$(date -j -f '%Y-%m-%dT%H:%M:%SZ' "$ctx_reset" '+%s' 2>/dev/null \
      || date -d "$ctx_reset" '+%s' 2>/dev/null)
    if [ -n "$reset_epoch" ] && [ "$reset_epoch" -gt "$now_epoch" ]; then
      secs_left=$(( reset_epoch - now_epoch ))
      hrs_left=$(( secs_left / 3600 ))
      mins_left=$(( (secs_left % 3600) / 60 ))
      if [ "$hrs_left" -gt 0 ]; then
        time_left="${hrs_left}h${mins_left}m"
      else
        time_left="${mins_left}m"
      fi
    fi
  fi

  if [ "$ctx_int" -ge 80 ]; then
    color='\033[1;31m'
  else
    color='\033[1;36m'
  fi

  if [ -n "$time_left" ]; then
    parts="$parts $(printf "${color}[session:%s%% (%s left)]\033[0m" "$ctx_int" "$time_left")"
  else
    parts="$parts $(printf "${color}[session:%s%%]\033[0m" "$ctx_int")"
  fi
fi

# Weekly usage limit
if [ -n "$week_used" ]; then
  week_int=$(printf '%.0f' "$week_used")
  if [ "$week_int" -ge 80 ]; then
    parts="$parts $(printf '\033[1;31m[week:%s%%]\033[0m' "$week_int")"
  else
    parts="$parts $(printf '\033[1;33m[week:%s%%]\033[0m' "$week_int")"
  fi
fi

printf '%s' "$parts"
