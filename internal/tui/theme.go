package tui

import "github.com/charmbracelet/lipgloss"

type srceryTheme struct {
	Black       lipgloss.Color
	Red         lipgloss.Color
	Green       lipgloss.Color
	Yellow      lipgloss.Color
	Blue        lipgloss.Color
	Magenta     lipgloss.Color
	Cyan        lipgloss.Color
	White       lipgloss.Color
	BrightBlack lipgloss.Color
}

var Theme = srceryTheme{
	Black:       lipgloss.Color("#1C1B19"),
	Red:         lipgloss.Color("#EF2F27"),
	Green:       lipgloss.Color("#519F50"),
	Yellow:      lipgloss.Color("#FBB829"),
	Blue:        lipgloss.Color("#2C78BF"),
	Magenta:     lipgloss.Color("#E02C6D"),
	Cyan:        lipgloss.Color("#0AAEB3"),
	White:       lipgloss.Color("#BAA67F"),
	BrightBlack: lipgloss.Color("#918175"),
}

type styles struct {
	Title     lipgloss.Style
	StatusBar lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Selected  lipgloss.Style
	Border    lipgloss.Style
	DiffAdd   lipgloss.Style
	DiffDel   lipgloss.Style
}

var Styles = styles{
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Yellow),

	StatusBar: lipgloss.NewStyle().
		Background(Theme.BrightBlack).
		Foreground(Theme.White).
		Padding(0, 1),

	Success: lipgloss.NewStyle().
		Foreground(Theme.Green),

	Error: lipgloss.NewStyle().
		Foreground(Theme.Red),

	Warning: lipgloss.NewStyle().
		Foreground(Theme.Yellow),

	Selected: lipgloss.NewStyle().
		Foreground(Theme.Cyan).
		Bold(true),

	Border: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BrightBlack),

	DiffAdd: lipgloss.NewStyle().
		Foreground(Theme.Green),

	DiffDel: lipgloss.NewStyle().
		Foreground(Theme.Red),
}

var Icons = struct {
	Success string
	Failure string
	Warning string
	Pending string
}{
	Success: Styles.Success.Render(""),
	Failure: Styles.Error.Render(""),
	Warning: Styles.Warning.Render(""),
	Pending: Styles.Warning.Render(""),
}
