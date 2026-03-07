package gitconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Setup holds the user's Git identity.
type Setup struct {
	Name  string
	Email string
}

// ReadExistingSetup attempts to read Git identity from ~/.gitconfig.local.
func ReadExistingSetup(homeDir string) (*Setup, error) {
	path := filepath.Join(homeDir, ".gitconfig.local")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	setup := &Setup{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name =") {
			setup.Name = strings.TrimSpace(strings.TrimPrefix(line, "name ="))
		}
		if strings.HasPrefix(line, "email =") {
			setup.Email = strings.TrimSpace(strings.TrimPrefix(line, "email ="))
		}
	}

	if setup.Name == "" && setup.Email == "" {
		return nil, nil
	}
	return setup, nil
}

// WriteGitConfigLocal writes the Git identity to ~/.gitconfig.local.
func WriteGitConfigLocal(homeDir string, setup *Setup) error {
	path := filepath.Join(homeDir, ".gitconfig.local")

	// We'll use a simple [user] section. If the file exists, we might want to preserve 
	// other sections, but for .gitconfig.local it's usually just identity.
	// For now, let's keep it simple and overwrite or create with the [user] section.
	
	content := fmt.Sprintf("[user]\n\tname = %s\n\temail = %s\n", setup.Name, setup.Email)
	
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
