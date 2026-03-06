package direnv_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/treycaliva/dotfiles/internal/direnv"
)

func TestWriteZshrcLocal_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	setup := &direnv.Setup{Context: "personal", OPAccount: "my.1password.com"}

	if err := direnv.WriteZshrcLocal(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmp, ".zshrc.local"))
	content := string(data)
	if !strings.Contains(content, "export DOTFILES_CONTEXT=personal") {
		t.Errorf("missing DOTFILES_CONTEXT, got:\n%s", content)
	}
	if !strings.Contains(content, "export DOTFILES_OP_ACCOUNT=my.1password.com") {
		t.Errorf("missing DOTFILES_OP_ACCOUNT, got:\n%s", content)
	}
}

func TestWriteZshrcLocal_UpdatesExisting(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".zshrc.local")
	existing := "# my config\nexport DOTFILES_CONTEXT=work\nexport OTHER=foo\n"
	os.WriteFile(path, []byte(existing), 0644)

	setup := &direnv.Setup{Context: "personal", OPAccount: "new.1password.com"}
	if err := direnv.WriteZshrcLocal(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if strings.Contains(content, "export DOTFILES_CONTEXT=work") {
		t.Error("old DOTFILES_CONTEXT should have been removed")
	}
	if !strings.Contains(content, "export DOTFILES_CONTEXT=personal") {
		t.Error("new DOTFILES_CONTEXT missing")
	}
	if !strings.Contains(content, "export OTHER=foo") {
		t.Error("unrelated line should be preserved")
	}
}

func TestPatchTemplate_WritesSecrets(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".config", "direnv", "templates")
	os.MkdirAll(dir, 0755)
	tplPath := filepath.Join(dir, "personal.env.tpl")
	os.WriteFile(tplPath, []byte("# Personal template\n# Format: export KEY={{ op://... }}\n"), 0644)

	setup := &direnv.Setup{
		Context: "personal",
		Secrets: []direnv.Secret{
			{Key: "GITHUB_TOKEN", OPRef: "op://Personal/GitHub/token"},
		},
	}
	if err := direnv.PatchTemplate(tmp, setup); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tplPath)
	content := string(data)
	if !strings.Contains(content, "export GITHUB_TOKEN={{ op://Personal/GitHub/token }}") {
		t.Errorf("secret missing from template:\n%s", content)
	}
	if !strings.Contains(content, "# Personal template") {
		t.Error("comments should be preserved")
	}
}
