package runner

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

	"github.com/TerrorSquad/forge/internal/forge/backend"
	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/git"
	"github.com/TerrorSquad/forge/internal/forge/ui"
)

var ticketRegex = regexp.MustCompile(`([A-Z]+-[0-9]+)`)
var conventionalRegex = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([^)]+\))?!?: .+`)

// RunOptions controls optional behaviour for a hook run.
type RunOptions struct {
	AllFiles   bool
	Source     string // git source arg for prepare-commit-msg (merge, squash, ...)
	NoCache    bool
	CheckMode  bool
	OnlyTools  []string
	OnlyGroups []string
	SkipTools  []string
}

func RunHook(hookName string, editFile string) error {
	return RunHookWithOptions(hookName, editFile, RunOptions{})
}

func RunHookWithOptions(hookName string, editFile string, opts RunOptions) error {
	repoRoot, err := git.DetectRepoRoot()
	if err != nil {
		return err
	}

	loadEnvFiles(repoRoot)

	if opts.AllFiles && hookName != "pre-commit" {
		return fmt.Errorf("--all-files is only valid for the pre-commit hook, not %q", hookName)
	}

	if isHookSkippedEnv(hookName) {
		fmt.Fprintf(ui.UI, "%s\n", ui.Yellow("~ "+hookName+" skipped (env skip set)"))
		return config.ErrHookSkipped
	}

	cfg, configPath, err := config.LoadConfig(repoRoot)
	if err != nil {
		return err
	}

	hookCfg, ok := cfg.Hooks[hookName]
	if !ok || !hookCfg.IsEnabled() {
		return config.ErrHookSkipped
	}

	fmt.Fprintf(ui.UI, "%s\n", ui.Dim("config: "+configPath))

	// Workspace mode: run hook for each affected member
	if len(cfg.Workspace.Members) > 0 && hookName != "commit-msg" {
		staged, err := git.StagedFiles(repoRoot)
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
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, opts)
	}

	if hookName == "post-merge" {
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, opts)
	}

	if hookName == "post-rewrite" {
		return runHookCfg(repoRoot, hookName, "", hookCfg, cfg.Execution, nil, opts)
	}

	files := []string{}
	if hookName == "pre-commit" {
		if opts.AllFiles {
			files, err = git.AllTrackedFiles(repoRoot)
			if err != nil {
				return err
			}
		} else {
			files, err = git.StagedFiles(repoRoot)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Fprintf(ui.UI, "%s\n", ui.Dim("no staged files — nothing to run"))
				return nil
			}
		}

		if safeStashEnabled(hookCfg) && !opts.AllFiles && !opts.CheckMode {
			_, stashed, stashErr := git.StashUnstagedChanges(repoRoot)
			if stashErr != nil {
				fmt.Fprintf(ui.UI, "%s\n", ui.Yellow("⚠ stash failed: "+stashErr.Error()+" — proceeding without stash"))
			} else if stashed {
				fmt.Fprintf(ui.UI, "  %s\n", ui.Dim("⬇  stashing unstaged changes..."))
				defer func() {
					fmt.Fprintf(ui.UI, "  %s\n", ui.Dim("⬆  restoring unstaged changes..."))
					if popErr := git.PopStash(repoRoot); popErr != nil {
						fmt.Fprintf(ui.UI, "%s\n", ui.Yellow("⚠ "+popErr.Error()))
					}
				}()
			}
		}
	}

	return runHookCfg(repoRoot, hookName, editFile, hookCfg, cfg.Execution, files, opts)
}

func runHookCfg(root, hookName, editFile string, hookCfg config.HookConfig, exec config.ExecutionConfig, files []string, opts RunOptions) error {
	allFiles := opts.AllFiles
	noCache := opts.NoCache
	checkMode := opts.CheckMode
	if IsParallelMode(hookCfg, exec) {
		return runHookCfgParallel(root, hookName, hookCfg, exec, files, opts)
	}
	toolNames := applyToolFilter(config.SortedToolNames(hookCfg.Tools), hookCfg.Tools, opts)
	if len(toolNames) == 0 {
		fmt.Fprintf(ui.UI, "%s\n", ui.Dim("no tools configured for "+hookName))
		return nil
	}

	ui.PrintHookHeaderCI(hookName)

	allowedGroups := parseAllowedGroups()
	var results []ui.ToolResult
	hookStart := time.Now()
	failed := false

	tc := loadCache(root)
	cacheUpdated := false

	for _, name := range toolNames {
		tool := hookCfg.Tools[name]

		if shouldSkipTool(name) {
			r := ui.ToolResult{Name: name, Status: "skip"}
			ui.PrintToolResult(r)
			results = append(results, r)
			continue
		}
		if shouldSkipGroup(tool.Group) {
			r := ui.ToolResult{Name: name, Status: "skip"}
			ui.PrintToolResult(r)
			results = append(results, r)
			continue
		}
		if len(allowedGroups) > 0 && tool.Group != "" {
			if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
				r := ui.ToolResult{Name: name, Status: "skip"}
				ui.PrintToolResult(r)
				results = append(results, r)
				continue
			}
		}
		if strings.TrimSpace(tool.Command) == "" {
			return fmt.Errorf("tool %s: command is required", name)
		}

		filesToRun := filterFiles(files, tool)
		if hookName == "pre-commit" && tool.PassFilesEnabled() && len(filesToRun) == 0 {
			r := ui.ToolResult{Name: name, Status: "skip"}
			ui.PrintToolResult(r)
			results = append(results, r)
			continue
		}

		b := backend.ResolveBackend(root, tool, exec.DefaultBackend)
		resolvedCmd := backend.ResolveCommandForBackend(root, tool, b)
		if !backend.ToolBinaryAvailable(root, resolvedCmd, b) {
			r := ui.ToolResult{Name: name, Status: "skip", Output: "binary not found: " + resolvedCmd}
			ui.PrintToolResult(r)
			results = append(results, r)
			continue
		}

		cacheEnabled := !noCache && (tool.Cache || exec.Cache)
		var cacheKey string
		if cacheEnabled {
			if k, err := toolCacheKey(tool, filesToRun); err == nil {
				cacheKey = k
				if isCacheHit(tc, cacheKey) {
					r := ui.ToolResult{Name: name, Status: "cached"}
					ui.PrintToolResult(r)
					results = append(results, r)
					continue
				}
			}
		}

		effectiveTool := toolConfigForCheck(tool, checkMode)

		ui.PrintRunning(name)
		start := time.Now()
		var toolOut string
		var err error
		if checkMode && tool.CheckFailIfOutput {
			toolOut, err = executeToolCaptureAll(root, effectiveTool, filesToRun, b, exec)
			if err == nil && strings.TrimSpace(toolOut) != "" {
				err = fmt.Errorf("check produced output (check_fail_if_output = true)")
			}
		} else {
			toolOut, err = executeToolCaptured(root, effectiveTool, filesToRun, b, exec)
		}
		dur := time.Since(start)
		ui.ClearRunning()

		if !checkMode && len(tool.StageOutputs) > 0 {
			_ = git.AddFiles(root, tool.StageOutputs)
		}

		if err != nil {
			status := "fail"
			if checkMode {
				status = "would-fail"
			}
			r := ui.ToolResult{Name: name, Status: status, Duration: dur, Output: toolOut}
			ui.PrintToolResult(r)
			results = append(results, r)
			isContinue := strings.EqualFold(strings.TrimSpace(tool.OnFailure), "continue")
			if !isContinue {
				failed = true
			}
			if !checkMode && !isContinue {
				if strings.EqualFold(strings.TrimSpace(tool.OnFailure), "stop") {
					ui.PrintSummary(results, time.Since(hookStart))
					return fmt.Errorf("tool %s failed and requested stop", name)
				}
			}
			continue
		}

		passOutput := ""
		if tool.ShowOutput {
			passOutput = toolOut
		}
		r := ui.ToolResult{Name: name, Status: "pass", Duration: dur, Output: passOutput}
		ui.PrintToolResult(r)
		results = append(results, r)

		if !checkMode && cacheEnabled && cacheKey != "" {
			updateCacheEntry(tc, cacheKey)
			cacheUpdated = true
		}

		if tool.Restage && tool.PassFilesEnabled() && !checkMode {
			if allFiles {
				fmt.Fprintf(ui.UI, "%s\n", ui.Yellow("  restage suppressed for "+name+" (--all-files mode)"))
			} else {
				if err := git.AddFiles(root, filesToRun); err != nil {
					return fmt.Errorf("tool %s restage failed: %w", name, err)
				}
			}
		}
	}

	ui.PrintSummaryCI(results, time.Since(hookStart), checkMode)

	if cacheUpdated {
		evictCache(tc, exec)
		saveCache(root, tc)
	}

	if failed {
		return errors.New("one or more tools failed")
	}

	return nil
}

func executeToolCaptured(repoRoot string, tool config.ToolConfig, files []string, b backend.Backend, execCfg config.ExecutionConfig) (string, error) {
	var buf bytes.Buffer
	err := executeToolWithContext(repoRoot, tool, files, b, execCfg, &buf)
	if err != nil {
		return buf.String(), err
	}
	return "", nil
}

func executeToolCaptureAll(repoRoot string, tool config.ToolConfig, files []string, b backend.Backend, execCfg config.ExecutionConfig) (string, error) {
	var buf bytes.Buffer
	err := executeToolWithContext(repoRoot, tool, files, b, execCfg, &buf)
	return buf.String(), err
}

func toolConfigForCheck(tool config.ToolConfig, checkMode bool) config.ToolConfig {
	if !checkMode || len(tool.CheckArgs) == 0 {
		return tool
	}
	t := tool
	t.Args = tool.CheckArgs
	return t
}

func applyCommitMessagePolicy(repoRoot string, policy *config.CommitMessagePolicy, editFile string) error {
	if policy == nil {
		return nil
	}

	if editFile == "" {
		return fmt.Errorf("commit-msg policy requires message file (pass --edit or let git pass the file path)")
	}

	branch, _ := git.CurrentBranch(repoRoot)

	for _, skip := range policy.SkippedBranches {
		if branch == skip {
			return nil
		}
	}

	if policy.ValidateBranchName && policy.BranchPattern != "" {
		if branch == "" {
			return fmt.Errorf("cannot determine current branch for branch_pattern validation")
		}
		branchRe, err := regexp.Compile(policy.BranchPattern)
		if err != nil {
			return fmt.Errorf("invalid branch_pattern %q: %w", policy.BranchPattern, err)
		}
		if !branchRe.MatchString(branch) {
			return fmt.Errorf("branch name %q does not match required pattern %q", branch, policy.BranchPattern)
		}
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

	ticket := ""
	if m := ticketRegex.FindStringSubmatch(branch); len(m) > 1 {
		ticket = m[1]
	}

	if policy.RequireTicket && ticket == "" {
		return fmt.Errorf("ticket required by policy but branch %q has no ticket (expected e.g. PRJ-123)", branch)
	}

	if policy.AppendTicketFooter && ticket != "" {
		label := policy.FooterLabel
		if label == "" {
			label = "Closes"
		}
		footer := fmt.Sprintf("%s: %s", label, ticket)
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

func applyPrepareCommitMsgPolicy(repoRoot string, policy *config.CommitMessagePolicy, editFile, source string) error {
	if policy == nil || !policy.PrependTicket {
		return nil
	}

	if editFile == "" {
		return nil
	}

	if policy.SkipOnMerge && (source == "merge" || source == "squash") {
		return nil
	}

	branch, err := git.CurrentBranch(repoRoot)
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

func executeTool(repoRoot string, tool config.ToolConfig, files []string, b backend.Backend, execCfg config.ExecutionConfig) error {
	_, err := executeToolCaptured(repoRoot, tool, files, b, execCfg)
	return err
}

func executeToolWithContext(repoRoot string, tool config.ToolConfig, files []string, b backend.Backend, execCfg config.ExecutionConfig, w io.Writer) error {
	ctx, cancel := buildToolContext(tool, execCfg)
	defer cancel()

	err := executeToolWithWriter(repoRoot, tool, files, b, w, ctx)
	if err != nil && ctx.Err() != nil {
		d := resolveToolTimeout(tool, execCfg)
		return fmt.Errorf("timed out after %s", d.Round(time.Millisecond))
	}
	return err
}

func resolveToolTimeout(tool config.ToolConfig, execCfg config.ExecutionConfig) time.Duration {
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

func buildToolContext(tool config.ToolConfig, execCfg config.ExecutionConfig) (context.Context, context.CancelFunc) {
	d := resolveToolTimeout(tool, execCfg)
	if d <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), d)
}

func executeToolWithWriter(repoRoot string, tool config.ToolConfig, files []string, b backend.Backend, w io.Writer, ctx context.Context) error {
	cmd := backend.ResolveCommandForBackend(repoRoot, tool, b)
	if tool.RunPerFile {
		for _, file := range files {
			args := append([]string{}, tool.Args...)
			if tool.PassFilesEnabled() {
				args = append(args, file)
			}
			if err := b.ExecWithContext(ctx, repoRoot, append([]string{cmd}, args...), tool.Env, w); err != nil {
				return err
			}
		}
		return nil
	}

	args := append([]string{}, tool.Args...)
	if tool.PassFilesEnabled() {
		args = append(args, files...)
	}
	return b.ExecWithContext(ctx, repoRoot, append([]string{cmd}, args...), tool.Env, w)
}

func filterFiles(files []string, tool config.ToolConfig) []string {
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

func loadEnvFiles(repoRoot string) {
	for _, name := range []string{".git-hooks.env", ".env"} {
		path := filepath.Join(repoRoot, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func safeStashEnabled(hookCfg config.HookConfig) bool {
	if hookCfg.SafeStash != nil {
		return *hookCfg.SafeStash
	}
	for _, tool := range hookCfg.Tools {
		if tool.Restage {
			return true
		}
	}
	return false
}

func shouldSkipTool(name string) bool {
	return ShouldSkipTool(name)
}

// ShouldSkipTool reports whether the SKIP_<NAME> environment variable is set,
// which the user can set to skip a specific tool during a hook run.
func ShouldSkipTool(name string) bool {
	key := "SKIP_" + sanitizeEnvKey(name)
	return isTruthy(os.Getenv(key))
}

// ResolveToolTimeout returns the effective timeout for a tool, falling back to
// the execution-level default. Returns 0 if no timeout is configured.
func ResolveToolTimeout(tool config.ToolConfig, execCfg config.ExecutionConfig) time.Duration {
	return resolveToolTimeout(tool, execCfg)
}

func shouldSkipGroup(group string) bool {
	if group == "" {
		return false
	}
	key := "SKIP_GROUP_" + sanitizeEnvKey(group)
	return isTruthy(os.Getenv(key))
}

func applyToolFilter(toolNames []string, tools map[string]config.ToolConfig, opts RunOptions) []string {
	if len(opts.OnlyTools) == 0 && len(opts.OnlyGroups) == 0 && len(opts.SkipTools) == 0 {
		return toolNames
	}
	onlyToolSet := setOf(opts.OnlyTools)
	onlyGroupSet := setOf(opts.OnlyGroups)
	skipToolSet := setOf(opts.SkipTools)

	filtered := make([]string, 0, len(toolNames))
	for _, name := range toolNames {
		if skipToolSet[name] {
			continue
		}
		tool := tools[name]
		if len(onlyToolSet) > 0 && !onlyToolSet[name] {
			if !isDependencyOf(name, opts.OnlyTools, tools) {
				continue
			}
		}
		if len(onlyGroupSet) > 0 && !onlyGroupSet[strings.ToLower(tool.Group)] {
			continue
		}
		filtered = append(filtered, name)
	}
	return filtered
}

func setOf(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, v := range items {
		m[strings.ToLower(v)] = true
	}
	return m
}

func isDependencyOf(name string, requestedTools []string, tools map[string]config.ToolConfig) bool {
	nameLower := strings.ToLower(name)
	for _, req := range requestedTools {
		t, ok := tools[req]
		if !ok {
			continue
		}
		for _, dep := range t.DependsOn {
			if strings.ToLower(dep) == nameLower {
				return true
			}
		}
	}
	return false
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

func parsePushContext(r io.Reader) PushContext {
	ctx := PushContext{
		Remote: os.Getenv("FORGE_PUSH_REMOTE"),
		URL:    os.Getenv("FORGE_PUSH_URL"),
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
	if ctx.Remote == "" && len(ctx.Refs) > 0 {
		ref := ctx.Refs[0].LocalRef
		if strings.HasPrefix(ref, "refs/heads/") {
			ctx.Remote = strings.TrimPrefix(ref, "refs/heads/")
		}
	}
	return ctx
}

func runHookCfgWithPushContext(root, hookName string, hookCfg config.HookConfig, execCfg config.ExecutionConfig, ctx PushContext) error {
	if ctx.Remote != "" {
		os.Setenv("FORGE_PUSH_REMOTE", ctx.Remote)
	}
	if ctx.URL != "" {
		os.Setenv("FORGE_PUSH_URL", ctx.URL)
	}
	if len(ctx.Refs) > 0 {
		ref := ctx.Refs[0].LocalRef
		if strings.HasPrefix(ref, "refs/heads/") {
			os.Setenv("FORGE_PUSH_BRANCH", strings.TrimPrefix(ref, "refs/heads/"))
		}
	}
	return runHookCfg(root, hookName, "", hookCfg, execCfg, nil, RunOptions{})
}
