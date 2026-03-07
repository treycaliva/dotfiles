package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type SummaryScreen struct {
	state           *AppState
	width           int
	height          int
	backupsExpanded bool
}

func NewSummaryScreen(state *AppState) *SummaryScreen {
	return &SummaryScreen{state: state}
}

func (s *SummaryScreen) Init() tea.Cmd { return nil }

func (s *SummaryScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	s.width = w
	s.height = h
}

func (s *SummaryScreen) StatusBar() []KeyBinding {
	return []KeyBinding{
		{Key: "r", Help: "start over"},
	}
}

func (s *SummaryScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return s, tea.Quit
		case "r":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenHome} }
		case "b":
			s.backupsExpanded = !s.backupsExpanded
		}
	}
	return s, nil
}

func (s *SummaryScreen) View() tea.View {
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

	// Banner
	if len(failed) == 0 {
		banner := lipgloss.NewStyle().
			Background(Theme.Green).
			Foreground(Theme.Black).
			Bold(true).
			Width(s.width - 4).
			Padding(0, 2).
			Render(fmt.Sprintf("✓  All %d packages installed successfully!", len(succeeded)))
		b.WriteString("  " + banner + "\n\n")
	} else {
		banner := lipgloss.NewStyle().
			Background(Theme.Red).
			Foreground(Theme.Black).
			Bold(true).
			Width(s.width - 4).
			Padding(0, 2).
			Render(fmt.Sprintf("✗  %d of %d tasks failed — see details below", len(failed), len(s.state.Selected)))
		b.WriteString("  " + banner + "\n\n")
	}

	// Score line
	scoreSuccess := Styles.Success.Bold(true).Render(fmt.Sprintf("  %d installed", len(succeeded)))
	scoreFail := ""
	if len(failed) > 0 {
		scoreFail = "  " + Styles.Error.Bold(true).Render(fmt.Sprintf("  %d failed", len(failed)))
	}
	b.WriteString("  " + scoreSuccess + scoreFail + "\n\n")

	// Details section
	if len(failed) > 0 {
		b.WriteString("  " + Styles.Title.Render("Package Status") + "\n")
		for _, pkg := range s.state.Selected {
			err := s.state.Results[pkg]
			if err == nil {
				b.WriteString(fmt.Sprintf("    %s %s\n", Icons.Success, pkg))
			} else {
				// Detail error box
				boxW := s.width - 10
				if boxW < 20 { boxW = 20 }
				
				errText := fmt.Sprintf("%v", err)
				errorBox := Styles.AccentBorderError.
					Width(boxW).
					Render(fmt.Sprintf("%s\n%s", 
						Styles.Error.Bold(true).Render(pkg),
						Styles.Dimmed.Render(errText)))
				
				b.WriteString("    " + Icons.Failure + " " + errorBox + "\n")
			}
		}
	} else {
		// Clean list of succeeded packages
		for _, pkg := range succeeded {
			b.WriteString(fmt.Sprintf("    %s %s\n", Icons.Success, pkg))
		}
	}

	// Collapsible backups
	if len(s.state.Backups) > 0 {
		b.WriteString("\n")
		if s.backupsExpanded {
			b.WriteString("  " + Styles.Title.Render("Backed up files") + "\n")
			for _, bak := range s.state.Backups {
				b.WriteString(fmt.Sprintf("    %s\n", Styles.Dimmed.Render(bak)))
			}
			b.WriteString("    " + Styles.Dimmed.Render("b: collapse") + "\n")
		} else {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				Styles.Warning.Render(fmt.Sprintf("↓ %d files backed up", len(s.state.Backups))),
				Styles.Dimmed.Render("— press b to expand")))
		}
	}

	return tea.NewView(b.String())
}
