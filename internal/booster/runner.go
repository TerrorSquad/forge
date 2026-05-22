package booster

import (
	"bufio"
	"bytes"
	"context"
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

// RunOptions controls optional behaviour for a hook run.
type RunOptions struct {
	AllFiles  bool   // run against all tracked files instead of only staged ones
	Source    string // git source arg for prepare-commit-msg (merge, squash, ...)
	NoCache   bool   // bypass run cache for this invocation
	CheckMode bool   // dry-run: use check_args, suppress restage, treat output as failure
}

func RunHook(hookName string, editFile string) error {
	return RunHookWithOptions(hookName, editFile, RunOptions{})
}

func RunHookWithOptions(hookName string, editFile string, opts RunOptions) error {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return err
	}

	if opts.AllFiles && hookName != "pre-commit" {
		return fmt.Errorf("--all-files is only valid for the pre-commit hook, not %q", hookName)
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

	if hookName == "prepare-commit-msg" {
		if err := applyPrepareCommitMsgPolicy(repoRoot, hookCfg.Policy, editFile, opts.Source); err != nil {
			return err
		}
	}

	if hookName == "pre-push" {
		pushCtx := parsePushContext(os.Stdin)
		return runHookCfgWithPushContext(repoRoot, hookName, hookCfg, cfg.Execution, pushCtx)
	}

	if hookName == "post-commit" {
		fmt.Fprintf(UI, "%s\n", dim("post-commit: informational — commit already saved"))
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, false, opts.NoCache, opts.CheckMode)
	}

	if hookName == "post-merge" {
		fmt.Fprintf(UI, "%s\n", dim("post-merge: informational — merge already complete"))
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, false, opts.NoCache, opts.CheckMode)
	}

	if hookName == "post-rewrite" {
		fmt.Fprintf(UI, "%s\n", dim("post-rewrite: informational — rewrite already complete (source: "+editFile+")"))
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, false, opts.NoCache, opts.CheckMode)
	}

	files := []string{}
	if hookName == "pre-commit" {
		if opts.AllFiles {
			files, err = allTrackedFiles(repoRoot)
			if err != nil {
				return err
			}
		} else {
			files, err = stagedFiles(repoRoot)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Fprintf(UI, "%s\n", dim("no staged files — nothing to run"))
				return nil
			}
		}
	}

	return runHookCfg(repoRoot, hookName, editFile, hookCfg, cfg.Execution, files, opts.AllFiles, opts.NoCache, opts.CheckMode)
}

// runHookCfg executes all tools in hookCfg for the given root / staged files.
// This is the inner loop used by both the root hook and workspace members.
func runHookCfg(root, hookName, editFile string, hookCfg HookConfig, exec ExecutionConfig, files []string, allFiles, noCache, checkMode bool) error {
	if isParallelMode(hookCfg, exec) {
		return runHookCfgParallel(root, hookName, hookCfg, exec, files, allFiles, noCache, checkMode)
	}
	toolNames := sortedToolNames(hookCfg.Tools)
	if isTUIMode(len(toolNames)) {
		return runHookCfgTUI(root, hookName, hookCfg, exec, files, allFiles, noCache, checkMode)
	}
	if len(toolNames) == 0 {
		fmt.Fprintf(UI, "%s\n", dim("no tools configured for "+hookName))
		return nil
	}

	PrintHookHeader(hookName)

	allowedGroups := parseAllowedGroups()
	var results []ToolResult
	hookStart := time.Now()
	failed := false

	// Load cache once for the hook run.
	tc := loadCache(root)
	cacheUpdated := false

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

		// Check run cache.
		cacheEnabled := !noCache && (tool.Cache || exec.Cache)
		var cacheKey string
		if cacheEnabled {
			if k, err := toolCacheKey(tool, filesToRun); err == nil {
				cacheKey = k
				if isCacheHit(tc, cacheKey) {
					r := ToolResult{Name: name, Status: "cached"}
					PrintToolResult(r)
					results = append(results, r)
					continue
				}
			}
		}

		// In check mode, substitute check_args and capture all output.
		effectiveTool := toolConfigForCheck(tool, checkMode)

		start := time.Now()
		var toolOut string
		var err error
		if checkMode && tool.CheckFailIfOutput {
			toolOut, err = executeToolCaptureAll(root, effectiveTool, filesToRun, backend, exec)
			if err == nil && strings.TrimSpace(toolOut) != "" {
				err = fmt.Errorf("check produced output (check_fail_if_output = true)")
			}
		} else {
			toolOut, err = executeToolCaptured(root, effectiveTool, filesToRun, backend, exec)
		}
		dur := time.Since(start)

		if err != nil {
			status := "fail"
			if checkMode {
				status = "would-fail"
			}
			r := ToolResult{Name: name, Status: status, Duration: dur, Output: toolOut}
			PrintToolResult(r)
			results = append(results, r)
			failed = true
			if !checkMode && strings.EqualFold(strings.TrimSpace(tool.OnFailure), "stop") {
				if checkMode {
					PrintCheckSummary(results, time.Since(hookStart))
				} else {
					PrintSummary(results, time.Since(hookStart))
				}
				return fmt.Errorf("tool %s failed and requested stop", name)
			}
			continue
		}

		r := ToolResult{Name: name, Status: "pass", Duration: dur}
		PrintToolResult(r)
		results = append(results, r)

		// Update cache on success (only in normal mode).
		if !checkMode && cacheEnabled && cacheKey != "" {
			updateCacheEntry(tc, cacheKey)
			cacheUpdated = true
		}

		if tool.Restage && tool.PassFilesEnabled() && !checkMode {
			if allFiles {
				fmt.Fprintf(UI, "%s\n", yellow("  restage suppressed for "+name+" (--all-files mode)"))
			} else {
				if err := addFiles(root, filesToRun); err != nil {
					return fmt.Errorf("tool %s restage failed: %w", name, err)
				}
			}
		}
	}

	if checkMode {
		PrintCheckSummary(results, time.Since(hookStart))
	} else {
		PrintSummary(results, time.Since(hookStart))
	}

	if cacheUpdated {
		saveCache(root, tc)
	}

	if failed {
		return errors.New("one or more tools failed")
	}

	return nil
}

