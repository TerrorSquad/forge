package booster

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// isParallelMode returns true if parallel execution is enabled for this hook.
// Hook-level setting overrides the global [execution] default.
func isParallelMode(hookCfg HookConfig, exec ExecutionConfig) bool {
	if hookCfg.Parallel != nil {
		return *hookCfg.Parallel
	}
	return exec.Parallel
}

// parallelToolResult holds the outcome of a single tool run in parallel mode.
type parallelToolResult struct {
	name       string
	result     ToolResult
	cacheKey   string
	filesToRun []string
	tool       ToolConfig
	err        error
}

// buildDependencyLevels groups tool names into levels where all dependencies of
// tools in level N are satisfied by tools in levels 0…N-1.
func buildDependencyLevels(toolNames []string, tools map[string]ToolConfig) [][]string {
	levels := make(map[string]int)
	for _, name := range toolNames {
		assignLevel(name, tools, levels, make(map[string]bool))
	}

	maxLevel := 0
	for _, l := range levels {
		if l > maxLevel {
			maxLevel = l
		}
	}

	result := make([][]string, maxLevel+1)
	// preserve sorted order within each level
	for _, name := range toolNames {
		l := levels[name]
		result[l] = append(result[l], name)
	}
	return result
}

func assignLevel(name string, tools map[string]ToolConfig, levels map[string]int, visiting map[string]bool) int {
	if l, ok := levels[name]; ok {
		return l
	}
	if visiting[name] {
		return 0 // cycle: treat as level 0
	}
	visiting[name] = true

	tool, ok := tools[name]
	if !ok {
		levels[name] = 0
		return 0
	}

	maxDep := -1
	for _, dep := range tool.DependsOn {
		if l := assignLevel(dep, tools, levels, visiting); l > maxDep {
			maxDep = l
		}
	}
	level := maxDep + 1
	levels[name] = level
	return level
}

// runHookCfgParallel runs hook tools in parallel, respecting depends_on groupings.
func runHookCfgParallel(root, hookName string, hookCfg HookConfig, exec ExecutionConfig, files []string, allFiles, noCache, checkMode bool) error {
	toolNames := sortedToolNames(hookCfg.Tools)
	if len(toolNames) == 0 {
		fmt.Fprintf(UI, "%s\n", dim("no tools configured for "+hookName))
		return nil
	}

	PrintHookHeader(hookName)

	allowedGroups := parseAllowedGroups()
	var allResults []ToolResult
	hookStart := time.Now()
	failed := false

	tc := loadCache(root)
	cacheUpdated := false

	levels := buildDependencyLevels(toolNames, hookCfg.Tools)

	for _, levelNames := range levels {
		if failed {
			break // stop_on_failure semantics: don't start new waves after a failure
		}

		waveResults := runToolWave(root, levelNames, hookCfg.Tools, files, exec, noCache, checkMode, allowedGroups, tc)

		for _, pr := range waveResults {
			PrintToolResult(pr.result)
			allResults = append(allResults, pr.result)

			if pr.err != nil {
				isContinue := strings.EqualFold(strings.TrimSpace(pr.tool.OnFailure), "continue")
				if !isContinue {
					failed = true
				}
				// Check if this specific tool requests stop — skip remaining waves
				if !isContinue && strings.EqualFold(strings.TrimSpace(pr.tool.OnFailure), "stop") {
					if checkMode {
						PrintCheckSummary(allResults, time.Since(hookStart))
					} else {
						PrintSummary(allResults, time.Since(hookStart))
					}
					return fmt.Errorf("tool %s failed and requested stop", pr.name)
				}
			} else if !checkMode && pr.cacheKey != "" {
				updateCacheEntry(tc, pr.cacheKey)
				cacheUpdated = true
			}
		}

		// Stage declared output artifacts and apply show_output after each wave (non-check mode only).
		if !checkMode {
			for _, pr := range waveResults {
				// Stage output files regardless of exit code (generated artifacts like diagrams).
				if len(pr.tool.StageOutputs) > 0 {
					_ = addFiles(root, pr.tool.StageOutputs) // best-effort; never block the hook
				}
			}
		}

		// Stage declared output artifacts after each wave (regardless of exit code).
		if !checkMode {
			for _, pr := range waveResults {
				if len(pr.tool.StageOutputs) > 0 {
					_ = addFiles(root, pr.tool.StageOutputs) // best-effort; never block the hook
				}
			}
		}

		// Restage after each wave completes (only in normal mode, not check mode)
		if !checkMode && !allFiles {
			for _, pr := range waveResults {
				if pr.err == nil && pr.tool.Restage && pr.tool.PassFilesEnabled() && len(pr.filesToRun) > 0 {
					if err := addFiles(root, pr.filesToRun); err != nil {
						return fmt.Errorf("tool %s restage failed: %w", pr.name, err)
					}
				}
			}
		}
	}

	if checkMode {
		PrintCheckSummary(allResults, time.Since(hookStart))
	} else {
		PrintSummary(allResults, time.Since(hookStart))
	}

	if cacheUpdated {
		saveCache(root, tc)
	}

	if failed {
		return errors.New("one or more tools failed")
	}
	return nil
}

