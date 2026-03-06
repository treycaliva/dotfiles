package platform

import (
	"bytes"
	"fmt"
	"os/exec"
)

// DepStatus holds whether a single binary dependency is installed.
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