// executeToolCaptured runs the tool and returns captured combined output on
// failure. On success the output is discarded (streamed to /dev/null is not
// quite right — we stream to a buffer and only surface it on error).
func executeToolCaptured(repoRoot string, tool ToolConfig, files []string, backend Backend, execCfg ExecutionConfig) (string, error) {
	var buf bytes.Buffer
	err := executeToolWithContext(repoRoot, tool, files, backend, execCfg, &buf)
	if err != nil {
		return buf.String(), err
	}
	return "", nil
}

// executeToolCaptureAll runs the tool and always returns stdout+stderr, even on success.
// Used in check mode when check_fail_if_output = true.
func executeToolCaptureAll(repoRoot string, tool ToolConfig, files []string, backend Backend, execCfg ExecutionConfig) (string, error) {
	var buf bytes.Buffer
	err := executeToolWithContext(repoRoot, tool, files, backend, execCfg, &buf)
	return buf.String(), err
}

// toolConfigForCheck returns a copy of tool with Args replaced by CheckArgs
// when in check mode and CheckArgs is non-empty.
func toolConfigForCheck(tool ToolConfig, checkMode bool) ToolConfig {
	if !checkMode || len(tool.CheckArgs) == 0 {
		return tool
	}
	t := tool
	t.Args = tool.CheckArgs
	return t
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

	if !policy.AppendTicketFooter && !policy.RequireTicket {
		return nil
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

// applyPrepareCommitMsgPolicy optionally prepends the ticket from the branch name.
func applyPrepareCommitMsgPolicy(repoRoot string, policy *CommitMessagePolicy, editFile, source string) error {
	if policy == nil || !policy.PrependTicket {
		return nil
	}

	if editFile == "" {
		return nil
	}

	if policy.SkipOnMerge && (source == "merge" || source == "squash") {
		return nil
	}

	branch, err := currentBranch(repoRoot)
	if err != nil {
		return err
	}

	ticket := ""
	if m := ticketRegex.FindStringSubmatch(branch); len(m) > 1 {
		ticket = m[1]
	}
	if ticket == "" {
		return nil
	}

	content, err := os.ReadFile(editFile)
	if err != nil {
		return err
	}

	if policy.SkipIfPresent && strings.Contains(string(content), ticket) {
		return nil
	}

	prefix := ticket + ": "
	rewritten := prefix + string(content)
	return os.WriteFile(editFile, []byte(rewritten), 0644)
}

func executeTool(repoRoot string, tool ToolConfig, files []string, backend Backend, execCfg ExecutionConfig) error {
	_, err := executeToolCaptured(repoRoot, tool, files, backend, execCfg)
	return err
}

// executeToolWithContext runs the tool with an optional timeout derived from execCfg.
func executeToolWithContext(repoRoot string, tool ToolConfig, files []string, backend Backend, execCfg ExecutionConfig, w io.Writer) error {
	ctx, cancel := buildToolContext(tool, execCfg)
	defer cancel()

	err := executeToolWithWriter(repoRoot, tool, files, backend, w, ctx)
	if err != nil && ctx.Err() != nil {
		d := resolveToolTimeout(tool, execCfg)
		return fmt.Errorf("timed out after %s", d.Round(time.Millisecond))
	}
	return err
}

// resolveToolTimeout returns the effective timeout for a tool; 0 means no limit.
func resolveToolTimeout(tool ToolConfig, execCfg ExecutionConfig) time.Duration {
	s := tool.Timeout
	if s == "" {
		s = execCfg.ToolTimeout
	}
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

// buildToolContext returns a context (and cancel) for the tool's timeout.
func buildToolContext(tool ToolConfig, execCfg ExecutionConfig) (context.Context, context.CancelFunc) {
	d := resolveToolTimeout(tool, execCfg)
	if d <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), d)
}

func executeToolWithWriter(repoRoot string, tool ToolConfig, files []string, backend Backend, w io.Writer, ctx context.Context) error {
	cmd := resolveCommandForBackend(repoRoot, tool, backend)
	if tool.RunPerFile {
		for _, file := range files {
			args := append([]string{}, tool.Args...)
			if tool.PassFilesEnabled() {
				args = append(args, file)
			}
			if err := backend.ExecWithContext(ctx, repoRoot, append([]string{cmd}, args...), w); err != nil {
				return err
			}
		}
		return nil
	}

	args := append([]string{}, tool.Args...)
	if tool.PassFilesEnabled() {
		args = append(args, files...)
	}
	return backend.ExecWithContext(ctx, repoRoot, append([]string{cmd}, args...), w)
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

// PushContext holds the ref info git passes to pre-push via stdin.
type PushContext struct {
	Remote string
	URL    string
	Refs   []PushRef
}

// PushRef is one pushed ref line from git's stdin.
type PushRef struct {
	LocalRef  string
	LocalSHA  string
	RemoteRef string
	RemoteSHA string
}

// parsePushContext reads git's pre-push stdin payload and the BOOSTER_PUSH_*
// env vars (which the shim should set from $1/$2).
func parsePushContext(r io.Reader) PushContext {
	ctx := PushContext{
		Remote: os.Getenv("BOOSTER_PUSH_REMOTE"),
		URL:    os.Getenv("BOOSTER_PUSH_URL"),
	}
	if r == nil {
		return ctx
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 4 {
			continue
		}
		ctx.Refs = append(ctx.Refs, PushRef{
			LocalRef:  parts[0],
			LocalSHA:  parts[1],
			RemoteRef: parts[2],
			RemoteSHA: parts[3],
		})
	}
	// Derive branch from first ref if not set
	if ctx.Remote == "" && len(ctx.Refs) > 0 {
		ref := ctx.Refs[0].LocalRef
		if strings.HasPrefix(ref, "refs/heads/") {
			ctx.Remote = strings.TrimPrefix(ref, "refs/heads/")
		}
	}
	return ctx
}

// runHookCfgWithPushContext runs pre-push tools, injecting push context into env.
func runHookCfgWithPushContext(root, hookName string, hookCfg HookConfig, execCfg ExecutionConfig, ctx PushContext) error {
	// Expose push context as env vars for tools
	if ctx.Remote != "" {
		os.Setenv("BOOSTER_PUSH_REMOTE", ctx.Remote)
	}
	if ctx.URL != "" {
		os.Setenv("BOOSTER_PUSH_URL", ctx.URL)
	}
	if len(ctx.Refs) > 0 {
		ref := ctx.Refs[0].LocalRef
		if strings.HasPrefix(ref, "refs/heads/") {
			os.Setenv("BOOSTER_PUSH_BRANCH", strings.TrimPrefix(ref, "refs/heads/"))
		}
	}
	// pre-push tools operate on no staged files (pass_files defaults to false)
	return runHookCfg(root, hookName, "", hookCfg, execCfg, nil, false, false, false)
}
