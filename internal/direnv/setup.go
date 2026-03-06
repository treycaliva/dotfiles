package direnv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Secret is a single op:// secret reference to inject via direnv.
type Secret struct {
	Key   string // environment variable name, e.g. GITHUB_TOKEN
	OPRef string // 1Password reference, e.g. op://Personal/GitHub/token
}

// Setup holds the user-supplied direnv configuration.
type Setup struct {
	Context   string   // "personal" or "work"
	OPAccount string   // op account shorthand, e.g. my.1password.com
	Secrets   []Secret
}

// WriteZshrcLocal writes DOTFILES_CONTEXT and DOTFILES_OP_ACCOUNT into
// ~/.zshrc.local (creating it if absent, updating existing entries in place).
func WriteZshrcLocal(homeDir string, setup *Setup) error {
	path := filepath.Join(homeDir, ".zshrc.local")

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	// Remove any existing managed lines.
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(line, "export DOTFILES_CONTEXT=") ||
			strings.HasPrefix(line, "export DOTFILES_OP_ACCOUNT=") {
			continue
		}
		kept = append(kept, line)
	}
	kept = append(kept,
		fmt.Sprintf("export DOTFILES_CONTEXT=%s", setup.Context),
		fmt.Sprintf("export DOTFILES_OP_ACCOUNT=%s", setup.OPAccount),
	)

	content := strings.Join(kept, "\n") + "\n"
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// PatchTemplate rewrites the op template for setup.Context, preserving
// existing comment lines and replacing all export lines with setup.Secrets.
func PatchTemplate(homeDir string, setup *Setup) error {
	path := filepath.Join(homeDir, ".config", "direnv", "templates", setup.Context+".env.tpl")

	var comments []string
	if data, err := os.ReadFile(path); err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				comments = append(comments, line)
			}
		}
	}

	var b strings.Builder
	for _, c := range comments {
		b.WriteString(c + "\n")
	}
	for _, s := range setup.Secrets {
		fmt.Fprintf(&b, "export %s={{ %s }}\n", s.Key, s.OPRef)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// AllowEnvrc runs `direnv allow ~/.envrc`.
func AllowEnvrc(homeDir string) error {
	return exec.Command("direnv", "allow", filepath.Join(homeDir, ".envrc")).Run()
}
