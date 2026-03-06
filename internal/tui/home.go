package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "charm.land/bubbletea/v2"
)

type HomeScreen struct {
	state  *AppState
	width  int
	height int
}

func NewHomeScreen(state *AppState) *HomeScreen {
	return &HomeScreen{state: state}
}

func (h *HomeScreen) Init() tea.Cmd { return nil }

func (h *HomeScreen) SetSize(w, height int) {
	if w < 10 {
		w = 10
	}
	if height < 3 {
		height = 3
	}
	h.width = w
	h.height = height
}

func (h *HomeScreen) StatusBar() []KeyBinding {
	return []KeyBinding{
		{Key: "enter", Help: "select packages"},
	}
}

func (h *HomeScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return h, func() tea.Msg {
				return NavigateMsg{Screen: ScreenSelect}
			}
		}
	}
	return h, nil
}

func (h *HomeScreen) View() tea.View {
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
		nameW := lipgloss.Width(nameStr)
		pillW := lipgloss.Width(pill)
		descW := lipgloss.Width(desc)
		available := h.width - nameW - pillW - 2
		if available < 0 {
			available = 0
		}
		pad := available - descW
		if pad < 0 {
			pad = 0
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			nameStr,
			desc+strings.Repeat(" ", pad),
			" ",
			pill,
		)
		b.WriteString(row + "\n")
	}

	return tea.NewView(b.String())
}
