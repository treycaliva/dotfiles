package tui

import (
	"fmt"
	"strings"

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
	return []KeyBinding{}
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

	b.WriteString(Styles.Title.Render("  dotfiles installer"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  OS:         %s (%s)\n", h.state.Platform.OS, h.state.Platform.PkgManager))
	if h.state.Platform.IsWSL {
		b.WriteString("  WSL:        yes\n")
	}
	b.WriteString(fmt.Sprintf("  Dotfiles:   %s\n", h.state.DotfilesDir))
	b.WriteString("\n")

	b.WriteString(Styles.Title.Render("  Packages"))
	b.WriteString("\n\n")

	for _, name := range h.state.Config.PackageNames() {
		pkg := h.state.Config.Packages[name]
		var status string
		if h.state.StowStatus[name] {
			status = Icons.Success + " installed"
		} else {
			status = Icons.Warning + " not installed"
		}
		b.WriteString(fmt.Sprintf("  %-12s %s  %s\n", name, status, Styles.StatusBar.Render(pkg.Description)))
	}

	return tea.NewView(b.String())
}
