# BubbleTea v2 Upgrade & UI Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate the dotfiles installer TUI from BubbleTea v1 to v2, add a persistent chrome (header + footer), and overhaul every screen with full-width layout, visual hierarchy, and an enriched progress experience.

**Architecture:** `App` becomes a chrome compositor — it owns a permanent header and footer, and renders each `ScreenModel` into the content area between them. Every screen implements `SetSize(w, h int)` and `StatusBar() []KeyBinding` so the chrome can size content correctly and render context-aware keybindings in the footer.

**Tech Stack:** Go 1.24, `github.com/charmbracelet/bubbletea/v2`, `github.com/charmbracelet/bubbles v1` (unchanged), `github.com/charmbracelet/lipgloss v1.1.0`

**Design doc:** `docs/plans/2026-03-05-bubbletea-v2-upgrade-design.md`

---

## Task 1: Upgrade bubbletea module to v2

**Files:**
- Modify: `go.mod`
- Modify: `go.sum` (auto-generated)

**Step 1: Update the module requirement**

Run:
```bash
go get github.com/charmbracelet/bubbletea/v2@latest
go mod tidy
```

Expected: `go.mod` now contains `github.com/charmbracelet/bubbletea/v2 v2.x.x`. No build errors yet — imports still reference old path.

**Step 2: Verify old import still present (to confirm step 3 is needed)**

Run:
```bash
grep -r "charmbracelet/bubbletea\"" internal/
```

Expected: Several matches across `tui/` files.

**Step 3: Replace all import paths**

Run:
```bash
find internal/ cmd/ -name "*.go" | xargs sed -i '' 's|github.com/charmbracelet/bubbletea"|github.com/charmbracelet/bubbletea/v2"|g'
```

**Step 4: Verify build compiles**

Run:
```bash
go build ./...
```

Expected: Compile errors about `tea.KeyMsg` undefined — that's correct, proceed to Task 2.

**Step 5: Commit**

```bash
git add go.mod go.sum $(git diff --name-only)
git commit -m "deps: upgrade bubbletea to v2 import path"
```

---

## Task 2: Migrate KeyMsg → KeyPressMsg

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/home.go`
- Modify: `internal/tui/select.go`
- Modify: `internal/tui/preview.go`
- Modify: `internal/tui/diffview.go`
- Modify: `internal/tui/progress.go`
- Modify: `internal/tui/summary.go`

**Step 1: Replace all occurrences**

Run:
```bash
find internal/tui -name "*.go" | xargs sed -i '' 's/tea\.KeyMsg/tea.KeyPressMsg/g'
```

**Step 2: Verify build compiles cleanly**

Run:
```bash
go build ./...
```

Expected: Clean build. If any errors remain, check for `case tea.KeyMsg:` patterns that weren't caught.

**Step 3: Run existing tests**

Run:
```bash
go test ./...
```

Expected: All existing tests pass (theme_test.go, config_test.go, etc.).

**Step 4: Commit**

```bash
git add -p
git commit -m "tui: migrate tea.KeyMsg to tea.KeyPressMsg for bubbletea v2"
```

---

## Task 3: Expand ScreenModel interface and add stubs

**Files:**
- Modify: `internal/tui/app.go` (interface definition)
- Modify: `internal/tui/home.go`
- Modify: `internal/tui/select.go`
- Modify: `internal/tui/preview.go`
- Modify: `internal/tui/diffview.go`
- Modify: `internal/tui/progress.go`
- Modify: `internal/tui/summary.go`

**Step 1: Update the interface in app.go**

Replace the `ScreenModel` interface:
```go
type ScreenModel interface {
    Init() tea.Cmd
    Update(tea.Msg) (ScreenModel, tea.Cmd)
    View() string
    SetSize(w, h int)
    StatusBar() []KeyBinding
}
```

Also add `contentW` and `contentH` fields to `App`:
```go
type App struct {
    state    *AppState
    screen   Screen
    current  ScreenModel
    width    int
    height   int
    showHelp bool
    contentW int
    contentH int
}
```

Define chrome height constants at the top of `app.go`:
```go
const (
    chromeHeaderLines = 3
    chromeFooterLines = 1
)
```

**Step 2: Add stub methods to every screen**

Add these methods to each screen struct (`HomeScreen`, `SelectScreen`, `PreviewScreen`, `DiffScreen`, `ProgressScreen`, `SummaryScreen`):

```go
func (x *XScreen) SetSize(w, h int) {
    if w < 10 { w = 10 }
    if h < 3  { h = 3  }
    x.width = w
    x.height = h
}

