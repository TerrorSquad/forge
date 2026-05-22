package booster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DoctorOptions controls doctor behaviour.
type DoctorOptions struct {
	Fix    bool
	DryRun bool
}

func Doctor() error { return DoctorWithOptions(DoctorOptions{}) }

func DoctorWithOptions(opts DoctorOptions) error {
	exe, err := os.Executable()
	if err != nil {
		exe = "unknown"
	}
	fmt.Printf("booster binary: %s\n", exe)

	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Printf("git repo: not detected (%v)\n", err)
		return nil
	}
	fmt.Printf("git repo root: %s\n", repoRoot)

	cfg, cfgPath, cfgErr := LoadConfig(repoRoot)
	if cfgErr != nil {
		fmt.Printf("config: missing (%v)\n", cfgErr)
		if opts.Fix && !opts.DryRun {
			fmt.Printf("  → run 'booster init' to create booster.toml\n")
		} else if opts.Fix && opts.DryRun {
			fmt.Printf("  [dry-run] would suggest: booster init\n")
		}
	} else {
		fmt.Printf("config: %s\n", cfgPath)
		fmt.Printf("configured hooks: %s\n", strings.Join(sortedHookNames(cfg.Hooks), ", "))
	}

	// Check core.hooksPath
	hooksPath, _ := localHooksPath(repoRoot)
	wantHooksPath := ".booster/hooks"
	if hooksPath == "" {
		fmt.Println("core.hooksPath: not set")
		if opts.Fix {
			if opts.DryRun {
				fmt.Printf("  [dry-run] would set core.hooksPath = %s\n", wantHooksPath)
			} else {
				fmt.Printf("  → setting core.hooksPath = %s ...", wantHooksPath)
				cmd := exec.Command("git", "-C", repoRoot, "config", "core.hooksPath", wantHooksPath)
				if out, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf(" ✗ (%v: %s)\n", err, out)
				} else {
					fmt.Println(" ✓")
					hooksPath = wantHooksPath
				}
			}
		}
	} else if hooksPath != wantHooksPath {
		fmt.Printf("core.hooksPath: %s (expected %s — skipping fix)\n", hooksPath, wantHooksPath)
	} else {
		fmt.Printf("core.hooksPath: %s\n", hooksPath)
	}

	// Check hook shims
	if hooksPath == wantHooksPath {
		anyMissing := false
		for _, hook := range supportedHooks {
			p := filepath.Join(repoRoot, ".booster", "hooks", hook)
			if _, err := os.Stat(p); err == nil {
				fmt.Printf("hook %s: installed\n", hook)
			} else {
				anyMissing = true
				fmt.Printf("hook %s: missing\n", hook)
			}
		}
		if anyMissing && opts.Fix {
			if opts.DryRun {
				fmt.Println("  [dry-run] would reinstall hook shims via InstallHooks()")
			} else {
				fmt.Print("  → reinstalling hook shims ...")
				if err := InstallHooks(); err != nil {
					fmt.Printf(" ✗ (%v)\n", err)
				} else {
					fmt.Println(" ✓")
				}
			}
		}
	}

	if cfgErr == nil {
		missing := checkToolAvailability(cfg)
		if len(missing) == 0 {
			fmt.Println("tool binaries: ok")
		} else {
			fmt.Println("tool binaries missing from PATH (cannot auto-fix):")
			for _, m := range missing {
				fmt.Printf("- %s\n", m)
			}
		}
	}

	return nil
}

func sortedHookNames(hooks map[string]HookConfig) []string {
	out := make([]string, 0, len(hooks))
	for k := range hooks {
		out = append(out, k)
	}
	sortStrings(out)
	return out
}

func sortStrings(items []string) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j] < items[i] {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func checkToolAvailability(cfg *Config) []string {
	set := map[string]struct{}{}
	for _, hook := range cfg.Hooks {
		for _, tool := range hook.Tools {
			cmd := strings.TrimSpace(tool.Command)
			if cmd == "" {
				continue
			}
			if _, ok := set[cmd]; ok {
				continue
			}
			if _, err := exec.LookPath(cmd); err != nil {
				set[cmd] = struct{}{}
			}
		}
	}
	missing := make([]string, 0, len(set))
	for cmd := range set {
		missing = append(missing, cmd)
	}
	sortStrings(missing)
	return missing
}
