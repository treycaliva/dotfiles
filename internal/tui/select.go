package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "charm.land/bubbletea/v2"
)

type SelectScreen struct {
	state     *AppState
	cursor    int
	checked   map[int]bool
	unstowing bool
	packages  []string
	width     int
	height    int
}

func NewSelectScreen(state *AppState) *SelectScreen {
	return &SelectScreen{
		state:    state,
		checked:  make(map[int]bool),
		packages: state.Config.PackageNames(),
	}
}

func (s *SelectScreen) Init() tea.Cmd { return nil }

func (s *SelectScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	s.width = w
	s.height = h
}

func (s *SelectScreen) StatusBar() []KeyBinding {
	return []KeyBinding{
		{Key: "j/k", Help: "move"},
		{Key: "space", Help: "toggle"},
		{Key: "enter", Help: "confirm"},
		{Key: "u", Help: "unstow mode"},
		{Key: "esc", Help: "back"},
	}
}

func (s *SelectScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenHome} }
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.packages)-1 {
				s.cursor++
			}
		case "tab", " ":
			s.checked[s.cursor] = !s.checked[s.cursor]
		case "m":
			s.applyProfile("minimal")
		case "s":
			s.applyProfile("server")
		case "f":
			s.applyProfile("full")
		case "u":
			s.unstowing = !s.unstowing
		case "a":
			if s.checkedCount() == len(s.packages) {
				s.checked = make(map[int]bool)
			} else {
				for i := range s.packages {
					s.checked[i] = true
				}
			}
		case "enter":
			var selected []string
			for i, name := range s.packages {
				if s.checked[i] {
					selected = append(selected, name)
				}
			}
			if len(selected) == 0 {
				return s, nil
			}
			s.state.Selected = selected
			s.state.Unstowing = s.unstowing
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		}
	}
	return s, nil
}

// checkedCount returns how many packages are currently selected.
func (s *SelectScreen) checkedCount() int {
	n := 0
	for _, v := range s.checked {
		if v {
			n++
		}
	}
	return n
}

func (s *SelectScreen) applyProfile(name string) {
	profile, ok := s.state.Config.Profiles[name]
	if !ok {
		return
	}
	s.checked = make(map[int]bool)
	profileSet := make(map[string]bool)
	for _, pkg := range profile.Packages {
		profileSet[pkg] = true
	}
	for i, pkg := range s.packages {
		if profileSet[pkg] {
			s.checked[i] = true
		}
	}
}

func (s *SelectScreen) View() tea.View {
	var b strings.Builder
	b.WriteString("\n")

	for i, name := range s.packages {
		isCursor := s.cursor == i

		var checked string
		if s.checked[i] {
			checked = Styles.Selected.Render("● ")
		} else if isCursor {
			// Use white so ○ is visible against the BrightBlack highlight background
			checked = lipgloss.NewStyle().Foreground(Theme.White).Render("○ ")
		} else {
			checked = Styles.Dimmed.Render("○ ")
		}

		var statusIcon string
		if s.state.StowStatus[name] {
			statusIcon = Icons.Success + " "
		} else {
			statusIcon = Icons.Warning + " "
		}

		var desc string
		if isCursor {
			desc = lipgloss.NewStyle().Foreground(Theme.White).Render(s.state.Config.Packages[name].Description)
		} else {
			desc = Styles.Dimmed.Render(s.state.Config.Packages[name].Description)
		}
		content := fmt.Sprintf("  %s%s%-14s %s", checked, statusIcon, name, desc)

		if isCursor {
			contentW := lipgloss.Width(content)
			pad := s.width - contentW
			if pad < 0 {
				pad = 0
			}
			padded := content + strings.Repeat(" ", pad)
			b.WriteString(Styles.HighlightRow.Render(padded) + "\n")
		} else {
			b.WriteString(content + "\n")
		}
	}

	b.WriteString("\n")

	// Profile pills
	profiles := []struct{ key, label string }{
		{"m", "minimal"}, {"s", "server"}, {"f", "full"},
	}
	var pillParts []string
	for _, p := range profiles {
		pill := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Cyan).
			Foreground(Theme.Cyan).
			Padding(0, 1).
			Render(fmt.Sprintf("[%s] %s", p.key, p.label))
		pillParts = append(pillParts, pill)
	}
	b.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Top, pillParts...) + "\n")

	return tea.NewView(b.String())
}
