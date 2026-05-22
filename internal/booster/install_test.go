package booster

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
		shimPath := filepath.Join(dir, ".booster", "hooks", hook)
		data, err := os.ReadFile(shimPath)
		if err != nil {
			t.Fatalf("missing shim %s: %v", hook, err)
		}
		if !strings.Contains(string(data), "booster") {
			t.Errorf("shim %q does not reference booster, got:\n%s", hook, data)
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
	if !strings.Contains(string(out), ".booster/hooks") {
		t.Errorf("core.hooksPath = %q, want .booster/hooks", strings.TrimSpace(string(out)))
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

	shimData, err := os.ReadFile(filepath.Join(dir, ".booster", "hooks", "pre-push"))
	if err != nil {
		t.Fatalf("ReadFile pre-push: %v", err)
	}
	shim := string(shimData)
	if !strings.Contains(shim, "BOOSTER_PUSH_REMOTE") {
		t.Errorf("pre-push shim missing BOOSTER_PUSH_REMOTE injection:\n%s", shim)
	}
	if !strings.Contains(shim, "BOOSTER_PUSH_URL") {
		t.Errorf("pre-push shim missing BOOSTER_PUSH_URL injection:\n%s", shim)
	}
}
