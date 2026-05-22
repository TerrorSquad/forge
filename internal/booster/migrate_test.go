package booster

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateConfig_BasicRoundtrip(t *testing.T) {
	dir := t.TempDir()

	passFilesTrue := true
	input := map[string]interface{}{
		"pre-commit": map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":       "eslint",
					"command":    "eslint",
					"args":       []string{"--fix"},
					"type":       "node",
					"group":      "lint",
					"restage":    true,
					"passFiles":  &passFilesTrue,
					"extensions": []string{".ts", ".js"},
				},
			},
		},
	}
	data, _ := json.Marshal(input)
	inputPath := filepath.Join(dir, ".git-hooks.config.json")
	writeFile(t, inputPath, string(data))

	outputPath := filepath.Join(dir, "booster.toml")
	if err := MigrateConfig(inputPath, outputPath); err != nil {
		t.Fatalf("MigrateConfig: %v", err)
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(out)

	for _, want := range []string{
		"[hooks.pre-commit]",
		`command = "eslint"`,
		`group = "lint"`,
		"restage = true",
		`".ts"`,
		`".js"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in output, got:\n%s", want, content)
		}
	}
}

func TestMigrateConfig_StdoutMode(t *testing.T) {
	dir := t.TempDir()
	input := map[string]interface{}{
		"commit-msg": map[string]interface{}{
			"tools": []map[string]interface{}{
				{"name": "commitlint", "command": "commitlint", "args": []string{"--edit"}},
			},
		},
	}
	data, _ := json.Marshal(input)
	inputPath := filepath.Join(dir, "hooks.json")
	writeFile(t, inputPath, string(data))

	// Output to "-" means stdout — we can't easily capture that in a test
	// so we just verify no error is returned
	if err := MigrateConfig(inputPath, "-"); err != nil {
		t.Fatalf("MigrateConfig to stdout: %v", err)
	}
}

func TestMigrateConfig_NoInputFile(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// No .git-hooks.config.json in dir — should fail
	if err := MigrateConfig("", "-"); err == nil {
		t.Error("expected error when no input file found")
	}
}

func TestMigrateConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "bad.json")
	writeFile(t, inputPath, `{not valid json`)

	if err := MigrateConfig(inputPath, "-"); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMigrateConfig_PassFilesFalse(t *testing.T) {
	dir := t.TempDir()
	passFalse := false
	input := map[string]interface{}{
		"pre-commit": map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":      "govet",
					"command":   "go",
					"args":      []string{"vet", "./..."},
					"passFiles": &passFalse,
				},
			},
		},
	}
	data, _ := json.Marshal(input)
	inputPath := filepath.Join(dir, "hooks.json")
	writeFile(t, inputPath, string(data))

	outputPath := filepath.Join(dir, "booster.toml")
	if err := MigrateConfig(inputPath, outputPath); err != nil {
		t.Fatalf("MigrateConfig: %v", err)
	}

	out, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(out), "pass_files = false") {
		t.Errorf("expected 'pass_files = false' in output, got:\n%s", out)
	}
}
