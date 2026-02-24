#!/bin/sh
set -e

# ── Dotfiles directory ────────────────────────────────────────────────
DOTFILES_DIR="$(cd "$(dirname "$0")" && pwd)"

# ── Color / formatting helpers ────────────────────────────────────────
# Fall back to plain text when the terminal does not support colors or
# when stdout is not a tty (e.g., piped or running in CI).
if [ -t 1 ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]; then
    FMT_GREEN=$(tput setaf 2)
    FMT_RED=$(tput setaf 1)
    FMT_YELLOW=$(tput setaf 3)
    FMT_BOLD=$(tput bold)
    FMT_RESET=$(tput sgr0)
else
    FMT_GREEN=""
    FMT_RED=""
    FMT_YELLOW=""
    FMT_BOLD=""
    FMT_RESET=""
fi

CHECKMARK="${FMT_GREEN}✔${FMT_RESET}"
CROSS="${FMT_RED}✘${FMT_RESET}"
WARNING="${FMT_YELLOW}⚠${FMT_RESET}"

# ── Output helpers ────────────────────────────────────────────────────
info() {
    printf '%s %s\n' "$CHECKMARK" "$*"
}

warn() {
    printf '%s %s\n' "$WARNING" "$*" >&2
}

err() {
    printf '%s %s\n' "$CROSS" "$*" >&2
}

# ── Utility ───────────────────────────────────────────────────────────
has_cmd() {
    command -v "$1" >/dev/null 2>&1
}

# ── OS detection ──────────────────────────────────────────────────────
detect_os() {
    case "$(uname -s)" in
        Darwin) OS="macos" ;;
        Linux)  OS="linux" ;;
        *)
            err "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
}

# ── Package-manager detection ─────────────────────────────────────────
detect_pkg_manager() {
    case "$OS" in
        macos)
            if has_cmd brew; then
                PKG_MGR="brew"
            else
                err "Homebrew is not installed."
                err "Install it first: https://brew.sh"
                exit 1
            fi
            ;;
        linux)
            if has_cmd apt; then
                PKG_MGR="apt"
            elif has_cmd dnf; then
                PKG_MGR="dnf"
            else
                err "No supported package manager found (apt or dnf)."
                err "Please install one and try again."
                exit 1
            fi
            ;;
    esac
}

# ── Stow check ────────────────────────────────────────────────────────
ensure_stow() {
    if has_cmd stow; then
        info "stow is installed"
        return
    fi

    warn "stow is not installed."

    if [ ! -t 0 ]; then
        err "Non-interactive shell -- cannot prompt. Please install stow manually:"
        err "  $PKG_MGR install stow"
        exit 1
    fi

    printf '  Install stow via %s? [y/N] ' "$PKG_MGR"
    read -r answer
    case "$answer" in
        [Yy]|[Yy][Ee][Ss])
            info "Installing stow via $PKG_MGR ..."
            case "$PKG_MGR" in
                brew) brew install stow ;;
                apt)  sudo apt update && sudo apt install -y stow ;;
                dnf)  sudo dnf install -y stow ;;
            esac
            if ! has_cmd stow; then
                err "stow installation failed."
                exit 1
            fi
            info "stow installed successfully"
            ;;
        *)
            err "stow is required. Aborting."
            exit 1
            ;;
    esac
}

# ── Dependency map ───────────────────────────────────────────────────
# Maps each stow package name to the binaries it requires.
# Returns a space-separated list on stdout; empty string means no deps.
deps_for_pkg() {
    case "$1" in
        zsh)       echo "zsh zoxide fzf" ;;
        tmux)      echo "tmux" ;;
        vim)       echo "vim" ;;
        nvim)      echo "nvim" ;;
        alacritty) echo "alacritty" ;;
        git)       echo "git" ;;
        p10k)      echo "" ;;
        zprezto)   echo "" ;;
        *)         echo "" ;;
    esac
}

# ── Binary-to-package-name translation ───────────────────────────────
# Translates a binary name to the correct package name for the current
# package manager.  Most binaries share their package name; exceptions
# are handled explicitly.
pkg_name() {
    case "$1" in
        nvim) echo "neovim" ;;
        *)    echo "$1" ;;
    esac
}

# ── Install dependencies for a stow package ─────────────────────────
# Takes a stow package name, checks each required binary, and installs
# any that are missing via the detected package manager.
install_deps() {
    _stow_pkg="$1"
    _deps="$(deps_for_pkg "$_stow_pkg")"
    _apt_updated=0

    # Nothing to install
    [ -z "$_deps" ] && return 0

    # Word splitting is intentional — deps are single-word binary names
    for _bin in $_deps; do
        if has_cmd "$_bin"; then
            info "$_bin is already installed"
        else
            _pkg="$(pkg_name "$_bin")"
            warn "$_bin is not installed. Installing $_pkg ..."
            case "$PKG_MGR" in
                brew) brew install "$_pkg" ;;
                apt)
                    if [ "$_apt_updated" -eq 0 ]; then
                        sudo apt update
                        _apt_updated=1
                    fi
                    sudo apt install -y "$_pkg" ;;
                dnf)  sudo dnf install -y "$_pkg" ;;
            esac
            if has_cmd "$_bin"; then
                info "$_bin installed successfully"
            else
                err "Failed to install $_bin"
                return 1
            fi
        fi
    done
}

# ── Main ──────────────────────────────────────────────────────────────
main() {
    info "Dotfiles directory: ${FMT_BOLD}${DOTFILES_DIR}${FMT_RESET}"

    detect_os
    info "Operating system:   ${FMT_BOLD}${OS}${FMT_RESET}"

    detect_pkg_manager
    info "Package manager:    ${FMT_BOLD}${PKG_MGR}${FMT_RESET}"

    ensure_stow

    info "All prerequisites satisfied."
}

main "$@"
