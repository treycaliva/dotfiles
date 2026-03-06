package tui

import (
	"fmt"
	"strings"

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
	return []KeyBinding{}
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

	mode := "Install"
	if s.unstowing {
		mode = "Unstow"
	}
	b.WriteString(Styles.Title.Render(fmt.Sprintf("  Select Packages (%s mode)", mode)))
	b.WriteString("\n\n")

	for i, name := range s.packages {
		cursor := "  "
		if s.cursor == i {
			cursor = Styles.Selected.Render("> ")
		}

		check := "[ ]"
		if s.checked[i] {
			check = Styles.Selected.Render("[x]")
		}

		var status string
		if s.state.StowStatus[name] {
			status = Icons.Success
		} else {
			status = Icons.Warning
		}

		desc := s.state.Config.Packages[name].Description
		b.WriteString(fmt.Sprintf("%s%s %s %-12s %s\n", cursor, check, status, name, desc))
	}

	b.WriteString("\n")
	b.WriteString("  Profiles: ")
	b.WriteString(Styles.Selected.Render("m") + "=minimal  ")
	b.WriteString(Styles.Selected.Render("s") + "=server  ")
	b.WriteString(Styles.Selected.Render("f") + "=full  ")
	b.WriteString(Styles.Selected.Render("a") + "=toggle all\n")
	return tea.NewView(b.String())
}
