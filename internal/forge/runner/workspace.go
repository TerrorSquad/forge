package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/git"
)

func matchingMembers(repoRoot string, members []string, stagedFiles []string) ([]string, error) {
	var matched []string
	seen := map[string]bool{}

	for _, pattern := range members {
		abs := filepath.Join(repoRoot, pattern)
		expanded, err := filepath.Glob(abs)
		if err != nil {
			return nil, fmt.Errorf("invalid workspace member pattern %q: %w", pattern, err)
		}
		for _, dir := range expanded {
			rel, err := filepath.Rel(repoRoot, dir)
			if err != nil {
				continue
			}
			if !isDir(dir) {
				continue
			}
			if seen[rel] {
				continue
			}
			for _, file := range stagedFiles {
				if pathHasPrefix(file, rel) {
					seen[rel] = true
					matched = append(matched, rel)
					break
				}
			}
		}
	}

	return matched, nil
}

func runWorkspaceHook(repoRoot, hookName, editFile string, members []string) error {
	for _, member := range members {
		memberRoot := filepath.Join(repoRoot, member)
		fmt.Printf("[workspace] %s\n", member)

		localConfig := filepath.Join(memberRoot, "forge.toml")
		if _, err := os.Stat(localConfig); err != nil {
			localConfig = filepath.Join(repoRoot, "forge.toml")
		}

		cfg, _, err := config.LoadConfigFromPath(localConfig)
		if err != nil {
			return fmt.Errorf("[%s] config error: %w", member, err)
		}

		hookCfg, ok := cfg.Hooks[hookName]
		if !ok || !hookCfg.IsEnabled() {
			fmt.Printf("[workspace] %s: hook %s not configured\n", member, hookName)
			continue
		}

		staged, err := stagedFilesForMember(repoRoot, member)
		if err != nil {
			return err
		}
		if err := runHookCfg(memberRoot, hookName, editFile, hookCfg, cfg.Execution, staged, RunOptions{}); err != nil {
			return fmt.Errorf("[%s] %w", member, err)
		}
	}
	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func pathHasPrefix(filePath, prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	p := prefix
	if p[len(p)-1] != filepath.Separator {
		p += string(filepath.Separator)
	}
	return len(filePath) >= len(p) && filePath[:len(p)] == p
}

func stagedFilesForMember(repoRoot, member string) ([]string, error) {
	all, err := git.StagedFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	prefix := member + string(filepath.Separator)
	var filtered []string
	for _, f := range all {
		if len(f) > len(prefix) && f[:len(prefix)] == prefix {
			filtered = append(filtered, f[len(prefix):])
		}
	}
	return filtered, nil
}
