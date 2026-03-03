# Bootstrap Script Design

## Overview

A single POSIX shell script (`install.sh`) at the repo root that automates setting up a new machine from this dotfiles repo. It detects the OS, installs missing dependencies, and interactively links Stow packages into `$HOME`.

## Prerequisites & OS Detection

- Detects macOS vs Linux via `uname -s`
- macOS: requires Homebrew (`brew`) — exits with a clear message if missing
- Linux: detects `apt` or `dnf` — exits if neither found
- If `stow` is not installed, offers to install it via the detected package manager

The script uses `#!/bin/sh` for POSIX compatibility on minimal Linux systems.

## Dependency Map

Each Stow package maps to system dependencies the script installs:

| Package     | Dependencies          |
|-------------|-----------------------|
| `zsh`       | zsh, zoxide, fzf      |
| `tmux`      | tmux                  |
| `vim`       | vim                   |
| `nvim`      | neovim                |
| `alacritty` | alacritty             |
| `git`       | git                   |
| `p10k`      | (none — sourced by zsh) |
| `zprezto`   | (none — sourced by zsh) |

Binary names are mapped to platform-specific package names (e.g., `neovim` on Homebrew vs `neovim` on apt).

## Interactive Menu

When run without arguments, the script shows:

```
[1] zsh        (installed)
[2] tmux       (installed)
[3] vim        (not installed)
[4] nvim       (not installed)
[5] alacritty  (installed)
[6] git        (installed)
[7] p10k       (installed)
[8] zprezto    (installed)

[a] Install all    [q] Quit
Enter choices (e.g. 3 4 or a):
```

"Installed" means the stow symlinks exist in `$HOME` and point into this repo. The user enters numbers separated by spaces, or `a` for all.

For each selected package, the script:
1. Installs missing system dependencies via the detected package manager
2. Runs `stow <package>` to create symlinks
3. Reports success/failure per package

## Conflict Handling

If `stow` detects an existing file that isn't a symlink (e.g., a real `~/.zshrc`), the script:
1. Backs it up to `~/.dotfiles-backup/<filename>.bak.<timestamp>`
2. Prompts the user to confirm before replacing
3. Re-runs stow after backup

## Post-Install Validation

After stowing, runs basic validation where possible:
- `zsh -n zsh/.zshrc` — shell syntax check
- `tmux -f tmux/.tmux.conf new -d && tmux kill-server` — tmux config validation

These match the workflows documented in AGENTS.md.

## Additional Changes

- Add `.gitignore` to exclude `firebase-debug.log`, `.DS_Store`, and `.claude/`
- Script is `chmod +x` at the repo root as `install.sh`

## Out of Scope

- No uninstall/remove command
- No Brewfile or package lockfile
- No automatic shell switching (`chsh`)
