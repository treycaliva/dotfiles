package stow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffFiles(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.txt")
	fileB := filepath.Join(dir, "b.txt")
	os.WriteFile(fileA, []byte("line1\nline2\n"), 0o644)
	os.WriteFile(fileB, []byte("line1\nline3\n"), 0o644)

	diff, err := DiffFiles(fileA, fileB)
	if err != nil {
		t.Fatalf("DiffFiles error: %v", err)
	}
	if !strings.Contains(diff, "line2") || !strings.Contains(diff, "line3") {
		t.Errorf("diff output missing expected lines:\n%s", diff)
	}
}

func TestDiffFiles_Identical(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.txt")
	os.WriteFile(fileA, []byte("same\n"), 0o644)

	diff, err := DiffFiles(fileA, fileA)
	if err != nil {
		t.Fatalf("DiffFiles error: %v", err)
	}
	if diff != "" {
		t.Errorf("expected empty diff for identical files, got:\n%s", diff)
	}
}
