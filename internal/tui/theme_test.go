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

func TestGradientTitle(t *testing.T) {
	result := GradientTitle("hi")
	if result == "" {
		t.Fatal("expected non-empty gradient string")
	}
	if len(result) < len("hi") {
		t.Errorf("gradient result too short: %q", result)
	}
}

func TestGradientTitleEmpty(t *testing.T) {
	if GradientTitle("") != "" {
		t.Fatal("expected empty string for empty input")
	}
}

func TestStylesHasNewFields(t *testing.T) {
	_ = Styles.Pill
	_ = Styles.PillSuccess
	_ = Styles.PillWarning
	_ = Styles.PillError
	_ = Styles.AccentBorderSuccess
	_ = Styles.AccentBorderWarning
	_ = Styles.AccentBorderError
	_ = Styles.HighlightRow
	_ = Styles.Dimmed
	_ = Styles.Header
	_ = Styles.Breadcrumb
}
