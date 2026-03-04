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
