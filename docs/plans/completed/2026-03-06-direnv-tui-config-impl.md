# direnv TUI Config Screen — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an interactive `ScreenDirenvConfig` to the Go TUI installer that collects 1Password context, account, and secrets when direnv is selected, then writes `~/.zshrc.local`, patches the op template, and runs `direnv allow` post-stow.

**Architecture:** New `internal/direnv` package holds data types and file-write helpers (testable in isolation). New `ScreenDirenvConfig` in `internal/tui/direnvconfig.go` implements the existing `ScreenModel` interface as a multi-step form. Preview routes to DirenvConfig instead of Progress when direnv is selected; Progress applies the config post-stow.

**Tech Stack:** Go 1.24, BubbleTea v2 (`charm.land/bubbletea/v2`), Bubbles v1 textinput (`github.com/charmbracelet/bubbles/textinput`) wrapped via existing `wrapV1Cmd`, lipgloss for styling.

---

### Task 1: Add `internal/direnv` package with types and file-write helpers

**Files:**
- Create: `internal/direnv/setup.go`
- Create: `internal/direnv/setup_test.go`

**Step 1: Write failing tests**

```go
// internal/direnv/setup_test.go
package direnv_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/treycaliva/dotfiles/internal/direnv"
)

func TestWriteZshrcLocal_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	setup := &direnv.Setup{Context: "personal", OPAccount: "my.1password.com"}

	if err := direnv.WriteZshrcLocal(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmp, ".zshrc.local"))
	content := string(data)
	if !strings.Contains(content, "export DOTFILES_CONTEXT=personal") {
		t.Errorf("missing DOTFILES_CONTEXT, got:\n%s", content)
	}
	if !strings.Contains(content, "export DOTFILES_OP_ACCOUNT=my.1password.com") {
		t.Errorf("missing DOTFILES_OP_ACCOUNT, got:\n%s", content)
	}
}

func TestWriteZshrcLocal_UpdatesExisting(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".zshrc.local")
	existing := "# my config\nexport DOTFILES_CONTEXT=work\nexport OTHER=foo\n"
	os.WriteFile(path, []byte(existing), 0644)

	setup := &direnv.Setup{Context: "personal", OPAccount: "new.1password.com"}
	if err := direnv.WriteZshrcLocal(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if strings.Contains(content, "export DOTFILES_CONTEXT=work") {
		t.Error("old DOTFILES_CONTEXT should have been removed")
	}
	if !strings.Contains(content, "export DOTFILES_CONTEXT=personal") {
		t.Error("new DOTFILES_CONTEXT missing")
	}
	if !strings.Contains(content, "export OTHER=foo") {
		t.Error("unrelated line should be preserved")
	}
}

func TestPatchTemplate_WritesSecrets(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".config", "direnv", "templates")
	os.MkdirAll(dir, 0755)
	tplPath := filepath.Join(dir, "personal.env.tpl")
	os.WriteFile(tplPath, []byte("# Personal template\n# Format: export KEY={{ op://... }}\n"), 0644)

	setup := &direnv.Setup{
		Context: "personal",
		Secrets: []direnv.Secret{
			{Key: "GITHUB_TOKEN", OPRef: "op://Personal/GitHub/token"},
		},
	}
	if err := direnv.PatchTemplate(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tplPath)
	content := string(data)
	if !strings.Contains(content, "export GITHUB_TOKEN={{ op://Personal/GitHub/token }}") {
		t.Errorf("secret missing from template:\n%s", content)
	}
	if !strings.Contains(content, "# Personal template") {
		t.Error("comments should be preserved")
	}
}
```

**Step 2: Run tests to confirm they fail**

```bash
cd /Users/treycaliva/dotfiles && go test ./internal/direnv/...
```
Expected: compile error (package does not exist yet)

**Step 3: Implement `internal/direnv/setup.go`**

