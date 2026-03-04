package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateProfileSelection state = iota
	stateGitName
	stateGitEmail
	stateInstalling
	stateDone
)

type Profile struct {
	Name        string
	Description string
	Packages    []string
}

var profiles = []Profile{
	{Name: "Base", Description: "Core CLI tools (zsh, tmux, vim, git)", Packages: []string{"zsh", "tmux", "vim", "git", "p10k"}},
	{Name: "Desktop", Description: "Base + GUI terminals (ghostty, alacritty)", Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty"}},
	{Name: "Dev", Description: "Desktop + Dev tools (nvim)", Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "nvim"}},
}

var (
	titleStyle = lipgloss.NewStyle().Margin(1, 0).Padding(0, 1).Background(lipgloss.Color("63")).Foreground(lipgloss.Color("230")).Bold(true)
	itemStyle  = lipgloss.NewStyle().PaddingLeft(2)
	selStyle   = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("205"))
	descStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	succStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
)

type model struct {
	state          state
	cursor         int
	selectedProf   *Profile
	nameInput      textinput.Model
	emailInput     textinput.Model
	installLog     []string
	err            error
}

func initialModel() model {
	tiName := textinput.New()
	tiName.Placeholder = "Jane Doe"
	tiName.Focus()

	tiEmail := textinput.New()
	tiEmail.Placeholder = "jane@example.com"

	return model{
		state:      stateProfileSelection,
		nameInput:  tiName,
		emailInput: tiEmail,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit keys
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

		// If we're in an error state or stateDone, any key exits
		if m.err != nil || m.state == stateDone {
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateProfileSelection:
		return m.updateProfileSelection(msg)
	case stateGitName:
		return m.updateGitName(msg)
	case stateGitEmail:
		return m.updateGitEmail(msg)
	case stateInstalling:
		return m.updateInstalling(msg)
	case stateDone:
		// Any key handling is now at the top of Update()
		return m, nil
	}

	return m, cmd
}

func (m model) updateProfileSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(profiles)-1 {
				m.cursor++
			}
		case "enter":
			m.selectedProf = &profiles[m.cursor]
			m.state = stateGitName
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateGitName(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if m.nameInput.Value() != "" {
				m.state = stateGitEmail
				m.emailInput.Focus()
				return m, textinput.Blink
			}
		}
	}
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m model) updateGitEmail(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if m.emailInput.Value() != "" {
				m.state = stateInstalling
				return m, m.doInstall()
			}
		}
	}
	m.emailInput, cmd = m.emailInput.Update(msg)
	return m, cmd
}

type installMsg string
type errMsg struct{ err error }

