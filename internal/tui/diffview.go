package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"

	"github.com/treycaliva/dotfiles/internal/stow"
)

// diffContentMsg carries the raw diff output to the screen.
type diffContentMsg struct {
	content string
}

// DiffScreen displays a colored unified diff in a scrollable viewport.
type DiffScreen struct {
	state    *AppState
	viewport viewport.Model
	pkg      string
	file     string
	ready    bool
}

func NewDiffScreen(state *AppState, pkg, file string) *DiffScreen {
	return &DiffScreen{
		state: state,
		pkg:   pkg,
		file:  file,
	}
}

func (d *DiffScreen) Init() tea.Cmd {
	state := d.state
	pkg := d.pkg
	file := d.file
	return func() tea.Msg {
		diff, err := stow.DiffConflict(state.HomeDir, state.DotfilesDir, pkg, file)
		if err != nil {
			return diffContentMsg{content: fmt.Sprintf("Error generating diff: %v", err)}
		}
		if diff == "" {
			return diffContentMsg{content: "(files are identical)"}
		}
		return diffContentMsg{content: diff}
	}
}

func (d *DiffScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 3 // title + blank line
		footerHeight := 2 // status bar + blank line
		height := msg.Height - headerHeight - footerHeight
		if height < 1 {
			height = 1
		}
		if !d.ready {
			d.viewport = viewport.New(msg.Width, height)
			d.ready = true
		} else {
			d.viewport.Width = msg.Width
			d.viewport.Height = height
		}
		return d, nil
	case diffContentMsg:
		styled := d.styleDiff(msg.content)
		if d.ready {
			d.viewport.SetContent(styled)
		} else {
			// Store content; it will be set when viewport initializes
			d.viewport.SetContent(styled)
		}
		return d, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return d, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		}
	}

	if d.ready {
		var v1cmd v1tea.Cmd
		d.viewport, v1cmd = d.viewport.Update(msg)
		return d, wrapV1Cmd(v1cmd)
	}
	return d, nil
}

// styleDiff colors diff lines: green for additions, red for deletions,
// cyan for hunk headers.
func (d *DiffScreen) styleDiff(raw string) string {
	var b strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			b.WriteString(Styles.DiffAdd.Render(line))
		case strings.HasPrefix(line, "-"):
			b.WriteString(Styles.DiffDel.Render(line))
		case strings.HasPrefix(line, "@@"):
			b.WriteString(Styles.Selected.Render(line)) // cyan
		default:
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func (d *DiffScreen) View() tea.View {
	var b strings.Builder

	title := fmt.Sprintf("  Diff: %s — ~/%s", d.pkg, d.file)
	b.WriteString(Styles.Title.Render(title))
	b.WriteString("\n\n")

	if !d.ready {
		b.WriteString("  Loading...\n")
		return tea.NewView(b.String())
	}

	b.WriteString(d.viewport.View())
	b.WriteString("\n")

	info := fmt.Sprintf(" %3.f%% ", d.viewport.ScrollPercent()*100)
	b.WriteString(Styles.StatusBar.Render(fmt.Sprintf("  q/esc: back  %s", info)))
	b.WriteString("\n")

	return tea.NewView(b.String())
}
