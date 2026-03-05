#!/bin/sh
# alert-sessions.sh - Emit tmux format badges for sessions with bell alerts.
# Usage: alert-sessions.sh <current_session_name>
# Outputs nothing if no other sessions have bell alerts.

current="$1"
YELLOW="colour214"

tmux list-windows -a -F '#{session_name}|#{window_bell_flag}' 2>/dev/null | while IFS='|' read -r name bell_flag; do
    [ "$name" = "$current" ] && continue
    [ "$bell_flag" = "1" ] || continue
    printf "#[fg=%s] 󰂚 %s " "$YELLOW" "$name"
done
