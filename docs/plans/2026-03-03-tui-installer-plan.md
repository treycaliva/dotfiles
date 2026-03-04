# TUI Dotfiles Installer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go TUI dotfiles installer using Bubble Tea that replaces install.sh with a polished, feature-rich experience.

**Architecture:** Elm-architecture TUI (Model-View-Update) with screen-based routing. Non-TUI logic lives in `internal/` packages (config, platform, stow, validate) that are independently testable. The TUI layer in `internal/tui/` composes these packages and renders screens.

**Tech Stack:** Go 1.22+, Bubble Tea v1 (`github.com/charmbracelet/bubbletea`), Lip Gloss v1 (`github.com/charmbracelet/lipgloss`), Bubbles v1 (`github.com/charmbracelet/bubbles`), `gopkg.in/yaml.v3` for config parsing.

**Reference:** See `docs/plans/2026-03-03-tui-installer-design.md` for the full design document.

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/installer/main.go`
- Create: `Makefile`
- Modify: `.gitignore`

**Step 1: Initialize Go module**

Run: `cd /Users/treycaliva/dotfiles && go mod init github.com/treycaliva/dotfiles`

**Step 2: Create minimal main.go**

Create `cmd/installer/main.go`:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("dotfiles installer")
	os.Exit(0)
}
```

**Step 3: Create Makefile**

Create `Makefile`:

```makefile
BINARY    := dotfiles-installer
CMD       := ./cmd/installer
GOFLAGS   := -trimpath -ldflags="-s -w"

.PHONY: build install clean test

build:
	go build $(GOFLAGS) -o $(BINARY) $(CMD)

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test ./...
```

**Step 4: Add binary to .gitignore**

Append to `.gitignore`:

```
dotfiles-installer
```

**Step 5: Build and verify**

Run: `make build && ./dotfiles-installer`
Expected: prints "dotfiles installer" and exits 0.

**Step 6: Commit**

```bash
git add go.mod cmd/installer/main.go Makefile .gitignore
git commit -m "installer: scaffold Go project with Makefile"
```

---

## Task 2: Config Package — YAML Parsing

**Files:**
- Create: `internal/config/config.yaml`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the config YAML**

Create `internal/config/config.yaml`:

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

**Step 2: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Check packages
	if len(cfg.Packages) == 0 {
		t.Fatal("expected packages, got none")
	}
	zsh, ok := cfg.Packages["zsh"]
	if !ok {
		t.Fatal("expected zsh package")
	}
	if zsh.Description != "Zsh shell config with zinit plugins" {
		t.Errorf("zsh description = %q", zsh.Description)
	}
	if len(zsh.Deps) != 3 {
		t.Errorf("zsh deps = %v, want 3 items", zsh.Deps)
	}

	// Check pkg_names
	nvim := cfg.Packages["nvim"]
	if nvim.PkgNames["nvim"] != "neovim" {
		t.Errorf("nvim pkg_names = %v", nvim.PkgNames)
	}

	// Check profiles
	if len(cfg.Profiles) == 0 {
		t.Fatal("expected profiles, got none")
	}
	minimal := cfg.Profiles["minimal"]
	if len(minimal.Packages) != 3 {
		t.Errorf("minimal packages = %v, want 3", minimal.Packages)
	}
}

