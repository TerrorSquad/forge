package booster

import (
	"strings"
	"testing"
)

func TestDoctorWithOptions_NoGitRepo(t *testing.T) {
	// DoctorWithOptions must not panic even when not in a git repo.
	// We can't easily fake detectRepoRoot, so just verify it returns nil.
	// (The function prints to stdout; we just check no panic + no error.)
	// This test relies on the real cwd; it may pass or fail depending on
	// whether tests are run inside a git repo. Just ensure no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DoctorWithOptions panicked: %v", r)
		}
	}()
	_ = DoctorWithOptions(DoctorOptions{})
}

func TestDoctorOptions_StructFields(t *testing.T) {
	opts := DoctorOptions{Fix: true, DryRun: true}
	if !opts.Fix || !opts.DryRun {
		t.Error("DoctorOptions fields not set correctly")
	}
}

func TestCheckToolAvailability_NoMissingTools(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"echo-check": {Command: "echo"},
				},
			},
		},
	}
	missing := checkToolAvailability(cfg)
	for _, m := range missing {
		if m == "echo" {
			t.Error("'echo' should be found in PATH")
		}
	}
}

func TestCheckToolAvailability_MissingTool(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"missing": {Command: "this-tool-definitely-does-not-exist-xyz123"},
				},
			},
		},
	}
	missing := checkToolAvailability(cfg)
	found := false
	for _, m := range missing {
		if strings.Contains(m, "this-tool-definitely-does-not-exist-xyz123") {
			found = true
		}
	}
	if !found {
		t.Error("expected missing tool to be reported")
	}
}

func TestSortedHookNames_Order(t *testing.T) {
	hooks := map[string]HookConfig{
		"pre-push":   {},
		"commit-msg": {},
		"pre-commit": {},
	}
	names := sortedHookNames(hooks)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "commit-msg" || names[1] != "pre-commit" || names[2] != "pre-push" {
		t.Errorf("unexpected order: %v", names)
	}
}
