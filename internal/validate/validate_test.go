package validate

import "testing"

func TestRun_EmptyCommand(t *testing.T) {
	result := Run("testpkg", "")
	if !result.Skipped {
		t.Error("expected skipped for empty command")
	}
}

func TestRun_TrueCommand(t *testing.T) {
	result := Run("testpkg", "true")
	if result.Err != nil {
		t.Errorf("expected success, got: %v", result.Err)
	}
	if result.Skipped {
		t.Error("should not be skipped")
	}
}

func TestRun_FalseCommand(t *testing.T) {
	result := Run("testpkg", "false")
	if result.Err == nil {
		t.Error("expected failure")
	}
}
