package forge

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall_CreatesHookShims(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := InstallHooks(); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	for _, hook := range []string{"pre-commit", "commit-msg", "pre-push", "prepare-commit-msg", "post-commit", "post-merge", "post-rewrite"} {
		shimPath := filepath.Join(dir, ".forge", "hooks", hook)
		data, err := os.ReadFile(shimPath)
		if err != nil {
			t.Fatalf("missing shim %s: %v", hook, err)
		}
		if !strings.Contains(string(data), "forge") {
			t.Errorf("shim %q does not reference forge, got:\n%s", hook, data)
		}
		info, err := os.Stat(shimPath)
		if err != nil {
			t.Fatalf("stat %s: %v", hook, err)
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("shim %q is not executable", hook)
		}
	}
}

func TestInstall_SetsHooksPath(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := InstallHooks(); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	cmd := exec.Command("git", "config", "--local", "core.hooksPath")
	cmd.Dir = dir
	out, _ := cmd.Output()
	if !strings.Contains(string(out), ".forge/hooks") {
		t.Errorf("core.hooksPath = %q, want .forge/hooks", strings.TrimSpace(string(out)))
	}
}

func TestUninstall_RemovesHooksPath(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := InstallHooks(); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}
	if err := UninstallHooks(); err != nil {
		t.Fatalf("UninstallHooks: %v", err)
	}

	cmd := exec.Command("git", "config", "--local", "core.hooksPath")
	cmd.Dir = dir
	out, _ := cmd.Output()
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected core.hooksPath to be unset after uninstall, got: %q", strings.TrimSpace(string(out)))
	}
}

func TestInstall_Idempotent(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := InstallHooks(); err != nil {
		t.Fatalf("first InstallHooks: %v", err)
	}
	if err := InstallHooks(); err != nil {
		t.Fatalf("second InstallHooks (idempotent): %v", err)
	}
}

func TestInstall_PrePushShimHasEnvInjection(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := InstallHooks(); err != nil {
		t.Fatalf("InstallHooks: %v", err)
	}

	shimData, err := os.ReadFile(filepath.Join(dir, ".forge", "hooks", "pre-push"))
	if err != nil {
		t.Fatalf("ReadFile pre-push: %v", err)
	}
	shim := string(shimData)
	if !strings.Contains(shim, "FORGE_PUSH_REMOTE") {
		t.Errorf("pre-push shim missing FORGE_PUSH_REMOTE injection:\n%s", shim)
	}
	if !strings.Contains(shim, "FORGE_PUSH_URL") {
		t.Errorf("pre-push shim missing FORGE_PUSH_URL injection:\n%s", shim)
	}
}

func TestBuildHookScript_MisePathRespectsMiseDataDir(t *testing.T) {
	script := buildHookScript("forge", "pre-commit")
	if !strings.Contains(script, "MISE_DATA_DIR") {
		t.Errorf("shim should use MISE_DATA_DIR, got:\n%s", script)
	}
	// Must not fall back to a hardcoded path that ignores MISE_DATA_DIR
	if strings.Contains(script, `$HOME/.local/share/mise/shims"`) {
		t.Errorf("shim should not hardcode the mise shims path; use MISE_DATA_DIR instead")
	}
}

func TestBuildHookScript_MisePathRespectsXDGDataHome(t *testing.T) {
	script := buildHookScript("forge", "pre-commit")
	if !strings.Contains(script, "XDG_DATA_HOME") {
		t.Errorf("shim should respect XDG_DATA_HOME, got:\n%s", script)
	}
}

func TestBuildHookScript_MiseFallsBackToDefault(t *testing.T) {
	// The default fallback must still be $HOME/.local/share.
	script := buildHookScript("forge", "pre-commit")
	if !strings.Contains(script, `$HOME/.local/share`) {
		t.Errorf("shim should retain $HOME/.local/share as final fallback, got:\n%s", script)
	}
}
