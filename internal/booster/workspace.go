package booster

import (
	"fmt"
	"os"
	"path/filepath"
)

// matchingMembers returns workspace member directories whose path prefix matches
// at least one of the provided staged file paths.
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

// runWorkspaceHook runs the given hook for each matching workspace member.
// Each member must have a booster.toml (or the global config will be used).
func runWorkspaceHook(repoRoot, hookName, editFile string, members []string) error {
	for _, member := range members {
		memberRoot := filepath.Join(repoRoot, member)
		fmt.Printf("[workspace] %s\n", member)

		// Look for a member-local config; fall back to root config
		localConfig := filepath.Join(memberRoot, "booster.toml")
		if _, err := os.Stat(localConfig); err != nil {
			localConfig = filepath.Join(repoRoot, "booster.toml")
		}

		cfg, _, err := loadConfigFromPath(localConfig)
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
		if err := runHookCfg(memberRoot, hookName, editFile, hookCfg, cfg.Execution, staged, false, false); err != nil {
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
	// Ensure prefix ends with separator so "apps/foo" doesn't match "apps/foobar"
	if len(prefix) == 0 {
		return true
	}
	p := prefix
	if p[len(p)-1] != filepath.Separator {
		p += string(filepath.Separator)
	}
	return len(filePath) >= len(p) && filePath[:len(p)] == p
}

// stagedFilesForMember returns staged files that belong to the given member sub-directory.
func stagedFilesForMember(repoRoot, member string) ([]string, error) {
	all, err := stagedFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	prefix := member + string(filepath.Separator)
	var filtered []string
	for _, f := range all {
		if len(f) > len(prefix) && f[:len(prefix)] == prefix {
			// Strip the member prefix — tools operate within the member root
			filtered = append(filtered, f[len(prefix):])
		}
	}
	return filtered, nil
}
