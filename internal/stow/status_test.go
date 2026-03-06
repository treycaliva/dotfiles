package stow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsStowed_NotStowed(t *testing.T) {
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
