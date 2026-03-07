# direnv TUI Configuration Screen — Design

## Overview

Add interactive direnv setup to the Go TUI installer. When the user selects the `direnv` package, a new `ScreenDirenvConfig` is inserted between the Preview and Progress screens to collect 1Password account details and secrets before installation begins.

## Data Model

### `internal/config/config.yaml`

Add direnv as a managed package with `op` as a required dependency:

```yaml
direnv:
  description: "direnv + 1Password secret injection"
  deps: [direnv, op]
```

### New type: `internal/direnv/setup.go`

```go
package direnv

type Secret struct {
    Key   string // e.g. GITHUB_TOKEN
    OPRef string // e.g. op://Personal/GitHub Token/token
}

type Setup struct {
    Context   string   // "personal" or "work"
    OPAccount string   // e.g. my.1password.com
    Secrets   []Secret
}
```

### `AppState` addition (`internal/tui/app.go`)

```go
DirenvConfig *direnv.Setup // nil if direnv not selected
```

## Navigation Flow

```
Home → Select → Preview → [DirenvConfig] → Progress → Summary
                                ↑
                    Only when direnv is in state.Selected
                    and state.Unstowing == false
```

The `navigate` function in `app.go` checks whether direnv is selected when transitioning from Preview; if so, it goes to `ScreenDirenvConfig` instead of `ScreenProgress`.

## ScreenDirenvConfig

### Steps (in order)

1. **Context** — radio-style toggle: `personal` / `work` (default: `personal`)
2. **OP Account** — free-text `textinput` for the `op` account shorthand (e.g. `my.1password.com`)
3. **Secrets loop** — repeating pair:
   - Secret key name (e.g. `GITHUB_TOKEN`)
   - `op://` reference (e.g. `op://Personal/GitHub Token/token`)
   - Prompt: "Add another secret? (y/n)"
4. **Confirm** — read-only summary of what will be written; `enter` proceeds to Progress

### Keybindings

| Key | Action |
|-----|--------|
| `enter` | Advance step / confirm |
| `esc` | Go back to Preview (discards all input) |
| `tab` | Toggle context (step 1) |

### State stored in AppState

On confirm, the collected values are written to `state.DirenvConfig`. No files are touched until the Progress screen runs.

## Progress Screen Changes

After successfully stowing `direnv`, if `state.DirenvConfig != nil`, the install worker runs these additional steps (logged as `[direnv] ...`):

1. **Write `~/.zshrc.local`** — append (or create) the file with:
   ```sh
   export DOTFILES_CONTEXT=<context>
   export DOTFILES_OP_ACCOUNT=<account>
   ```
   Existing entries for these keys are updated rather than duplicated.

2. **Patch the template** — write secrets as `op://` references into
   `~/.config/direnv/templates/<context>.env.tpl`, preserving any existing
   comment header lines.

3. **`direnv allow ~/.envrc`** — run so direnv activates the home `.envrc`
   without manual approval.

All three steps are non-fatal warnings if they fail (matching the existing `validate_pkg` behaviour); the package is still marked as installed.

## Error Handling

- If `op` is not installed, the dependency install step in Progress handles it (same as all other deps).
- If `direnv allow` fails, log a warning and continue — the user can run it manually.
- `~/.zshrc.local` writes are atomic (write to temp file, rename).

## Files Changed / Created

| File | Change |
|------|--------|
| `internal/config/config.yaml` | Add `direnv` package entry |
| `internal/direnv/setup.go` | New package with `Setup` and `Secret` types |
| `internal/tui/app.go` | Add `DirenvConfig` to `AppState`; add `ScreenDirenvConfig` to navigation |
| `internal/tui/direnvconfig.go` | New screen implementation |
| `internal/tui/progress.go` | Post-stow direnv configure step |
