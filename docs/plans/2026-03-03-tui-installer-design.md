# TUI Dotfiles Installer Design

**Date:** 2026-03-03
**Status:** Approved

## Summary

Replace the current POSIX sh `install.sh` interactive menu with a full-featured TUI built in Go using Bubble Tea (Charm ecosystem). The TUI provides checkbox-based package selection, profile presets, dry-run previews, config diff viewing, real-time installation progress, and unstow support. Compiles to a single binary with no runtime dependencies.

## Stack

- **Language:** Go
- **TUI framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm architecture)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss) (Srcery palette)
- **Components:** [Bubbles](https://github.com/charmbracelet/bubbles) + [Huh?](https://github.com/charmbracelet/huh) for form screens
- **Platforms:** macOS, Linux (apt/dnf), WSL

## Project Structure

```
dotfiles/
├── cmd/
│   └── installer/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Package/profile definitions, loaded from YAML
│   │   └── config.yaml          # Declarative package & profile config (go:embed)
│   ├── platform/
│   │   ├── detect.go            # OS, package manager, WSL detection
│   │   └── deps.go              # Dependency installation via brew/apt/dnf
│   ├── stow/
│   │   ├── stow.go              # Stow/unstow operations
│   │   ├── status.go            # Check if package is stowed
│   │   ├── conflict.go          # Conflict detection & backup resolution
│   │   └── diff.go              # Config diff preview
│   ├── tui/
│   │   ├── app.go               # Root Bubble Tea model, screen router
│   │   ├── theme.go             # Srcery color palette via Lip Gloss
│   │   ├── screens/
│   │   │   ├── home.go          # Landing: OS info, package status overview
│   │   │   ├── select.go        # Package selection with profiles
│   │   │   ├── preview.go       # Dry-run: changes, conflicts, deps
│   │   │   ├── diff.go          # Side-by-side diff viewer
│   │   │   ├── progress.go      # Per-package spinners + log viewport
│   │   │   └── summary.go       # Results report
│   │   └── components/
│   │       ├── pkglist.go        # Reusable package list with status
│   │       └── statusbar.go     # Bottom bar: keybindings, current action
│   └── validate/
│       └── validate.go          # Post-install validation
├── go.mod
├── go.sum
├── Makefile
└── install.sh                    # Kept as lightweight fallback
```

## Screen Flow

```
Home → Select → Preview (Dry-run) → Progress → Summary
                  ↕                    ↑
               Diff View               │
                                       │
         Esc/Back navigates ←──────────┘
```

### Home
OS info, package manager, dotfiles path, package status table. `Enter` to proceed.

### Select
Checkbox list of packages with installed/not-installed status. Toggle at top switches between Install and Unstow modes. Profile presets (minimal, full, server) pre-check groups. `Tab` toggles items, `Enter` proceeds.

### Preview (Dry-run)
Shows dependencies to install, files to symlink, conflicts detected, files to back up. Press `d` on a conflict to view its diff. `Enter` confirms, `Esc` goes back.

### Diff View
Unified or side-by-side diff of existing file vs dotfile. `Esc` returns to preview.

### Progress
Per-package row with spinner (active), checkmark (success), or cross (failure). Scrolling log viewport below. Packages install sequentially, UI stays responsive.

### Summary
Succeeded/failed counts, backed-up file paths, validation warnings. `q` quits, `r` returns to home.

### Global Keybindings
- `q` / `Ctrl+C` — quit from any screen
- `Esc` — back one screen (except during progress)
- `?` — toggle help overlay

## Config Format

```yaml
packages:
  zsh:
    description: "Zsh shell config with zinit plugins"
    deps: [zsh, zoxide, fzf]
    validate: "zsh -n ~/.zshrc"
  tmux:
    description: "Tmux config with vim-navigator integration"
    deps: [tmux]
    validate: "tmux -f ~/.tmux.conf new-session -d -s _validate && tmux kill-session -t _validate"
  vim:
    description: "Vim config with vim-plug"
    deps: [vim]
  nvim:
    description: "Neovim (sources ~/.vimrc)"
    deps: [nvim]
    pkg_names:
      nvim: neovim
  ghostty:
    description: "Ghostty terminal emulator config"
    deps: [ghostty]
  git:
    description: "Git config and aliases"
    deps: [git]
  p10k:
    description: "Powerlevel10k prompt theme"
    deps: []

profiles:
  minimal:
    description: "Shell essentials"
    packages: [zsh, git, p10k]
  server:
    description: "Headless / SSH machines"
    packages: [zsh, tmux, vim, git, p10k]
  full:
    description: "Everything"
    packages: [zsh, tmux, vim, nvim, ghostty, git, p10k]
```

- `deps` lists binary names checked with `exec.LookPath`
- `pkg_names` maps binary → package manager name when they differ
- `validate` is an optional shell command run post-stow
- YAML is embedded via `go:embed` — no separate file to distribute
- Profiles are named lists of packages

## Theme (Srcery Palette)

```go
Black       = "#1C1B19"
Red         = "#EF2F27"
Green       = "#519F50"
Yellow      = "#FBB829"
Blue        = "#2C78BF"
Magenta     = "#E02C6D"
Cyan        = "#0AAEB3"
White       = "#BAA67F"
BrightBlack = "#918175"
```

- Borders: rounded, `BrightBlack`, active panel `Cyan`
- Status: `Green` installed, `Yellow` not installed, `Red` failed
- Selection highlight: `Cyan` + bold
- Spinners: `Yellow` active, `Green`/`Red` complete
- Diff: `Green` additions, `Red` deletions
- Status bar: `BrightBlack` background, `White` text
- Headers: bold + `Yellow`
- Nerd Font icons (`, , , `) with plain-text fallback

## Core Operations

### Platform Detection
- OS via `runtime.GOOS`; WSL via `/proc/version`
- Package manager: brew (macOS), apt or dnf (Linux)
- External commands via `os/exec` with output streamed to TUI

### Stow Operations
- Shells out to `stow` / `stow -D` — does not reimplement GNU Stow
- Status check ports `is_stowed()` logic from install.sh
- Conflict detection via `stow --no` (dry-run), parses stderr
- Backup to `~/.dotfiles-backup/` with timestamps
- Diff via `diff -u` between existing file and dotfile

### Dependency Installation
- `exec.LookPath` checks binary presence
- Installs via detected package manager, streams output to progress screen
- `pkg_names` config handles binary→package name translation

### Validation
- Runs optional `validate` command from config via `os/exec`
- Success → checkmark, failure → warning (never blocks)

## Build & Distribution

```makefile
BINARY    := dotfiles-installer
CMD       := ./cmd/installer
GOFLAGS   := -trimpath -ldflags="-s -w"

build:
	go build $(GOFLAGS) -o $(BINARY) $(CMD)

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
```

- Single static binary, no CGO
- `install.sh` kept as minimal fallback for machines without Go
- Binary not committed — `.gitignore`d, built from source
