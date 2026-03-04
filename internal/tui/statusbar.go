package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyBinding pairs a key name with a short description for the status bar.
type KeyBinding struct {
	Key  string
	Help string
}

// StatusBar renders a consistent bottom status bar with keybinding hints.
func StatusBar(width int, bindings []KeyBinding) string {
	var parts []string
	for _, b := range bindings {
		key := lipgloss.NewStyle().Bold(true).Foreground(Theme.Cyan).Render(b.Key)
		parts = append(parts, key+":"+b.Help)
	}
	content := "  " + strings.Join(parts, "  ") + "  "

	style := lipgloss.NewStyle().
		Background(Theme.BrightBlack).
		Foreground(Theme.White).
		Width(width)

	return style.Render(content)
}
