package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// ErrHookSkipped is returned when a hook run is intentionally skipped.
var ErrHookSkipped = errors.New("hook skipped")

type Config struct {
	Hooks     map[string]HookConfig `toml:"hooks"`
	Execution ExecutionConfig       `toml:"execution"`
	Workspace WorkspaceConfig       `toml:"workspace"`
	Update    UpdateConfig          `toml:"update"`
}

// UpdateConfig controls self-update behaviour.
type UpdateConfig struct {
	// PinVersion warns when the running binary differs from this version.
	PinVersion string `toml:"pin_version"`
	// Channel selects the release channel: "stable" (default) or "rc".
	Channel string `toml:"channel"`
}

type HookConfig struct {
	Enabled   *bool                 `toml:"enabled"`
	Parallel  *bool                 `toml:"parallel"`
	Tools     map[string]ToolConfig `toml:"tools"`
	Policy    *CommitMessagePolicy  `toml:"policy"`
	toolOrder []string
}

type CommitMessagePolicy struct {
	ConventionalCommits bool     `toml:"conventional_commits"`
	AppendTicketFooter  bool     `toml:"append_ticket_footer"`
	RequireTicket       bool     `toml:"require_ticket"`
	BranchPattern       string   `toml:"branch_pattern"`
	ValidateBranchName  bool     `toml:"validate_branch_name"`
	SkippedBranches     []string `toml:"skipped_branches"`
	FooterLabel         string   `toml:"footer_label"`
	PrependTicket       bool     `toml:"prepend_ticket"`
	SkipOnMerge         bool     `toml:"skip_on_merge"`
	SkipIfPresent       bool     `toml:"skip_if_present"`
	// TicketPattern is a regexp matched against the branch name to extract a
	// ticket ID. Defaults to [A-Z]+-[0-9]+ (Jira/Linear style) when empty.
	TicketPattern string `toml:"ticket_pattern"`
	// AllowedTypes restricts Conventional Commits to a custom type list.
	// Defaults to: feat, fix, docs, style, refactor, perf, test, build, ci,
	// chore, revert.
	AllowedTypes []string `toml:"allowed_types"`
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
	Timeout           string            `toml:"timeout"`
	Cache             bool              `toml:"cache"`
	CheckArgs         []string          `toml:"check_args"`
	CheckFailIfOutput bool              `toml:"check_fail_if_output"`
	DependsOn         []string          `toml:"depends_on"`
	StageOutputs      []string          `toml:"stage_outputs"`
	ShowOutput        bool              `toml:"show_output"`
	Env               map[string]string `toml:"env"`
}

// ExecutionConfig holds repository-wide execution defaults.
type ExecutionConfig struct {
	DefaultBackend string `toml:"default_backend"`
	ToolTimeout    string `toml:"tool_timeout"`
	Cache          bool   `toml:"cache"`
	Parallel       bool   `toml:"parallel"`
	// CacheTTL is the maximum age of a cache entry before it is evicted.
	// Accepts Go duration strings (e.g. "24h", "7d" is not valid — use "168h").
	// Zero or empty means entries never expire.
	CacheTTL string `toml:"cache_ttl"`
	// CacheMaxSize is the maximum number of entries kept in cache.json.
	// When the limit is exceeded the oldest entries are evicted first.
	// Zero means unlimited.
	CacheMaxSize int `toml:"cache_max_size"`
}

// WorkspaceConfig defines monorepo member discovery.
type WorkspaceConfig struct {
	Members []string `toml:"members"`
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

// SortedToolNames returns the sorted keys of the tools map.
func SortedToolNames(tools map[string]ToolConfig) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// OrderedToolNames returns tool names in the order declared in the config file.
// Tools not present in the declaration order are appended alphabetically.
func (h HookConfig) OrderedToolNames() []string {
	if len(h.toolOrder) == 0 {
		return SortedToolNames(h.Tools)
	}

	names := make([]string, 0, len(h.Tools))
	seen := map[string]struct{}{}
	for _, name := range h.toolOrder {
		if _, ok := h.Tools[name]; ok {
			if _, already := seen[name]; !already {
				names = append(names, name)
				seen[name] = struct{}{}
			}
		}
	}

	if len(names) != len(h.Tools) {
		extra := make([]string, 0, len(h.Tools)-len(names))
		for name := range h.Tools {
			if _, ok := seen[name]; !ok {
				extra = append(extra, name)
			}
		}
		sort.Strings(extra)
		names = append(names, extra...)
	}
	return names
}

var toolSectionRE = regexp.MustCompile(`(?m)^\s*\[\s*hooks\.([^\.\s\]]+)\.tools\.(?:(?:"([^"]+)")|(?:'([^']+)')|([^\]\s]+))\s*\]\s*$`)

func parseHookToolOrder(data []byte) map[string][]string {
	matches := toolSectionRE.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return nil
	}

	order := map[string][]string{}
	seen := map[string]map[string]struct{}{}
	for _, match := range matches {
		hook := string(match[1])
		tool := ""
		switch {
		case len(match[2]) > 0:
			tool = string(match[2])
		case len(match[3]) > 0:
			tool = string(match[3])
		case len(match[4]) > 0:
			tool = string(match[4])
		default:
			continue
		}
		if seen[hook] == nil {
			seen[hook] = map[string]struct{}{}
		}
		if _, ok := seen[hook][tool]; ok {
			continue
		}
		seen[hook][tool] = struct{}{}
		order[hook] = append(order[hook], tool)
	}
	return order
}

