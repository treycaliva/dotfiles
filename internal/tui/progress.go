package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"
	"github.com/treycaliva/dotfiles/internal/validate"
)

type pkgStatus int

const (
	statusPending pkgStatus = iota
	statusActive
	statusDone
	statusFailed
)

type pkgProgress struct {
	name   string
	status pkgStatus
}

// installStepMsg is sent after each package finishes processing.
type installStepMsg struct {
	pkg  string
	log  string
	done bool
	err  error
}

// ProgressScreen shows real-time installation progress with per-package
// spinners and a scrolling log viewport.
type ProgressScreen struct {
	state   *AppState
	items   []pkgProgress
	current int
	done    bool
	spinner spinner.Model
	logView viewport.Model
	allLogs []string
	ready   bool
	width   int
	height  int
}

func NewProgressScreen(state *AppState) *ProgressScreen {
	items := make([]pkgProgress, len(state.Selected))
	for i, pkg := range state.Selected {
		items[i] = pkgProgress{name: pkg, status: statusPending}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Theme.Yellow)

	if len(items) > 0 {
		items[0].status = statusActive
	}

	return &ProgressScreen{
		state:   state,
		items:   items,
		spinner: s,
	}
}

func (p *ProgressScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	p.width = w
	p.height = h
}

func (p *ProgressScreen) StatusBar() []KeyBinding {
	return []KeyBinding{}
}

func (p *ProgressScreen) Init() tea.Cmd {
	return tea.Batch(wrapV1Cmd(p.spinner.Tick), p.processNext())
}

// processNext returns a tea.Cmd that processes the next pending package.
func (p *ProgressScreen) processNext() tea.Cmd {
	if p.current >= len(p.items) {
		return nil
	}

	idx := p.current
	pkg := p.items[idx].name
	state := p.state

	return func() tea.Msg {
		var logs []string

		if state.Unstowing {
			logs = append(logs, fmt.Sprintf("[%s] unstowing...", pkg))
			result := stow.Unstow(state.DotfilesDir, state.HomeDir, pkg)
			if result.Output != "" {
				logs = append(logs, result.Output)
			}
			if result.Err != nil {
				logs = append(logs, fmt.Sprintf("[%s] error: %v", pkg, result.Err))
				return installStepMsg{
					pkg:  pkg,
					log:  strings.Join(logs, "\n"),
					done: idx >= len(state.Selected)-1,
					err:  result.Err,
				}
			}
			logs = append(logs, fmt.Sprintf("[%s] unstowed", pkg))
		} else {
			// Install mode: deps, conflicts, stow, validate
			cfg := state.Config.Packages[pkg]

			// Check and install missing deps
			if len(cfg.Deps) > 0 {
				statuses := platform.CheckDeps(cfg.Deps)
				for _, ds := range statuses {
					if !ds.Installed {
						logs = append(logs, fmt.Sprintf("[%s] installing dep: %s", pkg, ds.Binary))
						ir := platform.InstallDep(state.Platform.PkgManager, ds.Binary, cfg.PkgNames)
						if ir.Output != "" {
							logs = append(logs, ir.Output)
						}
						if ir.Err != nil {
							logs = append(logs, fmt.Sprintf("[%s] dep install failed: %s: %v", pkg, ds.Binary, ir.Err))
							return installStepMsg{
								pkg:  pkg,
								log:  strings.Join(logs, "\n"),
								done: idx >= len(state.Selected)-1,
								err:  fmt.Errorf("dep %s: %w", ds.Binary, ir.Err),
							}
						}
					}
				}
			}

			// Handle conflicts by backing up
			if conflicts, ok := state.Conflicts[pkg]; ok {
				for _, cf := range conflicts {
					logs = append(logs, fmt.Sprintf("[%s] backing up conflict: %s", pkg, cf))
					backupPath, err := stow.BackupConflict(state.HomeDir, cf)
					if err != nil {
						logs = append(logs, fmt.Sprintf("[%s] backup failed: %v", pkg, err))
						return installStepMsg{
							pkg:  pkg,
							log:  strings.Join(logs, "\n"),
							done: idx >= len(state.Selected)-1,
							err:  err,
						}
					}
					state.Backups = append(state.Backups, backupPath)
					logs = append(logs, fmt.Sprintf("[%s] backed up to %s", pkg, backupPath))
				}
			}

			// Stow the package
			logs = append(logs, fmt.Sprintf("[%s] stowing...", pkg))
			result := stow.Stow(state.DotfilesDir, state.HomeDir, pkg)
			if result.Output != "" {
				logs = append(logs, result.Output)
			}
			if result.Err != nil {
				logs = append(logs, fmt.Sprintf("[%s] stow failed: %v", pkg, result.Err))
				return installStepMsg{
					pkg:  pkg,
					log:  strings.Join(logs, "\n"),
					done: idx >= len(state.Selected)-1,
					err:  result.Err,
				}
			}
			logs = append(logs, fmt.Sprintf("[%s] stowed", pkg))

			// Run validation
			if cfg.Validate != "" {
				logs = append(logs, fmt.Sprintf("[%s] validating...", pkg))
				vr := validate.Run(pkg, cfg.Validate)
				if vr.Output != "" {
					logs = append(logs, vr.Output)
				}
				if vr.Err != nil {
					logs = append(logs, fmt.Sprintf("[%s] validation warning: %v", pkg, vr.Err))
				} else {
					logs = append(logs, fmt.Sprintf("[%s] validation passed", pkg))
				}
			}
		}

		return installStepMsg{
			pkg:  pkg,
			log:  strings.Join(logs, "\n"),
			done: idx >= len(state.Selected)-1,
			err:  nil,
		}
	}
}

