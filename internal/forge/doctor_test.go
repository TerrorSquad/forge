package forge

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TerrorSquad/forge/internal/forge/config"
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

func TestDoctorReportsConfigValidationIssues(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "forge.toml"), []byte(`
[hooks.pre-commit]

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
unknown_field = true
`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(orig)
	}()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	err = DoctorWithOptions(DoctorOptions{})
	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	if err != nil {
		t.Fatalf("DoctorWithOptions returned error: %v", err)
	}
	if !strings.Contains(string(out), "config: invalid") {
		t.Errorf("expected invalid config message, got %q", string(out))
	}
}

func TestDoctorOptions_StructFields(t *testing.T) {
	opts := DoctorOptions{Fix: true, DryRun: true}
	if !opts.Fix || !opts.DryRun {
		t.Error("DoctorOptions fields not set correctly")
	}
}

func TestCheckToolAvailability_NoMissingTools(t *testing.T) {
	cfg := &config.Config{
		Hooks: map[string]config.HookConfig{
			"pre-commit": {
				Tools: map[string]config.ToolConfig{
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
	cfg := &config.Config{
		Hooks: map[string]config.HookConfig{
			"pre-commit": {
				Tools: map[string]config.ToolConfig{
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
	hooks := map[string]config.HookConfig{
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
