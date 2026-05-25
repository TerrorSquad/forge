package forge

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return 0
	case "version", "-v", "--version":
		fmt.Printf("booster %s (commit: %s, built: %s)\n", Version, Commit, Date)
		return 0
	case "init":
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		force := fs.Bool("force", false, "overwrite forge.toml if it already exists")
		preset := fs.String("preset", "", "starter preset (node, php, php-node, go, minimal) or https:// URL")
		listPresets := fs.Bool("list-presets", false, "list available presets and exit")
		yes := fs.Bool("yes", false, "skip confirmation prompt (also set by CI=true)")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if *listPresets {
			fmt.Println("Available presets:")
			for _, p := range ListPresets() {
				fmt.Printf("  %s\n", p)
			}
			return 0
		}
		if err := InitConfigWithOptions(*force, *yes, *preset); err != nil {
			fmt.Fprintf(os.Stderr, "init failed: %v\n", err)
			return 1
		}
		fmt.Println("Created forge.toml")
		return 0
	case "install":
		if err := InstallHooks(); err != nil {
			fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
			return 1
		}
		return 0
	case "uninstall":
		if err := UninstallHooks(); err != nil {
			fmt.Fprintf(os.Stderr, "uninstall failed: %v\n", err)
			return 1
		}
		return 0
	case "doctor":
		dfs := flag.NewFlagSet("doctor", flag.ContinueOnError)
		fix := dfs.Bool("fix", false, "automatically fix detected issues")
		dryRun := dfs.Bool("dry-run", false, "print what would be fixed without applying")
		if err := dfs.Parse(args[1:]); err != nil {
			return 2
		}
		opts := DoctorOptions{Fix: *fix, DryRun: *dryRun}
		if err := DoctorWithOptions(opts); err != nil {
			fmt.Fprintf(os.Stderr, "doctor failed: %v\n", err)
			return 1
		}
		return 0
	case "completion":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: booster completion <bash|zsh|fish>")
			return 2
		}
		script, err := GenerateCompletion(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "completion failed: %v\n", err)
			return 1
		}
		fmt.Print(script)
		return 0
	case "cache":
		return cacheCommand(args[1:])
	case "list":
		return listCommand()
	case "ci":
		return ciCommand()
	case "validate":
		return validateCommand()
	case "run":
		return runCommand(args[1:])
	case "migrate":
		return migrateCommand(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printHelp()
		return 2
	}
}

func runCommand(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: booster run <hook> [--edit FILE]")
		return 2
	}

	hook := args[0]
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	edit := fs.String("edit", "", "path to commit message file (for commit-msg hook)")
	allFiles := fs.Bool("all-files", false, "run against all tracked files (pre-commit only)")
	noCache := fs.Bool("no-cache", false, "bypass run cache for this invocation")
	checkMode := fs.Bool("check", false, "dry-run mode: use check_args, suppress restage, treat output as failure")
	tool := fs.String("tool", "", "only run these tools, comma-separated (e.g. phpstan,psalm)")
	group := fs.String("group", "", "only run tools in this group, comma-separated")
	skipTool := fs.String("skip-tool", "", "skip these tools by name, comma-separated")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	if *edit == "" {
		extra := fs.Args()
		if len(extra) > 0 {
			*edit = extra[0]
		}
	}

	opts := RunOptions{
		AllFiles:   *allFiles,
		NoCache:    *noCache,
		CheckMode:  *checkMode,
		OnlyTools:  splitCSV(*tool),
		OnlyGroups: splitCSV(*group),
		SkipTools:  splitCSV(*skipTool),
	}
	// Capture second positional arg as source (used by prepare-commit-msg)
	if extra := fs.Args(); len(extra) > 1 {
		opts.Source = extra[1]
	}
	if err := RunHookWithOptions(hook, *edit, opts); err != nil {
		if errors.Is(err, ErrHookSkipped) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		return 1
	}

	return 0
}

func validateCommand() int {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate: %v\n", err)
		return 1
	}
	cfg, _, err := LoadConfig(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate: %v\n", err)
		return 1
	}
	issues := ValidateConfig(cfg)
	if len(issues) == 0 {
		fmt.Fprintf(UI, "%s forge.toml is valid\n", green("✓"))
		return 0
	}
	hasError := PrintValidationIssues(issues)
	if hasError {
		return 1
	}
	return 0
}

