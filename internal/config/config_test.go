package config

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Packages) == 0 {
		t.Fatal("expected packages, got none")
	}
	zsh, ok := cfg.Packages["zsh"]
	if !ok {
		t.Fatal("expected zsh package")
	}
	if zsh.Description != "Zsh shell config with zinit plugins" {
		t.Errorf("zsh description = %q", zsh.Description)
	}
	if len(zsh.Deps) != 3 {
		t.Errorf("zsh deps = %v, want 3 items", zsh.Deps)
	}

	nvim := cfg.Packages["nvim"]
	if nvim.PkgNames["nvim"] != "neovim" {
		t.Errorf("nvim pkg_names = %v", nvim.PkgNames)
	}

	if len(cfg.Profiles) == 0 {
		t.Fatal("expected profiles, got none")
	}
	minimal := cfg.Profiles["minimal"]
	if len(minimal.Packages) != 3 {
		t.Errorf("minimal packages = %v, want 3", minimal.Packages)
	}
}

func TestPackageOrder(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	names := cfg.PackageNames()
	if len(names) != len(cfg.Packages) {
		t.Errorf("PackageNames() returned %d, want %d", len(names), len(cfg.Packages))
	}
}
