package direnv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Secret is a single op:// secret reference to inject via direnv.
type Secret struct {
	Key   string // environment variable name, e.g. GITHUB_TOKEN
	OPRef string // 1Password reference, e.g. op://Personal/GitHub/token
}

// Setup holds the user-supplied direnv configuration.
type Setup struct {
	Context       string // "personal" or "work"
	OPAccount     string // op account shorthand, e.g. my.1password.com
	Secrets       []Secret
	InheritGlobal bool // Whether to source_up in project mode
}

// ReadExistingSetup attempts to read the current direnv setup from ~/.zshrc.local
// and the corresponding op template. Returns nil if no setup is found.
func ReadExistingSetup(homeDir string) (*Setup, error) {
	setup := &Setup{
		Context: "personal", // Default
	}

	zshrcLocalPath := filepath.Join(homeDir, ".zshrc.local")
	data, err := os.ReadFile(zshrcLocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No existing setup
		}
		return nil, err
	}

	foundConfig := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export DOTFILES_CONTEXT=") {
			setup.Context = strings.TrimPrefix(line, "export DOTFILES_CONTEXT=")
			setup.Context = strings.Trim(setup.Context, `"'`)
			foundConfig = true
		}
		if strings.HasPrefix(line, "export DOTFILES_OP_ACCOUNT=") {
			setup.OPAccount = strings.TrimPrefix(line, "export DOTFILES_OP_ACCOUNT=")
			setup.OPAccount = strings.Trim(setup.OPAccount, `"'`)
			foundConfig = true
		}
	}

	if !foundConfig {
		return nil, nil
	}

	// Try to read secrets from template
	tplPath := filepath.Join(homeDir, ".config", "direnv", "templates", setup.Context+".env.tpl")
	tplData, err := os.ReadFile(tplPath)
	if err == nil {
		re := regexp.MustCompile(`^export\s+([A-Za-z0-9_]+)\s*=\s*\{\{\s*(op://.+?)\s*\}\}$`)
		for _, line := range strings.Split(string(tplData), "\n") {
			line = strings.TrimSpace(line)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				setup.Secrets = append(setup.Secrets, Secret{
					Key:   matches[1],
					OPRef: matches[2],
				})
			}
		}
	}

	return setup, nil
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

// ReadProjectSetup attempts to read the current direnv setup from a project's
// local .env.tpl. Returns nil if no setup is found.
func ReadProjectSetup(projectDir string) (*Setup, error) {
	tplPath := filepath.Join(projectDir, ".env.tpl")
	tplData, err := os.ReadFile(tplPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	setup := &Setup{
		Context: "project",
	}

	re := regexp.MustCompile(`^export\s+([A-Za-z0-9_]+)\s*=\s*\{\{\s*(op://.+?)\s*\}\}$`)
	for _, line := range strings.Split(string(tplData), "\n") {
		line = strings.TrimSpace(line)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			setup.Secrets = append(setup.Secrets, Secret{
				Key:   matches[1],
				OPRef: matches[2],
			})
		}
	}

	return setup, nil
}

// WriteProjectEnvrc ensures the project-level .envrc exists and contains
// the correct `watch_file` and `op_inject` boilerplate.
func WriteProjectEnvrc(projectDir string, setup *Setup) error {
	path := filepath.Join(projectDir, ".envrc")

	var b strings.Builder
	b.WriteString("# Project-level .envrc — loads secrets from .env.tpl via 1Password.\n")

	if setup.InheritGlobal {
		b.WriteString("\n# Inherit from global ~/.envrc\n")
		b.WriteString("source_up\n")
	}

	b.WriteString("\n# Load local project-specific secrets\n")
	b.WriteString("watch_file .env.tpl\n")
	b.WriteString("op_inject .env.tpl\n")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// PatchProjectTemplate rewrites the project-level .env.tpl, preserving
// existing comment lines and replacing all export lines with setup.Secrets.
func PatchProjectTemplate(projectDir string, setup *Setup) error {
	path := filepath.Join(projectDir, ".env.tpl")

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