func (x *XScreen) StatusBar() []KeyBinding {
    return []KeyBinding{} // placeholder, filled in per-screen tasks
}
```

Add `width` and `height` fields to each screen struct.

**Step 3: Update App.Update to call SetSize on WindowSizeMsg**

In `app.go` `Update()`, update the `tea.WindowSizeMsg` case:
```go
case tea.WindowSizeMsg:
    a.width = msg.Width
    a.height = msg.Height
    a.contentW = msg.Width
    a.contentH = msg.Height - chromeHeaderLines - chromeFooterLines
    if a.contentH < 3 {
        a.contentH = 3
    }
    a.current.SetSize(a.contentW, a.contentH)
```

**Step 4: Update navigate() to call SetSize**

In `navigate()`, after setting `a.current`, add:
```go
a.current.SetSize(a.contentW, a.contentH)
```

**Step 5: Verify build**

```bash
go build ./...
```

Expected: Clean build.

**Step 6: Commit**

```bash
git add -p
git commit -m "tui: expand ScreenModel interface with SetSize and StatusBar"
```

---

## Task 4: Expand theme with new styles and gradient helper

**Files:**
- Modify: `internal/tui/theme.go`
- Modify: `internal/tui/theme_test.go`

**Step 1: Write failing tests for the gradient helper**

Add to `theme_test.go`:
```go
func TestGradientTitle(t *testing.T) {
    result := GradientTitle("hi")
    if result == "" {
        t.Fatal("expected non-empty gradient string")
    }
    // Should have at least as many characters as input (ANSI codes wrap each rune)
    if len(result) < len("hi") {
        t.Errorf("gradient result too short: %q", result)
    }
}

func TestGradientTitleEmpty(t *testing.T) {
    if GradientTitle("") != "" {
        t.Fatal("expected empty string for empty input")
    }
}

func TestStylesHasNewFields(t *testing.T) {
    _ = Styles.Pill
    _ = Styles.PillSuccess
    _ = Styles.PillWarning
    _ = Styles.PillError
    _ = Styles.AccentBorderSuccess
    _ = Styles.AccentBorderWarning
    _ = Styles.AccentBorderError
    _ = Styles.HighlightRow
    _ = Styles.Dimmed
    _ = Styles.Header
    _ = Styles.Breadcrumb
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/tui/... -run TestGradient -v
go test ./internal/tui/... -run TestStylesHasNewFields -v
```

Expected: Compile error — `GradientTitle`, `Styles.Pill`, etc. undefined.

**Step 3: Implement in theme.go**

Add `GradientTitle` function and new style fields. Full implementation:

```go
// GradientTitle renders text with a per-character foreground color interpolated
// from Srcery Yellow (#FBB829) to Srcery Cyan (#0AAEB3).
func GradientTitle(s string) string {
    runes := []rune(s)
    if len(runes) == 0 {
        return ""
    }
    // Start: Yellow #FBB829, End: Cyan #0AAEB3
    sr, sg, sb := 0xFB, 0xB8, 0x29
    er, eg, eb := 0x0A, 0xAE, 0xB3
    var b strings.Builder
    n := len(runes)
    for i, r := range runes {
        t := 0.0
        if n > 1 {
            t = float64(i) / float64(n-1)
        }
        ri := sr + int(float64(er-sr)*t)
        gi := sg + int(float64(eg-sg)*t)
        bi := sb + int(float64(eb-sb)*t)
        color := lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", ri, gi, bi))
        b.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(r)))
    }
    return b.String()
}
```

Add new style fields to the `styles` struct and `Styles` var:

```go
type styles struct {
    // existing fields...
    Title     lipgloss.Style
    StatusBar lipgloss.Style
    Success   lipgloss.Style
    Error     lipgloss.Style
    Warning   lipgloss.Style
    Selected  lipgloss.Style
    Border    lipgloss.Style
    DiffAdd   lipgloss.Style
    DiffDel   lipgloss.Style
    // new fields:
    Pill              lipgloss.Style
    PillSuccess       lipgloss.Style
    PillWarning       lipgloss.Style
    PillError         lipgloss.Style
    AccentBorderSuccess lipgloss.Style
    AccentBorderWarning lipgloss.Style
    AccentBorderError   lipgloss.Style
    HighlightRow      lipgloss.Style
    Dimmed            lipgloss.Style
    Header            lipgloss.Style
    Breadcrumb        lipgloss.Style
}