func TestPackageOrder(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	names := cfg.PackageNames()
	if len(names) != len(cfg.Packages) {
		t.Errorf("PackageNames() returned %d, want %d", len(names), len(cfg.Packages))
	}
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/treycaliva/dotfiles && go test ./internal/config/ -v`
Expected: FAIL — `Load` not defined.

**Step 4: Write implementation**

Create `internal/config/config.go`:

```go
package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configData []byte

type Package struct {
	Description string            `yaml:"description"`
	Deps        []string          `yaml:"deps"`
	PkgNames    map[string]string `yaml:"pkg_names"`
	Validate    string            `yaml:"validate"`
}

type Profile struct {
	Description string   `yaml:"description"`
	Packages    []string `yaml:"packages"`
}

type Config struct {
	Packages map[string]Package `yaml:"packages"`
	Profiles map[string]Profile `yaml:"profiles"`

	// order preserves YAML key order for stable display
	order []string
}

// rawConfig is used for initial unmarshalling before we extract order.
type rawConfig struct {
	Packages yaml.Node `yaml:"packages"`
	Profiles yaml.Node `yaml:"profiles"`
}

func Load() (*Config, error) {
	// First pass: extract key order from YAML mapping node
	var raw rawConfig
	if err := yaml.Unmarshal(configData, &raw); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	var order []string
	if raw.Packages.Kind == yaml.MappingNode {
		for i := 0; i < len(raw.Packages.Content)-1; i += 2 {
			order = append(order, raw.Packages.Content[i].Value)
		}
	}

	// Second pass: unmarshal into typed struct
	var cfg Config
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.order = order

	return &cfg, nil
}

// PackageNames returns package names in the order defined in config.yaml.
func (c *Config) PackageNames() []string {
	return c.order
}
```

**Step 5: Install yaml dependency and run tests**

Run: `cd /Users/treycaliva/dotfiles && go get gopkg.in/yaml.v3 && go test ./internal/config/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "installer: add config package with YAML parsing and go:embed"
```

---

## Task 3: Platform Detection

**Files:**
- Create: `internal/platform/detect.go`
- Create: `internal/platform/detect_test.go`

**Step 1: Write the failing test**

Create `internal/platform/detect_test.go`:

```go
package platform

import (
	"runtime"
	"testing"
)

func TestDetectOS(t *testing.T) {
	info := DetectOS()

	switch runtime.GOOS {
	case "darwin":
		if info.OS != "macos" {
			t.Errorf("OS = %q, want macos", info.OS)
		}
	case "linux":
		if info.OS != "linux" {
			t.Errorf("OS = %q, want linux", info.OS)
		}
	}

	if info.OS == "" {
		t.Error("OS is empty")
	}
}

func TestDetectPkgManager(t *testing.T) {
	info := DetectOS()
	if info.PkgManager == "" {
		t.Error("PkgManager is empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/ -v`
Expected: FAIL — `DetectOS` not defined.

**Step 3: Write implementation**

Create `internal/platform/detect.go`:

```go
package platform

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Info struct {
	OS         string // "macos", "linux"
	PkgManager string // "brew", "apt", "dnf"
	IsWSL      bool
}

func DetectOS() Info {
	var info Info

	switch runtime.GOOS {
	case "darwin":
		info.OS = "macos"
	case "linux":
		info.OS = "linux"
		info.IsWSL = detectWSL()
	default:
		info.OS = runtime.GOOS
	}

	info.PkgManager = detectPkgManager(info.OS)
	return info
}

func detectWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

func detectPkgManager(osName string) string {
	switch osName {
	case "macos":
		if hasCmd("brew") {
			return "brew"
		}
	case "linux":
		if hasCmd("apt") {
			return "apt"
		}
		if hasCmd("dnf") {
			return "dnf"
		}
	}
	return ""
}

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// HasCmd is the exported version for use by other packages.
func HasCmd(name string) bool {
	return hasCmd(name)
}
```

**Step 4: Run tests**

Run: `go test ./internal/platform/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/
git commit -m "installer: add platform detection (OS, package manager, WSL)"
```

---

## Task 4: Dependency Installation

**Files:**
- Create: `internal/platform/deps.go`
- Create: `internal/platform/deps_test.go`

**Step 1: Write the failing test**

Create `internal/platform/deps_test.go`:

```go
package platform

import (
	"testing"

	"github.com/treycaliva/dotfiles/internal/config"
)

func TestCheckDeps(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	// git should be installed on any dev machine
	gitPkg := cfg.Packages["git"]
	result := CheckDeps(gitPkg.Deps)
	for _, dep := range result {
		if dep.Binary == "git" && !dep.Installed {
			t.Error("git should be detected as installed")
		}
	}
}

func TestResolvePkgName(t *testing.T) {
	names := map[string]string{"nvim": "neovim"}
	if got := ResolvePkgName("nvim", names); got != "neovim" {
		t.Errorf("ResolvePkgName(nvim) = %q, want neovim", got)
	}
	if got := ResolvePkgName("tmux", names); got != "tmux" {
		t.Errorf("ResolvePkgName(tmux) = %q, want tmux", got)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/ -v`
Expected: FAIL — `CheckDeps` not defined.

**Step 3: Write implementation**

Create `internal/platform/deps.go`:

```go
package platform

import (
	"bytes"
	"fmt"
	"os/exec"
)

type DepStatus struct {
	Binary    string
	Installed bool
}

// CheckDeps checks which binaries from a dep list are installed.
func CheckDeps(deps []string) []DepStatus {
	var results []DepStatus
	for _, bin := range deps {
		results = append(results, DepStatus{
			Binary:    bin,
			Installed: hasCmd(bin),
		})
	}
	return results
}

// ResolvePkgName translates a binary name to a package manager package name.
func ResolvePkgName(binary string, pkgNames map[string]string) string {
	if name, ok := pkgNames[binary]; ok {
		return name
	}
	return binary
}

// InstallResult holds the outcome of a single package install.
type InstallResult struct {
	Binary  string
	PkgName string
	Output  string
	Err     error
}

// InstallDep installs a single binary via the given package manager.
// Returns the combined output and any error.
func InstallDep(pkgManager, binary string, pkgNames map[string]string) InstallResult {
	pkg := ResolvePkgName(binary, pkgNames)
	result := InstallResult{Binary: binary, PkgName: pkg}

	var cmd *exec.Cmd
	switch pkgManager {
	case "brew":
		cmd = exec.Command("brew", "install", pkg)
	case "apt":
		cmd = exec.Command("sudo", "apt", "install", "-y", pkg)
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", pkg)
	default:
		result.Err = fmt.Errorf("unsupported package manager: %s", pkgManager)
		return result
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	result.Err = cmd.Run()
	result.Output = buf.String()
	return result
}
```

**Step 4: Run tests**

Run: `go test ./internal/platform/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/platform/deps.go internal/platform/deps_test.go
git commit -m "installer: add dependency checking and installation"
```

---

## Task 5: Stow Status Detection

**Files:**
- Create: `internal/stow/status.go`
- Create: `internal/stow/status_test.go`

**Step 1: Write the failing test**

Create `internal/stow/status_test.go`:

```go
package stow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsStowed_NotStowed(t *testing.T) {
	// Create a temp "dotfiles" dir with a package
	dotfiles := t.TempDir()
	home := t.TempDir()

	pkgDir := filepath.Join(dotfiles, "testpkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, ".testrc"), []byte("test"), 0o644)

	stowed, err := IsStowed("testpkg", dotfiles, home)
	if err != nil {
		t.Fatalf("IsStowed error: %v", err)
	}
	if stowed {
		t.Error("expected not stowed")
	}
}

func TestIsStowed_Stowed(t *testing.T) {
	dotfiles := t.TempDir()
	home := t.TempDir()

	pkgDir := filepath.Join(dotfiles, "testpkg")
	os.MkdirAll(pkgDir, 0o755)
	testFile := filepath.Join(pkgDir, ".testrc")
	os.WriteFile(testFile, []byte("test"), 0o644)

	// Simulate stow: create symlink in home pointing into dotfiles
	os.Symlink(testFile, filepath.Join(home, ".testrc"))

	stowed, err := IsStowed("testpkg", dotfiles, home)
	if err != nil {
		t.Fatalf("IsStowed error: %v", err)
	}
	if !stowed {
		t.Error("expected stowed")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/stow/ -v`
Expected: FAIL — `IsStowed` not defined.

**Step 3: Write implementation**

Create `internal/stow/status.go`:

```go
package stow

import (
	"os"
	"path/filepath"
	"strings"
)

// IsStowed checks whether a stow package has at least one leaf file
// symlinked into the home directory pointing back to the dotfiles dir.
func IsStowed(pkg, dotfilesDir, homeDir string) (bool, error) {
	pkgDir := filepath.Join(dotfilesDir, pkg)

	info, err := os.Stat(pkgDir)
	if err != nil || !info.IsDir() {
		return false, nil
	}

	found := false
	err = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == ".DS_Store" {
			return nil
		}

		rel, _ := filepath.Rel(pkgDir, path)
		homePath := filepath.Join(homeDir, rel)

		if resolvesToDotfiles(homePath, dotfilesDir) {
			found = true
		}
		return nil
	})

	return found, err
}

// resolvesToDotfiles checks if a path (or any parent) is a symlink
// resolving into the dotfiles directory.
func resolvesToDotfiles(path, dotfilesDir string) bool {
	// Walk from the leaf up to check each segment
	current := path
	home := filepath.Dir(path)
	// We need to walk up checking each segment
	for current != home {
		target, err := os.Readlink(current)
		if err == nil {
			// It's a symlink — resolve it
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(current), target)
			}
			resolved, err := filepath.EvalSymlinks(current)
			if err == nil && strings.HasPrefix(resolved, dotfilesDir+string(os.PathSeparator)) {
				return true
			}
			// Also check the raw target
			if strings.HasPrefix(target, dotfilesDir+string(os.PathSeparator)) {
				return true
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return false
}
```

**Step 4: Run tests**

Run: `go test ./internal/stow/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/stow/
git commit -m "installer: add stow status detection"
```

---

## Task 6: Stow & Unstow Operations

**Files:**
- Create: `internal/stow/stow.go`
- Create: `internal/stow/conflict.go`
- Create: `internal/stow/stow_test.go`

**Step 1: Write the failing test**

Create `internal/stow/stow_test.go`:

```go
package stow

import "testing"

func TestBuildStowArgs(t *testing.T) {
	args := buildStowArgs("/home/user/dotfiles", "/home/user", "zsh")
	expected := []string{"-d", "/home/user/dotfiles", "-t", "/home/user", "zsh"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestBuildUnstowArgs(t *testing.T) {
	args := buildUnstowArgs("/home/user/dotfiles", "/home/user", "zsh")
	expected := []string{"-D", "-d", "/home/user/dotfiles", "-t", "/home/user", "zsh"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
}

func TestParseConflicts(t *testing.T) {
	output := `* existing target is not owned by stow: .zshrc
* existing target is not owned by stow: .tmux.conf`

	conflicts := parseConflicts(output)
	if len(conflicts) != 2 {
		t.Fatalf("conflicts = %v, want 2", conflicts)
	}
	if conflicts[0] != ".zshrc" {
		t.Errorf("conflicts[0] = %q, want .zshrc", conflicts[0])
	}
}

func TestParseConflicts_OverExisting(t *testing.T) {
	output := `cannot stow zsh/.zshrc over existing target .zshrc since neither is a symlink`
	conflicts := parseConflicts(output)
	if len(conflicts) != 1 || conflicts[0] != ".zshrc" {
		t.Errorf("conflicts = %v, want [.zshrc]", conflicts)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/stow/ -v -run TestBuild`
Expected: FAIL — `buildStowArgs` not defined.

**Step 3: Write stow.go**

Create `internal/stow/stow.go`:

```go
package stow

import (
	"bytes"
	"fmt"
	"os/exec"
)

// StowResult holds the outcome of a stow/unstow operation.
type StowResult struct {
	Package string
	Output  string
	Err     error
}

func buildStowArgs(dotfilesDir, homeDir, pkg string) []string {
	return []string{"-d", dotfilesDir, "-t", homeDir, pkg}
}

func buildUnstowArgs(dotfilesDir, homeDir, pkg string) []string {
	return []string{"-D", "-d", dotfilesDir, "-t", homeDir, pkg}
}

// DryRun runs stow with --no (simulate) to detect conflicts without changes.
func DryRun(dotfilesDir, homeDir, pkg string) (conflicts []string, output string, err error) {
	args := append([]string{"--no"}, buildStowArgs(dotfilesDir, homeDir, pkg)...)
	var buf bytes.Buffer
	cmd := exec.Command("stow", args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run() // stow exits non-zero on conflicts
	out := buf.String()
	return parseConflicts(out), out, nil
}

// Stow links a package into the home directory.
func Stow(dotfilesDir, homeDir, pkg string) StowResult {
	args := buildStowArgs(dotfilesDir, homeDir, pkg)
	return runStow(pkg, args)
}

// Unstow removes a package's symlinks from the home directory.
func Unstow(dotfilesDir, homeDir, pkg string) StowResult {
	args := buildUnstowArgs(dotfilesDir, homeDir, pkg)
	return runStow(pkg, args)
}

func runStow(pkg string, args []string) StowResult {
	var buf bytes.Buffer
	cmd := exec.Command("stow", args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return StowResult{
		Package: pkg,
		Output:  buf.String(),
		Err:     err,
	}
}

// HasStow checks if GNU Stow is installed.
func HasStow() bool {
	_, err := exec.LookPath("stow")
	return err == nil
}

// InstallStow attempts to install stow via the given package manager.
func InstallStow(pkgManager string) error {
	var cmd *exec.Cmd
	switch pkgManager {
	case "brew":
		cmd = exec.Command("brew", "install", "stow")
	case "apt":
		cmd = exec.Command("sudo", "apt", "install", "-y", "stow")
	case "dnf":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "stow")
	default:
		return fmt.Errorf("unsupported package manager: %s", pkgManager)
	}
	return cmd.Run()
}
```

**Step 4: Write conflict.go**

Create `internal/stow/conflict.go`:

```go
package stow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// parseConflicts extracts conflicting file paths from stow's error output.
func parseConflicts(output string) []string {
	var conflicts []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		var file string

		if strings.Contains(line, "existing target is") && strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 3)
			if len(parts) >= 3 {
				file = strings.TrimSpace(parts[2])
			}
		} else if strings.Contains(line, "over existing target ") {
			after := strings.SplitN(line, "over existing target ", 2)
			if len(after) == 2 {
				file = strings.SplitN(after[1], " since ", 2)[0]
				file = strings.TrimSpace(file)
			}
		}

		if file != "" {
			conflicts = append(conflicts, file)
		}
	}
	return conflicts
}

// BackupConflict moves a conflicting file to ~/.dotfiles-backup/.
func BackupConflict(homeDir, relPath string) (backupPath string, err error) {
	fullPath := filepath.Join(homeDir, relPath)
	backupDir := filepath.Join(homeDir, ".dotfiles-backup")

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	bakName := fmt.Sprintf("%s.bak.%s", filepath.Base(relPath), timestamp)
	dest := filepath.Join(backupDir, bakName)

	if err := os.Rename(fullPath, dest); err != nil {
		return "", fmt.Errorf("backup %s: %w", relPath, err)
	}
	return dest, nil
}
```

**Step 5: Run tests**

Run: `go test ./internal/stow/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/stow/stow.go internal/stow/conflict.go internal/stow/stow_test.go
git commit -m "installer: add stow/unstow operations and conflict handling"
```

---

## Task 7: Diff Preview

**Files:**
- Create: `internal/stow/diff.go`
- Create: `internal/stow/diff_test.go`

**Step 1: Write the failing test**

Create `internal/stow/diff_test.go`:

```go
package stow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffFiles(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.txt")
	fileB := filepath.Join(dir, "b.txt")
	os.WriteFile(fileA, []byte("line1\nline2\n"), 0o644)
	os.WriteFile(fileB, []byte("line1\nline3\n"), 0o644)

	diff, err := DiffFiles(fileA, fileB)
	if err != nil {
		t.Fatalf("DiffFiles error: %v", err)
	}
	if !strings.Contains(diff, "line2") || !strings.Contains(diff, "line3") {
		t.Errorf("diff output missing expected lines:\n%s", diff)
	}
}

func TestDiffFiles_Identical(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.txt")
	os.WriteFile(fileA, []byte("same\n"), 0o644)

	diff, err := DiffFiles(fileA, fileA)
	if err != nil {
		t.Fatalf("DiffFiles error: %v", err)
	}
	if diff != "" {
		t.Errorf("expected empty diff for identical files, got:\n%s", diff)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/stow/ -v -run TestDiff`
Expected: FAIL — `DiffFiles` not defined.

**Step 3: Write implementation**

Create `internal/stow/diff.go`:

```go
package stow

import (
	"bytes"
	"os/exec"
	"path/filepath"
)

// DiffFiles returns a unified diff between two files.
// Returns empty string if files are identical.
func DiffFiles(pathA, pathB string) (string, error) {
	cmd := exec.Command("diff", "-u",
		"--label", filepath.Base(pathA)+" (current)",
		"--label", filepath.Base(pathB)+" (dotfile)",
		pathA, pathB,
	)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		// diff exits 1 when files differ — that's not an error for us
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return buf.String(), nil
		}
		return "", err
	}
	// Exit 0 means identical
	return "", nil
}

// DiffConflict generates a diff for a conflict: existing file vs dotfile source.
func DiffConflict(homeDir, dotfilesDir, pkg, relPath string) (string, error) {
	existing := filepath.Join(homeDir, relPath)
	dotfile := filepath.Join(dotfilesDir, pkg, relPath)
	return DiffFiles(existing, dotfile)
}
```

**Step 4: Run tests**

Run: `go test ./internal/stow/ -v -run TestDiff`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/stow/diff.go internal/stow/diff_test.go
git commit -m "installer: add config diff preview"
```

---

## Task 8: Post-Install Validation

**Files:**
- Create: `internal/validate/validate.go`
- Create: `internal/validate/validate_test.go`

**Step 1: Write the failing test**

Create `internal/validate/validate_test.go`:

```go
package validate

import "testing"

func TestRun_EmptyCommand(t *testing.T) {
	result := Run("testpkg", "")
	if !result.Skipped {
		t.Error("expected skipped for empty command")
	}
}

func TestRun_TrueCommand(t *testing.T) {
	result := Run("testpkg", "true")
	if result.Err != nil {
		t.Errorf("expected success, got: %v", result.Err)
	}
	if result.Skipped {
		t.Error("should not be skipped")
	}
}

func TestRun_FalseCommand(t *testing.T) {
	result := Run("testpkg", "false")
	if result.Err == nil {
		t.Error("expected failure")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/validate/ -v`
Expected: FAIL — `Run` not defined.

**Step 3: Write implementation**

Create `internal/validate/validate.go`:

```go
package validate

import (
	"bytes"
	"os/exec"
)

type Result struct {
	Package string
	Skipped bool
	Output  string
	Err     error
}

// Run executes a validation command for a package.
// If the command string is empty, returns a skipped result.
func Run(pkg, command string) Result {
	if command == "" {
		return Result{Package: pkg, Skipped: true}
	}

	cmd := exec.Command("sh", "-c", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	return Result{
		Package: pkg,
		Output:  buf.String(),
		Err:     err,
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/validate/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/validate/
git commit -m "installer: add post-install validation"
```

---

## Task 9: Srcery Theme

**Files:**
- Create: `internal/tui/theme.go`
- Create: `internal/tui/theme_test.go`

**Step 1: Write the failing test**

Create `internal/tui/theme_test.go`:

```go
package tui

import "testing"

func TestThemeColorsExist(t *testing.T) {
	// Verify theme constants are defined and non-empty
	colors := []struct {
		name  string
		color string
	}{
		{"Black", string(Theme.Black)},
		{"Red", string(Theme.Red)},
		{"Green", string(Theme.Green)},
		{"Yellow", string(Theme.Yellow)},
		{"Cyan", string(Theme.Cyan)},
	}
	for _, c := range colors {
		if c.color == "" {
			t.Errorf("Theme.%s is empty", c.name)
		}
	}
}

func TestThemeStyles(t *testing.T) {
	// Verify styles can render without panic
	_ = Styles.Title.Render("test")
	_ = Styles.StatusBar.Render("test")
	_ = Styles.Success.Render("test")
	_ = Styles.Error.Render("test")
	_ = Styles.Warning.Render("test")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/ -v`
Expected: FAIL — `Theme` not defined.

**Step 3: Write implementation**

Create `internal/tui/theme.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

type srceryTheme struct {
	Black       lipgloss.Color
	Red         lipgloss.Color
	Green       lipgloss.Color
	Yellow      lipgloss.Color
	Blue        lipgloss.Color
	Magenta     lipgloss.Color
	Cyan        lipgloss.Color
	White       lipgloss.Color
	BrightBlack lipgloss.Color
}

var Theme = srceryTheme{
	Black:       lipgloss.Color("#1C1B19"),
	Red:         lipgloss.Color("#EF2F27"),
	Green:       lipgloss.Color("#519F50"),
	Yellow:      lipgloss.Color("#FBB829"),
	Blue:        lipgloss.Color("#2C78BF"),
	Magenta:     lipgloss.Color("#E02C6D"),
	Cyan:        lipgloss.Color("#0AAEB3"),
	White:       lipgloss.Color("#BAA67F"),
	BrightBlack: lipgloss.Color("#918175"),
}

type styles struct {
	Title     lipgloss.Style
	StatusBar lipgloss.Style
	Success   lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Selected  lipgloss.Style
	Border    lipgloss.Style
	DiffAdd   lipgloss.Style
	DiffDel   lipgloss.Style
}

var Styles = styles{
	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(Theme.Yellow),

	StatusBar: lipgloss.NewStyle().
		Background(Theme.BrightBlack).
		Foreground(Theme.White).
		Padding(0, 1),

	Success: lipgloss.NewStyle().
		Foreground(Theme.Green),

	Error: lipgloss.NewStyle().
		Foreground(Theme.Red),

	Warning: lipgloss.NewStyle().
		Foreground(Theme.Yellow),

	Selected: lipgloss.NewStyle().
		Foreground(Theme.Cyan).
		Bold(true),

	Border: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Theme.BrightBlack),

	DiffAdd: lipgloss.NewStyle().
		Foreground(Theme.Green),

	DiffDel: lipgloss.NewStyle().
		Foreground(Theme.Red),
}

// Icons with Nerd Font glyphs. These render correctly with
// BlexMono Nerd Font Mono; plain-text fallbacks in parentheses.
var Icons = struct {
	Success string
	Failure string
	Warning string
	Pending string
}{
	Success: Styles.Success.Render(""),
	Failure: Styles.Error.Render(""),
	Warning: Styles.Warning.Render(""),
	Pending: Styles.Warning.Render(""),
}
```

**Step 4: Install lipgloss dependency and run tests**

Run: `cd /Users/treycaliva/dotfiles && go get github.com/charmbracelet/lipgloss && go test ./internal/tui/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/ go.mod go.sum
git commit -m "installer: add Srcery theme with Lip Gloss styles"
```

---

## Task 10: TUI App Shell & Screen Router

**Files:**
- Create: `internal/tui/app.go`
- Create: `internal/tui/screens/home.go`

This task wires up the root Bubble Tea model with screen-based routing and a minimal home screen so we can run the TUI end-to-end for the first time.

**Step 1: Install Bubble Tea and Bubbles dependencies**

Run: `cd /Users/treycaliva/dotfiles && go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles`

**Step 2: Create the screen enum and app model**

Create `internal/tui/app.go`:

```go
package tui

import (
	"github.com/treycaliva/dotfiles/internal/config"
	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenSelect
	ScreenPreview
	ScreenDiff
	ScreenProgress
	ScreenSummary
)

// ScreenModel is implemented by each screen.
type ScreenModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (ScreenModel, tea.Cmd)
	View() string
}

// AppState holds shared state passed between screens.
type AppState struct {
	Config      *config.Config
	Platform    platform.Info
	DotfilesDir string
	HomeDir     string

	// Selection
	Selected  []string // package names to install/unstow
	Unstowing bool     // true = unstow mode

	// Preview results
	Conflicts map[string][]string // pkg -> conflicting relative paths

	// Progress results
	Results map[string]error // pkg -> nil on success, error on failure
	Backups []string         // backup paths created

	// Stow status cache
	StowStatus map[string]bool // pkg -> is stowed
}

// RefreshStowStatus updates the stow status for all packages.
func (s *AppState) RefreshStowStatus() {
	s.StowStatus = make(map[string]bool)
	for _, name := range s.Config.PackageNames() {
		stowed, _ := stow.IsStowed(name, s.DotfilesDir, s.HomeDir)
		s.StowStatus[name] = stowed
	}
}

type App struct {
	state   *AppState
	screen  Screen
	current ScreenModel
	width   int
	height  int
}

func NewApp(cfg *config.Config, plat platform.Info, dotfilesDir, homeDir string) App {
	state := &AppState{
		Config:      cfg,
		Platform:    plat,
		DotfilesDir: dotfilesDir,
		HomeDir:     homeDir,
		Results:     make(map[string]error),
		Conflicts:   make(map[string][]string),
	}
	state.RefreshStowStatus()

	return App{
		state:   state,
		screen:  ScreenHome,
		current: NewHomeScreen(state),
	}
}

func (a App) Init() tea.Cmd {
	return a.current.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			if a.screen != ScreenProgress {
				return a, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case NavigateMsg:
		return a.navigate(msg)
	}

	updated, cmd := a.current.Update(msg)
	a.current = updated
	return a, cmd
}

func (a App) View() string {
	return a.current.View()
}

// NavigateMsg tells the app to switch screens.
type NavigateMsg struct {
	Screen Screen
}

func (a App) navigate(msg NavigateMsg) (tea.Model, tea.Cmd) {
	a.screen = msg.Screen

	switch msg.Screen {
	case ScreenHome:
		a.state.RefreshStowStatus()
		a.current = NewHomeScreen(a.state)
	// Other screens will be added in subsequent tasks
	}

	return a, a.current.Init()
}
```

**Step 3: Create minimal home screen**

Create `internal/tui/screens/home.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type HomeScreen struct {
	state *AppState
}

func NewHomeScreen(state *AppState) *HomeScreen {
	return &HomeScreen{state: state}
}

func (h *HomeScreen) Init() tea.Cmd {
	return nil
}

func (h *HomeScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return h, func() tea.Msg {
				return NavigateMsg{Screen: ScreenSelect}
			}
		}
	}
	return h, nil
}

func (h *HomeScreen) View() string {
	var b strings.Builder

	b.WriteString(Styles.Title.Render("  dotfiles installer"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  OS:         %s (%s)\n", h.state.Platform.OS, h.state.Platform.PkgManager))
	if h.state.Platform.IsWSL {
		b.WriteString("  WSL:        yes\n")
	}
	b.WriteString(fmt.Sprintf("  Dotfiles:   %s\n", h.state.DotfilesDir))
	b.WriteString("\n")

	// Package status table
	b.WriteString(Styles.Title.Render("  Packages"))
	b.WriteString("\n\n")

	for _, name := range h.state.Config.PackageNames() {
		pkg := h.state.Config.Packages[name]
		var status string
		if h.state.StowStatus[name] {
			status = Icons.Success + " installed"
		} else {
			status = Icons.Warning + " not installed"
		}
		b.WriteString(fmt.Sprintf("  %-12s %s  %s\n", name, status, Styles.StatusBar.Render(pkg.Description)))
	}

	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  enter: select packages  q: quit  ?: help  "))
	b.WriteString("\n")

	return b.String()
}
```

Note: The home screen file lives at `internal/tui/screens/home.go` but uses `package tui` because the screens are part of the `tui` package. Alternatively, if you prefer a `screens` subpackage, adjust imports accordingly. For simplicity, keep all TUI code in the `tui` package and use the `screens/` directory purely for file organization.

Actually, the simpler approach: put all screen files directly in `internal/tui/` to avoid circular imports:

Move `home.go` to `internal/tui/home.go` instead.

**Step 4: Wire up main.go**

Modify `cmd/installer/main.go`:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treycaliva/dotfiles/internal/config"
	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	plat := platform.DetectOS()

	// Determine dotfiles directory (where this binary's source lives)
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding executable: %v\n", err)
		os.Exit(1)
	}
	dotfilesDir := filepath.Dir(exe)

	// If run via `go run` or from the repo, use working directory
	if _, err := os.Stat(filepath.Join(dotfilesDir, "go.mod")); err != nil {
		dotfilesDir, _ = os.Getwd()
	}

	homeDir, _ := os.UserHomeDir()

	app := tui.NewApp(cfg, plat, dotfilesDir, homeDir)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 5: Build and run**

Run: `cd /Users/treycaliva/dotfiles && go mod tidy && make build && ./dotfiles-installer`
Expected: TUI launches in alt screen showing OS info and package status. Press `q` to quit.

**Step 6: Commit**

```bash
git add internal/tui/ cmd/installer/main.go go.mod go.sum
git commit -m "installer: add TUI app shell with home screen"
```

---

## Task 11: Select Screen

**Files:**
- Create: `internal/tui/select.go`

**Step 1: Write the select screen**

Create `internal/tui/select.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SelectScreen struct {
	state     *AppState
	cursor    int
	checked   map[int]bool
	unstowing bool
	packages  []string // ordered package names
}

func NewSelectScreen(state *AppState) *SelectScreen {
	return &SelectScreen{
		state:    state,
		checked:  make(map[int]bool),
		packages: state.Config.PackageNames(),
	}
}

func (s *SelectScreen) Init() tea.Cmd { return nil }

func (s *SelectScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenHome} }
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.packages)-1 {
				s.cursor++
			}
		case "tab", " ":
			s.checked[s.cursor] = !s.checked[s.cursor]
		case "m":
			s.applyProfile("minimal")
		case "s":
			s.applyProfile("server")
		case "f":
			s.applyProfile("full")
		case "u":
			s.unstowing = !s.unstowing
		case "a":
			allChecked := len(s.checked) == len(s.packages)
			if !allChecked {
				for i := range s.packages {
					s.checked[i] = true
				}
			} else {
				s.checked = make(map[int]bool)
			}
		case "enter":
			var selected []string
			for i, name := range s.packages {
				if s.checked[i] {
					selected = append(selected, name)
				}
			}
			if len(selected) == 0 {
				return s, nil
			}
			s.state.Selected = selected
			s.state.Unstowing = s.unstowing
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		}
	}
	return s, nil
}

func (s *SelectScreen) applyProfile(name string) {
	profile, ok := s.state.Config.Profiles[name]
	if !ok {
		return
	}
	s.checked = make(map[int]bool)
	profileSet := make(map[string]bool)
	for _, pkg := range profile.Packages {
		profileSet[pkg] = true
	}
	for i, pkg := range s.packages {
		if profileSet[pkg] {
			s.checked[i] = true
		}
	}
}

func (s *SelectScreen) View() string {
	var b strings.Builder

	mode := "Install"
	if s.unstowing {
		mode = "Unstow"
	}
	b.WriteString(Styles.Title.Render(fmt.Sprintf("  Select Packages (%s mode)", mode)))
	b.WriteString("\n\n")

	for i, name := range s.packages {
		cursor := "  "
		if s.cursor == i {
			cursor = Styles.Selected.Render("> ")
		}

		check := "[ ]"
		if s.checked[i] {
			check = Styles.Selected.Render("[x]")
		}

		var status string
		if s.state.StowStatus[name] {
			status = Icons.Success
		} else {
			status = Icons.Warning
		}

		desc := s.state.Config.Packages[name].Description
		b.WriteString(fmt.Sprintf("%s%s %s %-12s %s\n", cursor, check, status, name, desc))
	}

	b.WriteString("\n")
	b.WriteString("  Profiles: ")
	b.WriteString(Styles.Selected.Render("m") + "=minimal  ")
	b.WriteString(Styles.Selected.Render("s") + "=server  ")
	b.WriteString(Styles.Selected.Render("f") + "=full  ")
	b.WriteString(Styles.Selected.Render("a") + "=toggle all\n")
	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  j/k: move  space: toggle  u: unstow mode  enter: confirm  esc: back  "))
	b.WriteString("\n")

	return b.String()
}
```

**Step 2: Register select screen in app.go navigate()**

Add to the `navigate` switch in `internal/tui/app.go`:

```go
case ScreenSelect:
    a.current = NewSelectScreen(a.state)
```

**Step 3: Build and test interactively**

Run: `make build && ./dotfiles-installer`
Expected: Home screen → press Enter → Select screen with checkbox list. Toggle items with space, apply profiles with m/s/f, press Enter to proceed.

**Step 4: Commit**

```bash
git add internal/tui/select.go internal/tui/app.go
git commit -m "installer: add package selection screen with profiles"
```

---

## Task 12: Preview (Dry-Run) Screen

**Files:**
- Create: `internal/tui/preview.go`

**Step 1: Write the preview screen**

Create `internal/tui/preview.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"

	tea "github.com/charmbracelet/bubbletea"
)

type previewItem struct {
	pkg        string
	missingDeps []platform.DepStatus
	conflicts  []string
}

type PreviewScreen struct {
	state   *AppState
	items   []previewItem
	cursor  int
	loading bool
}

type previewReadyMsg struct {
	items []previewItem
}

func NewPreviewScreen(state *AppState) *PreviewScreen {
	return &PreviewScreen{state: state, loading: true}
}

func (p *PreviewScreen) Init() tea.Cmd {
	return p.runDryRun
}

func (p *PreviewScreen) runDryRun() tea.Msg {
	var items []previewItem
	for _, pkg := range p.state.Selected {
		item := previewItem{pkg: pkg}

		if !p.state.Unstowing {
			// Check deps
			cfgPkg := p.state.Config.Packages[pkg]
			deps := platform.CheckDeps(cfgPkg.Deps)
			for _, d := range deps {
				if !d.Installed {
					item.missingDeps = append(item.missingDeps, d)
				}
			}
			// Check conflicts
			conflicts, _, _ := stow.DryRun(p.state.DotfilesDir, p.state.HomeDir, pkg)
			item.conflicts = conflicts
		}

		items = append(items, item)
	}
	return previewReadyMsg{items: items}
}

func (p *PreviewScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case previewReadyMsg:
		p.items = msg.items
		p.loading = false

		// Store conflicts in state for use during progress
		p.state.Conflicts = make(map[string][]string)
		for _, item := range p.items {
			if len(item.conflicts) > 0 {
				p.state.Conflicts[item.pkg] = item.conflicts
			}
		}
		return p, nil
	case tea.KeyMsg:
		if p.loading {
			return p, nil
		}
		switch msg.String() {
		case "esc":
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenSelect} }
		case "enter":
			return p, func() tea.Msg { return NavigateMsg{Screen: ScreenProgress} }
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		case "d":
			// Open diff for current item's first conflict
			if p.cursor < len(p.items) && len(p.items[p.cursor].conflicts) > 0 {
				// Store diff context in state and navigate
				return p, func() tea.Msg { return NavigateMsg{Screen: ScreenDiff} }
			}
		}
	}
	return p, nil
}

