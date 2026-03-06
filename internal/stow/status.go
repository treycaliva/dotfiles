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
	current := path
	home := filepath.Dir(path)
	for current != home {
		target, err := os.Readlink(current)
		if err == nil {
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(current), target)
			}
			resolved, err := filepath.EvalSymlinks(current)
			if err == nil && strings.HasPrefix(resolved, dotfilesDir+string(os.PathSeparator)) {
				return true
			}
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