// In Styles var, add:
Pill: lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(0, 1),

PillSuccess: lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(Theme.Green).
    Foreground(Theme.Green).
    Padding(0, 1),

PillWarning: lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(Theme.Yellow).
    Foreground(Theme.Yellow).
    Padding(0, 1),

PillError: lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(Theme.Red).
    Foreground(Theme.Red).
    Padding(0, 1),

AccentBorderSuccess: lipgloss.NewStyle().
    Border(lipgloss.ThickBorder(), false, false, false, true).
    BorderForeground(Theme.Green).
    PaddingLeft(1),

AccentBorderWarning: lipgloss.NewStyle().
    Border(lipgloss.ThickBorder(), false, false, false, true).
    BorderForeground(Theme.Yellow).
    PaddingLeft(1),

AccentBorderError: lipgloss.NewStyle().
    Border(lipgloss.ThickBorder(), false, false, false, true).
    BorderForeground(Theme.Red).
    PaddingLeft(1),

HighlightRow: lipgloss.NewStyle().
    Background(Theme.BrightBlack).
    Bold(true),

Dimmed: lipgloss.NewStyle().
    Foreground(Theme.BrightBlack),

Header: lipgloss.NewStyle().
    Background(Theme.Black).
    Foreground(Theme.White).
    Bold(true),

Breadcrumb: lipgloss.NewStyle().
    Foreground(Theme.Cyan),
```

**Step 4: Add `"fmt"` and `"strings"` to theme.go imports if not present**

```go
import (
    "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
)
```

**Step 5: Run tests**

```bash
go test ./internal/tui/... -v
```

Expected: All pass including new gradient and styles tests.

**Step 6: Commit**

```bash
git add internal/tui/theme.go internal/tui/theme_test.go
git commit -m "tui: add gradient helper and expanded style set for redesign"
```

---

## Task 5: Implement persistent chrome in App

**Files:**
- Modify: `internal/tui/app.go`
- Delete: `internal/tui/statusbar.go` (logic absorbed into App)

**Step 1: Add chrome rendering helpers to app.go**

Add these private methods to `App`:

```go
// renderHeader builds the 3-line persistent header.
func (a App) renderHeader() string {
    // Line 1: gradient title left, platform info right
    title := " " + GradientTitle(" dotfiles installer")
    platform := Styles.Dimmed.Render(a.state.Platform.OS + " · " + a.state.Platform.PkgManager)
    if a.state.Platform.IsWSL {
        platform = Styles.Dimmed.Render("WSL · " + a.state.Platform.PkgManager)
    }
    spacer := strings.Repeat(" ", max(0, a.width-lipgloss.Width(title)-lipgloss.Width(platform)-1))
    line1 := title + spacer + platform

    // Line 2: breadcrumb
    line2 := Styles.Breadcrumb.Render("  ▸ " + screenName(a.screen))

    // Line 3: separator
    line3 := Styles.Dimmed.Render(strings.Repeat("─", a.width))

    return lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
}

// renderFooter builds the 1-line persistent footer.
func (a App) renderFooter() string {
    bindings := a.current.StatusBar()
    // Always append help and quit
    bindings = append(bindings,
        KeyBinding{Key: "?", Help: "help"},
        KeyBinding{Key: "q", Help: "quit"},
    )
    var parts []string
    for _, b := range bindings {
        key := lipgloss.NewStyle().Bold(true).Foreground(Theme.Yellow).Render(b.Key)
        parts = append(parts, key+":"+b.Help)
    }
    content := "  " + strings.Join(parts, "  ") + "  "
    return lipgloss.NewStyle().
        Background(Theme.Black).
        Foreground(Theme.White).
        Width(a.width).
        Render(content)
}