// runToolWave executes a slice of tools concurrently and returns results in the
// same order as the input names. Output from each tool is buffered and only
// flushed when the tool finishes, preventing interleaving.
func runToolWave(root string, names []string, tools map[string]ToolConfig, files []string, exec ExecutionConfig, noCache, checkMode bool, allowedGroups map[string]struct{}, tc toolCache) []parallelToolResult {
	results := make([]parallelToolResult, len(names))
	var wg sync.WaitGroup

	for i, name := range names {
		wg.Add(1)
		go func(idx int, toolName string) {
			defer wg.Done()
			tool := tools[toolName]
			pr := parallelToolResult{name: toolName, tool: tool}

			if shouldSkipTool(toolName) {
				pr.result = ToolResult{Name: toolName, Status: "skip"}
				results[idx] = pr
				return
			}
			if len(allowedGroups) > 0 && tool.Group != "" {
				if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
					pr.result = ToolResult{Name: toolName, Status: "skip"}
					results[idx] = pr
					return
				}
			}
			if strings.TrimSpace(tool.Command) == "" {
				pr.result = ToolResult{Name: toolName, Status: "fail", Output: "command is required"}
				pr.err = fmt.Errorf("tool %s: command is required", toolName)
				results[idx] = pr
				return
			}

			filesToRun := filterFiles(files, tool)
			pr.filesToRun = filesToRun

			backend := ResolveBackend(root, tool, exec.DefaultBackend)

			// Skip tool if its binary is not available.
			resolvedCmd := resolveCommandForBackend(root, tool, backend)
			if !toolBinaryAvailable(root, resolvedCmd, backend) {
				pr.result = ToolResult{Name: toolName, Status: "skip"}
				results[idx] = pr
				return
			}

			cacheEnabled := !noCache && (tool.Cache || exec.Cache)
			if cacheEnabled && !checkMode {
				if k, err := toolCacheKey(tool, filesToRun); err == nil {
					pr.cacheKey = k
					if isCacheHit(tc, k) {
						pr.result = ToolResult{Name: toolName, Status: "cached"}
						results[idx] = pr
						return
					}
				}
			}

			effectiveTool := toolConfigForCheck(tool, checkMode)

			// Buffer output per tool to prevent interleaving.
			var buf bytes.Buffer
			start := time.Now()
			var runErr error
			if checkMode && tool.CheckFailIfOutput {
				runErr = executeToolWithContext(root, effectiveTool, filesToRun, backend, exec, &buf)
				if runErr == nil && strings.TrimSpace(buf.String()) != "" {
					runErr = fmt.Errorf("check produced output (check_fail_if_output = true)")
				}
			} else {
				runErr = executeToolWithContext(root, effectiveTool, filesToRun, backend, exec, &buf)
			}
			dur := time.Since(start)
			pr.err = runErr

			if runErr != nil {
				status := "fail"
				if checkMode {
					status = "would-fail"
				}
				pr.result = ToolResult{Name: toolName, Status: status, Duration: dur, Output: buf.String()}
			} else {
				passOutput := ""
				if tool.ShowOutput {
					passOutput = buf.String()
				}
				pr.result = ToolResult{Name: toolName, Status: "pass", Duration: dur, Output: passOutput}
			}
			results[idx] = pr
		}(i, name)
	}

	wg.Wait()
	return results
}
