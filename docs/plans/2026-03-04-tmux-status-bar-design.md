# tmux Status Bar Redesign — Srcery Powerline Theme

## Goal

Consolidate the existing two-bar tmux status into a single, polished Powerline-style bar that matches the Srcery color theme. Retain session alert notifications from the old second status line.

## Decisions

- **Single bar** — remove `status 2`, drop `status-format[1]`
- **Powerline style** — chevron separators (``, ``) via Nerd Font
- **Right side** — time only, no hostname
- **Alert visibility** — other sessions with active alerts shown as badges before the time segment; sessions without alerts hidden

## Layout

```
 dotfiles   1:vim  2:shell  3:notes   ···   󰂚 caprock-hoops   22:10
└─────────┘ └──────────────────────┘     └──────────────────┘ └──────┘
  session       window tabs (middle)       alert badges (script)  time
```

## Color Segments

| Segment            | BG         | FG         | Notes                                  |
|--------------------|------------|------------|----------------------------------------|
| Session pill       | colour214  | colour234  | yellow, bold, ` #S `           |
| Status bar BG      | colour234  | —          | Srcery black                           |
| Inactive window    | colour234  | colour244  | dim grey text, no bg highlight         |
| Active window      | colour10   | colour234  | bright green pill, Powerline both ends |
| Alert badges       | colour234  | colour214  | yellow text, bell icon, from script    |
| Time pill          | colour239  | colour223  | grey pill, `  %H:%M `          |

## Implementation

### Files changed

1. **`tmux/.tmux.conf`**
   - Set `status 1` (remove second bar)
   - Rewrite `status-left`: session pill with Powerline separator
   - Rewrite `status-right`: alert script output + time pill
   - Rewrite `window-status-format` and `window-status-current-format` with Powerline arrows

2. **`tmux/.tmux/scripts/alert-sessions.sh`** (new, replaces `session-info.sh`)
   - List sessions, filter to only those with `#{session_alerts}` non-empty
   - Skip current session (passed as `$1`)
   - Output ` 󰂚 session_name` badges separated by spaces
   - Output nothing if no alerts (status-right collapses cleanly)
