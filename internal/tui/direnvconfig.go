package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/treycaliva/dotfiles/internal/direnv"
)

type direnvStep int

const (
	direnvStepContext direnvStep = iota
	direnvStepOPAccount
	direnvStepSecretKey
	direnvStepSecretRef
	direnvStepAddAnother
	direnvStepConfirm
)

// DirenvConfigScreen collects 1Password configuration before direnv is stowed.
type DirenvConfigScreen struct {
	state     *AppState
	step      direnvStep
	context   string // "personal" or "work"
	account   textinput.Model
	secretKey textinput.Model
	secretRef textinput.Model
	secrets   []direnv.Secret
	width     int
	height    int
}

func NewDirenvConfigScreen(state *AppState) *DirenvConfigScreen {
	account := textinput.New()
	account.Placeholder = "e.g. my.1password.com"
	account.CharLimit = 128

	secretKey := textinput.New()
	secretKey.Placeholder = "e.g. GITHUB_TOKEN"
	secretKey.CharLimit = 128

	secretRef := textinput.New()
	secretRef.Placeholder = "e.g. op://Personal/GitHub/token"
	secretRef.CharLimit = 256

	return &DirenvConfigScreen{
		state:     state,
		step:      direnvStepContext,
		context:   "personal",
		account:   account,
		secretKey: secretKey,
		secretRef: secretRef,
	}
}

func (d *DirenvConfigScreen) Init() tea.Cmd { return nil }

func (d *DirenvConfigScreen) SetSize(w, h int) {
	if w < 10 {
		w = 10
	}
	if h < 3 {
		h = 3
	}
	d.width = w
	d.height = h
}

func (d *DirenvConfigScreen) StatusBar() []KeyBinding {
	switch d.step {
	case direnvStepContext:
		return []KeyBinding{{Key: "tab", Help: "toggle"}, {Key: "enter", Help: "next"}, {Key: "esc", Help: "back"}}
	case direnvStepConfirm:
		return []KeyBinding{{Key: "enter", Help: "install"}, {Key: "esc", Help: "back"}}
	default:
		return []KeyBinding{{Key: "enter", Help: "next"}, {Key: "esc", Help: "back"}}
	}
}

func (d *DirenvConfigScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			d.state.DirenvConfig = nil
			return d, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		case "tab":
			if d.step == direnvStepContext {
				if d.context == "personal" {
					d.context = "work"
				} else {
					d.context = "personal"
				}
			}
		case "enter":
			return d.advance()
		case "y", "Y":
			if d.step == direnvStepAddAnother {
				d.secretKey.Reset()
				d.secretRef.Reset()
				d.step = direnvStepSecretKey
				d.secretKey.Focus()
				return d, wrapV1Cmd(textinput.Blink)
			}
		case "n", "N":
			if d.step == direnvStepAddAnother {
				d.step = direnvStepConfirm
			}
		}
	}

	// Convert v2 messages to v1 before forwarding to bubbles textinput.
	var v1msg v1tea.Msg = msg
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		v1msg = toV1KeyMsg(m)
	case tea.PasteMsg:
		v1msg = v1tea.KeyMsg(v1tea.Key{
			Type:  v1tea.KeyRunes,
			Runes: []rune(m.Content),
			Paste: true,
		})
	}

	var v1cmd v1tea.Cmd
	switch d.step {
	case direnvStepOPAccount:
		d.account, v1cmd = d.account.Update(v1msg)
	case direnvStepSecretKey:
		d.secretKey, v1cmd = d.secretKey.Update(v1msg)
	case direnvStepSecretRef:
		d.secretRef, v1cmd = d.secretRef.Update(v1msg)
	}

	return d, wrapV1Cmd(v1cmd)
}