func loadConfigFromBytes(data []byte) (*Config, error) {
	var cfg Config
	decoder := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string]HookConfig{}
	}

	for hookName, toolNames := range parseHookToolOrder(data) {
		hook, ok := cfg.Hooks[hookName]
		if !ok {
			continue
		}
		hook.toolOrder = toolNames
		cfg.Hooks[hookName] = hook
	}

	return &cfg, nil
}

func InitConfig(force bool, preset string) error {
	return InitConfigWithOptions(force, false, preset)
}

func InitConfigWithOptions(force, yes bool, preset string) error {
	path := "forge.toml"
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

// LoadConfigFromPath loads config from the exact path given, without candidate search.
func LoadConfigFromPath(p string) (*Config, string, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, "", err
	}
	cfg, err := loadConfigFromBytes(data)
	if err != nil {
		return nil, "", fmt.Errorf("invalid config at %s: %w", p, err)
	}
	return cfg, p, nil
}

func LoadConfig(repoRoot string) (*Config, string, error) {
	candidates := []string{}
	if envPath := strings.TrimSpace(os.Getenv("FORGE_CONFIG")); envPath != "" {
		if filepath.IsAbs(envPath) {
			candidates = append(candidates, envPath)
		} else {
			candidates = append(candidates, filepath.Join(repoRoot, envPath))
		}
	}
	candidates = append(candidates, filepath.Join(repoRoot, "forge.toml"))

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", err
		}

		cfg, err := loadConfigFromBytes(data)
		if err != nil {
			return nil, "", fmt.Errorf("invalid config at %s: %w", p, err)
		}

		if global, err := loadGlobalConfig(); err == nil && global != nil {
			mergeGlobalConfig(global, cfg)
		}

		return cfg, p, nil
	}

	return nil, "", fmt.Errorf("no config found; run 'forge init' to create forge.toml")
}

// GlobalConfigPath returns the path to the user-level global config.
// Override with FORGE_GLOBAL_CONFIG env var.
func GlobalConfigPath() string {
	if v := os.Getenv("FORGE_GLOBAL_CONFIG"); v != "" {
		return v
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "forge", "config.toml")
}

func loadGlobalConfig() (*Config, error) {
	p := GlobalConfigPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	cfg, err := loadConfigFromBytes(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: invalid global config at %s: %v\n", p, err)
		return nil, nil
	}
	return cfg, nil
}

func mergeGlobalConfig(global, repo *Config) {
	if repo.Execution.DefaultBackend == "" {
		repo.Execution.DefaultBackend = global.Execution.DefaultBackend
	}
	if repo.Execution.ToolTimeout == "" {
		repo.Execution.ToolTimeout = global.Execution.ToolTimeout
	}

	for hookName, globalHook := range global.Hooks {
		repoHook, exists := repo.Hooks[hookName]
		if !exists {
			continue
		}

		if repoHook.Tools == nil {
			repoHook.Tools = map[string]ToolConfig{}
		}
		for toolName, globalTool := range globalHook.Tools {
			if _, ok := repoHook.Tools[toolName]; !ok {
				repoHook.Tools[toolName] = globalTool
			}
		}

		if len(globalHook.toolOrder) > 0 {
			if len(repoHook.toolOrder) == 0 {
				repoHook.toolOrder = append([]string(nil), globalHook.toolOrder...)
			} else {
				seen := map[string]struct{}{}
				for _, name := range repoHook.toolOrder {
					seen[name] = struct{}{}
				}
				for _, name := range globalHook.toolOrder {
					if _, ok := seen[name]; !ok {
						repoHook.toolOrder = append(repoHook.toolOrder, name)
						seen[name] = struct{}{}
					}
				}
			}
		}

		if globalHook.Policy != nil && repoHook.Policy == nil {
			p := *globalHook.Policy
			repoHook.Policy = &p
		}

		repo.Hooks[hookName] = repoHook
	}
}
