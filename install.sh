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

# ── Stow status detection ────────────────────────────────────────────
# Checks whether a stow package is already linked into $HOME.
# A package counts as "stowed" if at least one of its top-level entries
# is a symlink under $HOME that resolves to a path inside $DOTFILES_DIR.
is_stowed() {
    _pkg="$1"
    _pkg_dir="$DOTFILES_DIR/$_pkg"

    [ -d "$_pkg_dir" ] || return 1

    for _entry in "$_pkg_dir"/*; do
        # Skip glob that matched nothing (no files in dir)
        [ -e "$_entry" ] || continue

        _basename="$(basename "$_entry")"

        # Skip .DS_Store
        [ "$_basename" = ".DS_Store" ] && continue

        _target="$HOME/$_basename"

        # Check if the corresponding file in $HOME is a symlink
        if [ -L "$_target" ]; then
            # Resolve the symlink — macOS readlink has no -f, so we
            # resolve it by cd-ing into the link's directory and using pwd.
            _link_dest="$(readlink "$_target")"

            # Handle both absolute and relative symlink targets
            case "$_link_dest" in
                /*)
                    # Absolute path — use as-is
                    _resolved="$_link_dest"
                    ;;
                *)
                    # Relative path — resolve from the symlink's parent dir
                    _link_parent="$(dirname "$_target")"
                    _resolved="$(cd "$_link_parent" && cd "$(dirname "$_link_dest")" && pwd)/$(basename "$_link_dest")"
                    ;;
            esac

            # Check if the resolved path is inside the dotfiles directory
            case "$_resolved" in
                "$DOTFILES_DIR"/*)
                    return 0
                    ;;
            esac
        fi
    done

    # Also check hidden files (glob * does not match dotfiles)
    for _entry in "$_pkg_dir"/.*; do
        _basename="$(basename "$_entry")"

        # Skip . , .. , and .DS_Store
        case "$_basename" in
            .|..|.DS_Store) continue ;;
        esac

        _target="$HOME/$_basename"

        if [ -L "$_target" ]; then
            _link_dest="$(readlink "$_target")"

            case "$_link_dest" in
                /*)
                    _resolved="$_link_dest"
                    ;;
                *)
                    _link_parent="$(dirname "$_target")"
                    _resolved="$(cd "$_link_parent" && cd "$(dirname "$_link_dest")" && pwd)/$(basename "$_link_dest")"
                    ;;
            esac

            case "$_resolved" in
                "$DOTFILES_DIR"/*)
                    return 0
                    ;;
            esac
        fi
    done

    return 1
}

# ── Conflict handling ────────────────────────────────────────────────
# Backs up a conflicting file so stow can replace it.
# Takes the conflicting file path (relative to $HOME, e.g. ".zshrc").
handle_conflict() {
    _conflict_file="$1"
    _full_path="$HOME/$_conflict_file"
    _backup_dir="$HOME/.dotfiles-backup"

    mkdir -p "$_backup_dir"

    if [ ! -t 0 ]; then
        warn "Non-interactive shell — cannot prompt about $_conflict_file. Skipping."
        return 1
    fi

    printf '  ~/%s already exists and is not a symlink. Back up and replace? [y/N] ' "$_conflict_file"
    read -r _answer
    case "$_answer" in
        [Yy]|[Yy][Ee][Ss])
            _timestamp="$(date +%Y%m%d%H%M%S)"
            _bak_name="$(basename "$_conflict_file").bak.$_timestamp"
            mv "$_full_path" "$_backup_dir/$_bak_name"
            info "Backed up ~/$_conflict_file → ~/.dotfiles-backup/$_bak_name"
            return 0
            ;;
        *)
            warn "Skipping $_conflict_file"
            return 1
            ;;
    esac
}

# ── Stow a single package ───────────────────────────────────────────
# Runs GNU Stow for one package, handling conflicts automatically.
stow_pkg() {
    _pkg="$1"

    # Capture both stdout and stderr; do not let set -e kill us on failure
    _stow_output=""
    _stow_output="$(stow -d "$DOTFILES_DIR" -t "$HOME" "$_pkg" 2>&1)" || {
        _stow_rc=$?

        # Check for conflict indicators in the output
        case "$_stow_output" in
            *"existing target"*|*"conflict"*|*"CONFLICT"*)
                # Extract conflicting filenames from stow's error output.
                # Stow prints lines like: "* existing target is not owned by stow: .zshrc"
                echo "$_stow_output" | while IFS= read -r _line; do
                    case "$_line" in
                        *"existing target"*)
                            # Pull the filename from the end of the line
                            _cfile="$(echo "$_line" | sed 's/.*: //')"
                            if [ -n "$_cfile" ]; then
                                if handle_conflict "$_cfile"; then
                                    : # handled — will re-stow below
                                fi
                            fi
                            ;;
                    esac
                done

                # Re-attempt stow after handling conflicts
                if stow -d "$DOTFILES_DIR" -t "$HOME" "$_pkg" 2>/dev/null; then
                    info "Stowed $_pkg (after resolving conflicts)"
                    return 0
                else
                    err "Failed to stow $_pkg even after conflict resolution"
                    return 1
                fi
                ;;
            *)
                err "stow $_pkg failed: $_stow_output"
                return 1
                ;;
        esac
    }

    info "Stowed $_pkg"
    return 0
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
