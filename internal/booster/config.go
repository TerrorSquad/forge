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
	Enabled   *bool                 `toml:"enabled"`
	Parallel  *bool                 `toml:"parallel"`   // nil = use [execution] parallel default
	SafeStash *bool                 `toml:"safe_stash"` // stash unstaged changes before fixers; auto when any tool has restage=true
	Tools     map[string]ToolConfig `toml:"tools"`
	Policy    *CommitMessagePolicy  `toml:"policy"`
}

type CommitMessagePolicy struct {
	ConventionalCommits bool     `toml:"conventional_commits"`
	AppendTicketFooter  bool     `toml:"append_ticket_footer"`
	RequireTicket       bool     `toml:"require_ticket"`
	BranchPattern       string   `toml:"branch_pattern"`
	ValidateBranchName  bool     `toml:"validate_branch_name"`
	SkippedBranches     []string `toml:"skipped_branches"`
	FooterLabel         string   `toml:"footer_label"`
	// prepare-commit-msg specific
	PrependTicket bool `toml:"prepend_ticket"`
	SkipOnMerge   bool `toml:"skip_on_merge"`
	SkipIfPresent bool `toml:"skip_if_present"`
}

type ToolConfig struct {
	Command           string            `toml:"command"`
	Args              []string          `toml:"args"`
	Type              string            `toml:"type"`
	Backend           string            `toml:"backend"`
	Extensions        []string          `toml:"extensions"`
	IncludePatterns   []string          `toml:"include_patterns"`
	ExcludePatterns   []string          `toml:"exclude_patterns"`
	PassFiles         *bool             `toml:"pass_files"`
	RunPerFile        bool              `toml:"run_per_file"`
	Restage           bool              `toml:"restage"`
	OnFailure         string            `toml:"on_failure"`
	Group             string            `toml:"group"`
	When              string            `toml:"when"`
	Timeout           string            `toml:"timeout"`              // e.g. "120s", "2m"; empty = use global default
	Cache             bool              `toml:"cache"`                // enable run cache for this tool
	CheckArgs         []string          `toml:"check_args"`           // args override used with --check flag
	CheckFailIfOutput bool              `toml:"check_fail_if_output"` // treat any stdout output as failure in --check mode
	DependsOn         []string          `toml:"depends_on"`           // tool names that must complete before this one (parallel mode)
	StageOutputs      []string          `toml:"stage_outputs"`        // files to git add after this tool runs (regardless of exit code); useful for generated artifacts
	ShowOutput        bool              `toml:"show_output"`          // print stdout/stderr even on success (e.g. test result counts)
	Env               map[string]string `toml:"env"`                  // extra environment variables injected only for this tool's process
}

// ExecutionConfig holds repository-wide execution defaults.
type ExecutionConfig struct {
	DefaultBackend string `toml:"default_backend"`
	ToolTimeout    string `toml:"tool_timeout"` // global default, e.g. "300s"; empty = no limit
	Cache          bool   `toml:"cache"`        // enable run cache globally
	Parallel       bool   `toml:"parallel"`     // run hook tools concurrently (opt-in)
}

// WorkspaceConfig defines monorepo member discovery.
type WorkspaceConfig struct {
	Members []string `toml:"members"`
}

func InitConfig(force bool, preset string) error {
	return InitConfigWithOptions(force, false, preset)
}

func InitConfigWithOptions(force, yes bool, preset string) error {
	path := "booster.toml"
	_, err := os.Stat(path)
	if err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var content string
	if strings.HasPrefix(preset, "https://") {
		fetched, err := fetchRemotePreset(preset, yes)
		if err != nil {
			return err
		}
		content = fetched
	} else if strings.HasPrefix(preset, "http://") {
		return fmt.Errorf("remote presets require an https:// URL, got %q", preset)
	} else if preset != "" {
		p, ok := presets[strings.ToLower(preset)]
		if !ok {
			return fmt.Errorf("unknown preset %q; available: %s", preset, strings.Join(ListPresets(), ", "))
		}
		content = p
	} else {
		content = defaultConfig
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

		// Merge global config under repo config (repo wins on conflict).
		if global, err := loadGlobalConfig(); err == nil && global != nil {
			mergeGlobalConfig(global, &cfg)
		}

		return &cfg, p, nil
	}

	return nil, "", fmt.Errorf("no config found; run 'booster init' to create booster.toml")
}

// globalConfigPath returns the path to the user-level global config.
// Override with BOOSTER_GLOBAL_CONFIG env var.
func globalConfigPath() string {
	if v := os.Getenv("BOOSTER_GLOBAL_CONFIG"); v != "" {
		return v
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "booster", "config.toml")
}

// loadGlobalConfig loads the user-level config; returns nil if it doesn't exist.
// A parse error produces a warning but is not fatal.
func loadGlobalConfig() (*Config, error) {
	p := globalConfigPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: invalid global config at %s: %v\n", p, err)
		return nil, nil
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string]HookConfig{}
	}
	return &cfg, nil
}

// mergeGlobalConfig merges global (user) config into repo config.
// Repo values always win. Only execution scalars and per-hook tools/policy are merged.
func mergeGlobalConfig(global, repo *Config) {
	// Execution scalars: only fill if repo has empty value.
	if repo.Execution.DefaultBackend == "" {
		repo.Execution.DefaultBackend = global.Execution.DefaultBackend
	}
	if repo.Execution.ToolTimeout == "" {
		repo.Execution.ToolTimeout = global.Execution.ToolTimeout
	}

	// Hooks: merge tools and policy from global if hook exists in repo.
	for hookName, globalHook := range global.Hooks {
		repoHook, exists := repo.Hooks[hookName]
		if !exists {
			continue // only merge into hooks that exist in repo config
		}

		// Merge tools: global tools not present in repo are added.
		if repoHook.Tools == nil {
			repoHook.Tools = map[string]ToolConfig{}
		}
		for toolName, globalTool := range globalHook.Tools {
			if _, ok := repoHook.Tools[toolName]; !ok {
				repoHook.Tools[toolName] = globalTool
			}
		}

		// Merge policy: fill zero-value fields from global.
		if globalHook.Policy != nil && repoHook.Policy == nil {
			p := *globalHook.Policy
			repoHook.Policy = &p
		}

		repo.Hooks[hookName] = repoHook
	}
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
