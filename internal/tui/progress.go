package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/treycaliva/dotfiles/internal/direnv"
	"github.com/treycaliva/dotfiles/internal/gitconfig"
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
	phase  string
}

// installPhaseMsg is sent before a major operation to give live phase feedback.
type installPhaseMsg struct {
	pkg   string
	phase string
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
	prog    progress.Model
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

	prog := progress.New(
		progress.WithScaledGradient("#FBB829", "#0AAEB3"),
		progress.WithWidth(40),
	)

	if len(items) > 0 {
		items[0].status = statusActive
	}

	return &ProgressScreen{
		state:   state,
		items:   items,
		spinner: s,
		prog:    prog,
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
	if p.ready {
		p.logView.Width = w - 4
		headerHeight := 2 + len(p.items) + 1
		logH := h - headerHeight - 2
		if logH < 3 {
			logH = 3
		}
		p.logView.Height = logH
	}
}

func (p *ProgressScreen) StatusBar() []KeyBinding {
	if p.done {
		return []KeyBinding{{Key: "enter", Help: "view summary"}}
	}
	return []KeyBinding{{Key: "j/k", Help: "scroll log"}}
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

			// Post-stow direnv configuration (writes ~/.zshrc.local, patches template, allows .envrc).
			if pkg == "direnv" && state.DirenvConfig != nil {
				logs = append(logs, fmt.Sprintf("[%s] writing ~/.zshrc.local ...", pkg))
				if err := direnv.WriteZshrcLocal(state.HomeDir, state.DirenvConfig); err != nil {
					logs = append(logs, fmt.Sprintf("[%s] warning: could not write ~/.zshrc.local: %v", pkg, err))
				} else {
					logs = append(logs, fmt.Sprintf("[%s] ~/.zshrc.local updated", pkg))
				}

				logs = append(logs, fmt.Sprintf("[%s] patching template ...", pkg))
				if err := direnv.PatchTemplate(state.HomeDir, state.DirenvConfig); err != nil {
					logs = append(logs, fmt.Sprintf("[%s] warning: could not patch template: %v", pkg, err))
				} else {
					logs = append(logs, fmt.Sprintf("[%s] template updated", pkg))
				}

				logs = append(logs, fmt.Sprintf("[%s] running direnv allow ~/.envrc ...", pkg))
				if err := direnv.AllowEnvrc(state.HomeDir); err != nil {
					logs = append(logs, fmt.Sprintf("[%s] warning: direnv allow failed: %v", pkg, err))
				} else {
					logs = append(logs, fmt.Sprintf("[%s] direnv allow done", pkg))
				}
			}

			// Post-stow git configuration (writes ~/.gitconfig.local).
			if pkg == "git" && state.GitConfig != nil {
				logs = append(logs, fmt.Sprintf("[%s] writing ~/.gitconfig.local ...", pkg))
				if err := gitconfig.WriteGitConfigLocal(state.HomeDir, state.GitConfig); err != nil {
					logs = append(logs, fmt.Sprintf("[%s] warning: could not write ~/.gitconfig.local: %v", pkg, err))
				} else {
					logs = append(logs, fmt.Sprintf("[%s] ~/.gitconfig.local updated", pkg))
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

	case progress.FrameMsg:
		m, v1cmd := p.prog.Update(msg)
		p.prog = m.(progress.Model)
		return p, wrapV1Cmd(v1cmd)

	case installPhaseMsg:
		for i := range p.items {
			if p.items[i].name == msg.pkg {
				p.items[i].phase = msg.phase
				break
			}
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
				p.items[i].phase = ""
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

		// Update progress bar
		doneCount := 0
		for _, item := range p.items {
			if item.status == statusDone || item.status == statusFailed {
				doneCount++
			}
		}
		v1progCmd := p.prog.SetPercent(float64(doneCount) / float64(len(p.items)))

		if msg.done {
			p.done = true
			return p, wrapV1Cmd(v1progCmd)
		}

		// Advance to the next package
		p.current++
		if p.current < len(p.items) {
			p.items[p.current].status = statusActive
		}
		return p, tea.Batch(wrapV1Cmd(v1progCmd), p.processNext())

	case tea.WindowSizeMsg:
		headerHeight := 2 + len(p.items) + 1 // progress bar + items + blank
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
		// Manual scrolling since v1 viewport can't handle v2 KeyPressMsg
		if p.ready {
			switch msg.String() {
			case "down", "j":
				p.logView.ScrollDown(1)
			case "up", "k":
				p.logView.ScrollUp(1)
			}
		}
	}

	return p, nil
}

func (p *ProgressScreen) View() tea.View {
	var b strings.Builder
	b.WriteString("\n")

	// Count completed
	doneCount := 0
	for _, item := range p.items {
		if item.status == statusDone || item.status == statusFailed {
			doneCount++
		}
	}
	total := len(p.items)

	// Progress bar + count
	countStr := Styles.Dimmed.Render(fmt.Sprintf("%d of %d complete", doneCount, total))
	progressLine := lipgloss.JoinHorizontal(lipgloss.Top,
		"  ",
		p.prog.View(),
		"  ",
		countStr,
	)
	b.WriteString(progressLine + "\n\n")

	// Per-package rows
	for _, item := range p.items {
		var icon string
		var nameStyle lipgloss.Style
		switch item.status {
		case statusPending:
			icon = Styles.Dimmed.Render("  ")
			nameStyle = Styles.Dimmed
		case statusActive:
			icon = p.spinner.View()
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(Theme.Cyan)
		case statusDone:
			icon = Icons.Success
			nameStyle = Styles.Success
		case statusFailed:
			icon = Icons.Failure
			nameStyle = Styles.Error.Bold(true)
		}

		row := fmt.Sprintf("  %s %-14s", icon, nameStyle.Render(item.name))
		if item.status == statusActive && item.phase != "" {
			row += Styles.Dimmed.Render(item.phase + "...")
		} else if item.status == statusPending {
			row += Styles.Dimmed.Render("pending")
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")

	// Scrolling log viewport
	if p.ready {
		viewW := p.width - 4
		if viewW < 10 {
			viewW = 10
		}
		
		// Use a distinct border style for the log
		logStyle := Styles.Border.
			BorderForeground(Theme.BrightBlack).
			Width(viewW).
			Padding(0, 1)
			
		b.WriteString(logStyle.Render(p.logView.View()) + "\n")
	}

	if p.done {
		successBanner := lipgloss.NewStyle().
			Foreground(Theme.Green).
			Bold(true).
			PaddingLeft(2).
			Render("✓  All tasks finished!")
		b.WriteString("\n" + successBanner + "\n")
	}

	return tea.NewView(b.String())
}
