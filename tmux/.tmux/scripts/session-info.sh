#!/bin/sh
# session-info.sh - Show tmux sessions with active pane context
# Usage: session-info.sh <current_session_name>
# Designed for use in tmux status-format via:
#   #(~/.tmux/scripts/session-info.sh "#{session_name}")

current="$1"

tmux list-sessions -F '#{session_name}|#{session_attached}|#{session_windows}' 2>/dev/null | {
    first=true
    while IFS='|' read -r name attached win_count; do
        pane_title=$(tmux display-message -t "${name}" -p '#{pane_title}' 2>/dev/null)
        window_name=$(tmux display-message -t "${name}" -p '#{window_name}' 2>/dev/null)

        $first || printf "  |  "
        first=false

        # [current]  (attached)   other
        if [ "$name" = "$current" ]; then
            printf "[%s]" "$name"
        elif [ "$attached" = "1" ]; then
            printf "(%s)" "$name"
        else
            printf " %s " "$name"
        fi

        # Show pane title if meaningful, fall back to window name
        if [ -n "$pane_title" ] && [ "$pane_title" != "$window_name" ] && [ "$pane_title" != "$name" ]; then
            printf " %s" "$pane_title"
        elif [ -n "$window_name" ] && [ "$window_name" != "zsh" ] && [ "$window_name" != "bash" ] && [ "$window_name" != "$name" ]; then
            printf " %s" "$window_name"
        fi

        # Show window count when there are multiple
        if [ "$win_count" -gt 1 ] 2>/dev/null; then
            printf " (%dw)" "$win_count"
        fi
    done
}
