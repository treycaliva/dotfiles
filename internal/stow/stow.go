package stow

import (
	"bytes"
	"fmt"
	"os/exec"
)

// StowResult captures the outcome of a stow or unstow operation.
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
	_ = cmd.Run()
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
