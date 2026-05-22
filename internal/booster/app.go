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
	case "init":
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		force := fs.Bool("force", false, "overwrite booster.toml if it already exists")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if err := InitConfig(*force); err != nil {
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
	case "run":
		return runCommand(args[1:])
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
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	if *edit == "" {
		extra := fs.Args()
		if len(extra) > 0 {
			*edit = extra[0]
		}
	}

	if err := RunHook(hook, *edit); err != nil {
		if errors.Is(err, ErrHookSkipped) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		return 1
	}

	return 0
}

func printHelp() {
	fmt.Println(`booster - policy-driven git hook runner

Usage:
  booster init [--force]
  booster install
  booster uninstall
  booster run <hook> [--edit FILE]
  booster doctor

Examples:
  booster init
  booster install
  booster run pre-commit
  booster run commit-msg --edit .git/COMMIT_EDITMSG`)
}