func screenName(s Screen) string {
    switch s {
    case ScreenHome:     return "Home"
    case ScreenSelect:   return "Select Packages"
    case ScreenPreview:  return "Preview"
    case ScreenDiff:     return "Diff"
    case ScreenProgress: return "Installing"
    case ScreenSummary:  return "Summary"
    default:             return ""
    }
}

func max(a, b int) int {
    if a > b { return a }
    return b
}
```

**Step 2: Replace App.View() with chrome compositor**

```go
func (a App) View() string {
    if a.showHelp {
        help := Styles.Border.Padding(1, 2).Render(helpText())
        return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, help)
    }
    header := a.renderHeader()
    footer := a.renderFooter()
    content := a.current.View()
    return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}
```

**Step 3: Remove statusbar.go**

```bash
git rm internal/tui/statusbar.go
```

Update any screen that imports or calls `StatusBar(width, bindings)` directly — those will be replaced in per-screen tasks. For now, add a build-time check:

```bash
go build ./...
```

Expected: Errors only where `StatusBar(...)` function is called. Note the files and fix those calls by removing them (screens' old hardcoded status bar lines) — each screen task will replace them properly.

**Step 4: Add `"strings"` import to app.go if not present**

**Step 5: Verify build**

```bash
go build ./...
```

Expected: Clean build (screens' `StatusBar()` stubs return empty slices for now — footer renders with just `?:help q:quit`).

**Step 6: Commit**

```bash
git add internal/tui/app.go
git commit -m "tui: add persistent chrome compositor (header + footer) to App"
```

---

## Task 6: Redesign Home screen

**Files:**
- Modify: `internal/tui/home.go`

**Step 1: Update StatusBar() to return real bindings**

```go
func (h *HomeScreen) StatusBar() []KeyBinding {
    return []KeyBinding{
        {Key: "enter", Help: "select packages"},
    }
}
```

**Step 2: Rewrite View() for full-width table layout**

```go
func (h *HomeScreen) View() string {
    var b strings.Builder

    b.WriteString("\n")

    installed := 0
    for _, name := range h.state.Config.PackageNames() {
        if h.state.StowStatus[name] {
            installed++
        }
    }
    total := len(h.state.Config.PackageNames())
    countLine := Styles.Dimmed.Render(fmt.Sprintf("  %d packages · %d installed", total, installed))
    b.WriteString(countLine + "\n\n")

    for _, name := range h.state.Config.PackageNames() {
        pkg := h.state.Config.Packages[name]
        nameStr := fmt.Sprintf("  %-14s", name)

        var pill string
        if h.state.StowStatus[name] {
            pill = Styles.PillSuccess.Render("installed")
        } else {
            pill = Styles.PillWarning.Render("not installed")
        }

        desc := Styles.Dimmed.Render(pkg.Description)
        available := h.width - lipgloss.Width(nameStr) - lipgloss.Width(pill) - 2
        if available < 0 {
            available = 0
        }
        descPadded := desc + strings.Repeat(" ", max(0, available-lipgloss.Width(desc)))

        row := lipgloss.JoinHorizontal(lipgloss.Top,
            nameStr,
            descPadded,
            " ",
            pill,
        )
        b.WriteString(row + "\n")
    }

    return b.String()
}
```

**Step 3: Build and verify**

```bash
go build ./...
```

**Step 4: Commit**

```bash
git add internal/tui/home.go
git commit -m "tui: redesign home screen with full-width table and status pills"
```

---

## Task 7: Redesign Select screen

**Files:**
- Modify: `internal/tui/select.go`

**Step 1: Update StatusBar()**

```go
func (s *SelectScreen) StatusBar() []KeyBinding {
    return []KeyBinding{
        {Key: "j/k", Help: "move"},
        {Key: "space", Help: "toggle"},
        {Key: "enter", Help: "confirm"},
        {Key: "u", Help: "unstow mode"},
        {Key: "esc", Help: "back"},
    }
}
```

**Step 2: Rewrite View() with highlight-bar cursor and pill checkboxes**

```go
func (s *SelectScreen) View() string {
    var b strings.Builder
    b.WriteString("\n")

    for i, name := range s.packages {
        checked := "○ "
        if s.checked[i] {
            checked = Styles.Selected.Render("● ")
        }

        var statusIcon string
        if s.state.StowStatus[name] {
            statusIcon = Icons.Success + " "
        } else {
            statusIcon = Icons.Warning + " "
        }

        desc := Styles.Dimmed.Render(s.state.Config.Packages[name].Description)
        content := fmt.Sprintf("  %s%s%-14s %s", checked, statusIcon, name, desc)

        if s.cursor == i {
            // Pad to full width for the highlight bar
            padded := content + strings.Repeat(" ", max(0, s.width-lipgloss.Width(content)))
            b.WriteString(Styles.HighlightRow.Render(padded) + "\n")
        } else {
            b.WriteString(content + "\n")
        }
    }

    b.WriteString("\n")

    // Profile pills
    profiles := []struct{ key, name string }{
        {"m", "minimal"}, {"s", "server"}, {"f", "full"},
    }
    var pillParts []string
    for _, p := range profiles {
        pill := lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(Theme.Cyan).
            Foreground(Theme.Cyan).
            Padding(0, 1).
            Render(fmt.Sprintf("[%s] %s", p.key, p.name))
        pillParts = append(pillParts, pill)
    }
    b.WriteString("  " + strings.Join(pillParts, "  ") + "\n")

    return b.String()
}
```

**Step 3: Build and verify**

```bash
go build ./...
```

**Step 4: Commit**

```bash
git add internal/tui/select.go
git commit -m "tui: redesign select screen with highlight-bar cursor and pill checkboxes"
```

---

## Task 8: Redesign Preview screen

**Files:**
- Modify: `internal/tui/preview.go`

**Step 1: Add spinner field to PreviewScreen**

```go
import "github.com/charmbracelet/bubbles/spinner"

