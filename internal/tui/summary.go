package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type SummaryScreen struct {
	state *AppState
}

func NewSummaryScreen(state *AppState) *SummaryScreen {
	return &SummaryScreen{state: state}
}

func (s *SummaryScreen) Init() tea.Cmd { return nil }

func (s *SummaryScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return s, tea.Quit
		case "r":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenHome} }
		}
	}
	return s, nil
}

func (s *SummaryScreen) View() tea.View {
	var b strings.Builder

	b.WriteString(Styles.Title.Render("  Summary"))
	b.WriteString("\n\n")

	succeeded := 0
	failed := 0
	for _, pkg := range s.state.Selected {
		err := s.state.Results[pkg]
		if err == nil {
			b.WriteString(fmt.Sprintf("  %s %s\n", Icons.Success, pkg))
			succeeded++
		} else {
			b.WriteString(fmt.Sprintf("  %s %s — %v\n", Icons.Failure, pkg, err))
			failed++
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Succeeded: %s\n", Styles.Success.Render(fmt.Sprintf("%d", succeeded))))
	if failed > 0 {
		b.WriteString(fmt.Sprintf("  Failed:    %s\n", Styles.Error.Render(fmt.Sprintf("%d", failed))))
	}

	if len(s.state.Backups) > 0 {
		b.WriteString("\n")
		b.WriteString(Styles.Title.Render("  Backed up files"))
		b.WriteString("\n")
		for _, bak := range s.state.Backups {
			b.WriteString(fmt.Sprintf("    %s\n", bak))
		}
	}

	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  q: quit  r: start over  "))
	b.WriteString("\n")

	return tea.NewView(b.String())
}