// advance validates the current step and moves to the next.
func (d *DirenvConfigScreen) advance() (ScreenModel, tea.Cmd) {
	switch d.step {
	case direnvStepContext:
		d.step = direnvStepOPAccount
		d.account.Focus()
		return d, wrapV1Cmd(textinput.Blink)

	case direnvStepOPAccount:
		if strings.TrimSpace(d.account.Value()) == "" {
			return d, nil
		}
		d.step = direnvStepSecretKey
		d.secretKey.Focus()
		return d, wrapV1Cmd(textinput.Blink)

	case direnvStepSecretKey:
		if strings.TrimSpace(d.secretKey.Value()) == "" {
			return d, nil
		}
		d.step = direnvStepSecretRef
		d.secretRef.Focus()
		return d, wrapV1Cmd(textinput.Blink)

	case direnvStepSecretRef:
		if strings.TrimSpace(d.secretRef.Value()) == "" {
			return d, nil
		}
		d.secrets = append(d.secrets, direnv.Secret{
			Key:   strings.TrimSpace(d.secretKey.Value()),
			OPRef: strings.TrimSpace(d.secretRef.Value()),
		})
		d.secretKey.Blur()
		d.secretRef.Blur()
		d.step = direnvStepAddAnother

	case direnvStepAddAnother:
		d.step = direnvStepConfirm

	case direnvStepConfirm:
		d.state.DirenvConfig = &direnv.Setup{
			Context:   d.context,
			OPAccount: strings.TrimSpace(d.account.Value()),
			Secrets:   d.secrets,
		}
		return d, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
	}

	return d, nil
}

func (d *DirenvConfigScreen) View() tea.View {
	var b strings.Builder
	b.WriteString("\n")

	label := lipgloss.NewStyle().Bold(true).Foreground(Theme.Cyan)
	dim := Styles.Dimmed

	switch d.step {
	case direnvStepContext:
		b.WriteString("  " + label.Render("Context") + "\n\n")
		for _, ctx := range []string{"personal", "work"} {
			if ctx == d.context {
				b.WriteString("  " + Styles.Selected.Render("● "+ctx) + "\n")
			} else {
				b.WriteString("  " + dim.Render("○ "+ctx) + "\n")
			}
		}
		b.WriteString("\n  " + dim.Render("tab: toggle  enter: next") + "\n")

	case direnvStepOPAccount:
		b.WriteString("  " + label.Render("1Password account shorthand") + "\n")
		b.WriteString("  " + dim.Render("Run `op account list` to find this value") + "\n\n")
		b.WriteString("  " + d.account.View() + "\n")

	case direnvStepSecretKey:
		b.WriteString("  " + label.Render("Secret — environment variable name") + "\n\n")
		b.WriteString("  " + d.secretKey.View() + "\n")

	case direnvStepSecretRef:
		b.WriteString("  " + label.Render(fmt.Sprintf("Secret — op:// reference for %s", d.secretKey.Value())) + "\n\n")
		b.WriteString("  " + d.secretRef.View() + "\n")

	case direnvStepAddAnother:
		b.WriteString("  " + label.Render("Add another secret?") + "\n\n")
		for _, s := range d.secrets {
			b.WriteString(fmt.Sprintf("  %s %s = %s\n", Icons.Success, s.Key, dim.Render(s.OPRef)))
		}
		b.WriteString("\n  " + dim.Render("y: add another   n/enter: done") + "\n")

	case direnvStepConfirm:
		b.WriteString("  " + label.Render("Ready to install") + "\n\n")
		b.WriteString(fmt.Sprintf("  Context:    %s\n", Styles.Success.Render(d.context)))
		b.WriteString(fmt.Sprintf("  OP Account: %s\n", Styles.Success.Render(strings.TrimSpace(d.account.Value()))))
		if len(d.secrets) > 0 {
			b.WriteString("\n  " + label.Render("Secrets") + "\n")
			for _, s := range d.secrets {
				b.WriteString(fmt.Sprintf("  %s %s\n", Icons.Success, s.Key))
				b.WriteString(fmt.Sprintf("      %s\n", dim.Render(s.OPRef)))
			}
		}
		b.WriteString("\n  " + dim.Render("Writes ~/.zshrc.local and ~/.config/direnv/templates/"+d.context+".env.tpl") + "\n")
	}

	return tea.NewView(b.String())
}
