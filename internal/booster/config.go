package booster

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

const defaultConfig = `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.prettier]
command = "prettier"
args = ["--write", "--ignore-unknown"]
type = "node"
extensions = [".js", ".ts", ".json", ".md", ".yml", ".yaml"]
restage = true
group = "format"

[hooks.pre-commit.tools.eslint]
command = "eslint"
args = ["--fix", "--cache", "--no-warn-ignored"]
type = "node"
extensions = [".js", ".jsx", ".ts", ".tsx", ".vue"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`

// presets is the set of named starter configurations.
var presets = map[string]string{
	"node": defaultConfig,
	"php": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.ecs]
command = "vendor/bin/ecs"
args = ["check", "--fix"]
type = "php"
extensions = [".php"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"php-node": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.ecs]
command = "vendor/bin/ecs"
args = ["check", "--fix"]
type = "php"
extensions = [".php"]
restage = true
group = "lint"

[hooks.pre-commit.tools.prettier]
command = "prettier"
args = ["--write", "--ignore-unknown"]
type = "node"
extensions = [".js", ".ts", ".json", ".md", ".yml", ".yaml"]
restage = true
group = "format"

[hooks.pre-commit.tools.eslint]
command = "eslint"
args = ["--fix", "--cache", "--no-warn-ignored"]
type = "node"
extensions = [".js", ".jsx", ".ts", ".tsx", ".vue"]
restage = true
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"go": `[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
args = ["-w"]
type = "system"
extensions = [".go"]
restage = true
group = "format"

[hooks.pre-commit.tools.govet]
command = "go"
args = ["vet", "./..."]
type = "system"
extensions = [".go"]
pass_files = false
group = "lint"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = false
require_ticket = false

[hooks.pre-push]
enabled = false
`,
	"minimal": `[hooks.pre-commit]
enabled = true

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true

[hooks.pre-push]
enabled = false
`,
}

// ListPresets returns the sorted preset names.
func ListPresets() []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var ErrHookSkipped = errors.New("hook skipped")

// Version information injected at build time via ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

type Config struct {
	Hooks     map[string]HookConfig `toml:"hooks"`
	Execution ExecutionConfig       `toml:"execution"`
	Workspace WorkspaceConfig       `toml:"workspace"`
}

type HookConfig struct {
	Enabled *bool                 `toml:"enabled"`
	Tools   map[string]ToolConfig `toml:"tools"`
	Policy  *CommitMessagePolicy  `toml:"policy"`
}

type CommitMessagePolicy struct {
	ConventionalCommits bool   `toml:"conventional_commits"`
	AppendTicketFooter  bool   `toml:"append_ticket_footer"`
	RequireTicket       bool   `toml:"require_ticket"`
	BranchPattern       string `toml:"branch_pattern"`
}

type ToolConfig struct {
	Command         string   `toml:"command"`
	Args            []string `toml:"args"`
	Type            string   `toml:"type"`
	Backend         string   `toml:"backend"`
	Extensions      []string `toml:"extensions"`
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`
	PassFiles       *bool    `toml:"pass_files"`
	RunPerFile      bool     `toml:"run_per_file"`
	Restage         bool     `toml:"restage"`
	OnFailure       string   `toml:"on_failure"`
	Group           string   `toml:"group"`
	When            string   `toml:"when"`
	Timeout         string   `toml:"timeout"` // e.g. "120s", "2m"; empty = use global default
}

// ExecutionConfig holds repository-wide execution defaults.
type ExecutionConfig struct {
	DefaultBackend string `toml:"default_backend"`
	ToolTimeout    string `toml:"tool_timeout"` // global default, e.g. "300s"; empty = no limit
}

// WorkspaceConfig defines monorepo member discovery.
type WorkspaceConfig struct {
	Members []string `toml:"members"`
}

func InitConfig(force bool, preset string) error {
	path := "booster.toml"
	_, err := os.Stat(path)
	if err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	content := defaultConfig
	if preset != "" {
		p, ok := presets[strings.ToLower(preset)]
		if !ok {
			return fmt.Errorf("unknown preset %q; available: %s", preset, strings.Join(ListPresets(), ", "))
		}
		content = p
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func loadConfigFromPath(p string) (*Config, string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, "", err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, "", fmt.Errorf("invalid config at %s: %w", p, err)
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string]HookConfig{}
	}
	return &cfg, p, nil
}

func LoadConfig(repoRoot string) (*Config, string, error) {
	candidates := []string{}
	if envPath := strings.TrimSpace(os.Getenv("BOOSTER_CONFIG")); envPath != "" {
		if filepath.IsAbs(envPath) {
			candidates = append(candidates, envPath)
		} else {
			candidates = append(candidates, filepath.Join(repoRoot, envPath))
		}
	}
	candidates = append(candidates, filepath.Join(repoRoot, "booster.toml"))

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", err
		}

		var cfg Config
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, "", fmt.Errorf("invalid config at %s: %w", p, err)
		}

		if cfg.Hooks == nil {
			cfg.Hooks = map[string]HookConfig{}
		}

		return &cfg, p, nil
	}

	return nil, "", fmt.Errorf("no config found; run 'booster init' to create booster.toml")
}

func sortedToolNames(tools map[string]ToolConfig) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (h HookConfig) IsEnabled() bool {
	if h.Enabled == nil {
		return true
	}
	return *h.Enabled
}

func (t ToolConfig) PassFilesEnabled() bool {
	if t.PassFiles == nil {
		return true
	}
	return *t.PassFiles
}