type PreviewScreen struct {
    state   *AppState
    items   []previewItem
    cursor  int
    loading bool
    width   int
    height  int
    spinner spinner.Model
}

func NewPreviewScreen(state *AppState) *PreviewScreen {
    s := spinner.New()
    s.Spinner = spinner.MiniDot
    s.Style = lipgloss.NewStyle().Foreground(Theme.Cyan)
    return &PreviewScreen{
        state:   state,
        loading: true,
        spinner: s,
    }
}
```

**Step 2: Update Init() to tick the spinner**

```go
func (p *PreviewScreen) Init() tea.Cmd {
    state := p.state
    return tea.Batch(p.spinner.Tick, func() tea.Msg {
        // existing analysis logic unchanged
        var items []previewItem
        for _, pkg := range state.Selected {
            // ... same as before
        }
        return previewReadyMsg{items: items}
    })
}
```

**Step 3: Handle spinner.TickMsg in Update()**

```go
case spinner.TickMsg:
    if p.loading {
        var cmd tea.Cmd
        p.spinner, cmd = p.spinner.Update(msg)
        return p, cmd
    }
    return p, nil
```

**Step 4: Update StatusBar()**

```go
func (p *PreviewScreen) StatusBar() []KeyBinding {
    if p.loading {
        return nil
    }
    return []KeyBinding{
        {Key: "j/k", Help: "move"},
        {Key: "d", Help: "view diff"},
        {Key: "enter", Help: "confirm"},
        {Key: "esc", Help: "back"},
    }
}
```

**Step 5: Rewrite View() with accent-border cards**

```go
func (p *PreviewScreen) View() string {
    var b strings.Builder
    b.WriteString("\n")

    if p.loading {
        b.WriteString(fmt.Sprintf("  %s Analyzing packages...\n", p.spinner.View()))
        return b.String()
    }

    for i, item := range p.items {
        hasIssues := len(item.missingDeps) > 0 || len(item.conflicts) > 0
        cursor := "  "
        if p.cursor == i {
            cursor = Styles.Selected.Render("▸ ")
        }

        var cardStyle lipgloss.Style
        var statusIcon string
        if len(item.conflicts) > 0 {
            cardStyle = Styles.AccentBorderError
            statusIcon = Icons.Failure
        } else if hasIssues {
            cardStyle = Styles.AccentBorderWarning
            statusIcon = Icons.Warning
        } else {
            cardStyle = Styles.AccentBorderSuccess
            statusIcon = Icons.Success
        }

        var cardLines strings.Builder
        cardLines.WriteString(fmt.Sprintf("%s%s %s\n", cursor, statusIcon, item.pkg))

        if len(item.missingDeps) > 0 {
            deps := make([]string, len(item.missingDeps))
            for j, d := range item.missingDeps {
                deps[j] = d.Binary
            }
            cardLines.WriteString(fmt.Sprintf("  deps to install: %s\n",
                Styles.Warning.Render(strings.Join(deps, ", "))))
        }

        if len(item.conflicts) > 0 {
            cardLines.WriteString(fmt.Sprintf("  %s\n",
                Styles.Error.Render(fmt.Sprintf("%d conflict(s)", len(item.conflicts)))))
            for _, c := range item.conflicts {
                cardLines.WriteString(fmt.Sprintf("  ┆ %s\n", Styles.Error.Render(c)))
            }
        }

        if !hasIssues {
            cardLines.WriteString("  " + Styles.Success.Render("ready") + "\n")
        }

        b.WriteString(cardStyle.Render(cardLines.String()) + "\n")
    }

    return b.String()
}
```

**Step 6: Build and verify**

```bash
go build ./...
```

**Step 7: Commit**

```bash
git add internal/tui/preview.go
git commit -m "tui: redesign preview screen with animated spinner and accent-border cards"
```

---

## Task 9: Redesign Progress screen

**Files:**
- Modify: `internal/tui/progress.go`

**Step 1: Add phase field to installStepMsg and import progress component**

```go
import "github.com/charmbracelet/bubbles/progress"

