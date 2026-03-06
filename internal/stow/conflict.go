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
			if len(parts) >= 2 {
				file = strings.TrimSpace(parts[len(parts)-1])
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
