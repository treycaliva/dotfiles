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
		// diff exits 1 when files differ -- that's not an error for us
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