// Extend installStepMsg:
type installStepMsg struct {
    pkg   string
    log   string
    phase string // "installing deps" | "backing up" | "stowing" | "validating" | ""
    done  bool
    err   error
}
```

**Step 2: Emit phase labels from processNext()**

In the `processNext()` command function, before each major operation, return an intermediate message. The simplest approach: emit one `installStepMsg` per phase with `done: false, err: nil` and no log, just a phase label — then the final one with the full log.

Actually, simpler: pass phase via a `phaseMsg` struct and update `pkgProgress` to carry a `phase` string:

```go
type pkgProgress struct {
    name   string
    status pkgStatus
    phase  string
}
```

Update `installStepMsg` to set `phase` on the active package when the step message arrives. Emit separate step messages for each phase — or just set the phase in the final step message when the package completes. For a simpler implementation that still shows live phase updates, use a `tea.Cmd` that sends a phase update before starting each sub-step.

Simplest working approach — set phase in the final installStepMsg (shows last completed phase):

In `processNext()`, set the phase string based on what completed last:
```go
return installStepMsg{pkg: pkg, log: ..., phase: "stowing", done: ..., err: nil}
```

And emit intermediate `installPhaseMsg` structs for live updates:
```go
type installPhaseMsg struct {
    pkg   string
    phase string
}
```

Send these with `tea.Sequence` before the blocking work in `processNext()`:
```go
return tea.Sequence(
    func() tea.Msg { return installPhaseMsg{pkg: pkg, phase: "installing deps"} },
    // ... actual work command
)
```

Handle `installPhaseMsg` in `Update()` to update the active item's phase label.

**Step 3: Add progress.Model to ProgressScreen**

```go
type ProgressScreen struct {
    state    *AppState
    items    []pkgProgress
    current  int
    done     bool
    spinner  spinner.Model
    logView  viewport.Model
    allLogs  []string
    ready    bool
    width    int
    height   int
    progress progress.Model
}

func NewProgressScreen(state *AppState) *ProgressScreen {
    // ... existing init ...
    prog := progress.New(
        progress.WithDefaultGradient(),
        progress.WithWidth(40),
    )
    // Use Srcery colors:
    prog = progress.New(
        progress.WithScaledGradient("#FBB829", "#0AAEB3"),
        progress.WithWidth(40),
    )
    return &ProgressScreen{..., progress: prog}
}
```

**Step 4: Update Update() to handle progress bar and phase messages**

```go
case installStepMsg:
    // existing logic ...
    // Update progress bar
    cmd = tea.Batch(cmd, p.progress.SetPercent(float64(doneCount)/float64(len(p.items))))

case installPhaseMsg:
    for i := range p.items {
        if p.items[i].name == msg.pkg {
            p.items[i].phase = msg.phase
            break
        }
    }
    return p, nil
