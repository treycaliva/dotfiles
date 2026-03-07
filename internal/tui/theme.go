package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
	Title              lipgloss.Style
	StatusBar          lipgloss.Style
	Success            lipgloss.Style
	Error              lipgloss.Style
	Warning            lipgloss.Style
	Selected           lipgloss.Style
	Border             lipgloss.Style
	DiffAdd            lipgloss.Style
	DiffDel            lipgloss.Style
	Pill               lipgloss.Style
	PillSuccess        lipgloss.Style
	PillWarning        lipgloss.Style
	PillError          lipgloss.Style
	AccentBorderSuccess lipgloss.Style
	AccentBorderWarning lipgloss.Style
	AccentBorderError  lipgloss.Style
	HighlightRow       lipgloss.Style
	Dimmed             lipgloss.Style
	Header             lipgloss.Style
	Breadcrumb         lipgloss.Style
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

	Pill: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1),

	PillSuccess: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Green).
		Foreground(Theme.Green).
		Padding(0, 1),

	PillWarning: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Yellow).
		Foreground(Theme.Yellow).
		Padding(0, 1),

	PillError: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.Red).
		Foreground(Theme.Red).
		Padding(0, 1),

	AccentBorderSuccess: lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Theme.Green).
		PaddingLeft(1),

	AccentBorderWarning: lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Theme.Yellow).
		PaddingLeft(1),

	AccentBorderError: lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Theme.Red).
		PaddingLeft(1),

	HighlightRow: lipgloss.NewStyle().
		Background(Theme.BrightBlack).
		Bold(true),

	Dimmed: lipgloss.NewStyle().
		Foreground(Theme.BrightBlack),

	Header: lipgloss.NewStyle().
		Background(Theme.Black).
		Foreground(Theme.White).
		Bold(true),

	Breadcrumb: lipgloss.NewStyle().
		Foreground(Theme.Cyan),
}

// GradientTitle renders s with per-character foreground color interpolated
// from Srcery Yellow (#FBB829) to Srcery Cyan (#0AAEB3).
func GradientTitle(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	sr, sg, sb := 0xFB, 0xB8, 0x29
	er, eg, eb := 0x0A, 0xAE, 0xB3
	var b strings.Builder
	n := len(runes)
	for i, r := range runes {
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		ri := sr + int(float64(er-sr)*t)
		gi := sg + int(float64(eg-sg)*t)
		bi := sb + int(float64(eb-sb)*t)
		color := lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", ri, gi, bi))
		b.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(r)))
	}
	return b.String()
}

var Icons = struct {
	Success   string
	Failure   string
	Warning   string
	Pending   string
	Checked   string
	Unchecked string
	Cursor    string
}{
	Success:   Styles.Success.Render(" "),
	Failure:   Styles.Error.Render(" "),
	Warning:   Styles.Warning.Render(" "),
	Pending:   Styles.Warning.Render("󰔟 "),
	Checked:   Styles.Selected.Render("󰄲 "),
	Unchecked: Styles.Dimmed.Render("󰄱 "),
	Cursor:    lipgloss.NewStyle().Foreground(Theme.Yellow).Render("▶ "),
}
