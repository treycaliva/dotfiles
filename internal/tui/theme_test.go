package tui

import "testing"

func TestThemeColorsExist(t *testing.T) {
	colors := []struct {
		name  string
		color string
	}{
		{"Black", string(Theme.Black)},
		{"Red", string(Theme.Red)},
		{"Green", string(Theme.Green)},
		{"Yellow", string(Theme.Yellow)},
		{"Cyan", string(Theme.Cyan)},
	}
	for _, c := range colors {
		if c.color == "" {
			t.Errorf("Theme.%s is empty", c.name)
		}
	}
}

func TestThemeStyles(t *testing.T) {
	_ = Styles.Title.Render("test")
	_ = Styles.StatusBar.Render("test")
	_ = Styles.Success.Render("test")
	_ = Styles.Error.Render("test")
	_ = Styles.Warning.Render("test")
}
