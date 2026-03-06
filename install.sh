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
        direnv)    echo "direnv" ;;
        zsh)       echo "zsh zoxide fzf" ;;
        tmux)      echo "tmux" ;;
        vim)       echo "vim" ;;
        nvim)      echo "nvim" ;;
        alacritty) echo "alacritty" ;;
        ghostty)   echo "ghostty" ;;
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

# Check if a symlink resolves to a path inside $DOTFILES_DIR.
# Returns 0 if yes, 1 if no or if the path is not a symlink.
_resolves_to_dotfiles() {
    _target="$1"
    [ -L "$_target" ] || return 1

    # Resolve the symlink — macOS readlink has no -f, so we
    # resolve it by cd-ing into the link's directory and using pwd.
    _link_dest="$(readlink "$_target")"

    case "$_link_dest" in
        /*) _resolved="$_link_dest" ;;
        *)
            _link_parent="$(dirname "$_target")"
            _resolved="$(cd "$_link_parent" && cd "$(dirname "$_link_dest")" && pwd)/$(basename "$_link_dest")"
            ;;
    esac

    case "$_resolved" in
        "$DOTFILES_DIR"/*) return 0 ;;
    esac
    return 1
}

# Checks whether a stow package is already linked into $HOME.
# A package counts as "stowed" if at least one of its leaf files has a
# corresponding symlink (at any path level) under $HOME that resolves
# into $DOTFILES_DIR.  This handles both shallow packages (zsh/.zshrc)
# and deep directory trees (ghostty/Library/…, nvim/.config/nvim/…).
is_stowed() {
    _pkg_dir="$DOTFILES_DIR/$1"

    [ -d "$_pkg_dir" ] || return 1

    _is_tmpfile="$(mktemp)"
    find "$_pkg_dir" ! -type d > "$_is_tmpfile" 2>/dev/null
    while IFS= read -r _file; do
        _is_rel="${_file#"$_pkg_dir"/}"
        case "$_is_rel" in .DS_Store) continue ;; esac
        # Walk each path segment from the leaf up toward $HOME
        _seg="$HOME/$_is_rel"
        while [ "$_seg" != "$HOME" ]; do
            if _resolves_to_dotfiles "$_seg"; then
                rm -f "$_is_tmpfile"
                return 0
            fi
            _seg="$(dirname "$_seg")"
        done
    done < "$_is_tmpfile"
    rm -f "$_is_tmpfile"

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

    if [ ! -e /dev/tty ]; then
        warn "No terminal available -- cannot prompt about $_conflict_file. Skipping."
        return 1
    fi

    printf '  ~/%s already exists and is not a symlink. Back up and replace? [y/N] ' "$_conflict_file"
    read -r _answer </dev/tty
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
        # Check for conflict indicators in the output
        case "$_stow_output" in
            *"existing target"*|*"conflict"*|*"CONFLICT"*)
                # Extract conflicting filenames from stow's error output.
                # Stow prints lines like: "* existing target is not owned by stow: .zshrc"
                # Use a temp file to avoid a subshell (pipe would break
                # handle_conflict's interactive prompt).
                _tmpfile="$(mktemp)"
                echo "$_stow_output" > "$_tmpfile"
                while IFS= read -r _line; do
                    case "$_line" in
                        *"existing target is"*": "*)
                            # Format: "existing target is not owned by stow: <path>"
                            _cfile="${_line##*: }"
                            ;;
                        *"over existing target "*)
                            # Format: "cannot stow <src> over existing target <path> since ..."
                            _cfile="${_line#*over existing target }"
                            _cfile="${_cfile%% since *}"
                            ;;
                        *) continue ;;
                    esac
                    if [ -n "$_cfile" ]; then
                        handle_conflict "$_cfile" || true
                    fi
                done < "$_tmpfile"
                rm -f "$_tmpfile"

                # Re-attempt stow after handling conflicts
                _retry_output=""
                if _retry_output="$(stow -d "$DOTFILES_DIR" -t "$HOME" "$_pkg" 2>&1)"; then
                    info "Stowed $_pkg (after resolving conflicts)"
                    return 0
                else
                    err "Failed to stow $_pkg even after conflict resolution"
                    [ -n "$_retry_output" ] && err "$_retry_output"
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

# ── Post-install validation ───────────────────────────────────────────
# Runs lightweight config checks after a package is stowed.
# Warnings are informational and never block installation (always returns 0).
validate_pkg() {
    _pkg="$1"

    case "$_pkg" in
        zsh)
            if zsh -n "$DOTFILES_DIR/zsh/.zshrc" 2>/dev/null; then
                info "Validation passed for $_pkg"
            else
                warn "Validation warning: zsh syntax check failed for .zshrc"
            fi
            ;;
        tmux)
            if tmux -f "$DOTFILES_DIR/tmux/.tmux.conf" new-session -d -s _validate 2>/dev/null \
               && tmux kill-session -t _validate 2>/dev/null; then
                info "Validation passed for $_pkg"
            else
                warn "Validation warning: tmux config load test failed for .tmux.conf"
            fi
            ;;
    esac

    return 0
}

# ── Package list ──────────────────────────────────────────────────────
PACKAGES="zsh tmux vim nvim alacritty ghostty git direnv p10k zprezto"

# ── Interactive menu ─────────────────────────────────────────────────
show_menu() {
    printf '\n%s%s Available packages:%s\n\n' "$FMT_BOLD" "$FMT_GREEN" "$FMT_RESET"

    _i=1
    for _pkg in $PACKAGES; do
        if is_stowed "$_pkg"; then
            _status="${FMT_GREEN}(installed)${FMT_RESET}"
        else
            _status="${FMT_YELLOW}(not installed)${FMT_RESET}"
        fi
        printf '  [%d] %-12s %s\n' "$_i" "$_pkg" "$_status"
        _i=$(( _i + 1 ))
    done

    printf '\n  [a] Install all    [q] Quit\n\n'
    printf 'Enter choices (e.g. 3 4 or a): '
}

# ── Tmux theme installation ──────────────────────────────────────────
# Installs the Srcery tmux theme if it is missing.
install_tmux_theme() {
    _theme_dir="$HOME/.tmux/themes/srcery-tmux"
    if [ ! -d "$_theme_dir" ]; then
        info "Installing Srcery tmux theme ..."
        mkdir -p "$(dirname "$_theme_dir")"
        if has_cmd git; then
            git clone --depth 1 https://github.com/srcery-colors/srcery-tmux "$_theme_dir"
            info "Srcery tmux theme installed successfully"
        else
            err "git is required to install the tmux theme. Skipping."
        fi
    else
        info "Srcery tmux theme is already installed"
    fi
}

# ── Main ──────────────────────────────────────────────────────────────
main() {
    detect_os
    detect_pkg_manager

    printf '\n%sdotfiles installer%s\n' "$FMT_BOLD" "$FMT_RESET"
    printf '==================\n'
    printf 'OS:       %s (%s)\n' "$OS" "$PKG_MGR"
    printf 'Dotfiles: %s\n\n' "$DOTFILES_DIR"

    ensure_stow

    info "All prerequisites satisfied."

    show_menu
    read -r _choices </dev/tty

    # Handle quit
    case "$_choices" in
        q|Q) info "Goodbye."; exit 0 ;;
    esac

    # Build the selection list
    _selected=""
    case "$_choices" in
        a|A)
            _selected="$PACKAGES"
            ;;
        *)
            for _num in $_choices; do
                _j=1
                _found=0
                for _pkg in $PACKAGES; do
                    if [ "$_j" -eq "$_num" ] 2>/dev/null; then
                        _selected="$_selected $_pkg"
                        _found=1
                        break
                    fi
                    _j=$(( _j + 1 ))
                done
                [ "$_found" -eq 0 ] && warn "Ignoring invalid choice: $_num"
            done
            ;;
    esac

    # Trim leading space
    _selected="${_selected# }"

    if [ -z "$_selected" ]; then
        warn "No valid packages selected."
        exit 1
    fi

    # Process each selected package
    _ok=0
    _fail=0
    printf '\n'
    for _pkg in $_selected; do
        printf '%s Setting up %s%s%s ...\n' "$FMT_BOLD" "$FMT_GREEN" "$_pkg" "$FMT_RESET"

        if install_deps "$_pkg"; then
            if stow_pkg "$_pkg"; then
                if [ "$_pkg" = "tmux" ]; then
                    install_tmux_theme
                fi
                validate_pkg "$_pkg"
                info "$_pkg setup complete"
                _ok=$(( _ok + 1 ))
            else
                err "$_pkg stow failed"
                _fail=$(( _fail + 1 ))
            fi
        else
            err "$_pkg dependency installation failed"
            _fail=$(( _fail + 1 ))
        fi

        printf '\n'
    done

    # Summary
    printf '%s── Summary ──────────────────────────────────────────────────────%s\n' "$FMT_BOLD" "$FMT_RESET"
    info "Succeeded: $_ok"
    if [ "$_fail" -gt 0 ]; then
        err "Failed:    $_fail"
        exit 1
    fi
}

main "$@"