func (m model) doInstall() tea.Cmd {
	return func() tea.Msg {
		home, err := os.UserHomeDir()
		if err != nil {
			return errMsg{err}
		}

		// 1. Check for stow
		if _, err := exec.LookPath("stow"); err != nil {
			return errMsg{fmt.Errorf("stow is not installed. Please install it first (e.g., brew install stow or sudo apt install stow)")}
		}

		// 2. Generate Git Config Local
		gitLocalPath := home + "/.gitconfig.local"
		tmplContent := `[user]
	name = {{.Name}}
	email = {{.Email}}
`
		t, err := template.New("gitconfig").Parse(tmplContent)
		if err != nil {
			return errMsg{err}
		}

		f, err := os.Create(gitLocalPath)
		if err != nil {
			return errMsg{err}
		}
		err = t.Execute(f, struct {
			Name  string
			Email string
		}{
			Name:  m.nameInput.Value(),
			Email: m.emailInput.Value(),
		})
		f.Close()
		if err != nil {
			return errMsg{err}
		}

		// 3. Stow Packages
		cwd, err := os.Getwd()
		if err != nil {
			return errMsg{err}
		}

		for _, pkg := range m.selectedProf.Packages {
			m.installLog = append(m.installLog, fmt.Sprintf("Stowing %s...", pkg))
			cmd := exec.Command("stow", "-d", cwd, "-t", home, "--restow", pkg)
			output, err := cmd.CombinedOutput()
			if err != nil {
				outStr := string(output)
				if strings.Contains(outStr, "conflicts:") || strings.Contains(outStr, "not owned by stow") {
					lines := strings.Split(outStr, "\n")
					conflictsResolved := false
					for _, line := range lines {
						var conflictFile string
						if strings.Contains(line, "over existing target") {
							// Format: "* cannot stow ... over existing target .gitconfig since ..."
							parts := strings.Split(line, "over existing target")
							if len(parts) > 1 {
								after := strings.TrimSpace(parts[1])
								fileParts := strings.Fields(after)
								if len(fileParts) > 0 {
									conflictFile = fileParts[0]
								}
							}
						} else if strings.Contains(line, "existing target is not owned by stow") {
							// Format: "* existing target is not owned by stow: .zshrc"
							parts := strings.Split(line, ":")
							if len(parts) > 1 {
								conflictFile = strings.TrimSpace(parts[len(parts)-1])
							}
						}

						if conflictFile != "" {
							fullPath := home + "/" + conflictFile
							repoPath := cwd + "/" + pkg + "/" + conflictFile

							// Check if files are identical
							isIdentical := false
							repoData, err1 := os.ReadFile(repoPath)
							homeData, err2 := os.ReadFile(fullPath)
							if err1 == nil && err2 == nil && string(repoData) == string(homeData) {
								isIdentical = true
							}

							backupDir := home + "/.dotfiles-backup"
							os.MkdirAll(backupDir, 0755)

							if isIdentical {
								// If identical, we can just remove it to let stow create the link
								if removeErr := os.Remove(fullPath); removeErr == nil {
									m.installLog = append(m.installLog, fmt.Sprintf("  - Resolved conflict: Removed identical file %s", conflictFile))
									conflictsResolved = true
								}
							} else {
								// If different, show a diff summary and backup
								diffCmd := exec.Command("diff", "-u", repoPath, fullPath)
								diffOut, _ := diffCmd.CombinedOutput()
								diffStr := string(diffOut)
								if len(diffStr) > 500 {
									diffStr = diffStr[:500] + "\n... (diff truncated)"
								}

								backupPath := fmt.Sprintf("%s/%s.bak.%d", backupDir, conflictFile, os.Getpid())
								if moveErr := os.Rename(fullPath, backupPath); moveErr == nil {
									m.installLog = append(m.installLog, fmt.Sprintf("  - Resolved conflict: Files differed! Backed up %s", conflictFile))
									if diffStr != "" {
										m.installLog = append(m.installLog, "    Diff preview:\n"+diffStr)
									}
									conflictsResolved = true
								}
							}
						}
					}
					if conflictsResolved {
						// Final attempt for this package
						cmd = exec.Command("stow", "-d", cwd, "-t", home, "--restow", pkg)
						output, err = cmd.CombinedOutput()
						if err == nil {
							m.installLog = append(m.installLog, fmt.Sprintf("  - Successfully stowed %s after resolution", pkg))
							continue
						}
					}
				}
				return errMsg{fmt.Errorf("failed to stow %s: %v\nOutput: %s", pkg, err, string(output))}
			}
			m.installLog = append(m.installLog, fmt.Sprintf("  - Successfully stowed %s", pkg))
		}

		return installMsg("Installation Complete!")
	}
}

func (m model) updateInstalling(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installMsg:
		m.installLog = append(m.installLog, string(msg))
		m.state = stateDone
		return m, nil
	case errMsg:
		m.err = msg.err
		m.state = stateDone
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return errStyle.Render(fmt.Sprintf("Error: %v\n\nPress any key to exit.", m.err))
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Dotfiles Installer"))
	b.WriteString("\n\n")

	switch m.state {
	case stateProfileSelection:
		b.WriteString("Select an installation profile:\n\n")
		for i, p := range profiles {
			cursor := "  "
			style := itemStyle
			if m.cursor == i {
				cursor = "> "
				style = selStyle
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(p.Name)))
			b.WriteString(fmt.Sprintf("   %s\n", descStyle.Render(p.Description)))
		}
		b.WriteString("\n" + infoStyle.Render("Press Enter to select, Esc to quit."))

	case stateGitName:
		b.WriteString("Let's set up your Git identity.\n\n")
		b.WriteString("What is your Name?\n")
		b.WriteString(m.nameInput.View() + "\n\n")
		b.WriteString(infoStyle.Render("Press Enter to continue."))

	case stateGitEmail:
		b.WriteString("Let's set up your Git identity.\n\n")
		b.WriteString("What is your Email?\n")
		b.WriteString(m.emailInput.View() + "\n\n")
		b.WriteString(infoStyle.Render("Press Enter to continue."))

	case stateInstalling:
		b.WriteString("Installing packages and configuring dotfiles...\n")

	case stateDone:
		b.WriteString(succStyle.Render("Done!"))
		b.WriteString("\n\nDetails:\n")
		for _, log := range m.installLog {
			b.WriteString("- " + log + "\n")
		}
		b.WriteString("\n" + infoStyle.Render("Press any key to exit."))
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}