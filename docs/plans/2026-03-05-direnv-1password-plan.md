# direnv + 1Password Integration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a `direnv` stow package with context-aware 1Password secret injection, zsh hook, and a new Go TUI screen that configures the setup on first install.

**Architecture:** A new `direnv/` stow package provides `direnvrc` (with an `op_inject` helper), starter `.env.tpl` templates for personal/work contexts, and a global `~/.envrc`. The Go TUI gains a `stateDirenvSetup` screen that collects context and 1Password account shorthand, then writes them to `~/.zshrc.local`. All installer actions (stow, direnv allow, op check) are sequenced within the existing `doInstall` flow.

**Tech Stack:** GNU Stow, direnv, 1Password CLI (`op`), Go 1.24, Bubble Tea (charmbracelet/bubbletea), Bubbles textinput

**Design doc:** `docs/plans/2026-03-05-direnv-1password-design.md`

---

### Task 1: Create the direnv stow package

**Files:**
- Create: `direnv/.config/direnv/direnvrc`
- Create: `direnv/.config/direnv/templates/personal.env.tpl`
- Create: `direnv/.config/direnv/templates/work.env.tpl`
- Create: `direnv/.envrc`

**Step 1: Create the package directory tree**

```bash
mkdir -p direnv/.config/direnv/templates
```

**Step 2: Create `direnv/.config/direnv/direnvrc`**

This is sourced automatically by direnv before every `.envrc`. It defines the `op_inject` helper.

```bash
# direnvrc — sourced by direnv before every .envrc
# Provides op_inject: pulls secrets from 1Password into the environment.
#
# Required env vars (set in ~/.zshrc.local by the dotfiles installer):
#   DOTFILES_CONTEXT     — "personal" or "work" (default: personal)
#   DOTFILES_OP_ACCOUNT  — op account shorthand, e.g. my.1password.com

op_inject() {
  local context="${DOTFILES_CONTEXT:-personal}"
  local account="${DOTFILES_OP_ACCOUNT:-}"
  local template="${HOME}/.config/direnv/templates/${context}.env.tpl"

  if [ ! -f "$template" ]; then
    echo "direnv: template not found: $template" >&2
    echo "direnv: set DOTFILES_CONTEXT to 'personal' or 'work' in ~/.zshrc.local" >&2
    return 1
  fi

  if [ -z "$account" ]; then
    echo "direnv: DOTFILES_OP_ACCOUNT is not set" >&2
    echo "direnv: run 'op account list' and set DOTFILES_OP_ACCOUNT in ~/.zshrc.local" >&2
    return 1
  fi

  eval "$(op inject --account "$account" -i "$template" | sed 's/^/export /')"
}
```

**Step 3: Create `direnv/.config/direnv/templates/personal.env.tpl`**

```bash
# Personal 1Password secret template
# Uncomment and populate with your op:// references.
# Format: export KEY={{ op://VaultName/ItemName/field }}
#
# Examples:
# export GITHUB_TOKEN={{ op://Personal/GitHub Token/token }}
# export AWS_ACCESS_KEY_ID={{ op://Personal/AWS Personal/access_key_id }}
# export AWS_SECRET_ACCESS_KEY={{ op://Personal/AWS Personal/secret_access_key }}
```

**Step 4: Create `direnv/.config/direnv/templates/work.env.tpl`**

```bash
# Work 1Password secret template
# Uncomment and populate with your op:// references.
# Format: export KEY={{ op://VaultName/ItemName/field }}
#
# Examples:
# export GITHUB_TOKEN={{ op://Work/GitHub Token/token }}
# export AWS_ACCESS_KEY_ID={{ op://Work/AWS Work/access_key_id }}
# export AWS_SECRET_ACCESS_KEY={{ op://Work/AWS Work/secret_access_key }}
```

**Step 5: Create `direnv/.envrc`**

```bash
# Global ~/.envrc — loads context-specific secrets from 1Password.
# Managed by dotfiles. Edit templates in ~/.config/direnv/templates/.

# Watch the active template so direnv reloads when it changes.
context="${DOTFILES_CONTEXT:-personal}"
watch_file "${HOME}/.config/direnv/templates/${context}.env.tpl"

op_inject
```

**Step 6: Verify stow would work (dry run)**

```bash
stow -d . -t "$HOME" --no direnv
```

Expected: no errors. If conflicts exist, note the conflicting files — the TUI handles them automatically.

