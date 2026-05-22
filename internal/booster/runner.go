package booster

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ticketRegex = regexp.MustCompile(`([A-Z]+-[0-9]+)`)
var conventionalRegex = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([^)]+\))?!?: .+`)

func RunHook(hookName string, editFile string) error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	if isHookSkippedEnv(hookName) {
		fmt.Printf("Skipping %s (env skip set)\n", hookName)
		return ErrHookSkipped
	}

	cfg, configPath, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}
	fmt.Printf("Using config: %s\n", configPath)

	hookCfg, ok := cfg.Hooks[hookName]
	if !ok || !hookCfg.IsEnabled() {
		fmt.Printf("Hook %s disabled or not configured\n", hookName)
		return ErrHookSkipped
	}

	if hookName == "commit-msg" {
		if err := applyCommitMessagePolicy(repoRoot, hookCfg.Policy, editFile); err != nil {
			return err
		}
	}

	files := []string{}
	if hookName == "pre-commit" {
		files, err = stagedFiles(repoRoot)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Println("No staged files. Nothing to run.")
			return nil
		}
	}

	toolNames := sortedToolNames(hookCfg.Tools)
	if len(toolNames) == 0 {
		fmt.Printf("No tools configured for %s\n", hookName)
		return nil
	}

	allowedGroups := parseAllowedGroups()
	failed := false

	for _, name := range toolNames {
		tool := hookCfg.Tools[name]

		if shouldSkipTool(name) {
			fmt.Printf("- %s: skipped\n", name)
			continue
		}
		if len(allowedGroups) > 0 && tool.Group != "" {
			if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
				fmt.Printf("- %s: skipped by HOOKS_ONLY\n", name)
				continue
			}
		}
		if strings.TrimSpace(tool.Command) == "" {
			return fmt.Errorf("tool %s: command is required", name)
		}

		filesToRun := filterFiles(files, tool)
		if hookName == "pre-commit" && tool.PassFilesEnabled() && len(filesToRun) == 0 {
			fmt.Printf("- %s: no matching files\n", name)
			continue
		}

		fmt.Printf("- %s: running\n", name)
		err := executeTool(repoRoot, tool, filesToRun)
		if err != nil {
			fmt.Printf("  %s failed: %v\n", name, err)
			failed = true
			if strings.EqualFold(strings.TrimSpace(tool.OnFailure), "stop") {
				return fmt.Errorf("tool %s failed and requested stop", name)
			}
			continue
		}

		if tool.Restage && tool.PassFilesEnabled() {
			if err := addFiles(repoRoot, filesToRun); err != nil {
				return fmt.Errorf("tool %s restage failed: %w", name, err)
			}
		}
	}

	if failed {
		return errors.New("one or more tools failed")
	}

	return nil
}

func applyCommitMessagePolicy(repoRoot string, policy *CommitMessagePolicy, editFile string) error {
	if policy == nil {
		return nil
	}

	if editFile == "" {
		return fmt.Errorf("commit-msg policy requires message file (pass --edit or let git pass the file path)")
	}

	content, err := os.ReadFile(editFile)
	if err != nil {
		return err
	}
	lines := splitLines(string(content))
	if len(lines) == 0 {
		return fmt.Errorf("empty commit message")
	}

	subject := strings.TrimSpace(lines[0])
	if policy.ConventionalCommits && !conventionalRegex.MatchString(subject) {
		return fmt.Errorf("commit subject does not follow conventional commits: %q", subject)
	}

	branch, err := currentBranch(repoRoot)
	if err != nil {
		return err
	}

	ticket := ""
	if m := ticketRegex.FindStringSubmatch(branch); len(m) > 1 {
		ticket = m[1]
	}

	if policy.RequireTicket && ticket == "" {
		return fmt.Errorf("ticket required by policy but branch %q has no ticket (expected e.g. PRJ-123)", branch)
	}

	if policy.AppendTicketFooter && ticket != "" {
		footer := fmt.Sprintf("Closes: %s", ticket)
		if !containsLine(lines, footer) {
			text := strings.TrimRight(string(content), "\n") + "\n\n" + footer + "\n"
			if err := os.WriteFile(editFile, []byte(text), 0644); err != nil {
				return err
			}
			fmt.Printf("Appended commit footer: %s\n", footer)
		}
	}

	return nil
}

func executeTool(repoRoot string, tool ToolConfig, files []string) error {
	if tool.RunPerFile {
		for _, file := range files {
			args := append([]string{}, tool.Args...)
			if tool.PassFilesEnabled() {
				args = append(args, file)
			}
			if err := runToolCommand(repoRoot, tool.Command, args); err != nil {
				return err
			}
		}
		return nil
	}

	args := append([]string{}, tool.Args...)
	if tool.PassFilesEnabled() {
		args = append(args, files...)
	}
	return runToolCommand(repoRoot, tool.Command, args)
}

func runToolCommand(repoRoot, cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func filterFiles(files []string, tool ToolConfig) []string {
	if len(files) == 0 {
		return files
	}

	matches := make([]string, 0, len(files))
	for _, f := range files {
		if !matchExt(f, tool.Extensions) {
			continue
		}
		if !matchPatterns(f, tool.IncludePatterns, true) {
			continue
		}
		if matchPatterns(f, tool.ExcludePatterns, false) {
			continue
		}
		matches = append(matches, filepath.ToSlash(f))
	}
	sort.Strings(matches)
	return matches
}

func matchExt(file string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	for _, ext := range extensions {
		if strings.EqualFold(filepath.Ext(file), ext) {
			return true
		}
	}
	return false
}

func matchPatterns(file string, patterns []string, defaultWhenEmpty bool) bool {
	if len(patterns) == 0 {
		return defaultWhenEmpty
	}
	for _, p := range patterns {
		ok, err := filepath.Match(p, file)
		if err == nil && ok {
			return true
		}
	}
	return false
}

func parseAllowedGroups() map[string]struct{} {
	raw := strings.TrimSpace(os.Getenv("HOOKS_ONLY"))
	if raw == "" {
		return nil
	}
	res := map[string]struct{}{}
	s := bufio.NewScanner(strings.NewReader(raw))
	s.Split(splitComma)
	for s.Scan() {
		v := strings.TrimSpace(strings.ToLower(s.Text()))
		if v != "" {
			res[v] = struct{}{}
		}
	}
	return res
}

func splitComma(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == ',' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func isHookSkippedEnv(hook string) bool {
	key := "SKIP_" + strings.ToUpper(strings.ReplaceAll(hook, "-", ""))
	return isTruthy(os.Getenv(key))
}

func shouldSkipTool(name string) bool {
	key := "SKIP_" + sanitizeEnvKey(name)
	return isTruthy(os.Getenv(key))
}

func sanitizeEnvKey(name string) string {
	up := strings.ToUpper(name)
	b := strings.Builder{}
	lastUnderscore := false
	for _, r := range up {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func isTruthy(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func containsLine(lines []string, line string) bool {
	for _, l := range lines {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}
