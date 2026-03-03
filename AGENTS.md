# Repository Guidelines

## Project Structure & Module Organization
This repository is a GNU Stow-style dotfiles collection. Each top-level directory is a package that maps files into `$HOME`:
- `zsh/` (`.zshrc`) and `zprezto/` (`zpreztorc`) for shell behavior
- `tmux/` (`.tmux.conf`, `tat`) for terminal multiplexing
- `vim/` (`.vimrc`) and `nvim/` (`init.vim`) for editor config
- `alacritty/` (`alacritty.toml`) for terminal appearance
- `git/` (`.gitconfig`) and `p10k/` (`p10k.zsh`) for tooling and prompt

Keep package boundaries clear: add new config under the matching tool directory instead of mixing files across packages.

## Build, Test, and Development Commands
There is no build step. Typical workflows are:
- `stow zsh tmux vim nvim alacritty git p10k zprezto` to symlink packages into `$HOME`
- `stow -D zsh && stow zsh` to safely restow one package
- `zsh -n zsh/.zshrc` to syntax-check shell config
- `tmux -f tmux/.tmux.conf new -d && tmux kill-server` to validate tmux config
- `nvim +PlugInstall +qa` to install/update Vim/Neovim plugins

## Coding Style & Naming Conventions
Use existing style per file type:
- Shell: POSIX/Zsh-compatible syntax, lowercase function names, quote variables (`"$VAR"`)
- Vimscript/Tmux: keep current indentation and option style (`set`, `let g:`)
- TOML: preserve section grouping and key ordering where practical

Prefer small, focused edits. Keep comments short and only where behavior is non-obvious.

## Testing Guidelines
No automated test suite exists. Validate changes by loading the target tool:
- Shell changes: open a new shell and run core aliases/functions
- Tmux changes: reload config (`tmux source-file ~/.tmux.conf`) and check keybinds
- Vim/Neovim changes: open files for affected languages and verify plugins/mappings

## Commit & Pull Request Guidelines
Recent history favors short, imperative commit messages (for example, `Update vimrc`, `Updated dotfiles`). Use:
- `Area: concise action` when possible (example: `tmux: tune pane resize keys`)
- One logical change per commit

For PRs, include:
- What changed and why
- Any manual verification steps run
- Screenshots or terminal captures for visual/UI-facing changes (prompt/theme/layout)
