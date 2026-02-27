# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A GNU Stow-based dotfiles collection. Each top-level directory is a **stow package** whose contents get symlinked into `$HOME`. The interactive `install.sh` script handles dependency installation, stowing, and conflict resolution across macOS and Linux.

**Packages:** `zsh`, `tmux`, `vim`, `nvim`, `alacritty`, `ghostty`, `git`, `p10k`, `zprezto`

## Common Commands

```bash
# Bootstrap installer (interactive menu)
./install.sh

# Stow/unstow packages manually
stow zsh tmux vim nvim alacritty ghostty git p10k zprezto
stow -D zsh && stow zsh          # restow a single package

# Validation
zsh -n zsh/.zshrc                 # syntax-check zshrc
tmux -f tmux/.tmux.conf new -d && tmux kill-server  # validate tmux config
nvim +PlugInstall +qa             # install/update vim plugins
```

## Architecture

- **install.sh** — POSIX sh bootstrap script (~440 lines). Detects OS/package manager, maps each stow package to its binary dependencies (`deps_for_pkg()`), installs missing deps, runs stow with conflict resolution (backs up existing files to `~/.dotfiles-backup/`), and runs post-install validation.
- **Srcery color theme** is used consistently across tmux, vim, alacritty, and ghostty configs.
- **Plugin managers auto-bootstrap**: Zinit (zsh) and vim-plug (vim/nvim) both download themselves on first run if missing.
- **nvim/init.vim** simply sources `~/.vimrc` — vim and nvim share one config.
- **ghostty/** uses a deeper directory tree (`Library/Application Support/com.mitchellh.ghostty/config`) because stow mirrors the full path into `$HOME`.

## Style

- Shell scripts: POSIX-compatible, lowercase function names, always quote variables (`"$VAR"`)
- install.sh must stay POSIX sh (no bashisms) for portability
- Small, focused edits; comments only where behavior is non-obvious
- Commit messages: `area: concise imperative action` (e.g., `tmux: tune pane resize keys`)
- One logical change per commit