```

**Step 5: Rewrite View() with progress bar and phase labels**

```go
func (p *ProgressScreen) View() string {
    var b strings.Builder
    b.WriteString("\n")

    doneCount := 0
    for _, item := range p.items {
        if item.status == statusDone || item.status == statusFailed {
            doneCount++
        }
    }
    total := len(p.items)

    // Overall progress line
    countStr := Styles.Dimmed.Render(fmt.Sprintf("%d of %d complete", doneCount, total))
    progressLine := lipgloss.JoinHorizontal(lipgloss.Top,
        "  ",
        p.progress.View(),
        "  ",
        countStr,
    )
    b.WriteString(progressLine + "\n\n")

    // Per-package rows
    for _, item := range p.items {
        var icon string
        switch item.status {
        case statusPending:
            icon = Styles.Dimmed.Render("  ")
        case statusActive:
            icon = p.spinner.View()
        case statusDone:
            icon = Icons.Success
        case statusFailed:
            icon = Icons.Failure
        }

        row := fmt.Sprintf("  %s %-14s", icon, item.name)
        if item.status == statusActive && item.phase != "" {
            row += Styles.Dimmed.Render(item.phase+"...")
        } else if item.status == statusPending {
            row = Styles.Dimmed.Render(row + "pending")
        }
        b.WriteString(row + "\n")
    }

    b.WriteString("\n")

    // Log viewport
    if p.ready {
        bordered := Styles.Border.Width(p.width - 4).Render(p.logView.View())
        b.WriteString(bordered + "\n")
    }

    if p.done {
        b.WriteString("\n" + Styles.Success.Render("  All done!") + "\n")
    }

    return b.String()
}
```

**Step 6: Update StatusBar()**

```go
func (p *ProgressScreen) StatusBar() []KeyBinding {
    if p.done {
        return []KeyBinding{{Key: "enter", Help: "view summary"}}
    }
    return []KeyBinding{{Key: "j/k", Help: "scroll log"}}
}
```

**Step 7: Build and verify**

```bash
go build ./...
```

**Step 8: Commit**

```bash
git add internal/tui/progress.go
git commit -m "tui: redesign progress screen with progress bar and phase labels"
```

---

## Task 10: Redesign Summary screen

**Files:**
- Modify: `internal/tui/summary.go`

**Step 1: Update StatusBar()**

```go
func (s *SummaryScreen) StatusBar() []KeyBinding {
    return []KeyBinding{
        {Key: "r", Help: "start over"},
    }
}
```

**Step 2: Add backupsExpanded field**

```go
type SummaryScreen struct {
    state           *AppState
    width           int
    height          int
    backupsExpanded bool
}
```

Handle `"b"` key in `Update()`:
```go
case "b":
    s.backupsExpanded = !s.backupsExpanded
