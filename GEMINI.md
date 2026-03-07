# Dotfiles

This repository contains personal configuration files (dotfiles) managed using **GNU Stow**. It is designed to be portable across macOS and Linux, providing a consistent development environment.

## Project Overview

- **Management Strategy**: Uses GNU Stow to symlink configuration "packages" from this directory into `$HOME`.
- **Platform Support**: macOS (Homebrew) and Linux (apt, dnf).
- **Core Aesthetic**:
    - **Color Scheme**: [Srcery](https://srcery-colors.github.io/) (consistent across Vim, Tmux, and Terminal).
    - **Font**: BlexMono Nerd Font Mono.

## Repository Structure

Each top-level directory is a "stow package" whose contents are mirrored to `$HOME`:

| Package | Main Config File(s) | Description |
| :--- | :--- | :--- |
| `zsh/` | `.zshrc` | Shell config using `zinit` and `p10k`. |
| `tmux/` | `.tmux.conf` | Terminal multiplexer with `C-a` prefix and vim-navigation. |
| `vim/` | `.vimrc` | Base Vim configuration. |
| `nvim/` | `.config/nvim/init.vim` | Neovim config (sources `~/.vimrc`). |
| `ghostty/` | `.../com.mitchellh.ghostty/config` | Ghostty terminal configuration. |
| `alacritty/`| `alacritty.toml` | Alacritty terminal configuration. |
| `git/` | `.gitconfig` | Global git settings and aliases. |
| `direnv/` | `.envrc`, `direnvrc` | Environment management with 1Password templates. |
| `p10k/` | `p10k.zsh` | Powerlevel10k prompt configuration. |

## Usage & Commands

### Go TUI Installer (Recommended)
A modern terminal-based installer that validates dependencies, previews changes, and handles interactive configurations like `git` identity and `direnv` 1Password secrets.

```bash
# Build the installer
make build

# Run it
./dotfiles-installer
```

### Legacy Bootstrap (POSIX Shell)
The original `install.sh` script is maintained for environments where Go is not available.
```bash
./install.sh
```

### Manual Stow Management
```bash
# Link a package
stow <package_name>

# Unlink a package
stow -D <package_name>

# Restow (refresh links)
stow -R <package_name>
```

### Validation
- **Go Workspace**: `make test`
- **Zsh**: `zsh -n zsh/.zshrc` (syntax check)
- **Tmux**: `tmux -f tmux/.tmux.conf new -d && tmux kill-server`
- **Vim/Nvim**: `nvim +PlugInstall +qa`

## Development Conventions

- **Shell Scripts**: Must remain POSIX sh compatible (no bashisms) for portability, especially `install.sh`.
- **Logic**: Use lowercase function names and always quote variables (`"$VAR"`).
- **Architecture**: Keep package boundaries clean. Tool-specific configurations should reside in their respective directories.
- **Commit Messages**: Follow the format `area: concise imperative action` (e.g., `tmux: remap prefix to C-a`).
- **Plugins**: Both `zinit` (Zsh) and `vim-plug` (Vim/Nvim) are configured to auto-bootstrap if missing.

## Active Projects
- **Go TUI Installer**: A modern, interactive replacement for `install.sh` built with Go, Bubble Tea v2, and LipGloss. It supports dependency validation, previewing changes, and custom setup screens for `git` (identity) and `direnv` (1Password integration). See `docs/plans/` for architectural details.
