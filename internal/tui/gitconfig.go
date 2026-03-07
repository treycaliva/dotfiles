package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/treycaliva/dotfiles/internal/gitconfig"
)

type gitStep int

const (
	gitStepExisting gitStep = iota
	gitStepName
	gitStepEmail
	gitStepConfirm
)

// GitConfigScreen collects Git identity before git is stowed.
type GitConfigScreen struct {
	state  *AppState
	step   gitStep
	name   textinput.Model
	email  textinput.Model
	width  int
	height int
}

func NewGitConfigScreen(state *AppState) *GitConfigScreen {
	name := textinput.New()
	name.Placeholder = "Your Name"
	name.CharLimit = 128

	email := textinput.New()
	email.Placeholder = "your.email@example.com"
	email.CharLimit = 128

	screen := &GitConfigScreen{
		state: state,
		step:  gitStepName,
		name:  name,
		email: email,
	}

	// Load existing if available
	if setup, _ := gitconfig.ReadExistingSetup(state.HomeDir); setup != nil {
		screen.name.SetValue(setup.Name)
		screen.email.SetValue(setup.Email)
		screen.step = gitStepExisting
	}

	return screen
}

func (g *GitConfigScreen) Init() tea.Cmd { return nil }

func (g *GitConfigScreen) SetSize(w, h int) {
	g.width = w
	g.height = h
}

func (g *GitConfigScreen) StatusBar() []KeyBinding {
	switch g.step {
	case gitStepExisting:
		return []KeyBinding{{Key: "enter", Help: "keep"}, {Key: "e", Help: "edit"}, {Key: "esc", Help: "back"}}
	case gitStepConfirm:
		return []KeyBinding{{Key: "enter", Help: "install"}, {Key: "esc", Help: "back"}}
	default:
		return []KeyBinding{{Key: "enter", Help: "next"}, {Key: "esc", Help: "back"}}
	}
}

func (g *GitConfigScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			g.state.GitConfig = nil
			return g, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		case "e", "E":
			if g.step == gitStepExisting {
				g.step = gitStepName
				g.name.Focus()
				return g, wrapV1Cmd(textinput.Blink)
			}
		case "enter":
			return g.advance()
		}
	}

	var v1msg v1tea.Msg = msg
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		v1msg = toV1KeyMsg(m)
	}

	var v1cmd v1tea.Cmd
	switch g.step {
	case gitStepName:
		g.name, v1cmd = g.name.Update(v1msg)
	case gitStepEmail:
		g.email, v1cmd = g.email.Update(v1msg)
	}

	return g, wrapV1Cmd(v1cmd)
}

func (g *GitConfigScreen) advance() (ScreenModel, tea.Cmd) {
	switch g.step {
	case gitStepExisting:
		g.step = gitStepConfirm
		return g, nil
	case gitStepName:
		if strings.TrimSpace(g.name.Value()) == "" {
			return g, nil
		}
		g.step = gitStepEmail
		g.email.Focus()
		return g, wrapV1Cmd(textinput.Blink)
	case gitStepEmail:
		if strings.TrimSpace(g.email.Value()) == "" {
			return g, nil
		}
		g.step = gitStepConfirm
	case gitStepConfirm:
		g.state.GitConfig = &gitconfig.Setup{
			Name:  strings.TrimSpace(g.name.Value()),
			Email: strings.TrimSpace(g.email.Value()),
		}
		
		// If direnv is also selected, go there next. Otherwise Progress.
		nextScreen := ScreenProgress
		for _, pkg := range g.state.Selected {
			if pkg == "direnv" {
				nextScreen = ScreenDirenvConfig
				break
			}
		}
		return g, func() tea.Msg { return NavigateMsg{Screen: nextScreen} }
	}
	return g, nil
}

func (g *GitConfigScreen) View() tea.View {
	var b strings.Builder
	b.WriteString("\n")

	label := lipgloss.NewStyle().Bold(true).Foreground(Theme.Cyan)
	dim := Styles.Dimmed

	switch g.step {
	case gitStepExisting:
		b.WriteString("  " + label.Render("Existing Git Identity Found") + "\n\n")
		b.WriteString(fmt.Sprintf("  Name:  %s\n", Styles.Success.Render(g.name.Value())))
		b.WriteString(fmt.Sprintf("  Email: %s\n", Styles.Success.Render(g.email.Value())))
		b.WriteString("\n  " + dim.Render("enter: keep and proceed   e: edit identity") + "\n")

	case gitStepName:
		b.WriteString("  " + label.Render("Git Name") + "\n")
		b.WriteString("  " + dim.Render("Your full name for git commits") + "\n\n")
		b.WriteString("  " + g.name.View() + "\n")

	case gitStepEmail:
		b.WriteString("  " + label.Render("Git Email") + "\n")
		b.WriteString("  " + dim.Render("Your email address for git commits") + "\n\n")
		b.WriteString("  " + g.email.View() + "\n")

	case gitStepConfirm:
		b.WriteString("  " + label.Render("Ready to configure Git") + "\n\n")
		b.WriteString(fmt.Sprintf("  Name:  %s\n", Styles.Success.Render(g.name.Value())))
		b.WriteString(fmt.Sprintf("  Email: %s\n", Styles.Success.Render(g.email.Value())))
		b.WriteString("\n  " + dim.Render("Writes ~/.gitconfig.local") + "\n")
	}

	return tea.NewView(b.String())
}
