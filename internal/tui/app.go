package tui

import (
	"github.com/treycaliva/dotfiles/internal/config"
	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenSelect
	ScreenPreview
	ScreenDiff
	ScreenProgress
	ScreenSummary
)

// ScreenModel is implemented by each screen.
type ScreenModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (ScreenModel, tea.Cmd)
	View() string
}

// AppState holds shared state passed between screens.
type AppState struct {
	Config      *config.Config
	Platform    platform.Info
	DotfilesDir string
	HomeDir     string

	// Selection
	Selected  []string
	Unstowing bool

	// Preview results
	Conflicts map[string][]string

	// Progress results
	Results map[string]error
	Backups []string

	// Diff target
	DiffPkg  string
	DiffFile string

	// Stow status cache
	StowStatus map[string]bool
}

func (s *AppState) RefreshStowStatus() {
	s.StowStatus = make(map[string]bool)
	for _, name := range s.Config.PackageNames() {
		stowed, _ := stow.IsStowed(name, s.DotfilesDir, s.HomeDir)
		s.StowStatus[name] = stowed
	}
}

// NavigateMsg tells the app to switch screens.
type NavigateMsg struct {
	Screen Screen
}

type App struct {
	state    *AppState
	screen   Screen
	current  ScreenModel
	width    int
	height   int
	showHelp bool
}

func NewApp(cfg *config.Config, plat platform.Info, dotfilesDir, homeDir string) App {
	state := &AppState{
		Config:      cfg,
		Platform:    plat,
		DotfilesDir: dotfilesDir,
		HomeDir:     homeDir,
		Results:     make(map[string]error),
		Conflicts:   make(map[string][]string),
	}
	state.RefreshStowStatus()

	return App{
		state:   state,
		screen:  ScreenHome,
		current: NewHomeScreen(state),
	}
}

func (a App) Init() tea.Cmd {
	return a.current.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "?":
			a.showHelp = !a.showHelp
			return a, nil
		case "q":
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			if a.screen != ScreenProgress && a.screen != ScreenDiff {
				return a, tea.Quit
			}
		}
		// Swallow all other keys while help is visible.
		if a.showHelp {
			return a, nil
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case NavigateMsg:
		return a.navigate(msg)
	}

	updated, cmd := a.current.Update(msg)
	a.current = updated
	return a, cmd
}

func (a App) View() string {
	view := a.current.View()
	if a.showHelp {
		help := Styles.Border.Padding(1, 2).Render(helpText())
		return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, help)
	}
	return view
}

func (a App) navigate(msg NavigateMsg) (tea.Model, tea.Cmd) {
	a.screen = msg.Screen

	switch msg.Screen {
	case ScreenHome:
		a.state.RefreshStowStatus()
		a.current = NewHomeScreen(a.state)
	case ScreenSelect:
		a.current = NewSelectScreen(a.state)
	case ScreenPreview:
		a.current = NewPreviewScreen(a.state)
	case ScreenDiff:
		a.current = NewDiffScreen(a.state, a.state.DiffPkg, a.state.DiffFile)
	case ScreenProgress:
		a.current = NewProgressScreen(a.state)
	case ScreenSummary:
		a.state.RefreshStowStatus()
		a.current = NewSummaryScreen(a.state)
	}

	return a, a.current.Init()
}

func helpText() string {
	return `  Keyboard Shortcuts

  Navigation
    enter     Proceed / confirm
    esc       Go back one screen
    q         Quit
    ?         Toggle this help

  Select Screen
    j/k       Move cursor
    space     Toggle package
    m/s/f     Apply profile (minimal/server/full)
    a         Toggle all
    u         Switch install/unstow mode

  Preview Screen
    d         View diff for conflicting file

  Diff View
    j/k       Scroll
    esc       Back to preview

  Summary
    r         Start over`
}
