# BubbleTea v2 Upgrade & UI Redesign

**Date:** 2026-03-05
**Branch:** upgrade-bubbletea-v2
**Status:** Approved — ready for implementation

## Goals

1. Migrate from `bubbletea v1.3.10` to `bubbletea/v2`
2. Fix flat visual hierarchy — add clear separation between title, content, and controls
3. Fill the terminal — all screens use full width/height, no fixed-width padding
4. Enrich the progress experience — overall progress bar, per-package phase labels, animated spinners

## Non-Goals

- Migrating `bubbles` to v2 (not yet stable)
- Adding new installer functionality
- Changing the 6-screen navigation flow

---

## Architecture

### Module changes

| Package | Before | After |
|---|---|---|
| bubbletea | `github.com/charmbracelet/bubbletea v1.3.10` | `github.com/charmbracelet/bubbletea/v2` |
| bubbles | `github.com/charmbracelet/bubbles v1.0.0` | unchanged |
| lipgloss | `github.com/charmbracelet/lipgloss v1.1.0` | unchanged |

**Breaking API change:** `tea.KeyMsg` → `tea.KeyPressMsg` across all 6 screen files and `app.go`.

### Persistent chrome

`App.View()` becomes a compositor that renders three layers:

```
┌─────────────────────────────────────────────────────┐  header (3 lines)
│   dotfiles installer                   macOS · brew │
│  ▸ Select Packages                                  │
├─────────────────────────────────────────────────────┤
│                                                     │
│   [current screen fills this area]                  │  content area
│                                                     │
├─────────────────────────────────────────────────────┤
│  j/k: move  space: toggle  enter: confirm  ?: help  │  footer (1 line)
└─────────────────────────────────────────────────────┘
```

- **Header:** app title with gradient + right-aligned platform info + breadcrumb line
- **Content area:** `terminalHeight - headerHeight(3) - footerHeight(1)` lines tall, full width
- **Footer:** full-width status bar showing current screen's keybindings

### `ScreenModel` interface

Add two methods:

```go
type ScreenModel interface {
    Init() tea.Cmd
    Update(tea.Msg) (ScreenModel, tea.Cmd)
    View() string
    SetSize(w, h int)             // called by App on WindowSizeMsg and navigate()
    StatusBar() []KeyBinding      // screen declares its own keybindings for the footer
}
```

`App` stores `contentW` and `contentH` (terminal dimensions minus chrome). On `WindowSizeMsg` and on every `navigate()`, it calls `current.SetSize(contentW, contentH)`.

---

## Visual Design

### Color roles (Srcery palette)

| Element | Color | Hex |
|---|---|---|
| Title gradient start | Yellow | `#FBB829` |
| Title gradient end | Cyan | `#0AAEB3` |
| Screen breadcrumb | Cyan | `#0AAEB3` |
| Active cursor / selected | Cyan bold | `#0AAEB3` |
| Success | Green | `#519F50` |
| Error | Red | `#EF2F27` |
| Warning | Yellow | `#FBB829` |
| Borders / separators | BrightBlack | `#918175` |
| Footer background | Black | `#1C1B19` |
| Footer key labels | Yellow bold | `#FBB829` |
| Footer descriptions | White | `#BAA67F` |

### Title gradient

The app title `dotfiles installer` renders with per-character foreground color interpolated from Yellow (`#FBB829`) to Cyan (`#0AAEB3`) using lipgloss `lipgloss.Color` on each rune. A Nerd Font icon prefixes the title.

### Package status pills

Replace plain `installed` / `not installed` text with bordered pills:

- `[ installed ]` — Green border, Green text
- `[ not installed ]` — BrightBlack border, BrightBlack text
- `[ conflict ]` — Yellow border, Yellow text

### Full-width rows

All list rows use `lipgloss.JoinHorizontal(lipgloss.Top, leftContent, spacer, rightContent)` to push status badges to the right edge of the content width.

---

## Screen-by-Screen Changes

### Home screen

- Title replaced with gradient + Nerd Font icon
- Package list becomes a full-width table: name left, status pill right
- Summary count line: `7 packages · 4 installed` in BrightBlack below the list

### Select screen

- Cursor row renders as a full-width highlight bar (BrightBlack background) instead of `> ` prefix
- Checked: `●` Cyan; unchecked: `○` BrightBlack
- Profile shortcuts render as pill buttons: `[m] minimal`  `[s] server`  `[f] full`
- Mode (install/unstow) shown as a right-aligned badge in the App header chrome

### Preview screen

- Loading state uses bubbles `spinner.MiniDot` (animated) instead of static icon
- Each package renders as a mini-card with a left accent border colored by status:
  - Green border = ready
  - Yellow border = warnings (missing deps or conflicts)
  - Red border = blocking conflicts
- Conflict file list indents under the package with a `┆` gutter line

### Diff screen

- Unchanged functionally; receives chrome treatment and consistent border/padding

### Progress screen *(major redesign)*

```
  Installing packages                        3 of 7 complete
  ████████████████████░░░░░░░░░░░░░░░  42%

  ✓ zsh          installed
  ✓ tmux         installed
  ✓ git          installed
  ⠸ nvim         installing deps...
    vim           pending
    ghostty       pending
    p10k          pending

 ┌─ log ──────────────────────────────────────────────────┐
 │ [nvim] installing dep: neovim                          │
 │ [nvim] neovim installed                                │
 └────────────────────────────────────────────────────────┘
```

Changes:
- `bubbles/progress` bar shows overall completion (`N of M packages`)
- `installStepMsg` gains a `phase` field: `"installing deps"` / `"backing up"` / `"stowing"` / `"validating"`
- Active package row shows phase label in BrightBlack italic
- Pending packages shown in BrightBlack (dimmed)
- Scrolling log viewport below, in a rounded border box

### Summary screen

- Wide terminals (>100 cols): two-column layout — succeeded left, failed right
- Score line: `✓ 6 installed  ✗ 1 failed` rendered large with color
- All-success state: full-width Green success banner
- Backed-up files section: collapsed by default, `b` to expand (shows count hint when collapsed)

---

## File Change Map

| File | Change |
|---|---|
| `go.mod` | Upgrade bubbletea to v2 import path |
| `internal/tui/app.go` | Add chrome compositor, `contentW/H`, `KeyPressMsg`, call `SetSize` |
| `internal/tui/theme.go` | Add gradient helper, expand `Styles` (pill, accent border, highlight row) |
| `internal/tui/statusbar.go` | Remove (logic moves into `App.View()` footer layer) |
| `internal/tui/home.go` | `SetSize`, `StatusBar()`, full-width table, gradient title, pills |
| `internal/tui/select.go` | `SetSize`, `StatusBar()`, highlight-bar cursor, pill checkboxes, profile pills |
| `internal/tui/preview.go` | `SetSize`, `StatusBar()`, animated spinner, accent-border cards |
| `internal/tui/diffview.go` | `SetSize`, `StatusBar()`, minor border cleanup |
| `internal/tui/progress.go` | `SetSize`, `StatusBar()`, bubbles progress bar, phase labels, extended `installStepMsg` |
| `internal/tui/summary.go` | `SetSize`, `StatusBar()`, two-column layout, success banner, collapsible backups |

---

## Testing Approach

- Existing unit tests for `theme_test.go` — verify `Styles` fields still exist after expansion
- Manual smoke-test each screen at 80-col, 120-col, and 200-col widths
- Verify `SetSize(0, 0)` doesn't panic (guard with minimums)
- Run `go vet ./...` and `go build ./...` after v2 module bump to catch import errors
