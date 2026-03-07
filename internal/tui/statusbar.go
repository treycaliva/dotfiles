package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders a consistent bottom status bar with keybinding hints.
func RenderStatusBar(width int, leftBindings []KeyBinding, rightBindings []KeyBinding) string {
	var leftParts []string
	for _, b := range leftBindings {
		key := lipgloss.NewStyle().Bold(true).Foreground(Theme.Yellow).Render(b.Key)
		leftParts = append(leftParts, key+":"+b.Help)
	}
	leftContent := "  " + strings.Join(leftParts, "  ")

	var rightParts []string
	for _, b := range rightBindings {
		key := lipgloss.NewStyle().Bold(true).Foreground(Theme.Yellow).Render(b.Key)
		rightParts = append(rightParts, key+":"+b.Help)
	}
	rightContent := strings.Join(rightParts, "  ") + "  "

	leftW := lipgloss.Width(leftContent)
	rightW := lipgloss.Width(rightContent)
	gap := width - leftW - rightW
	if gap < 0 {
		gap = 0
	}

	content := leftContent + strings.Repeat(" ", gap) + rightContent
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#282828")).
		Foreground(Theme.White).
		Width(width).
		Render(content)
}