func (p *PreviewScreen) View() string {
	var b strings.Builder

	action := "Install"
	if p.state.Unstowing {
		action = "Unstow"
	}
	b.WriteString(Styles.Title.Render(fmt.Sprintf("  Preview — %s %d package(s)", action, len(p.state.Selected))))
	b.WriteString("\n\n")

	if p.loading {
		b.WriteString("  Analyzing packages...\n")
		return b.String()
	}

	for i, item := range p.items {
		cursor := "  "
		if p.cursor == i {
			cursor = Styles.Selected.Render("> ")
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, Styles.Title.Render(item.pkg)))

		if len(item.missingDeps) > 0 {
			b.WriteString("    Dependencies to install:\n")
			for _, dep := range item.missingDeps {
				b.WriteString(fmt.Sprintf("      %s %s\n", Icons.Pending, dep.Binary))
			}
		}

		if len(item.conflicts) > 0 {
			b.WriteString(fmt.Sprintf("    Conflicts (%s to view diff):\n", Styles.Selected.Render("d")))
			for _, c := range item.conflicts {
				b.WriteString(fmt.Sprintf("      %s ~/%s\n", Icons.Warning, c))
			}
		}

		if len(item.missingDeps) == 0 && len(item.conflicts) == 0 {
			b.WriteString(fmt.Sprintf("    %s Ready\n", Icons.Success))
		}
		b.WriteString("\n")
	}

	b.WriteString(Styles.StatusBar.Render("  enter: proceed  d: diff  esc: back  "))
	b.WriteString("\n")

	return b.String()
}
```

**Step 2: Register preview screen in app.go navigate()**

Add to the `navigate` switch:

```go
case ScreenPreview:
    a.current = NewPreviewScreen(a.state)
