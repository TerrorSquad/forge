package forge

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to path, creating directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

func TestLoadConfig_ValidTOML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forge.toml"), `
[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
args = ["-w"]
type = "system"
extensions = [".go"]
restage = true
group = "format"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
`)

	cfg, path, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty config path")
	}

	hook, ok := cfg.Hooks["pre-commit"]
	if !ok {
		t.Fatal("expected pre-commit hook")
	}
	if !hook.IsEnabled() {
		t.Error("expected pre-commit to be enabled")
	}

	tool, ok := hook.Tools["gofmt"]
	if !ok {
		t.Fatal("expected gofmt tool")
	}
	if tool.Command != "gofmt" {
		t.Errorf("command = %q, want %q", tool.Command, "gofmt")
	}
	if !tool.Restage {
		t.Error("expected restage = true")
	}
	if len(tool.Extensions) != 1 || tool.Extensions[0] != ".go" {
		t.Errorf("extensions = %v, want [.go]", tool.Extensions)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, _, err := LoadConfig(dir)
	if err == nil {
		t.Error("expected error for missing config, got nil")
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forge.toml"), `this is not toml ][[[`)

	_, _, err := LoadConfig(dir)
	if err == nil {
		t.Error("expected error for invalid TOML, got nil")
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom.toml")
	writeFile(t, custom, `
[hooks.pre-commit]
enabled = false
`)
	// Also write a default that would be picked up without the env override
	writeFile(t, filepath.Join(dir, "forge.toml"), `
[hooks.pre-commit]
enabled = true
`)

	t.Setenv("FORGE_CONFIG", custom)

	cfg, path, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if path != custom {
		t.Errorf("path = %q, want %q", path, custom)
	}
	if cfg.Hooks["pre-commit"].IsEnabled() {
		t.Error("expected hook to be disabled (from custom config)")
	}
}

func TestHookConfig_IsEnabled_NilDefault(t *testing.T) {
	// When enabled is omitted, hook is enabled by default
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forge.toml"), `
[hooks.pre-commit]
# no enabled field
`)
	cfg, _, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.Hooks["pre-commit"].IsEnabled() {
		t.Error("expected hook with no 'enabled' field to default to enabled")
	}
}

func TestToolConfig_PassFilesEnabled_NilDefault(t *testing.T) {
	// When pass_files is omitted, it defaults to true
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forge.toml"), `
[hooks.pre-commit]

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
`)
	cfg, _, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	tool := cfg.Hooks["pre-commit"].Tools["gofmt"]
	if !tool.PassFilesEnabled() {
		t.Error("expected pass_files to default to true")
	}
}

func TestToolConfig_PassFilesEnabled_ExplicitFalse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forge.toml"), `
[hooks.pre-commit]

[hooks.pre-commit.tools.govet]
command = "go"
args = ["vet", "./..."]
pass_files = false
`)
	cfg, _, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	tool := cfg.Hooks["pre-commit"].Tools["govet"]
	if tool.PassFilesEnabled() {
		t.Error("expected pass_files = false to be respected")
	}
}

func TestInitConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := InitConfig(false, ""); err != nil {
		t.Fatalf("InitConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "forge.toml"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty forge.toml")
	}
}

func TestInitConfig_NoOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	writeFile(t, filepath.Join(dir, "forge.toml"), "existing")
	if err := InitConfig(false, ""); err == nil {
		t.Error("expected error when file exists and force=false")
	}
}

func TestInitConfig_Preset(t *testing.T) {
	for _, preset := range []string{"node", "php", "php-node", "go", "minimal"} {
		t.Run(preset, func(t *testing.T) {
			dir := t.TempDir()
			orig, _ := os.Getwd()
			if err := os.Chdir(dir); err != nil {
				t.Fatalf("Chdir: %v", err)
			}
			t.Cleanup(func() { os.Chdir(orig) })

			if err := InitConfig(false, preset); err != nil {
				t.Fatalf("InitConfig(%q): %v", preset, err)
			}
			data, err := os.ReadFile(filepath.Join(dir, "forge.toml"))
			if err != nil || len(data) == 0 {
				t.Fatalf("expected forge.toml content for preset %q", preset)
			}
		})
	}
}

func TestInitConfig_UnknownPreset(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	if err := InitConfig(false, "nonexistent"); err == nil {
		t.Error("expected error for unknown preset")
	}
}

func TestListPresets_ContainsExpected(t *testing.T) {
	presets := ListPresets()
	want := map[string]bool{"go": true, "node": true, "php": true, "php-node": true, "minimal": true}
	for _, p := range presets {
		delete(want, p)
	}
	if len(want) > 0 {
		t.Errorf("missing presets: %v", want)
	}
}

func TestGlobalConfigPath_EnvOverride(t *testing.T) {
	t.Setenv("FORGE_GLOBAL_CONFIG", "/tmp/my-global-forge.toml")
	got := globalConfigPath()
	if got != "/tmp/my-global-forge.toml" {
		t.Errorf("expected env override, got %s", got)
	}
}

func TestGlobalConfigPath_XDGDefault(t *testing.T) {
	t.Setenv("FORGE_GLOBAL_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	got := globalConfigPath()
	if got != "/custom/xdg/forge/config.toml" {
		t.Errorf("expected XDG path, got %s", got)
	}
}

func TestLoadGlobalConfig_Missing(t *testing.T) {
	t.Setenv("FORGE_GLOBAL_CONFIG", "/tmp/nonexistent-forge-global-xyz.toml")
	cfg, err := loadGlobalConfig()
	if err != nil {
		t.Errorf("missing global config should not error, got: %v", err)
	}
	if cfg != nil {
		t.Error("missing global config should return nil")
	}
}

func TestLoadGlobalConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	writeFile(t, p, `
[execution]
default_backend = "host"
tool_timeout = "90s"
`)
	t.Setenv("FORGE_GLOBAL_CONFIG", p)
	cfg, err := loadGlobalConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Execution.DefaultBackend != "host" {
		t.Errorf("expected default_backend=host, got %s", cfg.Execution.DefaultBackend)
	}
}

func TestMergeGlobalConfig_ExecutionFallback(t *testing.T) {
	global := &Config{
		Execution: ExecutionConfig{DefaultBackend: "host", ToolTimeout: "90s"},
	}
	repo := &Config{
		Hooks:     map[string]HookConfig{},
		Execution: ExecutionConfig{},
	}
	mergeGlobalConfig(global, repo)
	if repo.Execution.DefaultBackend != "host" {
		t.Errorf("expected fallback to global default_backend, got %s", repo.Execution.DefaultBackend)
	}
	if repo.Execution.ToolTimeout != "90s" {
		t.Errorf("expected fallback to global tool_timeout, got %s", repo.Execution.ToolTimeout)
	}
}

func TestMergeGlobalConfig_RepoWins(t *testing.T) {
	global := &Config{
		Execution: ExecutionConfig{DefaultBackend: "ddev", ToolTimeout: "30s"},
	}
	repo := &Config{
		Hooks:     map[string]HookConfig{},
		Execution: ExecutionConfig{DefaultBackend: "host", ToolTimeout: "60s"},
	}
	mergeGlobalConfig(global, repo)
	if repo.Execution.DefaultBackend != "host" {
		t.Errorf("repo should win, got %s", repo.Execution.DefaultBackend)
	}
}

func TestMergeGlobalConfig_ToolsMerged(t *testing.T) {
	global := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"my-global-tool": {Command: "global-check"},
				},
			},
		},
	}
	repo := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"repo-tool": {Command: "repo-check"},
				},
			},
		},
	}
	mergeGlobalConfig(global, repo)
	hook := repo.Hooks["pre-commit"]
	if _, ok := hook.Tools["my-global-tool"]; !ok {
		t.Error("global tool should have been merged into repo hook")
	}
	if _, ok := hook.Tools["repo-tool"]; !ok {
		t.Error("repo tool should still be present")
	}
}