**Step 7: Commit**

```bash
git add direnv/
git commit -m "direnv: add stow package with op_inject helper and starter templates"
```

---

### Task 2: Add direnv hook to zsh config

**Files:**
- Modify: `zsh/.zshrc` (near the end, before `~/.zshrc.local` sourcing)

**Step 1: Open `zsh/.zshrc` and locate the local overrides line**

The file ends with:
```zsh
# Load local overrides (e.g. machine-specific aliases or exports)
[[ -f ~/.zshrc.local ]] && source ~/.zshrc.local
```

**Step 2: Insert the direnv hook before that block**

Add these lines immediately before the `~/.zshrc.local` sourcing:

```zsh
# direnv — must come before .zshrc.local so DOTFILES_CONTEXT is available
eval "$(direnv hook zsh)"
```

The final lines of `.zshrc` should look like:

```zsh
# direnv — must come before .zshrc.local so DOTFILES_CONTEXT is available
eval "$(direnv hook zsh)"

# Load local overrides (e.g. machine-specific aliases or exports)
[[ -f ~/.zshrc.local ]] && source ~/.zshrc.local
```

**Step 3: Syntax-check**

```bash
zsh -n zsh/.zshrc
```

Expected: no output (no errors).

**Step 4: Commit**

```bash
git add zsh/.zshrc
git commit -m "zsh: add direnv hook"
```

---

### Task 3: Add direnv state constants and model fields to the TUI

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Add `stateDirenvSetup` to the `state` const block**

Current const block (lines 18-24):
```go
const (
    stateProfileSelection state = iota
    stateGitName
    stateGitEmail
    stateInstalling
    stateDone
)
```

Replace with:
```go
const (
    stateProfileSelection state = iota
    stateGitName
    stateGitEmail
    stateDirenvSetup
    stateInstalling
    stateDone
)
```

**Step 2: Add direnv-related fields to the `model` struct**

Current struct (lines 48-56):
```go
type model struct {
    state          state
    cursor         int
    selectedProf   *Profile
    nameInput      textinput.Model
    emailInput     textinput.Model
    installLog     []string
    err            error
}
```

Replace with:
```go
type model struct {
    state           state
    cursor          int
    selectedProf    *Profile
    nameInput       textinput.Model
    emailInput      textinput.Model
    direnvCtxCursor int    // 0=personal, 1=work
    opAccountInput  textinput.Model
    direnvStep      int    // 0=context selection, 1=account input
    installLog      []string
    err             error
}
```

**Step 3: Initialize `opAccountInput` in `initialModel`**

Current `initialModel` (lines 58-71):
```go
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
```

Replace with:
```go
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
```

**Step 4: Wire `stateDirenvSetup` into the `Update` switch**

In the `Update` function, the state switch (lines 94-108) currently goes:
```go
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
    return m, nil
}
```

Replace with:
```go
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
    return m, nil
}
```

**Step 5: Update `updateGitEmail` to transition to `stateDirenvSetup`**

Currently on successful Enter it sets `m.state = stateInstalling`. Change it to:
```go
m.state = stateDirenvSetup
m.opAccountInput.Focus()
return m, textinput.Blink
```

**Step 6: Build to check for compile errors**

```bash
go build ./cmd/installer/
```

Expected: compile error about missing `updateDirenvSetup` method — that's fine, we add it next.

---

### Task 4: Implement the direnv setup update and view logic

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Add `updateDirenvSetup` method**

Add this method after `updateGitEmail`:

```go
var direnvContexts = []string{"personal", "work"}

func (m model) updateDirenvSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.direnvStep == 0 {
            // Context selection via arrow keys
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
            // Account shorthand input
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
```

**Step 2: Add direnv setup view to the `View` method**

In the `View` switch, add a case for `stateDirenvSetup` after the `stateGitEmail` case:

```go
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
```

**Step 3: Build and verify it compiles**

```bash
go build ./cmd/installer/
```

Expected: clean build with no errors.

**Step 4: Smoke test the TUI flow manually**

```bash
./dotfiles-installer
```

Walk through: profile → git name → git email → direnv context selection → account input → (cancel before install). Verify each screen renders correctly and navigation works.

**Step 5: Commit**

```bash
git add cmd/installer/main.go
git commit -m "installer: add direnv setup screen with context and op account inputs"
```

---

### Task 5: Implement direnv installer actions in `doInstall`

**Files:**
- Modify: `cmd/installer/main.go`

