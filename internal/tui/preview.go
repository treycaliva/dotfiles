package tui

import (
	"fmt"
	"strings"

	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"

	tea "github.com/charmbracelet/bubbletea"
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
}

func NewPreviewScreen(state *AppState) *PreviewScreen {
	return &PreviewScreen{
		state:   state,
		loading: true,
	}
}

func (p *PreviewScreen) Init() tea.Cmd {
	state := p.state
	return func() tea.Msg {
		var items []previewItem
		for _, pkg := range state.Selected {
			item := previewItem{pkg: pkg}

			// Check dependencies
			cfg := state.Config.Packages[pkg]
			if len(cfg.Deps) > 0 {
				statuses := platform.CheckDeps(cfg.Deps)
				for _, s := range statuses {
					if !s.Installed {
						item.missingDeps = append(item.missingDeps, s)
					}
				}
			}

			// Run stow dry-run to detect conflicts (only for install, not unstow)
			if !state.Unstowing {
				conflicts, _, _ := stow.DryRun(state.DotfilesDir, state.HomeDir, pkg)
				item.conflicts = conflicts
			}

			items = append(items, item)
		}
		return previewReadyMsg{items: items}
	}
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
	case tea.KeyMsg:
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
			}
		case "enter":
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
		}
	}
	return p, nil
}

func (p *PreviewScreen) View() string {
	var b strings.Builder

	action := "Install"
	if p.state.Unstowing {
		action = "Unstow"
	}
	b.WriteString(Styles.Title.Render(fmt.Sprintf("  Preview (%s)", action)))
	b.WriteString("\n\n")

	if p.loading {
		b.WriteString(fmt.Sprintf("  %s Analyzing packages...\n", Icons.Pending))
		return b.String()
	}

	for i, item := range p.items {
		cursor := "  "
		if p.cursor == i {
			cursor = Styles.Selected.Render("> ")
		}

		status := Icons.Success
		if len(item.missingDeps) > 0 || len(item.conflicts) > 0 {
			status = Icons.Warning
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, status, item.pkg))

		// Show missing deps
		if len(item.missingDeps) > 0 {
			deps := make([]string, len(item.missingDeps))
			for j, d := range item.missingDeps {
				deps[j] = d.Binary
			}
			b.WriteString(fmt.Sprintf("      deps to install: %s\n",
				Styles.Warning.Render(strings.Join(deps, ", "))))
		}

		// Show conflicts
		if len(item.conflicts) > 0 {
			b.WriteString(fmt.Sprintf("      conflicts: %s\n",
				Styles.Error.Render(fmt.Sprintf("%d file(s)", len(item.conflicts)))))
			for _, c := range item.conflicts {
				b.WriteString(fmt.Sprintf("        %s %s\n", Styles.Error.Render("-"), c))
			}
		}

		if len(item.missingDeps) == 0 && len(item.conflicts) == 0 {
			b.WriteString("      " + Styles.Success.Render("ready") + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  j/k: move  d: diff  enter: confirm  esc: back  "))
	b.WriteString("\n")

	return b.String()
}
