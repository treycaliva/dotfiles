package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	stateDirenvSetup
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
	state           state
	cursor          int
	selectedProf    *Profile
	nameInput       textinput.Model
	emailInput      textinput.Model
	direnvCtxCursor int
	opAccountInput  textinput.Model
	direnvStep      int
	installLog      []string
	err             error
}

func initialModel() model {
	tiName := textinput.New()
	tiName.Placeholder = "Jane Doe"
	tiName.Focus()

	tiEmail := textinput.New()
	tiEmail.Placeholder = "jane@example.com"

	tiAccount := textinput.New()
	tiAccount.Placeholder = "my.1password.com"
	tiAccount.SetValue("my.1password.com")

	return model{
		state:          stateProfileSelection,
		nameInput:      tiName,
		emailInput:     tiEmail,
		opAccountInput: tiAccount,
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
	case stateDirenvSetup:
		return m.updateDirenvSetup(msg)
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
				m.state = stateDirenvSetup
				m.opAccountInput.Focus()
				return m, textinput.Blink
			}
		}
	}
	m.emailInput, cmd = m.emailInput.Update(msg)
	return m, cmd
}

var direnvContexts = []string{"personal", "work"}

func (m model) updateDirenvSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.direnvStep == 0 {
			switch msg.String() {
			case "up", "k":
				if m.direnvCtxCursor > 0 {
					m.direnvCtxCursor--
				}
			case "down", "j":
				if m.direnvCtxCursor < len(direnvContexts)-1 {
					m.direnvCtxCursor++
				}
			case "enter":
				m.direnvStep = 1
				return m, textinput.Blink
			}
		} else {
			if msg.Type == tea.KeyEnter {
				if m.opAccountInput.Value() != "" {
					m.state = stateInstalling
					return m, m.doInstall()
				}
			}
		}
	}
	if m.direnvStep == 1 {
		m.opAccountInput, cmd = m.opAccountInput.Update(msg)
	}
	return m, cmd
}

type installMsg string
type errMsg struct{ err error }

