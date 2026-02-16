# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a personal dotfiles repository managed with [GNU Stow](https://www.gnu.org/software/stow/). Each top-level directory is a stow package containing config files that get symlinked to `$HOME`.

## Stow Usage

To symlink a package (e.g., `zsh`):
```sh
stow zsh        # symlinks zsh/.zshrc -> ~/.zshrc
stow -D zsh     # removes the symlink
```

Each directory maps to a stow package. The directory structure mirrors the home directory layout — dotfiles inside each package are placed at the root level (e.g., `git/.gitconfig` becomes `~/.gitconfig`).

## Packages

- **alacritty** — Terminal emulator config (Srcery color scheme, BlexMono Nerd Font)
- **git** — Git user config and aliases (`ac`, `co`, `rename`)
- **nvim** — Neovim config (sources `~/.vimrc`)
- **p10k** — Powerlevel10k zsh prompt theme config
- **tmux** — Tmux config (prefix: `C-a`, vim-tmux-navigator integration) and `tat` script for session management
- **vim** — Vim/Neovim config using vim-plug; Srcery colorscheme, CoC for completion, fzf integration, leader is `<Space>`
- **zsh** — Zsh config using zinit plugin manager, Powerlevel10k prompt, with fzf, nvm, terraform, and gcloud setup

## Key Conventions

- Color scheme is **Srcery** across terminal, vim, and tmux
- Font is **BlexMono Nerd Font Mono**
- Vim plugin manager is **vim-plug** (`~/.vim/plugged`)
- Zsh plugin manager is **zinit** (migrated from Prezto)
- Tmux prefix is `C-a` (not default `C-b`)
- Vim-tmux-navigator provides seamless `C-h/j/k/l` pane/split navigation between tmux and vim
