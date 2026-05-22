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

var ErrHookSkipped = errors.New("hook skipped")

type Config struct {
	Hooks     map[string]HookConfig `toml:"hooks"`
	Execution ExecutionConfig       `toml:"execution"`
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
}

// ExecutionConfig holds repository-wide execution defaults.
type ExecutionConfig struct {
	DefaultBackend string `toml:"default_backend"`
}

func InitConfig(force bool) error {
	path := "booster.toml"
	_, err := os.Stat(path)
	if err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return os.WriteFile(path, []byte(defaultConfig), 0644)
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