```

**Step 3: Build and test interactively**

Run: `make build && ./dotfiles-installer`
Expected: Home → Select → check some packages → Enter → Preview shows deps/conflicts analysis.

**Step 4: Commit**

```bash
git add internal/tui/preview.go internal/tui/app.go
git commit -m "installer: add dry-run preview screen"
```

---

## Task 13: Diff Viewer Screen

**Files:**
- Create: `internal/tui/diff.go`

**Step 1: Write the diff screen**

Create `internal/tui/diff.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/treycaliva/dotfiles/internal/stow"
)

type DiffScreen struct {
	state    *AppState
	viewport viewport.Model
	pkg      string
	file     string
	ready    bool
}

type diffContentMsg struct {
	content string
}

func NewDiffScreen(state *AppState, pkg, file string) *DiffScreen {
	return &DiffScreen{
		state: state,
		pkg:   pkg,
		file:  file,
	}
}

func (d *DiffScreen) Init() tea.Cmd {
	return d.loadDiff
}

func (d *DiffScreen) loadDiff() tea.Msg {
	diff, err := stow.DiffConflict(d.state.HomeDir, d.state.DotfilesDir, d.pkg, d.file)
	if err != nil {
		return diffContentMsg{content: fmt.Sprintf("Error loading diff: %v", err)}
	}
	if diff == "" {
		return diffContentMsg{content: "(files are identical)"}
	}
	return diffContentMsg{content: diff}
}