func (p *ProgressScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if !p.done {
			var v1cmd v1tea.Cmd
			p.spinner, v1cmd = p.spinner.Update(msg)
			return p, wrapV1Cmd(v1cmd)
		}
		return p, nil

	case installStepMsg:
		// Update the item status
		for i := range p.items {
			if p.items[i].name == msg.pkg {
				if msg.err != nil {
					p.items[i].status = statusFailed
				} else {
					p.items[i].status = statusDone
				}
				break
			}
		}

		// Store result in shared state
		p.state.Results[msg.pkg] = msg.err

		// Add log output
		if msg.log != "" {
			p.allLogs = append(p.allLogs, msg.log)
			if p.ready {
				p.logView.SetContent(strings.Join(p.allLogs, "\n"))
				p.logView.GotoBottom()
			}
		}

		if msg.done {
			p.done = true
			return p, nil
		}

		// Advance to the next package
		p.current++
		if p.current < len(p.items) {
			p.items[p.current].status = statusActive
		}
		return p, p.processNext()

	case tea.WindowSizeMsg:
		headerHeight := 3 + len(p.items) + 1 // title + items + blank
		footerHeight := 3                     // blank + border padding + status bar
		height := msg.Height - headerHeight - footerHeight
		if height < 3 {
			height = 3
		}
		width := msg.Width - 4 // account for border
		if width < 10 {
			width = 10
		}
		if !p.ready {
			p.logView = viewport.New(width, height)
			p.ready = true
		} else {
			p.logView.Width = width
			p.logView.Height = height
		}
		if len(p.allLogs) > 0 {
			p.logView.SetContent(strings.Join(p.allLogs, "\n"))
			p.logView.GotoBottom()
		}
		return p, nil

	case tea.KeyPressMsg:
		if p.done && msg.String() == "enter" {
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenSummary} }
		}
		// Allow scrolling the log viewport
		if p.ready {
			var v1cmd v1tea.Cmd
			p.logView, v1cmd = p.logView.Update(msg)
			return p, wrapV1Cmd(v1cmd)
		}
	}

	return p, nil
}

func (p *ProgressScreen) View() tea.View {
	var b strings.Builder

	title := "Installing packages"
	if p.state.Unstowing {
		title = "Unstowing packages"
	}
	b.WriteString(Styles.Title.Render("  " + title))
	b.WriteString("\n\n")

	// Per-package progress rows
	for _, item := range p.items {
		var icon string
		switch item.status {
		case statusPending:
			icon = "  "
		case statusActive:
			icon = p.spinner.View()
		case statusDone:
			icon = Icons.Success
		case statusFailed:
			icon = Icons.Failure
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, item.name))
	}

	b.WriteString("\n")

	// Scrolling log viewport in a bordered box
	if p.ready {
		logContent := p.logView.View()
		bordered := Styles.Border.Render(logContent)
		b.WriteString(bordered)
		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}
