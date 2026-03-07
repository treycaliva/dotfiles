package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	v1tea "github.com/charmbracelet/bubbletea"
	tea "charm.land/bubbletea/v2"
)

// wrapV1Cmd converts a bubbles v1 Cmd (which returns a v1 Msg / interface{})
// into a bubbletea v2 Cmd so that bubbles components remain usable while we
// run on the v2 runtime.
func wrapV1Cmd(cmd v1tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg { return cmd() }
}

// toV1KeyMsg converts a v2 KeyPressMsg into a v1 KeyMsg so that bubbles v1
// components (e.g. textinput) can process keyboard events from the v2 runtime.
func toV1KeyMsg(k tea.KeyPressMsg) v1tea.KeyMsg {
	v1key := v1tea.Key{Alt: k.Mod.Contains(tea.ModAlt)}

	switch k.Code {
	case tea.KeyBackspace:
		v1key.Type = v1tea.KeyBackspace
	case tea.KeyDelete:
		v1key.Type = v1tea.KeyDelete
	case tea.KeyLeft:
		v1key.Type = v1tea.KeyLeft
	case tea.KeyRight:
		v1key.Type = v1tea.KeyRight
	case tea.KeyHome:
		v1key.Type = v1tea.KeyHome
	case tea.KeyEnd:
		v1key.Type = v1tea.KeyEnd
	case tea.KeyTab:
		v1key.Type = v1tea.KeyTab
	case tea.KeyEnter:
		v1key.Type = v1tea.KeyEnter
	case tea.KeyEscape:
		v1key.Type = v1tea.KeyEscape
	default:
		if k.Mod.Contains(tea.ModCtrl) && k.Code >= 'a' && k.Code <= 'z' {
			// v1 ctrl keys map to raw control codes (ctrl+a = 1, etc.)
			v1key.Type = v1tea.KeyType(k.Code - 'a' + 1)
		} else if k.Text != "" {
			v1key.Type = v1tea.KeyRunes
			v1key.Runes = []rune(k.Text)
		} else {
			v1key.Type = v1tea.KeyRunes
			v1key.Runes = []rune{rune(k.Code)}
		}
	}

	return v1tea.KeyMsg(v1key)
}

// V2Spinner encapsulates a v1 spinner.Model for use in a v2 app.
type V2Spinner struct {
	Model spinner.Model
}

func NewV2Spinner(s spinner.Model) V2Spinner {
	return V2Spinner{Model: s}
}

func (s V2Spinner) Tick() tea.Cmd {
	return wrapV1Cmd(s.Model.Tick)
}

func (s *V2Spinner) Update(msg tea.Msg) tea.Cmd {
	var cmd v1tea.Cmd
	s.Model, cmd = s.Model.Update(msg)
	return wrapV1Cmd(cmd)
}

func (s V2Spinner) View() string {
	return s.Model.View()
}

// V2Progress encapsulates a v1 progress.Model for use in a v2 app.
type V2Progress struct {
	Model progress.Model
}

func NewV2Progress(p progress.Model) V2Progress {
	return V2Progress{Model: p}
}

func (p V2Progress) SetPercent(percent float64) tea.Cmd {
	return wrapV1Cmd(p.Model.SetPercent(percent))
}

func (p *V2Progress) Update(msg tea.Msg) tea.Cmd {
	m, cmd := p.Model.Update(msg)
	p.Model = m.(progress.Model)
	return wrapV1Cmd(cmd)
}

func (p V2Progress) View() string {
	return p.Model.View()
}

// V2Viewport encapsulates a v1 viewport.Model for use in a v2 app.
type V2Viewport struct {
	Model viewport.Model
}

func NewV2Viewport(width, height int) V2Viewport {
	return V2Viewport{Model: viewport.New(width, height)}
}

func (v *V2Viewport) Update(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		// Try routing through normal update with v1 key msg for native viewport scrolling bindings
		v1Msg := toV1KeyMsg(k)
		m, cmd := v.Model.Update(v1Msg)
		v.Model = m
		return wrapV1Cmd(cmd)
	}
	m, cmd := v.Model.Update(msg)
	v.Model = m
	return wrapV1Cmd(cmd)
}

func (v V2Viewport) View() string {
	return v.Model.View()
}