func (d *DiffScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.viewport = viewport.New(msg.Width, msg.Height-4)
		d.ready = true
	case diffContentMsg:
		styled := d.styleDiff(msg.content)
		d.viewport.SetContent(styled)
		return d, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return d, func() tea.Msg { return NavigateMsg{Screen: ScreenPreview} }
		}
	}

	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d *DiffScreen) styleDiff(raw string) string {
	var b strings.Builder
	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			b.WriteString(Styles.DiffAdd.Render(line))
		case strings.HasPrefix(line, "-"):
			b.WriteString(Styles.DiffDel.Render(line))
		case strings.HasPrefix(line, "@@"):
			b.WriteString(Styles.Selected.Render(line))
		default:
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (d *DiffScreen) View() string {
	var b strings.Builder

	b.WriteString(Styles.Title.Render(fmt.Sprintf("  Diff: %s — ~/%s", d.pkg, d.file)))
	b.WriteString("\n\n")

	if !d.ready {
		b.WriteString("  Loading...\n")
		return b.String()
	}

	b.WriteString(d.viewport.View())
	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  j/k: scroll  esc: back  "))
	b.WriteString("\n")

	return b.String()
}
```

**Step 2: Update navigate() to pass diff context**

The diff screen needs to know which package/file to diff. Update `AppState` to include current diff target, and update the preview screen's `d` key handler to set it before navigating. Add to `AppState`:

```go
DiffPkg  string // current diff target package
DiffFile string // current diff target file
```

Update `navigate` for diff:

```go
case ScreenDiff:
    a.current = NewDiffScreen(a.state, a.state.DiffPkg, a.state.DiffFile)
```

Update preview screen's `d` handler to set `DiffPkg`/`DiffFile` before navigating.

**Step 3: Build and test**

Run: `make build && ./dotfiles-installer`
Expected: Preview → press `d` on a conflict → Diff screen shows colored unified diff. Esc returns to preview.

**Step 4: Commit**

```bash
git add internal/tui/diff.go internal/tui/app.go internal/tui/preview.go
git commit -m "installer: add diff viewer screen"
```

---

## Task 14: Progress Screen

**Files:**
- Create: `internal/tui/progress.go`

**Step 1: Write the progress screen**

Create `internal/tui/progress.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treycaliva/dotfiles/internal/platform"
	"github.com/treycaliva/dotfiles/internal/stow"
	"github.com/treycaliva/dotfiles/internal/validate"
)

