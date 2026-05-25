package forge

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var supportedHooks = []string{"pre-commit", "commit-msg", "pre-push", "prepare-commit-msg", "post-commit", "post-merge", "post-rewrite"}

func InstallHooks() error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(repoRoot, ".booster", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	exeName := "forge"
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

	if _, err := runGit(repoRoot, "config", "--local", "core.hooksPath", ".forge/hooks"); err != nil {
		return err
	}

	// Add .forge/cache.json to .git/info/exclude so it is never committed.
	excludeFile := filepath.Join(repoRoot, ".git", "info", "exclude")
	excludeEntry := ".forge/cache.json\n"
	addToExclude(excludeFile, excludeEntry)

	// Write the JSON Schema and Taplo config for IDE support.
	if err := installSchema(repoRoot); err != nil {
		// Non-fatal: schema is a convenience feature.
		fmt.Printf("warning: could not write schema: %v\n", err)
	}

	fmt.Println("Installed hook shims in .forge/hooks")
	fmt.Println("Configured git core.hooksPath=.forge/hooks")
	return nil
}

// addToExclude appends an entry to the given exclude file if not already present.
func addToExclude(path, entry string) {
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), strings.TrimSpace(entry)) {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(entry)
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
	if strings.TrimSpace(hooksPath) == ".forge/hooks" {
		if _, err := runGit(repoRoot, "config", "--local", "--unset", "core.hooksPath"); err != nil {
			return err
		}
		fmt.Println("Unset git core.hooksPath")
	} else {
		fmt.Println("core.hooksPath is not .forge/hooks; leaving git config untouched")
	}

	hooksDir := filepath.Join(repoRoot, ".booster", "hooks")
	if err := os.RemoveAll(hooksDir); err != nil {
		return err
	}
	fmt.Println("Removed .forge/hooks")
	return nil
}

func buildHookScript(exeName, hook string) string {
	// For pre-push, forward git's $1/$2 (remote name and URL) as env vars
	prePushEnv := ""
	if hook == "pre-push" {
		prePushEnv = `
FORGE_PUSH_REMOTE="$1" FORGE_PUSH_URL="$2" \`
	}

	return fmt.Sprintf(`#!/usr/bin/env sh
set -eu

# Prefer system-installed binary; fall back to repo-local binary (dev workflow)
if command -v %s >/dev/null 2>&1; then%s
  exec %s run %s "$@"
fi

# Local dev: repo root has a compiled binary
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo .)"
if [ -x "$REPO_ROOT/%s" ]; then%s
  exec "$REPO_ROOT/%s" run %s "$@"
fi

echo "booster not found on PATH and not found at $REPO_ROOT/%s." >&2
echo "Run: go build -o booster ./cmd/forge" >&2
exit 1
`, exeName, prePushEnv, exeName, hook, exeName, prePushEnv, exeName, hook, exeName)
}

// installSchema writes forge.schema.json to .forge/ and creates/updates
// .taplo.toml so that Even Better TOML (and any Taplo-based editor) automatically
// provides schema validation and completion for forge.toml.
func installSchema(repoRoot string) error {
	boosterDir := filepath.Join(repoRoot, ".booster")
	schemaPath := filepath.Join(boosterDir, "forge.schema.json")

	if err := os.WriteFile(schemaPath, []byte(SchemaJSON), 0644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}

	taplo := filepath.Join(repoRoot, ".taplo.toml")
	rule := `
[[rule]]
name = "forge"
include = ["**/forge.toml"]
url = "file:./.forge/forge.schema.json"
`
	data, _ := os.ReadFile(taplo)
	if strings.Contains(string(data), `name = "forge"`) {
		// Rule already present — update the schema path in case it moved.
		updated := strings.ReplaceAll(string(data),
			`url = "file:./.forge/forge.schema.json"`,
			`url = "file:./.forge/forge.schema.json"`,
		)
		return os.WriteFile(taplo, []byte(updated), 0644)
	}

	f, err := os.OpenFile(taplo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .taplo.toml: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(rule)
	return err
}
