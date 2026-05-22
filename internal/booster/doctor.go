package booster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Doctor() error {
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
	} else {
		fmt.Printf("config: %s\n", cfgPath)
		fmt.Printf("configured hooks: %s\n", strings.Join(sortedHookNames(cfg.Hooks), ", "))
	}

	hooksPath, _ := localHooksPath(repoRoot)
	if hooksPath == "" {
		fmt.Println("core.hooksPath: not set (git default .git/hooks)")
	} else {
		fmt.Printf("core.hooksPath: %s\n", hooksPath)
	}

	if hooksPath == ".booster/hooks" {
		for _, hook := range supportedHooks {
			p := filepath.Join(repoRoot, ".booster", "hooks", hook)
			if _, err := os.Stat(p); err == nil {
				fmt.Printf("hook %s: installed\n", hook)
			} else {
				fmt.Printf("hook %s: missing\n", hook)
			}
		}
	}

	if cfgErr == nil {
		missing := checkToolAvailability(cfg)
		if len(missing) == 0 {
			fmt.Println("tool binaries: ok")
		} else {
			fmt.Println("tool binaries missing from PATH:")
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