type pkgStatus int

const (
	statusPending pkgStatus = iota
	statusActive
	statusDone
	statusFailed
)

type pkgProgress struct {
	name   string
	status pkgStatus
	logs   []string
}

type ProgressScreen struct {
	state    *AppState
	items    []pkgProgress
	current  int
	done     bool
	spinner  spinner.Model
	logView  viewport.Model
	allLogs  []string
}

type installStepMsg struct {
	pkg  string
	log  string
	done bool
	err  error
}

func NewProgressScreen(state *AppState) *ProgressScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(Theme.Yellow)

	var items []pkgProgress
	for _, pkg := range state.Selected {
		items = append(items, pkgProgress{name: pkg, status: statusPending})
	}

	return &ProgressScreen{
		state:   state,
		items:   items,
		spinner: s,
		logView: viewport.New(80, 10),
	}
}

func (p *ProgressScreen) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, p.processNext())
}

func (p *ProgressScreen) processNext() tea.Cmd {
	if p.current >= len(p.items) {
		return nil
	}

	pkg := p.items[p.current].name

	return func() tea.Msg {
		cfgPkg := p.state.Config.Packages[pkg]

		if p.state.Unstowing {
			result := stow.Unstow(p.state.DotfilesDir, p.state.HomeDir, pkg)
			return installStepMsg{pkg: pkg, log: result.Output, done: true, err: result.Err}
		}

		// Install deps
		deps := platform.CheckDeps(cfgPkg.Deps)
		for _, dep := range deps {
			if !dep.Installed {
				result := platform.InstallDep(p.state.Platform.PkgManager, dep.Binary, cfgPkg.PkgNames)
				if result.Err != nil {
					return installStepMsg{pkg: pkg, log: result.Output, done: true, err: result.Err}
				}
			}
		}

		// Handle conflicts
		if conflicts, ok := p.state.Conflicts[pkg]; ok {
			for _, c := range conflicts {
				bakPath, err := stow.BackupConflict(p.state.HomeDir, c)
				if err != nil {
					return installStepMsg{pkg: pkg, log: fmt.Sprintf("backup failed: %v", err), done: true, err: err}
				}
				p.state.Backups = append(p.state.Backups, bakPath)
			}
		}

		// Stow
		result := stow.Stow(p.state.DotfilesDir, p.state.HomeDir, pkg)
		if result.Err != nil {
			return installStepMsg{pkg: pkg, log: result.Output, done: true, err: result.Err}
		}

		// Validate
		valResult := validate.Run(pkg, cfgPkg.Validate)
		var valLog string
		if !valResult.Skipped {
			if valResult.Err != nil {
				valLog = fmt.Sprintf("validation warning: %s", valResult.Output)
			} else {
				valLog = "validation passed"
			}
		}

		return installStepMsg{pkg: pkg, log: valLog, done: true, err: nil}
	}
}

