package booster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// legacyConfig mirrors the shape of .git-hooks.config.json / .git-hooks.config.dist.json
type legacyConfig struct {
	PreCommit *legacyHook `json:"pre-commit"`
	CommitMsg *legacyHook `json:"commit-msg"`
	PrePush   *legacyHook `json:"pre-push"`
}

type legacyHook struct {
	Tools []legacyTool `json:"tools"`
}

type legacyTool struct {
	Name       string   `json:"name"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Type       string   `json:"type"`
	Group      string   `json:"group"`
	PassFiles  *bool    `json:"passFiles"`
	RunPerFile bool     `json:"runPerFile"`
	Restage    bool     `json:"restage"`
	OnFailure  string   `json:"onFailure"`
	Extensions []string `json:"extensions"`
}

// MigrateConfig reads a .git-hooks.config.json (or its .dist variant) from
// the given path and emits an equivalent booster.toml to stdout (or the
// output file when outputPath != "").
func MigrateConfig(inputPath, outputPath string) error {
	// Default input candidates
	if inputPath == "" {
		for _, candidate := range []string{
			".git-hooks.config.json",
			".husky/.git-hooks.config.dist.json",
		} {
			if _, err := os.Stat(candidate); err == nil {
				inputPath = candidate
				break
			}
		}
	}
	if inputPath == "" {
		return fmt.Errorf("no .git-hooks.config.json found; provide path with --from")
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", inputPath, err)
	}

	var legacy legacyConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("parsing %s: %w", inputPath, err)
	}

	var sb strings.Builder
	sb.WriteString("# Migrated from ")
	sb.WriteString(filepath.Base(inputPath))
	sb.WriteString("\n# Review and adjust as needed.\n\n")

	migrateHookSection(&sb, "pre-commit", legacy.PreCommit)
	migrateHookSection(&sb, "commit-msg", legacy.CommitMsg)
	migrateHookSection(&sb, "pre-push", legacy.PrePush)

	out := sb.String()

	if outputPath == "" || outputPath == "-" {
		fmt.Print(out)
		return nil
	}

	if err := os.WriteFile(outputPath, []byte(out), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}
	fmt.Printf("Wrote %s\n", outputPath)
	return nil
}

func migrateHookSection(sb *strings.Builder, hookName string, hook *legacyHook) {
	if hook == nil {
		return
	}

	sb.WriteString("[hooks.")
	sb.WriteString(hookName)
	sb.WriteString("]\nenabled = true\n\n")

	for _, t := range hook.Tools {
		name := sanitizeEnvKey(t.Name)
		sb.WriteString("[hooks.")
		sb.WriteString(hookName)
		sb.WriteString(".tools.")
		sb.WriteString(strings.ToLower(name))
		sb.WriteString("]\n")

		cmd := t.Command
		if cmd == "" {
			cmd = t.Name
		}
		sb.WriteString(fmt.Sprintf("command = %q\n", cmd))

		if len(t.Args) > 0 {
			sb.WriteString("args = [")
			for i, a := range t.Args {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%q", a))
			}
			sb.WriteString("]\n")
		}

		if t.Type != "" {
			sb.WriteString(fmt.Sprintf("type = %q\n", t.Type))
		}
		if t.Group != "" {
			sb.WriteString(fmt.Sprintf("group = %q\n", t.Group))
		}
		if len(t.Extensions) > 0 {
			sb.WriteString("extensions = [")
			for i, ext := range t.Extensions {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%q", ext))
			}
			sb.WriteString("]\n")
		}
		if t.PassFiles != nil && !*t.PassFiles {
			sb.WriteString("pass_files = false\n")
		}
		if t.RunPerFile {
			sb.WriteString("run_per_file = true\n")
		}
		if t.Restage {
			sb.WriteString("restage = true\n")
		}
		if t.OnFailure != "" {
			sb.WriteString(fmt.Sprintf("on_failure = %q\n", t.OnFailure))
		}
		sb.WriteString("\n")
	}
}
