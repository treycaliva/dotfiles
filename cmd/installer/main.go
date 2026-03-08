package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/treycaliva/dotfiles/internal/config"
	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	plat := platform.DetectOS()

	// Determine dotfiles directory
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding executable: %v\n", err)
		os.Exit(1)
	}
	dotfilesDir := filepath.Dir(exe)

	// If run via `go run` or from the repo, use working directory
	if _, err := os.Stat(filepath.Join(dotfilesDir, "go.mod")); err != nil {
		dotfilesDir, _ = os.Getwd()
	}

	homeDir, _ := os.UserHomeDir()
	workingDir, _ := os.Getwd()

	// Mode selection logic: if we're not in the dotfiles repo, we're in project mode.
	mode := tui.ModeInstall
	if _, err := os.Stat(filepath.Join(workingDir, ".git")); err == nil {
		if workingDir != dotfilesDir {
			mode = tui.ModeProject
		}
	}

	app := tui.NewApp(cfg, plat, dotfilesDir, homeDir, workingDir, mode)
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