// upsertZshrcLocal sets KEY=value in ~/.zshrc.local, adding the line if absent
// or replacing it if already present.
func upsertZshrcLocal(path, key, value string) error {
	export := fmt.Sprintf("export %s=%s", key, value)

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "export "+key+"=") {
			lines[i] = export
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, export)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

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
				// ... (conflict resolution logic remains same)
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
							fullPath := filepath.Join(home, conflictFile)
							repoPath := filepath.Join(cwd, pkg, conflictFile)

							// Check if files are identical
							isIdentical := false
							repoData, err1 := os.ReadFile(repoPath)
							homeData, err2 := os.ReadFile(fullPath)
							if err1 == nil && err2 == nil && string(repoData) == string(homeData) {
								isIdentical = true
							}

							backupDir := filepath.Join(home, ".dotfiles-backup")
							
							if isIdentical {
								// If identical, we can just remove it to let stow create the link
								if removeErr := os.Remove(fullPath); removeErr == nil {
									m.installLog = append(m.installLog, fmt.Sprintf("  - Resolved conflict: Removed identical file %s", conflictFile))
									conflictsResolved = true
								}
							} else if _, statErr := os.Lstat(fullPath); statErr == nil {
								// If different or not identical, show a diff summary and backup
								diffCmd := exec.Command("diff", "-u", repoPath, fullPath)
								diffOut, _ := diffCmd.CombinedOutput()
								diffStr := string(diffOut)
								if len(diffStr) > 500 {
									diffStr = diffStr[:500] + "\n... (diff truncated)"
								}

								backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.bak.%d", conflictFile, os.Getpid()))
								// Ensure parent directory of backupPath exists
								os.MkdirAll(filepath.Dir(backupPath), 0755)

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
							if pkg == "tmux" {
								m.installTmuxTheme(home)
							}
							continue
						}
					}
				}
				return errMsg{fmt.Errorf("failed to stow %s: %v\nOutput: %s", pkg, err, string(output))}
			}
			m.installLog = append(m.installLog, fmt.Sprintf("  - Successfully stowed %s", pkg))
			if pkg == "tmux" {
				m.installTmuxTheme(home)
			}
		}

		// Write context and account to ~/.zshrc.local
		zshrcLocal := filepath.Join(home, ".zshrc.local")
		ctx := direnvContexts[m.direnvCtxCursor]
		account := m.opAccountInput.Value()

		if err := upsertZshrcLocal(zshrcLocal, "DOTFILES_CONTEXT", ctx); err != nil {
			return errMsg{fmt.Errorf("failed to write DOTFILES_CONTEXT to ~/.zshrc.local: %v", err)}
		}
		if err := upsertZshrcLocal(zshrcLocal, "DOTFILES_OP_ACCOUNT", account); err != nil {
			return errMsg{fmt.Errorf("failed to write DOTFILES_OP_ACCOUNT to ~/.zshrc.local: %v", err)}
		}
		m.installLog = append(m.installLog, fmt.Sprintf("  - Wrote DOTFILES_CONTEXT=%s and DOTFILES_OP_ACCOUNT=%s to ~/.zshrc.local", ctx, account))

		// Allow global ~/.envrc
		envrcPath := filepath.Join(home, ".envrc")
		if _, err := os.Stat(envrcPath); err == nil {
			allowCmd := exec.Command("direnv", "allow", envrcPath)
			if output, err := allowCmd.CombinedOutput(); err != nil {
				m.installLog = append(m.installLog, fmt.Sprintf("  - Warning: 'direnv allow' failed: %v\n    Output: %s\n    Run manually: direnv allow ~/.envrc", err, string(output)))
			} else {
				m.installLog = append(m.installLog, "  - Ran: direnv allow ~/.envrc")
			}
		} else {
			m.installLog = append(m.installLog, "  - Skipped direnv allow: ~/.envrc not found (stow may need to run first)")
		}

		// Check op is signed in (warn-only)
		opCheckCmd := exec.Command("op", "account", "list")
		if output, err := opCheckCmd.CombinedOutput(); err != nil || strings.TrimSpace(string(output)) == "" {
			m.installLog = append(m.installLog, "  - Warning: no 1Password account found. Run 'op signin' to authenticate.")
		} else {
			m.installLog = append(m.installLog, "  - 1Password CLI: account found")
		}

		return installMsg("Installation Complete!")
	}
}

func (m *model) installTmuxTheme(home string) {
	themeDir := filepath.Join(home, ".tmux/themes/srcery-tmux")
	if _, err := os.Stat(themeDir); os.IsNotExist(err) {
		m.installLog = append(m.installLog, "  - Installing Srcery tmux theme...")
		cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/srcery-colors/srcery-tmux", themeDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			m.installLog = append(m.installLog, fmt.Sprintf("    Error installing theme: %v\nOutput: %s", err, string(output)))
		} else {
			m.installLog = append(m.installLog, "    Successfully installed Srcery tmux theme")
		}
	} else {
		m.installLog = append(m.installLog, "  - Srcery tmux theme already installed")
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

	case stateDirenvSetup:
		if m.direnvStep == 0 {
			b.WriteString("Set up direnv + 1Password.\n\n")
			b.WriteString("Which context is this machine?\n\n")
			for i, ctx := range direnvContexts {
				cursor := "  "
				style := itemStyle
				if m.direnvCtxCursor == i {
					cursor = "> "
					style = selStyle
				}
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(ctx)))
			}
			b.WriteString("\n" + infoStyle.Render("Press Enter to confirm."))
		} else {
			b.WriteString("Set up direnv + 1Password.\n\n")
			b.WriteString("1Password account shorthand:\n")
			b.WriteString(infoStyle.Render("Run 'op account list' to find yours.\n\n"))
			b.WriteString(m.opAccountInput.View() + "\n\n")
			b.WriteString(infoStyle.Render("Press Enter to begin installation."))
		}

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