```go
package direnv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Secret is a single op:// secret reference to inject via direnv.
type Secret struct {
	Key   string // environment variable name, e.g. GITHUB_TOKEN
	OPRef string // 1Password reference, e.g. op://Personal/GitHub/token
}

// Setup holds the user-supplied direnv configuration.
type Setup struct {
	Context   string   // "personal" or "work"
	OPAccount string   // op account shorthand, e.g. my.1password.com
	Secrets   []Secret
}

// WriteZshrcLocal writes DOTFILES_CONTEXT and DOTFILES_OP_ACCOUNT into
// ~/.zshrc.local (creating it if absent, updating existing entries in place).
func WriteZshrcLocal(homeDir string, setup *Setup) error {
	path := filepath.Join(homeDir, ".zshrc.local")

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	// Remove any existing managed lines.
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(line, "export DOTFILES_CONTEXT=") ||
			strings.HasPrefix(line, "export DOTFILES_OP_ACCOUNT=") {
			continue
		}
		kept = append(kept, line)
	}
	kept = append(kept,
		fmt.Sprintf("export DOTFILES_CONTEXT=%s", setup.Context),
		fmt.Sprintf("export DOTFILES_OP_ACCOUNT=%s", setup.OPAccount),
	)

	content := strings.Join(kept, "\n") + "\n"
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// PatchTemplate rewrites the op template for setup.Context, preserving
// existing comment lines and replacing all export lines with setup.Secrets.
func PatchTemplate(homeDir string, setup *Setup) error {
	path := filepath.Join(homeDir, ".config", "direnv", "templates", setup.Context+".env.tpl")

	var comments []string
	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				comments = append(comments, line)
			}
		}
	}

	var b strings.Builder
	for _, c := range comments {
		b.WriteString(c + "\n")
	}
	for _, s := range setup.Secrets {
		fmt.Fprintf(&b, "export %s={{ %s }}\n", s.Key, s.OPRef)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// AllowEnvrc runs `direnv allow ~/.envrc`.
func AllowEnvrc(homeDir string) error {
	return exec.Command("direnv", "allow", filepath.Join(homeDir, ".envrc")).Run()
}
```

**Step 4: Run tests to confirm they pass**

```bash
cd /Users/treycaliva/dotfiles && go test ./internal/direnv/... -v
```
Expected: all 3 tests PASS

**Step 5: Commit**

```bash
git add internal/direnv/
git commit -m "direnv: add Setup types and zshrc/template write helpers"
```

---

### Task 2: Add direnv to `config.yaml`

**Files:**
- Modify: `internal/config/config.yaml`

**Step 1: Add the package entry**

Open `internal/config/config.yaml` and add after the `git` entry (before `p10k`):

```yaml
  direnv:
    description: "direnv + 1Password secret injection"
    deps: [direnv, op]
```

The `op` binary name matches the 1Password CLI (`brew install 1password-cli` installs `op`). No `pkg_names` override needed on macOS; Linux users using apt/dnf will need a separate note but `op` install on Linux typically needs the 1Password apt repo — for now just list the dep and let the missing-dep message guide them.

**Step 2: Verify config loads correctly**

```bash
cd /Users/treycaliva/dotfiles && go test ./internal/config/... -v
```
Expected: PASS (existing config tests still pass, direnv now appears in package list)

**Step 3: Commit**

```bash
git add internal/config/config.yaml
git commit -m "config: add direnv package with op dependency"
```

---

### Task 3: Add `ScreenDirenvConfig` constant and `DirenvConfig` to `AppState`

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add the screen constant**

In `app.go`, find the `Screen` iota block (line ~33) and add `ScreenDirenvConfig` after `ScreenPreview`:

```go
const (
    ScreenHome Screen = iota
    ScreenSelect
    ScreenPreview
    ScreenDirenvConfig   // ← add this
    ScreenDiff
    ScreenProgress
    ScreenSummary
)
```

**Step 2: Add `DirenvConfig` to `AppState`**

Add the import and field. At the top of the file add the import:
```go
"github.com/treycaliva/dotfiles/internal/direnv"
```

