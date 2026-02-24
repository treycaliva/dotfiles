# Bootstrap Script Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create `install.sh` — an interactive bootstrap script that detects the OS, installs dependencies, and stow-links dotfiles packages.

**Architecture:** Single POSIX shell script at repo root. OS detection via `uname -s`, package manager detection (brew/apt/dnf), interactive numbered menu, stow for symlinking. No external dependencies beyond a shell and a package manager.

**Tech Stack:** POSIX sh, GNU Stow, Homebrew (macOS), apt/dnf (Linux)

---

### Task 1: Create `.gitignore`

**Files:**
- Create: `.gitignore`

**Step 1: Write `.gitignore`**

```
firebase-debug.log
.DS_Store
.claude/
```

**Step 2: Commit**

```bash
git add .gitignore
git commit -m "Add .gitignore for firebase-debug.log, .DS_Store, .claude/"
```

---

### Task 2: Create `install.sh` scaffold with OS detection and prerequisite checks

**Files:**
- Create: `install.sh`

**Step 1: Write the script scaffold**

The script should include:

1. `#!/bin/sh` shebang + `set -e`
2. Color/formatting variables for output (green checkmark, red X, yellow warning — using tput or ANSI codes with a no-color fallback for non-interactive terminals)
3. Helper functions:
   - `info()`, `warn()`, `err()` — formatted output helpers
   - `has_cmd()` — checks if a binary exists on PATH via `command -v`
4. OS detection:
   - `detect_os()` — sets `OS` to `macos` or `linux` via `uname -s`
   - `detect_pkg_manager()` — sets `PKG_MGR` to `brew`, `apt`, or `dnf`
   - On macOS, if `brew` is missing: print message to install Homebrew first and exit 1
   - On Linux, if neither `apt` nor `dnf` found: print message and exit 1
5. Stow check:
   - If `stow` is missing, prompt to install it via `$PKG_MGR` (e.g., `brew install stow`)
   - Exit 1 if user declines
6. `DOTFILES_DIR` — set to the directory containing the script (`$(cd "$(dirname "$0")" && pwd)`)

**Step 2: Make executable and test**

```bash
chmod +x install.sh
sh install.sh
```

Expected: sees OS detection output, stow check passes (stow is installed), script exits cleanly (no menu yet).

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: scaffold with OS detection and prerequisite checks"
```

---

### Task 3: Add dependency map and install function

**Files:**
- Modify: `install.sh`

**Step 1: Add the dependency map and installer**

Add after the prerequisite checks:

1. Dependency map — a function `deps_for_pkg()` that maps each stow package name to its required binaries:
   - `zsh` → `zsh zoxide fzf`
   - `tmux` → `tmux`
   - `vim` → `vim`
   - `nvim` → `nvim`
   - `alacritty` → `alacritty`
   - `git` → `git`
   - `p10k` → (empty)
   - `zprezto` → (empty)

2. Platform package name mapping — a function `pkg_name()` that translates a binary name to the correct package name for the current package manager:
   - `nvim` → `neovim` (all platforms)
   - `zsh` → `zsh` (all platforms)
   - `fzf` → `fzf` (all platforms)
   - `zoxide` → `zoxide` (all platforms)
   - Everything else: binary name = package name

3. `install_deps()` — takes a stow package name, checks each binary, installs missing ones:
   - For each binary from `deps_for_pkg`, run `has_cmd`
   - If missing: run `$PKG_MGR install $(pkg_name $bin)`
   - Use `sudo` prefix for `apt`/`dnf`, no sudo for `brew`
   - Print status per dependency (already installed / installing / failed)

**Step 2: Test with a dry run**

Add a temporary call like `install_deps zsh` at the end. Run:

```bash
sh install.sh
```

Expected: checks for zsh, zoxide, fzf — reports each as already installed (assuming they are on this machine).

Remove the temporary call after verifying.

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: add dependency map and install function"
```

---

### Task 4: Add stow status detection and link function

**Files:**
- Modify: `install.sh`

**Step 1: Add status check and stow link functions**

