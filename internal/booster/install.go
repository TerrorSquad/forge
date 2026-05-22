package booster

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var supportedHooks = []string{"pre-commit", "commit-msg", "pre-push"}

func InstallHooks() error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(repoRoot, ".booster", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	exeName := "booster"
	if runtime.GOOS == "windows" {
		exeName = "booster.exe"
	}

	for _, hook := range supportedHooks {
		script := buildHookScript(exeName, hook)
		path := filepath.Join(hooksDir, hook)
		if err := os.WriteFile(path, []byte(script), 0755); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	if _, err := runGit(repoRoot, "config", "--local", "core.hooksPath", ".booster/hooks"); err != nil {
		return err
	}

	fmt.Println("Installed hook shims in .booster/hooks")
	fmt.Println("Configured git core.hooksPath=.booster/hooks")
	return nil
}

func UninstallHooks() error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	hooksPath, err := localHooksPath(repoRoot)
	if err != nil {
		return err
	}
	if strings.TrimSpace(hooksPath) == ".booster/hooks" {
		if _, err := runGit(repoRoot, "config", "--local", "--unset", "core.hooksPath"); err != nil {
			return err
		}
		fmt.Println("Unset git core.hooksPath")
	} else {
		fmt.Println("core.hooksPath is not .booster/hooks; leaving git config untouched")
	}

	hooksDir := filepath.Join(repoRoot, ".booster", "hooks")
	if err := os.RemoveAll(hooksDir); err != nil {
		return err
	}
	fmt.Println("Removed .booster/hooks")
	return nil
}

func buildHookScript(exeName, hook string) string {
	return fmt.Sprintf(`#!/usr/bin/env sh
set -eu

# Prefer system-installed binary; fall back to repo-local binary (dev workflow)
if command -v %s >/dev/null 2>&1; then
  exec %s run %s "$@"
fi

# Local dev: repo root has a compiled binary
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo .)"
if [ -x "$REPO_ROOT/%s" ]; then
  exec "$REPO_ROOT/%s" run %s "$@"
fi

echo "booster not found on PATH and not found at $REPO_ROOT/%s." >&2
echo "Run: go build -o booster ./cmd/booster" >&2
exit 1
`, exeName, exeName, hook, exeName, exeName, hook, exeName)
}