In the `AppState` struct, add after `Backups`:
```go
// direnv interactive setup — nil when direnv is not selected
DirenvConfig *direnv.Setup
```

**Step 3: Add `ScreenDirenvConfig` to `navigate()`**

In the `navigate` method, add a case after `ScreenPreview`:
```go
case ScreenDirenvConfig:
    a.current = NewDirenvConfigScreen(a.state)
    a.current.SetSize(a.contentW, a.contentH)
```

**Step 4: Add `ScreenDirenvConfig` to `screenName()`**

```go
case ScreenDirenvConfig:
    return "direnv Setup"
```

**Step 5: Verify it compiles** (NewDirenvConfigScreen doesn't exist yet — expect a compile error on that symbol only)

```bash
cd /Users/treycaliva/dotfiles && go build ./... 2>&1 | grep -v NewDirenvConfigScreen
```
Expected: no errors other than the missing `NewDirenvConfigScreen`

**Step 6: Commit**

```bash
git add internal/tui/app.go
git commit -m "tui: add ScreenDirenvConfig constant and AppState.DirenvConfig field"
```

---

### Task 4: Update `preview.go` to route through DirenvConfig

**Files:**
- Modify: `internal/tui/preview.go`

**Step 1: Find the enter key handler** (line ~146 in preview.go)

Replace:
```go
case "enter":
    return p, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
```

With:
```go
case "enter":
    // Route through direnv config screen when direnv is being installed.
    if !p.state.Unstowing {
        for _, pkg := range p.state.Selected {
            if pkg == "direnv" {
                return p, func() tea.Msg { return NavigateMsg{Screen: ScreenDirenvConfig} }
            }
        }
    }
    return p, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
```

**Step 2: Verify it compiles**

```bash
cd /Users/treycaliva/dotfiles && go build ./...
```
Expected: same single undefined-symbol error for `NewDirenvConfigScreen`

**Step 3: Commit**

```bash
git add internal/tui/preview.go
git commit -m "tui: route preview→DirenvConfig when direnv selected"
```

---

### Task 5: Implement `ScreenDirenvConfig`

**Files:**
- Create: `internal/tui/direnvconfig.go`

This is the multi-step form screen. Steps in order:

| step const | what it shows |
|---|---|
| `direnvStepContext` | Tab to toggle personal/work |
| `direnvStepOPAccount` | Free-text input for op account |
| `direnvStepSecretKey` | Free-text input for env var name |
| `direnvStepSecretRef` | Free-text input for op:// ref |
| `direnvStepAddAnother` | y/n prompt |
| `direnvStepConfirm` | Read-only summary, enter proceeds |

**Step 1: Create `internal/tui/direnvconfig.go`**

```go
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
			// Discard config and go back to preview.
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

	// Forward keyboard events to active textinput.
	var v1cmd v1tea.Cmd
	switch d.step {
	case direnvStepOPAccount:
		d.account, v1cmd = d.account.Update(msg)
	case direnvStepSecretKey:
		d.secretKey, v1cmd = d.secretKey.Update(msg)
	case direnvStepSecretRef:
		d.secretRef, v1cmd = d.secretRef.Update(msg)
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
			return d, nil // require non-empty
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
		// enter with no y/n typed = go to confirm
		d.step = direnvStepConfirm

	case direnvStepConfirm:
		// Commit to AppState and proceed to Progress.
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
```

**Step 2: Verify it compiles**

```bash
cd /Users/treycaliva/dotfiles && go build ./...
```
Expected: clean build (no errors)

**Step 3: Update `helpText()` in `app.go` to mention the direnv screen**

In the help text block, add under Navigation:
```
    esc       Go back one screen / discard direnv config
```
(The existing `esc` line already covers this; no change needed.)

Also update the `q` quit guard in `app.go` Update to also block quit on `ScreenDirenvConfig`:
```go
if a.screen != ScreenProgress && a.screen != ScreenDiff && a.screen != ScreenDirenvConfig {
    return a, tea.Quit
}
```

**Step 4: Build and run a quick smoke test**

```bash
cd /Users/treycaliva/dotfiles && go run ./cmd/installer/
```
Select direnv, proceed through preview, confirm the DirenvConfig screen appears and is navigable.

**Step 5: Commit**

```bash
git add internal/tui/direnvconfig.go internal/tui/app.go
git commit -m "tui: add DirenvConfigScreen multi-step form"
```

---

### Task 6: Apply direnv config in Progress post-stow

**Files:**
- Modify: `internal/tui/progress.go`

**Step 1: Add the import**

In `progress.go` imports, add:
```go
"github.com/treycaliva/dotfiles/internal/direnv"
```

**Step 2: Add post-stow direnv configure block**

In `processNext()`, after the validation block (after `validate.Run` call, still inside the `else` branch), add:

```go
// Post-stow direnv configuration (writes ~/.zshrc.local, patches template, allows .envrc).
if pkg == "direnv" && state.DirenvConfig != nil {
    logs = append(logs, fmt.Sprintf("[%s] writing ~/.zshrc.local ...", pkg))
    if err := direnv.WriteZshrcLocal(state.HomeDir, state.DirenvConfig); err != nil {
        logs = append(logs, fmt.Sprintf("[%s] warning: could not write ~/.zshrc.local: %v", pkg, err))
    } else {
        logs = append(logs, fmt.Sprintf("[%s] ~/.zshrc.local updated", pkg))
    }

    logs = append(logs, fmt.Sprintf("[%s] patching template ...", pkg))
    if err := direnv.PatchTemplate(state.HomeDir, state.DirenvConfig); err != nil {
        logs = append(logs, fmt.Sprintf("[%s] warning: could not patch template: %v", pkg, err))
    } else {
        logs = append(logs, fmt.Sprintf("[%s] template updated", pkg))
    }

    logs = append(logs, fmt.Sprintf("[%s] running direnv allow ~/.envrc ...", pkg))
    if err := direnv.AllowEnvrc(state.HomeDir); err != nil {
        logs = append(logs, fmt.Sprintf("[%s] warning: direnv allow failed: %v", pkg, err))
    } else {
        logs = append(logs, fmt.Sprintf("[%s] direnv allow done", pkg))
    }
}
```

Place this block **after** the `validate.Run` block and **before** the closing `}` of the install `else` branch.

**Step 2: Build to confirm no compile errors**

```bash
cd /Users/treycaliva/dotfiles && go build ./...
```
Expected: clean build

**Step 3: Run all tests**

```bash
cd /Users/treycaliva/dotfiles && go test ./...
```
Expected: all pass

**Step 4: Commit**

```bash
git add internal/tui/progress.go
git commit -m "tui: apply direnv config post-stow (zshrc.local, template, direnv allow)"
```

---

### Task 7: Final smoke test and cleanup

**Step 1: Full build + test pass**

```bash
cd /Users/treycaliva/dotfiles && go build ./... && go test ./... -v
```
Expected: clean build, all tests pass

**Step 2: Verify install.sh still has direnv in PACKAGES**

`install.sh` line 364 already lists `direnv` — no changes needed there. The shell script and the TUI are independent; the TUI is the preferred path.

**Step 3: Commit cleanup if anything was missed**

```bash
git add -p  # review any straggling changes
git commit -m "direnv: finalize integration and smoke test"
```

---

## Summary of all files changed

| File | Action |
|------|--------|
| `internal/direnv/setup.go` | Create — types + write helpers |
| `internal/direnv/setup_test.go` | Create — unit tests for helpers |
| `internal/config/config.yaml` | Modify — add direnv package entry |
| `internal/tui/app.go` | Modify — screen constant, AppState field, navigate case, screenName case |
| `internal/tui/preview.go` | Modify — route to DirenvConfig when direnv selected |
| `internal/tui/direnvconfig.go` | Create — multi-step form screen |
| `internal/tui/progress.go` | Modify — post-stow direnv configure block |