The `doInstall` function runs as a `tea.Cmd` (goroutine). We need to add three actions after the stow loop completes:
1. Write `DOTFILES_CONTEXT` and `DOTFILES_OP_ACCOUNT` to `~/.zshrc.local`
2. Run `direnv allow ~/.envrc`
3. Check `op account list` (warn-only)

**Step 1: Add a helper to write/update a variable in `~/.zshrc.local`**

Add this function before `doInstall`:

```go
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
```

**Step 2: Add direnv actions at the end of `doInstall`, after the stow loop**

Inside the `doInstall` closure, after the package stow loop and before `return installMsg(...)`, add:

```go
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
```

**Step 3: Build**

```bash
go build ./cmd/installer/
```

Expected: clean build.

**Step 4: Commit**

```bash
git add cmd/installer/main.go
git commit -m "installer: write zshrc.local vars and run direnv allow post-install"
```

---

### Task 6: Add direnv to all profiles

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Update the `profiles` var**

Current:
```go
var profiles = []Profile{
    {Name: "Base",    Description: "Core CLI tools (zsh, tmux, vim, git)",               Packages: []string{"zsh", "tmux", "vim", "git", "p10k"}},
    {Name: "Desktop", Description: "Base + GUI terminals (ghostty, alacritty)",           Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty"}},
    {Name: "Dev",     Description: "Desktop + Dev tools (nvim)",                          Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "nvim"}},
}
```

Replace with:
```go
var profiles = []Profile{
    {Name: "Base",    Description: "Core CLI tools (zsh, tmux, vim, git, direnv)",        Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "direnv"}},
    {Name: "Desktop", Description: "Base + GUI terminals (ghostty, alacritty)",           Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "direnv"}},
    {Name: "Dev",     Description: "Desktop + Dev tools (nvim)",                          Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "nvim", "direnv"}},
}
```

**Step 2: Build**

```bash
go build ./cmd/installer/
```

Expected: clean build.

**Step 3: Commit**

```bash
git add cmd/installer/main.go
git commit -m "installer: add direnv to all installation profiles"
```

---

### Task 7: End-to-end manual verification

**Goal:** Confirm the full flow works on a real machine before opening a PR.

**Step 1: Install direnv if not present**

```bash
brew install direnv
```

**Step 2: Restow your dotfiles to pick up the new direnv package**

```bash
stow -d ~/dotfiles -t "$HOME" --restow direnv zsh
```

**Step 3: Reload your shell**

```bash
exec zsh
```

**Step 4: Verify the hook is active**

```bash
direnv version
```

Expected: version string printed (e.g. `2.x.x`).

**Step 5: Verify `~/.envrc` exists and is a symlink**

```bash
ls -la ~/.envrc
```

Expected: `~/.envrc -> ~/dotfiles/direnv/.envrc`

**Step 6: Verify `direnvrc` is in place**

```bash
ls -la ~/.config/direnv/direnvrc
```

Expected: symlink to `~/dotfiles/direnv/.config/direnv/direnvrc`

**Step 7: Set context variables in `~/.zshrc.local` manually (simulate TUI output)**

```bash
echo 'export DOTFILES_CONTEXT=personal' >> ~/.zshrc.local
echo 'export DOTFILES_OP_ACCOUNT=your.1password.com' >> ~/.zshrc.local
```

Replace `your.1password.com` with your actual shorthand from `op account list`.

**Step 8: Allow the global envrc**

```bash
direnv allow ~/.envrc
```

**Step 9: Reload and verify op_inject runs without error**

```bash
exec zsh
```

Expected: if 1Password is signed in and your account shorthand is correct, secrets from the template load silently. If the template has no uncommented exports, direnv loads with no output — that's correct.

**Step 10: Run the TUI installer and walk through the full flow**

```bash
cd ~/dotfiles && go run ./cmd/installer/
```

Walk through all screens through to completion. Verify install log shows:
- `Stowing direnv...` + success
- `Wrote DOTFILES_CONTEXT=... and DOTFILES_OP_ACCOUNT=...`
- `Ran: direnv allow ~/.envrc`
- `1Password CLI: account found` (or the warning if not signed in)

**Step 11: Commit any fixes found during verification, then open PR**

```bash
git push origin add-direnv
gh pr create --title "direnv: add 1Password-backed secret injection" \
  --body "Adds direnv stow package with op_inject helper, context-aware templates, zsh hook, and TUI setup screen."
```
