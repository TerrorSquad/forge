package booster

import (
	"errors"
	"flag"
	"fmt"
	"os"
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
		force := fs.Bool("force", false, "overwrite booster.toml if it already exists")
		preset := fs.String("preset", "", "starter preset (node, php, php-node, go, minimal)")
		listPresets := fs.Bool("list-presets", false, "list available presets and exit")
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
		if err := InitConfig(*force, *preset); err != nil {
			fmt.Fprintf(os.Stderr, "init failed: %v\n", err)
			return 1
		}
		fmt.Println("Created booster.toml")
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
		if err := Doctor(); err != nil {
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
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	if *edit == "" {
		extra := fs.Args()
		if len(extra) > 0 {
			*edit = extra[0]
		}
	}

	opts := RunOptions{AllFiles: *allFiles}
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

func migrateCommand(args []string) int {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	from := fs.String("from", "", "path to .git-hooks.config.json (auto-detected if omitted)")
	to := fs.String("to", "-", "output path for booster.toml (- prints to stdout)")
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
  booster run <hook> [--edit FILE]
  booster migrate [--from FILE] [--to FILE]
  booster doctor

Presets:
  node, php, php-node, go, minimal

Examples:
  booster init --preset go
  booster init --list-presets
  booster install
  booster run pre-commit
  booster run commit-msg --edit .git/COMMIT_EDITMSG
  booster migrate --from .git-hooks.config.json --to booster.toml`)
}
