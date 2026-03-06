package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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
	width    int
	height   int
}

func NewDiffScreen(state *AppState, pkg, file string) *DiffScreen {
	return &DiffScreen{
		state: state,
		pkg:   pkg,
		file:  file,
	}
}

func (d *DiffScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	d.width = w
	d.height = h
	vw := w - 4
	vh := h - 2
	if vh < 3 {
		vh = 3
	}
	if !d.ready {
		d.viewport = viewport.New(vw, vh)
		d.ready = true
	} else {
		d.viewport.Width = vw
		d.viewport.Height = vh
	}
}

func (d *DiffScreen) StatusBar() []KeyBinding {
	return []KeyBinding{
		{Key: "j/k", Help: "scroll"},
		{Key: "esc", Help: "back"},
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
		// Sizing handled by App via SetSize()
		return d, nil
	case diffContentMsg:
		styled := d.styleDiff(msg.content)
		d.viewport.SetContent(styled)
		return d, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return d, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		case "down", "j":
			if d.ready {
				d.viewport.ScrollDown(1)
			}
		case "up", "k":
			if d.ready {
				d.viewport.ScrollUp(1)
			}
		case "d":
			if d.ready {
				d.viewport.HalfPageDown()
			}
		case "u":
			if d.ready {
				d.viewport.HalfPageUp()
			}
		}
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

	if !d.ready {
		b.WriteString("  Loading...\n")
		return tea.NewView(b.String())
	}

	b.WriteString(d.viewport.View())
	b.WriteString("\n")

	return tea.NewView(b.String())
}
