package stow

import "testing"

func TestBuildStowArgs(t *testing.T) {
	args := buildStowArgs("/home/user/dotfiles", "/home/user", "zsh")
	expected := []string{"-d", "/home/user/dotfiles", "-t", "/home/user", "zsh"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestBuildUnstowArgs(t *testing.T) {
	args := buildUnstowArgs("/home/user/dotfiles", "/home/user", "zsh")
	expected := []string{"-D", "-d", "/home/user/dotfiles", "-t", "/home/user", "zsh"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
}

func TestParseConflicts(t *testing.T) {
	output := `* existing target is not owned by stow: .zshrc
* existing target is not owned by stow: .tmux.conf`

	conflicts := parseConflicts(output)
	if len(conflicts) != 2 {
		t.Fatalf("conflicts = %v, want 2", conflicts)
	}
	if conflicts[0] != ".zshrc" {
		t.Errorf("conflicts[0] = %q, want .zshrc", conflicts[0])
	}
}

func TestParseConflicts_OverExisting(t *testing.T) {
	output := `cannot stow zsh/.zshrc over existing target .zshrc since neither is a symlink`
	conflicts := parseConflicts(output)
	if len(conflicts) != 1 || conflicts[0] != ".zshrc" {
		t.Errorf("conflicts = %v, want [.zshrc]", conflicts)
	}
}