func (p *ProgressScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.logView = viewport.New(msg.Width, msg.Height/3)
	case spinner.TickMsg:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		return p, cmd
	case installStepMsg:
		for i := range p.items {
			if p.items[i].name == msg.pkg {
				if msg.err != nil {
					p.items[i].status = statusFailed
					p.state.Results[msg.pkg] = msg.err
				} else {
					p.items[i].status = statusDone
					p.state.Results[msg.pkg] = nil
				}
				if msg.log != "" {
					p.allLogs = append(p.allLogs, fmt.Sprintf("[%s] %s", msg.pkg, msg.log))
					p.logView.SetContent(strings.Join(p.allLogs, "\n"))
					p.logView.GotoBottom()
				}
				break
			}
		}

		p.current++
		if p.current >= len(p.items) {
			p.done = true
			return p, nil
		}
		p.items[p.current].status = statusActive
		return p, p.processNext()

	case tea.KeyMsg:
		if p.done {
			switch msg.String() {
			case "enter":
				return p, func() tea.Msg { return NavigateMsg{Screen: ScreenSummary} }
			}
		}
	}

	var cmd tea.Cmd
	p.logView, cmd = p.logView.Update(msg)
	return p, cmd
}

func (p *ProgressScreen) View() string {
	var b strings.Builder

	action := "Installing"
	if p.state.Unstowing {
		action = "Unstowing"
	}
	b.WriteString(Styles.Title.Render(fmt.Sprintf("  %s packages", action)))
	b.WriteString("\n\n")

	for _, item := range p.items {
		var icon string
		switch item.status {
		case statusPending:
			icon = "  "
		case statusActive:
			icon = p.spinner.View()
		case statusDone:
			icon = Icons.Success
		case statusFailed:
			icon = Icons.Failure
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", icon, item.name))
	}

	b.WriteString("\n")
	b.WriteString(Styles.Border.Render(p.logView.View()))
	b.WriteString("\n")

	if p.done {
		b.WriteString("\n")
		b.WriteString(Styles.StatusBar.Render("  enter: view summary  "))
	}
	b.WriteString("\n")

	return b.String()
}
```

**Step 2: Register in navigate()**

```go
case ScreenProgress:
    a.current = NewProgressScreen(a.state)
```

**Step 3: Build and test**

Run: `make build && ./dotfiles-installer`
Expected: Select packages → Preview → Enter → Progress shows spinners for each package, logs scroll in viewport. Press Enter when done to go to summary.

**Step 4: Commit**

```bash
git add internal/tui/progress.go internal/tui/app.go
git commit -m "installer: add progress screen with spinners and log viewport"
```

---

## Task 15: Summary Screen

**Files:**
- Create: `internal/tui/summary.go`

**Step 1: Write the summary screen**

Create `internal/tui/summary.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SummaryScreen struct {
	state *AppState
}

func NewSummaryScreen(state *AppState) *SummaryScreen {
	return &SummaryScreen{state: state}
}

func (s *SummaryScreen) Init() tea.Cmd { return nil }

func (s *SummaryScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return s, tea.Quit
		case "r":
			return s, func() tea.Msg { return NavigateMsg{Screen: ScreenHome} }
		}
	}
	return s, nil
}

