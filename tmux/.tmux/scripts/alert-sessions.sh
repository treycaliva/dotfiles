#!/bin/sh
# alert-sessions.sh - Emit tmux format badges for sessions with active alerts.
# Usage: alert-sessions.sh <current_session_name>
# Outputs nothing if no other sessions have alerts.

current="$1"
YELLOW="colour214"

tmux list-sessions -F '#{session_name}|#{session_alerts}' 2>/dev/null | while IFS='|' read -r name alerts; do
    [ "$name" = "$current" ] && continue
    [ -z "$alerts" ] && continue
    printf "#[fg=%s] 󰂚 %s " "$YELLOW" "$name"
done
