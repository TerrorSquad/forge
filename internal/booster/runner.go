package booster

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ticketRegex = regexp.MustCompile(`([A-Z]+-[0-9]+)`)
var conventionalRegex = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([^)]+\))?!?: .+`)

func RunHook(hookName string, editFile string) error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	if isHookSkippedEnv(hookName) {
		fmt.Fprintf(UI, "%s\n", yellow("~ "+hookName+" skipped (env skip set)"))
		return ErrHookSkipped
	}

	cfg, configPath, err := LoadConfig(repoRoot)
	if err != nil {
		return err
	}
	fmt.Fprintf(UI, "%s\n", dim("config: "+configPath))

	hookCfg, ok := cfg.Hooks[hookName]
	if !ok || !hookCfg.IsEnabled() {
		fmt.Fprintf(UI, "%s\n", dim(hookName+" disabled or not configured"))
		return ErrHookSkipped
	}

	// Workspace mode: run hook for each affected member
	if len(cfg.Workspace.Members) > 0 && hookName != "commit-msg" {
		staged, err := stagedFiles(repoRoot)
		if err != nil {
			return err
		}
		members, err := matchingMembers(repoRoot, cfg.Workspace.Members, staged)
		if err != nil {
			return err
		}
		if len(members) > 0 {
			return runWorkspaceHook(repoRoot, hookName, editFile, members)
		}
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
			fmt.Fprintf(UI, "%s\n", dim("no staged files — nothing to run"))
			return nil
		}
	}

	return runHookCfg(repoRoot, hookName, editFile, hookCfg, cfg.Execution, files)
}

// runHookCfg executes all tools in hookCfg for the given root / staged files.
// This is the inner loop used by both the root hook and workspace members.
func runHookCfg(root, hookName, editFile string, hookCfg HookConfig, exec ExecutionConfig, files []string) error {
	toolNames := sortedToolNames(hookCfg.Tools)
	if len(toolNames) == 0 {
		fmt.Fprintf(UI, "%s\n", dim("no tools configured for "+hookName))
		return nil
	}

	PrintHookHeader(hookName)

	allowedGroups := parseAllowedGroups()
	var results []ToolResult
	hookStart := time.Now()
	failed := false

	for _, name := range toolNames {
		tool := hookCfg.Tools[name]

		if shouldSkipTool(name) {
			r := ToolResult{Name: name, Status: "skip"}
			PrintToolResult(r)
			results = append(results, r)
			continue
		}
		if len(allowedGroups) > 0 && tool.Group != "" {
			if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
				r := ToolResult{Name: name, Status: "skip"}
				PrintToolResult(r)
				results = append(results, r)
				continue
			}
		}
		if strings.TrimSpace(tool.Command) == "" {
			return fmt.Errorf("tool %s: command is required", name)
		}

		filesToRun := filterFiles(files, tool)
		if hookName == "pre-commit" && tool.PassFilesEnabled() && len(filesToRun) == 0 {
			r := ToolResult{Name: name, Status: "skip"}
			PrintToolResult(r)
			results = append(results, r)
			continue
		}

		backend := ResolveBackend(root, tool, exec.DefaultBackend)
		start := time.Now()
		toolOut, err := executeToolCaptured(root, tool, filesToRun, backend)
		dur := time.Since(start)

		if err != nil {
			r := ToolResult{Name: name, Status: "fail", Duration: dur, Output: toolOut}
			PrintToolResult(r)
			results = append(results, r)
			failed = true
			if strings.EqualFold(strings.TrimSpace(tool.OnFailure), "stop") {
				PrintSummary(results, time.Since(hookStart))
				return fmt.Errorf("tool %s failed and requested stop", name)
			}
			continue
		}

		r := ToolResult{Name: name, Status: "pass", Duration: dur}
		PrintToolResult(r)
		results = append(results, r)

		if tool.Restage && tool.PassFilesEnabled() {
			if err := addFiles(root, filesToRun); err != nil {
				return fmt.Errorf("tool %s restage failed: %w", name, err)
			}
		}
	}

	PrintSummary(results, time.Since(hookStart))

	if failed {
		return errors.New("one or more tools failed")
	}

	return nil
}

// executeToolCaptured runs the tool and returns captured combined output on
// failure. On success the output is discarded (streamed to /dev/null is not
// quite right — we stream to a buffer and only surface it on error).
func executeToolCaptured(repoRoot string, tool ToolConfig, files []string, backend Backend) (string, error) {
	var buf bytes.Buffer
	err := executeToolWithWriter(repoRoot, tool, files, backend, &buf)
	if err != nil {
		return buf.String(), err
	}
	return "", nil
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

func executeTool(repoRoot string, tool ToolConfig, files []string, backend Backend) error {
	_, err := executeToolCaptured(repoRoot, tool, files, backend)
	return err
}

func executeToolWithWriter(repoRoot string, tool ToolConfig, files []string, backend Backend, w io.Writer) error {
	cmd := resolveCommandForBackend(repoRoot, tool, backend)
	if tool.RunPerFile {
		for _, file := range files {
			args := append([]string{}, tool.Args...)
			if tool.PassFilesEnabled() {
				args = append(args, file)
			}
			if err := backend.ExecWithWriter(repoRoot, append([]string{cmd}, args...), w); err != nil {
				return err
			}
		}
		return nil
	}

	args := append([]string{}, tool.Args...)
	if tool.PassFilesEnabled() {
		args = append(args, files...)
	}
	return backend.ExecWithWriter(repoRoot, append([]string{cmd}, args...), w)
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