func (s *SummaryScreen) View() string {
	var b strings.Builder

	b.WriteString(Styles.Title.Render("  Summary"))
	b.WriteString("\n\n")

	succeeded := 0
	failed := 0
	for _, pkg := range s.state.Selected {
		err := s.state.Results[pkg]
		if err == nil {
			b.WriteString(fmt.Sprintf("  %s %s\n", Icons.Success, pkg))
			succeeded++
		} else {
			b.WriteString(fmt.Sprintf("  %s %s — %v\n", Icons.Failure, pkg, err))
			failed++
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Succeeded: %s\n", Styles.Success.Render(fmt.Sprintf("%d", succeeded))))
	if failed > 0 {
		b.WriteString(fmt.Sprintf("  Failed:    %s\n", Styles.Error.Render(fmt.Sprintf("%d", failed))))
	}

	if len(s.state.Backups) > 0 {
		b.WriteString("\n")
		b.WriteString(Styles.Title.Render("  Backed up files"))
		b.WriteString("\n")
		for _, bak := range s.state.Backups {
			b.WriteString(fmt.Sprintf("    %s\n", bak))
		}
	}

	b.WriteString("\n")
	b.WriteString(Styles.StatusBar.Render("  q: quit  r: start over  "))
	b.WriteString("\n")

	return b.String()
}
```

**Step 2: Register in navigate()**

```go
case ScreenSummary:
    a.state.RefreshStowStatus()
    a.current = NewSummaryScreen(a.state)
```

**Step 3: Build and test end-to-end**

Run: `make build && ./dotfiles-installer`
Expected: Full flow works: Home → Select → Preview → Progress → Summary. Summary shows results, `r` returns to home, `q` quits.

**Step 4: Commit**

```bash
git add internal/tui/summary.go internal/tui/app.go
git commit -m "installer: add summary screen"
```

---

## Task 16: Status Bar Component

**Files:**
- Create: `internal/tui/statusbar.go`

**Step 1: Write the status bar**

Create `internal/tui/statusbar.go`:

```go
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type KeyBinding struct {
	Key  string
	Help string
}

// StatusBar renders a consistent bottom status bar with keybinding hints.
func StatusBar(width int, bindings []KeyBinding) string {
	var parts []string
	for _, b := range bindings {
		key := lipgloss.NewStyle().Bold(true).Foreground(Theme.Cyan).Render(b.Key)
		parts = append(parts, key+":"+b.Help)
	}
	content := "  " + strings.Join(parts, "  ") + "  "

	style := lipgloss.NewStyle().
		Background(Theme.BrightBlack).
		Foreground(Theme.White).
		Width(width)

	return style.Render(content)
}
```

This component can then replace the inline `Styles.StatusBar.Render(...)` calls in each screen for a consistent look. Refactor each screen to use `StatusBar(width, bindings)` instead.

**Step 2: Build and verify**

Run: `make build && ./dotfiles-installer`
Expected: Status bar renders consistently at the bottom of each screen.

**Step 3: Commit**

```bash
git add internal/tui/statusbar.go
git commit -m "installer: add reusable status bar component"
```

---

## Task 17: Help Overlay

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add help toggle to app model**

Add a `showHelp bool` field to `App`. When `?` is pressed, toggle it. In `View()`, if `showHelp` is true, overlay a help panel on top of the current screen content.

```go
// In App.Update, before delegating to current screen:
case "?":
    a.showHelp = !a.showHelp
    return a, nil
```

```go
// In App.View:
func (a App) View() string {
    view := a.current.View()
    if a.showHelp {
        help := Styles.Border.Render(helpText())
        // Center the help overlay
        return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, help)
    }
    return view
}

func helpText() string {
    return `
  Keyboard Shortcuts

  Navigation
    enter     Proceed / confirm
    esc       Go back one screen
    q         Quit
    ?         Toggle this help

  Select Screen
    j/k       Move cursor
    space     Toggle package
    m/s/f     Apply profile (minimal/server/full)
    a         Toggle all
    u         Switch install/unstow mode

  Preview Screen
    d         View diff for conflicting file

  Diff View
    j/k       Scroll
    esc       Back to preview

  Summary
    r         Start over
`
}
```

**Step 2: Build and test**

Run: `make build && ./dotfiles-installer`
Expected: Press `?` on any screen → help overlay appears. Press `?` again → dismissed.

**Step 3: Commit**

```bash
git add internal/tui/app.go
git commit -m "installer: add help overlay"
```

---

## Task 18: Run All Tests & Final Polish

**Step 1: Run full test suite**

Run: `cd /Users/treycaliva/dotfiles && go test ./... -v`
Expected: All tests pass.

**Step 2: Run linter**

Run: `go vet ./...`
Expected: No issues.

**Step 3: Build release binary**

Run: `make clean && make build && ls -lh dotfiles-installer`
Expected: Binary exists, ~5-10MB.

**Step 4: Manual end-to-end test**

Run through the full flow:
1. Launch `./dotfiles-installer`
2. Home screen shows correct OS/package info and stow status
3. Press Enter → Select screen
4. Apply "minimal" profile → check correct packages
5. Press Enter → Preview shows dry-run analysis
6. Press Enter → Progress shows spinners, packages install
7. Summary shows results
8. Press `r` → back to home, statuses updated
9. Press `q` → clean exit

**Step 5: Final commit**

```bash
git add -A
git commit -m "installer: complete TUI dotfiles installer v1"
```

---

## Dependency Summary

Install these Go modules:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/bubbles`
- `gopkg.in/yaml.v3`

## Notes for the Implementer

- **All screen files go in `internal/tui/`** using `package tui` to avoid circular imports. The `screens/` subdirectory from the design doc is dropped in favor of flat files in `internal/tui/`.
- **Bubble Tea v1** (stable) is used, not v2 beta. `View()` returns `string`, messages use `tea.KeyMsg`.
- **Test the non-TUI packages thoroughly** (config, platform, stow, validate). TUI screens are tested manually since Bubble Tea testing requires tea.TestModel which is complex to set up.
- **The existing `install.sh` is not modified** — it stays as a fallback.
