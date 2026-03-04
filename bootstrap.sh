#!/usr/bin/env bash
set -e

DOTFILES_REPO="https://github.com/treycaliva/dotfiles.git"
DOTFILES_DIR="$HOME/dotfiles"

echo "=> Bootstrapping Dotfiles..."

# 1. Check for git
if ! command -v git >/dev/null 2>&1; then
    echo "Error: git is required to clone the dotfiles repository."
    exit 1
fi

# 2. Clone or update repository
if [ -d "$DOTFILES_DIR/.git" ]; then
    echo "=> Repository already exists at $DOTFILES_DIR. Pulling latest changes..."
    cd "$DOTFILES_DIR"
    git pull
else
    echo "=> Cloning repository to $DOTFILES_DIR..."
    git clone "$DOTFILES_REPO" "$DOTFILES_DIR"
    cd "$DOTFILES_DIR"
fi

# 3. Launch Installer
# If a pre-compiled binary exists (e.g. from GitHub Releases), we would download it here.
# For now, we check if Go is installed to build the TUI, otherwise fallback to legacy script.

if command -v go >/dev/null 2>&1; then
    echo "=> Go is installed. Building and launching the new TUI installer..."
    go run ./cmd/installer
else
    echo "=> Go is not installed. Falling back to the legacy shell installer..."
    if [ -f "./install.sh" ]; then
        ./install.sh
    else
        echo "Error: Could not find install.sh."
        exit 1
    fi
fi
