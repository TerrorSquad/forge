package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// legacyConfig mirrors the shape of .git-hooks.config.json / .git-hooks.config.dist.json
// Supports two formats:
//   - flat format: top-level "pre-commit", "commit-msg", "pre-push" (original)
//   - econnect format: nested "hooks.preCommit" / "hooks.prePush" / "hooks.commitMsg"
type legacyConfig struct {
	PreCommit *legacyHook `json:"pre-commit"`
	CommitMsg *legacyHook `json:"commit-msg"`
	PrePush   *legacyHook `json:"pre-push"`
	// econnect nested format
	Hooks *legacyHooksNested `json:"hooks"`
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

// legacyHooksNested is the econnect variant: "hooks": { "preCommit": { "tools": { "ESLint": {...} } } }
type legacyHooksNested struct {
	PreCommit *legacyHookNested `json:"preCommit"`
	CommitMsg *legacyHookNested `json:"commitMsg"`
	PrePush   *legacyHookNested `json:"prePush"`
}

type legacyHookNested struct {
	// Tools is a map — we use json.RawMessage + ordered key extraction to preserve declaration order.
	Tools json.RawMessage `json:"tools"`
}

// legacyToolNested is one entry in the econnect "tools" map.
type legacyToolNested struct {
	Command          string   `json:"command"`
	Args             []string `json:"args"`
	Type             string   `json:"type"`
	Group            string   `json:"group"`
	PassFiles        *bool    `json:"passFiles"`
	RunForEachFile   bool     `json:"runForEachFile"`
	StagesFilesAfter bool     `json:"stagesFilesAfter"`
	OnFailure        string   `json:"onFailure"`
	Extensions       []string `json:"extensions"`
}

// parseOrderedTools parses a JSON object of tools preserving declaration order.
func parseOrderedTools(raw json.RawMessage) ([]string, map[string]legacyToolNested, error) {
	if raw == nil {
		return nil, nil, nil
	}
	// Decode into a map for values.
	var toolMap map[string]legacyToolNested
	if err := json.Unmarshal(raw, &toolMap); err != nil {
		return nil, nil, err
	}
	// Use json.Decoder token scanning to get ordered keys.
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	var order []string
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch v := tok.(type) {
		case json.Delim:
			if v == '{' {
				depth++
			} else if v == '}' {
				depth--
			}
		case string:
			if depth == 1 {
				order = append(order, v)
				// skip the value token(s)
				var discard json.RawMessage
				if decErr := dec.Decode(&discard); decErr != nil {
					break
				}
			}
		}
	}
	return order, toolMap, nil
}

// MigrateConfig reads a .git-hooks.config.json (or its .dist variant) from
// the given path and emits an equivalent forge.toml to stdout (or the
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

	if legacy.Hooks != nil {
		// econnect nested format
		if err := migrateNestedHookSection(&sb, "pre-commit", legacy.Hooks.PreCommit); err != nil {
			return err
		}
		if err := migrateNestedHookSection(&sb, "commit-msg", legacy.Hooks.CommitMsg); err != nil {
			return err
		}
		if err := migrateNestedHookSection(&sb, "pre-push", legacy.Hooks.PrePush); err != nil {
			return err
		}
	} else {
		// original flat format
		migrateHookSection(&sb, "pre-commit", legacy.PreCommit)
		migrateHookSection(&sb, "commit-msg", legacy.CommitMsg)
		migrateHookSection(&sb, "pre-push", legacy.PrePush)
	}

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

// migrateNestedHookSection handles the econnect format where tools is a named map.
func migrateNestedHookSection(sb *strings.Builder, hookName string, hook *legacyHookNested) error {
	if hook == nil {
		return nil
	}

	order, toolMap, err := parseOrderedTools(hook.Tools)
	if err != nil {
		return fmt.Errorf("parsing tools for %s: %w", hookName, err)
	}
	if len(order) == 0 {
		return nil
	}

	sb.WriteString("[hooks.")
	sb.WriteString(hookName)
	sb.WriteString("]\nenabled = true\n\n")

	for _, toolName := range order {
		t, ok := toolMap[toolName]
		if !ok {
			continue
		}
		name := sanitizeEnvKey(toolName)
		sb.WriteString("[hooks.")
		sb.WriteString(hookName)
		sb.WriteString(".tools.")
		sb.WriteString(strings.ToLower(name))
		sb.WriteString("]\n")

		cmd := t.Command
		if cmd == "" {
			cmd = toolName
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
		if t.RunForEachFile {
			sb.WriteString("run_per_file = true\n")
		}
		if t.StagesFilesAfter {
			sb.WriteString("restage = true\n")
		}
		if t.OnFailure != "" {
			sb.WriteString(fmt.Sprintf("on_failure = %q\n", t.OnFailure))
		}
		sb.WriteString("\n")
	}
	return nil
}
