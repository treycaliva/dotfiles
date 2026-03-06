package platform

import (
	"testing"

	"github.com/treycaliva/dotfiles/internal/config"
)

func TestCheckDeps(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	// git should be installed on any dev machine
	gitPkg := cfg.Packages["git"]
	result := CheckDeps(gitPkg.Deps)
	for _, dep := range result {
		if dep.Binary == "git" && !dep.Installed {
			t.Error("git should be detected as installed")
		}
	}
}

func TestResolvePkgName(t *testing.T) {
	names := map[string]string{"nvim": "neovim"}
	if got := ResolvePkgName("nvim", names); got != "neovim" {
		t.Errorf("ResolvePkgName(nvim) = %q, want neovim", got)
	}
	if got := ResolvePkgName("tmux", names); got != "tmux" {
		t.Errorf("ResolvePkgName(tmux) = %q, want tmux", got)
	}
}
