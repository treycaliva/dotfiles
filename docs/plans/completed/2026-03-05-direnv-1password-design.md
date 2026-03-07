# direnv + 1Password Integration Design

**Date:** 2026-03-05
**Status:** Approved
**Branch:** add-direnv

## Summary

Add direnv as a stow package with context-aware 1Password secret injection via `op inject`. Supports personal and work machines (separate 1Password accounts) using a `DOTFILES_CONTEXT` environment variable set in the untracked `~/.zshrc.local`. The Go TUI installer gains a direnv setup screen to configure context and account on first run.

## Decisions

- **Injection pattern:** `op inject` with `.env.tpl` template files — keeps secrets out of the repo, vault references explicit in version-controlled files
- **Context selection:** `DOTFILES_CONTEXT` env var in `~/.zshrc.local` (consistent with existing local-override pattern)
- **Account model:** Supports both multi-account (personal + work as separate 1Password accounts) and single-account-multi-vault via the same `op://Vault/Item/Field` reference syntax
- **Scope:** Global `~/.envrc` for common tokens; per-project `.envrc` + `.env.tpl` files for project-specific secrets (standard direnv usage)

## Package Structure

```
direnv/
├── .config/
│   └── direnv/
│       ├── direnvrc                  # global lib — defines op_inject helper
│       └── templates/
│           ├── personal.env.tpl      # starter personal template
│           └── work.env.tpl          # starter work template
└── .envrc                            # global envrc → symlinked to ~/.envrc
```

### `direnvrc`

Defines a reusable `op_inject` shell function:
- Reads `DOTFILES_CONTEXT` (defaults to `personal`)
- Reads `DOTFILES_OP_ACCOUNT` for the `op` account shorthand
- Runs `op inject --account "$DOTFILES_OP_ACCOUNT" -i ~/.config/direnv/templates/"$DOTFILES_CONTEXT".env.tpl | dotenv`

### `~/.envrc` (global)

Calls `op_inject` and sets `watch_file` on the active template so direnv re-evaluates on template changes.

### Template files

Version-controlled stubs with commented `op://` reference examples. No secrets in the repo.

```bash
# Example: personal.env.tpl
# export GITHUB_TOKEN={{ op://Personal/GitHub/token }}
# export AWS_ACCESS_KEY_ID={{ op://Personal/AWS/access_key_id }}
```

## Zsh Integration

One addition to `zsh/.zshrc` (near the end):

```zsh
# direnv
eval "$(direnv hook zsh)"
```

Machine-specific values written to untracked `~/.zshrc.local` by the TUI:

```zsh
export DOTFILES_CONTEXT=personal          # or "work"
export DOTFILES_OP_ACCOUNT=my.1password.com
```

## Go TUI Integration

### New screen: Direnv Setup

Inserted between `stateGitEmail` and `stateInstalling`.

```
Profile Selection -> Git Name -> Git Email -> Direnv Setup -> Installing -> Done
```

**Step 1 - Context:** Arrow-key selection between `personal` and `work`.

**Step 2 - 1Password account shorthand:** Text input, pre-filled with `my.1password.com`. Helper note: "Run `op account list` to find your shorthand."

### Installer actions

1. Append/update `DOTFILES_CONTEXT` and `DOTFILES_OP_ACCOUNT` in `~/.zshrc.local`
2. Stow the `direnv` package (same conflict resolution as other packages)
3. Run `direnv allow ~/.envrc`
4. Check `op account list` — warn in install log if no account found, do not fail hard

### Profile updates

`direnv` added to all profiles:

```go
{Name: "Base",    Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "direnv"}},
{Name: "Desktop", Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "direnv"}},
{Name: "Dev",     Packages: []string{"zsh", "tmux", "vim", "git", "p10k", "ghostty", "alacritty", "nvim", "direnv"}},
```

## Per-Project Usage (Post-Install)

For project-specific secrets, users create two files in the project root (both git-ignored):

```bash
# .envrc
op_inject   # calls the helper from direnvrc
```

```bash
# .env.tpl
# export DATABASE_URL={{ op://Work/MyProject DB/url }}
```

Run `direnv allow` once per project. direnv handles the rest on `cd`.

## Error Handling

- `op` not signed in: warn in install log, suggest `op signin`; do not block installation
- `direnv allow` fails: warn in install log, user can run manually
- Template file missing for context: `direnvrc` prints a clear error message with the expected path

## Out of Scope

- Auto-generating per-project `.env.tpl` files
- Syncing 1Password item names across machines
- Option C (context directory tree) — can be grown into from this foundation
