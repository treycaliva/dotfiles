package platform

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Info holds detected platform details.
type Info struct {
	OS         string // "macos", "linux"
	PkgManager string // "brew", "apt", "dnf"
	IsWSL      bool
}

// DetectOS returns platform information for the current system.
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