1. `is_stowed()` — takes a package name, checks if that package's files are symlinked into `$HOME` and point into `$DOTFILES_DIR`:
   - For each file in `$DOTFILES_DIR/<pkg>/` (excluding `.DS_Store`), check the corresponding path under `$HOME`
   - A package is "installed" if at least one of its target files is a symlink resolving to a path within `$DOTFILES_DIR`
   - Uses `readlink` (with `-f` on Linux, no `-f` on macOS — use a helper for portability)

2. `stow_pkg()` — takes a package name, runs `stow` from `$DOTFILES_DIR`:
   - Runs `stow -d "$DOTFILES_DIR" -t "$HOME" <pkg>`
   - On failure (exit code != 0), captures stderr
   - If conflict detected (file already exists), calls `handle_conflict()`
   - Returns 0 on success, 1 on failure

3. `handle_conflict()` — handles stow conflicts:
   - Parses the conflicting file path from stow's error output
   - Creates `~/.dotfiles-backup/` if it doesn't exist
   - Prompts: "~/.zshrc already exists and is not a symlink. Back up and replace? [y/N]"
   - If yes: moves file to `~/.dotfiles-backup/<filename>.bak.<timestamp>`
   - If no: skips that package

**Step 2: Test**

Add a temporary call: `is_stowed zsh && echo "stowed" || echo "not stowed"`. Run:

```bash
sh install.sh
```

Expected: prints "stowed" (since ~/.zshrc is already a symlink into this repo).

Remove temporary call.

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: add stow status detection and conflict handling"
```

---

### Task 5: Add interactive menu

**Files:**
- Modify: `install.sh`

**Step 1: Add the menu and main logic**

1. Define the package list: `PACKAGES="zsh tmux vim nvim alacritty git p10k zprezto"`

2. `show_menu()` — displays the interactive menu:
   - Iterates over `$PACKAGES` with a counter
   - For each, calls `is_stowed` and shows `(installed)` or `(not installed)`
   - Shows `[a] Install all    [q] Quit` at the bottom
   - Prompts: `Enter choices (e.g. 3 4 or a): `

3. `main()` — the entry point:
   - Calls `detect_os`, `detect_pkg_manager`, checks stow
   - Calls `show_menu`
   - Reads user input
   - If `q`: exit 0
   - If `a`: set selection to all packages
   - Otherwise: map numbers to package names
   - For each selected package:
     a. Print "Setting up <pkg>..."
     b. Call `install_deps <pkg>`
     c. Call `stow_pkg <pkg>`
     d. Print success/failure
   - Print summary at end

4. Call `main "$@"` at bottom of script

**Step 2: Test interactively**

```bash
sh install.sh
```

Expected: see the full menu with correct install status for each package. Enter `q` to quit cleanly. Run again and enter `a` to verify all packages process (most will already be installed).

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: add interactive menu and main flow"
```

---

### Task 6: Add post-install validation

**Files:**
- Modify: `install.sh`

**Step 1: Add validation function**

1. `validate_pkg()` — runs after successful stow, per package:
   - `zsh`: `zsh -n "$DOTFILES_DIR/zsh/.zshrc"` — syntax check
   - `tmux`: `tmux -f "$DOTFILES_DIR/tmux/.tmux.conf" new-session -d -s _validate && tmux kill-session -t _validate` — config load test
   - All others: no validation (skip silently)
   - Print result: "Validation passed" or "Validation warning: ..." (warnings don't block)

2. Wire `validate_pkg` into the main loop — call it after `stow_pkg` succeeds.

**Step 2: Test**

```bash
sh install.sh
```

Select `1` (zsh). Expected: stow runs (or reports already installed), then validation runs and reports syntax check passed.

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: add post-install validation for zsh and tmux"
```

---

### Task 7: Final polish and test

**Files:**
- Modify: `install.sh`

**Step 1: Add usage header**

When the script starts, print:

```
dotfiles installer
==================
OS: macOS (brew)
Dotfiles: /Users/treycaliva/dotfiles
```

**Step 2: Full end-to-end test**

Run through these scenarios:
1. `sh install.sh` — see menu, press `q` to quit
2. `sh install.sh` — select `a`, verify all packages report installed/linked
3. Verify `sh -n install.sh` passes (POSIX syntax check)
4. Verify `shellcheck install.sh` passes (if shellcheck is available)

**Step 3: Commit**

```bash
git add install.sh
git commit -m "install.sh: add usage header and finalize"
```
