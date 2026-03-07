# tmux Status Bar Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the two-bar tmux status with a single Srcery Powerline bar that shows session name, windows, alert badges, and time.

**Architecture:** Two file changes — a new `alert-sessions.sh` script that emits tmux format strings for sessions with active alerts, and updated `tmux.conf` status segments using Powerline separator characters. The old `session-info.sh` is removed.

**Tech Stack:** tmux format strings, POSIX sh, Nerd Font glyphs (Powerline: ``, ``), BlexMono Nerd Font Mono

---

### Task 1: Create alert-sessions.sh

**Files:**
- Create: `tmux/.tmux/scripts/alert-sessions.sh`
- Delete: `tmux/.tmux/scripts/session-info.sh`

**Step 1: Write the new script**

Create `tmux/.tmux/scripts/alert-sessions.sh` with this exact content:

```sh
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
```

**Step 2: Make it executable**

```bash
chmod +x tmux/.tmux/scripts/alert-sessions.sh
```

**Step 3: Smoke-test the script manually**

Run it from the shell (outside tmux first, then inside a tmux session):

```bash
~/.tmux/scripts/alert-sessions.sh "dotfiles"
```

Expected when no other sessions have alerts: no output (empty string).
Expected when another session has an alert: `#[fg=colour214] 󰂚 session-name `

**Step 4: Delete the old script**

```bash
rm tmux/.tmux/scripts/session-info.sh
```

**Step 5: Commit**

```bash
git add tmux/.tmux/scripts/alert-sessions.sh tmux/.tmux/scripts/session-info.sh
git commit -m "tmux: replace session-info with alert-sessions script"
```

---

### Task 2: Rewrite status bar config in tmux.conf

**Files:**
- Modify: `tmux/.tmux.conf` (lines 28–37, 109–110, 121–122)

**Step 1: Replace the status bar section**

In `tmux/.tmux.conf`, replace everything from the `# --- Srcery Color Palette ---` comment through the end of the file (lines 25–123) with the following:

```tmux
# --- Srcery Color Palette ---
# Black: colour234  Grey: colour239  White: colour223  Yellow: colour214  Green: colour10

# Single status bar
set -g status 1
set-option -g status-style "bg=colour234,fg=colour223"
set-option -g status-position bottom

# Left: session name pill with Powerline separator
set-option -g status-left-length 40
set-option -g status-left "#[bg=colour214,fg=colour234,bold]  #S #[fg=colour214,bg=colour234,nobold]"

# Right: alert badges (from script) + time pill
set-option -g status-right-length 120
set-option -g status-right "#($HOME/.tmux/scripts/alert-sessions.sh '#{session_name}')#[fg=colour239,bg=colour234]#[bg=colour239,fg=colour223]  %H:%M "

# Window tab styling
setw -g window-status-separator ""
setw -g window-status-format "#[fg=colour244,bg=colour234] #I:#W "
setw -g window-status-current-format "#[fg=colour214,bg=colour10,bold]#[fg=colour234,bg=colour10,bold] #I:#W #[fg=colour10,bg=colour234,nobold]"

# Activity / Bell styling
set -g monitor-activity on
set -g visual-activity off
set -g bell-action any

setw -g window-status-activity-style "fg=colour214,bg=colour234,bold,underline"
setw -g window-status-bell-style "fg=colour1,bg=colour234,bold"
```

Key changes from old config:
- `status 2` → `status 1` (removes second bar)
- `status-format[1]` line removed entirely
- `status-right` now calls `alert-sessions.sh` instead of `session-info.sh`, drops hostname
- Active window gets a small yellow Powerline entry separator (colour214 → colour10) for contrast
- `window-status-separator ""` prevents double-spacing between tabs

**Step 2: Reload and visually verify**

Inside a running tmux session:

```bash
tmux source-file ~/.tmux.conf
```

Check:
- [ ] Only one status bar (no second row)
- [ ] Session name pill is yellow with Powerline `▶` separator
- [ ] Active window is green with Powerline arrows
- [ ] Inactive windows are dim grey
- [ ] Right side shows only time (no hostname)
- [ ] If you trigger an alert in another session, a `󰂚 name` badge appears before the time

**Step 3: Validate config syntax**

```bash
tmux -f tmux/.tmux.conf new-session -d -s test-validate && tmux kill-session -t test-validate && echo "OK"
```

Expected: `OK`

**Step 4: Commit**

```bash
git add tmux/.tmux.conf
git commit -m "tmux: redesign status bar as single Srcery Powerline bar"
```
