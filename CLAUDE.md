# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A GNU Stow-based dotfiles collection. Each top-level directory is a **stow package** whose contents get symlinked into `$HOME`. The interactive `install.sh` script handles dependency installation, stowing, and conflict resolution across macOS and Linux.

**Packages:** `zsh`, `tmux`, `vim`, `nvim`, `ghostty`, `git`, `p10k`

## Common Commands

```bash
# Bootstrap installer (interactive menu)
./install.sh

# Stow/unstow packages manually
stow zsh tmux vim nvim ghostty git p10k
stow -D zsh && stow zsh          # restow a single package

# Validation
zsh -n zsh/.zshrc                 # syntax-check zshrc
tmux -f tmux/.tmux.conf new -d && tmux kill-server  # validate tmux config
nvim +PlugInstall +qa             # install/update vim plugins
```

## Architecture

- **install.sh** — POSIX sh bootstrap script (~440 lines). Detects OS/package manager, maps each stow package to its binary dependencies (`deps_for_pkg()`), installs missing deps, runs stow with conflict resolution (backs up existing files to `~/.dotfiles-backup/`), and runs post-install validation.
- **Srcery color theme** is used consistently across tmux, vim, and ghostty configs.
- **Plugin managers auto-bootstrap**: Zinit (zsh) and vim-plug (vim/nvim) both download themselves on first run if missing.
- **nvim** and **ghostty** use deeper directory trees (`.config/nvim/init.vim` and `Library/Application Support/com.mitchellh.ghostty/config`) because stow mirrors the full path into `$HOME`.
- **nvim/.config/nvim/init.vim** simply sources `~/.vimrc` — vim and nvim share one config.

## Key Conventions

- Color scheme is **Srcery** across terminal, vim, and tmux
- Font is **BlexMono Nerd Font Mono**
- Vim plugin manager is **vim-plug** (`~/.vim/plugged`)
- Zsh plugin manager is **zinit** (migrated from Prezto)
- Tmux prefix is `C-a` (not default `C-b`)
- Vim-tmux-navigator provides seamless `C-h/j/k/l` pane/split navigation between tmux and vim

## Style

- Shell scripts: POSIX-compatible, lowercase function names, always quote variables (`"$VAR"`)
- install.sh must stay POSIX sh (no bashisms) for portability
- Small, focused edits; comments only where behavior is non-obvious
- Commit messages: `area: concise imperative action` (e.g., `tmux: tune pane resize keys`)
- One logical change per commit
