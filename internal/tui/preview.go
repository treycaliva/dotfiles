package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	v1tea "github.com/charmbracelet/bubbletea"
	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"

	tea "charm.land/bubbletea/v2"
)

// previewItem holds analysis results for a single package.
type previewItem struct {
	pkg         string
	missingDeps []platform.DepStatus
	conflicts   []string
}

// previewReadyMsg is sent when async analysis completes.
type previewReadyMsg struct {
	items []previewItem
}

// PreviewScreen shows a dry-run analysis before installation.
type PreviewScreen struct {
	state   *AppState
	items   []previewItem
	cursor  int
	loading bool
	spinner spinner.Model
	width   int
	height  int
	flash   string
}

// clearFlashMsg clears the flash message after a short delay.
type clearFlashMsg struct{}

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

func (p *PreviewScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	p.width = w
	p.height = h
}

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

func (p *PreviewScreen) Init() tea.Cmd {
	state := p.state
	analyzeCmd := func() tea.Msg {
		var items []previewItem
		for _, pkg := range state.Selected {
			item := previewItem{pkg: pkg}
			cfg := state.Config.Packages[pkg]
			if len(cfg.Deps) > 0 {
				statuses := platform.CheckDeps(cfg.Deps)
				for _, s := range statuses {
					if !s.Installed {
						item.missingDeps = append(item.missingDeps, s)
					}
				}
			}
			if !state.Unstowing {
				conflicts, _, _ := stow.DryRun(state.DotfilesDir, state.HomeDir, pkg)
				item.conflicts = conflicts
			}
			items = append(items, item)
		}
		return previewReadyMsg{items: items}
	}
	return tea.Batch(wrapV1Cmd(p.spinner.Tick), analyzeCmd)
}

func (p *PreviewScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case previewReadyMsg:
		p.loading = false
		p.items = msg.items
		// Store conflicts in shared state for use during progress
		p.state.Conflicts = make(map[string][]string)
		for _, item := range p.items {
			if len(item.conflicts) > 0 {
				p.state.Conflicts[item.pkg] = item.conflicts
			}
		}
		return p, nil
	case tea.KeyPressMsg:
		if p.loading {
			return p, nil
		}
		switch msg.String() {
		case "esc":
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenSelect} }
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		case "d":
			if p.cursor < len(p.items) {
				item := p.items[p.cursor]
				if len(item.conflicts) > 0 {
					p.state.DiffPkg = item.pkg
					p.state.DiffFile = item.conflicts[0]
					return p, func() tea.Msg { return NavigateMsg{Screen: ScreenDiff} }
				}
				p.flash = fmt.Sprintf("No conflicts for %s", item.pkg)
				return p, tea.Tick(time.Second*2, func(time.Time) tea.Msg {
					return clearFlashMsg{}
				})
			}
		case "enter":
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
		}
	case clearFlashMsg:
		p.flash = ""
		return p, nil
	case spinner.TickMsg:
		if p.loading {
			var v1cmd v1tea.Cmd
			p.spinner, v1cmd = p.spinner.Update(msg)
			return p, wrapV1Cmd(v1cmd)
		}
		return p, nil
	}
	return p, nil
}

func (p *PreviewScreen) View() tea.View {
	var b strings.Builder
	b.WriteString("\n")

	if p.loading {
		b.WriteString(fmt.Sprintf("  %s Analyzing packages...\n", p.spinner.View()))
		return tea.NewView(b.String())
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

	if p.flash != "" {
		b.WriteString("\n  " + Styles.Dimmed.Render(p.flash) + "\n")
	}

	return tea.NewView(b.String())
}