func migrateCommand(args []string) int {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	from := fs.String("from", "", "path to .git-hooks.config.json (auto-detected if omitted)")
	to := fs.String("to", "-", "output path for forge.toml (- prints to stdout)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := MigrateConfig(*from, *to); err != nil {
		fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
		return 1
	}
	return 0
}

func printHelp() {
	fmt.Println(`booster - policy-driven git hook runner

Usage:
  booster init [--force] [--preset PRESET] [--list-presets]
  booster install
  booster uninstall
  booster run <hook> [--edit FILE] [--all-files] [--check] [--no-cache]
                     [--tool NAMES] [--group NAMES] [--skip-tool NAMES]
  booster validate
  booster list
  booster ci
  booster migrate [--from FILE] [--to FILE]
  booster doctor [--fix] [--dry-run]
  booster cache clear
  booster completion <bash|zsh|fish>

Run flags:
  --tool phpstan,psalm       only run the named tools
  --group analysis           only run tools in this group
  --skip-tool psalm          run everything except the named tools

Env vars:
  SKIP_<TOOL>=1              skip a specific tool (e.g. SKIP_PHPSTAN=1)
  SKIP_GROUP_<GROUP>=1       skip an entire group (e.g. SKIP_GROUP_ANALYSIS=1)

Presets:
  node, php, php-node, go, minimal

Examples:
  booster init --preset go
  booster install
  booster run pre-commit
  booster run pre-commit --tool phpstan
  booster run pre-commit --group analysis --all-files
  booster run pre-commit --skip-tool psalm
  booster run pre-commit --check --all-files
  booster list
  booster ci
  booster cache clear`)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func cacheCommand(args []string) int {
	if len(args) == 0 || args[0] != "clear" {
		fmt.Fprintln(os.Stderr, "usage: booster cache clear")
		return 2
	}
	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cache clear failed: %v\n", err)
		return 1
	}
	if err := ClearCache(repoRoot); err != nil {
		fmt.Fprintf(os.Stderr, "cache clear failed: %v\n", err)
		return 1
	}
	fmt.Println("cache cleared")
	return 0
}

// listCommand prints all configured hooks and their tools.
func listCommand() int {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list failed: %v\n", err)
		return 1
	}
	cfg, configPath, err := LoadConfig(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(UI, "%s\n\n", dim("config: "+configPath))

	if len(cfg.Hooks) == 0 {
		fmt.Fprintf(UI, "%s\n", dim("no hooks configured"))
		return 0
	}

	hookNames := make([]string, 0, len(cfg.Hooks))
	for name := range cfg.Hooks {
		hookNames = append(hookNames, name)
	}
	sort.Strings(hookNames)

	for _, hookName := range hookNames {
		hookCfg := cfg.Hooks[hookName]
		enabled := hookCfg.IsEnabled()
		statusIcon := green("✓")
		if !enabled {
			statusIcon = dim("·")
		}
		parallel := ""
		if isParallelMode(hookCfg, cfg.Execution) {
			parallel = dim(" [parallel]")
		}
		fmt.Fprintf(UI, "%s %s%s\n", statusIcon, bold(hookName), parallel)

		toolNames := sortedToolNames(hookCfg.Tools)
		for _, toolName := range toolNames {
			tool := hookCfg.Tools[toolName]
			backend := ""
			if tool.Backend != "" {
				backend = dim(" [" + tool.Backend + "]")
			}
			group := ""
			if tool.Group != "" {
				group = dim(" group:" + tool.Group)
			}
			fmt.Fprintf(UI, "    %s  %s%s%s\n", cyan("→"), toolName, backend, group)
			fmt.Fprintf(UI, "       %s\n", dim(tool.Command))
		}
		fmt.Fprintln(UI)
	}
	return 0
}

// ciCommand is an opinionated shortcut for CI pipelines:
// runs pre-commit in check + all-files + no-cache mode.
func ciCommand() int {
	if err := RunHookWithOptions("pre-commit", "", RunOptions{
		AllFiles:  true,
		CheckMode: true,
		NoCache:   true,
	}); err != nil {
		if errors.Is(err, ErrHookSkipped) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "ci check failed: %v\n", err)
		return 1
	}
	return 0
}