```

**Step 3: Rewrite View() with two-column layout, success banner, collapsible backups**

```go
func (s *SummaryScreen) View() string {
    var b strings.Builder
    b.WriteString("\n")

    var succeeded, failed []string
    for _, pkg := range s.state.Selected {
        if s.state.Results[pkg] == nil {
            succeeded = append(succeeded, pkg)
        } else {
            failed = append(failed, pkg)
        }
    }

    // All-success banner
    if len(failed) == 0 {
        banner := lipgloss.NewStyle().
            Background(Theme.Green).
            Foreground(Theme.Black).
            Bold(true).
            Width(s.width).
            Padding(0, 2).
            Render(fmt.Sprintf("  ✓  All %d packages installed successfully!", len(succeeded)))
        b.WriteString(banner + "\n\n")
    }

    // Score line
    scoreSuccess := Styles.Success.Bold(true).Render(fmt.Sprintf("✓ %d installed", len(succeeded)))
    scoreFail := ""
    if len(failed) > 0 {
        scoreFail = "  " + Styles.Error.Bold(true).Render(fmt.Sprintf("✗ %d failed", len(failed)))
    }
    b.WriteString("  " + scoreSuccess + scoreFail + "\n\n")

    // Two-column layout if wide enough
    if s.width >= 100 && len(failed) > 0 {
        colW := (s.width - 4) / 2
        var left, right strings.Builder
        left.WriteString(Styles.Success.Render("Installed") + "\n")
        for _, pkg := range succeeded {
            left.WriteString(fmt.Sprintf("  %s %s\n", Icons.Success, pkg))
        }
        right.WriteString(Styles.Error.Render("Failed") + "\n")
        for _, pkg := range failed {
            right.WriteString(fmt.Sprintf("  %s %s — %v\n", Icons.Failure, pkg, s.state.Results[pkg]))
        }
        cols := lipgloss.JoinHorizontal(lipgloss.Top,
            lipgloss.NewStyle().Width(colW).Render(left.String()),
            lipgloss.NewStyle().Width(colW).Render(right.String()),
        )
        b.WriteString(cols)
    } else {
        for _, pkg := range succeeded {
            b.WriteString(fmt.Sprintf("  %s %s\n", Icons.Success, pkg))
        }
        for _, pkg := range failed {
            b.WriteString(fmt.Sprintf("  %s %s — %v\n", Icons.Failure, pkg, s.state.Results[pkg]))
        }
    }

    // Collapsible backups
    if len(s.state.Backups) > 0 {
        b.WriteString("\n")
        if s.backupsExpanded {
            b.WriteString(Styles.Title.Render("  Backed up files") + "\n")
            for _, bak := range s.state.Backups {
                b.WriteString(fmt.Sprintf("  %s\n", Styles.Dimmed.Render(bak)))
            }
            b.WriteString(Styles.Dimmed.Render("  b: collapse") + "\n")
        } else {
            b.WriteString(fmt.Sprintf("  %s  %s\n",
                Styles.Warning.Render(fmt.Sprintf("↓ %d files backed up", len(s.state.Backups))),
                Styles.Dimmed.Render("— press b to expand")))
        }
    }

    return b.String()
}
```

**Step 4: Build and verify**

```bash
go build ./...
```

**Step 5: Commit**

```bash
git add internal/tui/summary.go
git commit -m "tui: redesign summary screen with banner, two-column layout, collapsible backups"
```

---

## Task 11: Update Diff screen for chrome compatibility

**Files:**
- Modify: `internal/tui/diffview.go`

**Step 1: Read the current diffview.go to understand its structure**

(Read the file before editing.)

**Step 2: Add SetSize, StatusBar, and wire viewport to content dimensions**

```go
func (d *DiffScreen) SetSize(w, h int) {
    if w < 10 { w = 10 }
    if h < 3  { h = 3  }
    d.width = w
    d.height = h
    if d.ready {
        d.viewport.Width = w - 4
        d.viewport.Height = h - 2
    }
}

func (d *DiffScreen) StatusBar() []KeyBinding {
    return []KeyBinding{
        {Key: "j/k", Help: "scroll"},
        {Key: "esc", Help: "back"},
    }
}
```

Add `width` and `height` fields to `DiffScreen`.

Update `WindowSizeMsg` handling in `DiffScreen.Update()` to use `d.width` and `d.height` instead of `msg.Width`/`msg.Height`.

**Step 3: Build and verify**

```bash
go build ./...
go test ./...
```

Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/tui/diffview.go
git commit -m "tui: update diff screen for chrome-aware sizing"
```

---

## Task 12: End-to-end smoke test

**Step 1: Build the binary**

```bash
go build -o /tmp/dotfiles-installer ./cmd/installer/
```

Expected: Binary produced, no errors.

**Step 2: Run at 80-column width**

```bash
COLUMNS=80 /tmp/dotfiles-installer
```

Manually verify:
- [ ] Header renders with gradient title and platform info
- [ ] Footer shows context-sensitive keybindings
- [ ] Home screen package rows don't overflow
- [ ] Select screen highlight bar spans full width
- [ ] No layout artifacts at narrow width

**Step 3: Run at 120-column width**

```bash
COLUMNS=120 /tmp/dotfiles-installer
```

Manually verify:
- [ ] Summary screen uses two-column layout when failed packages exist
- [ ] Progress bar scales to available width

**Step 4: Verify no panics on resize**

While running the app, resize the terminal rapidly. Verify no panics occur (SetSize guards handle 0-dimension edge cases).

**Step 5: Run all tests one final time**

```bash
go test ./...
```

Expected: All pass.

**Step 6: Final commit**

```bash
git add -p
git commit -m "tui: complete BubbleTea v2 upgrade and UI redesign"
```
