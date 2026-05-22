package booster

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// runHookCfgTUI executes hook tools with a live in-place terminal progress view.
// It handles both sequential and parallel modes, routing to the appropriate
// execution strategy while feeding updates to the tuiDisplay.
func runHookCfgTUI(root, hookName string, hookCfg HookConfig, exec ExecutionConfig, files []string, allFiles, noCache, checkMode bool) error {
	toolNames := sortedToolNames(hookCfg.Tools)

	disp := newTUIDisplay(hookName, toolNames)
	stopSpinner := make(chan struct{})
	var stopOnce sync.Once
	closeSpinner := func() { stopOnce.Do(func() { close(stopSpinner) }) }
	go disp.spinnerLoop(stopSpinner)
	defer closeSpinner()

	allowedGroups := parseAllowedGroups()
	hookStart := time.Now()
	failed := false

	tc := loadCache(root)
	cacheUpdated := false
	var allResults []ToolResult

	if isParallelMode(hookCfg, exec) {
		levels := buildDependencyLevels(toolNames, hookCfg.Tools)
		for _, levelNames := range levels {
			if failed {
				break
			}
			waveResults := runTUIWave(root, levelNames, hookCfg.Tools, files, exec, noCache, checkMode, allowedGroups, tc, disp)
			for _, pr := range waveResults {
				allResults = append(allResults, pr.result)
				if pr.err != nil {
					failed = true
					if strings.EqualFold(strings.TrimSpace(pr.tool.OnFailure), "stop") {
						closeSpinner()
						disp.finish(allResults, time.Since(hookStart), checkMode)
						return fmt.Errorf("tool %s failed and requested stop", pr.name)
					}
				} else if !checkMode && pr.cacheKey != "" {
					updateCacheEntry(tc, pr.cacheKey)
					cacheUpdated = true
				}
			}
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
	} else {
		for _, name := range toolNames {
			tool := hookCfg.Tools[name]

			if shouldSkipTool(name) {
				r := ToolResult{Name: name, Status: "skip"}
				disp.doneTool(r)
				allResults = append(allResults, r)
				continue
			}
			if len(allowedGroups) > 0 && tool.Group != "" {
				if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
					r := ToolResult{Name: name, Status: "skip"}
					disp.doneTool(r)
					allResults = append(allResults, r)
					continue
				}
			}
			if strings.TrimSpace(tool.Command) == "" {
				return fmt.Errorf("tool %s: command is required", name)
			}

			filesToRun := filterFiles(files, tool)
			if hookName == "pre-commit" && tool.PassFilesEnabled() && len(filesToRun) == 0 {
				r := ToolResult{Name: name, Status: "skip"}
				disp.doneTool(r)
				allResults = append(allResults, r)
				continue
			}

			backend := ResolveBackend(root, tool, exec.DefaultBackend)

			// Skip tool if its binary is not available.
			resolvedCmd := resolveCommandForBackend(root, tool, backend)
			if !toolBinaryAvailable(root, resolvedCmd, backend) {
				r := ToolResult{Name: name, Status: "skip"}
				disp.doneTool(r)
				allResults = append(allResults, r)
				continue
			}

			cacheEnabled := !noCache && !checkMode && (tool.Cache || exec.Cache)
			var cacheKey string
			if cacheEnabled {
				if k, err := toolCacheKey(tool, filesToRun); err == nil {
					cacheKey = k
					if isCacheHit(tc, cacheKey) {
						r := ToolResult{Name: name, Status: "cached"}
						disp.doneTool(r)
						allResults = append(allResults, r)
						continue
					}
				}
			}

			disp.startTool(name)
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
				disp.doneTool(r)
				allResults = append(allResults, r)
				failed = true
				if strings.EqualFold(strings.TrimSpace(tool.OnFailure), "stop") {
					closeSpinner()
					disp.finish(allResults, time.Since(hookStart), checkMode)
					return fmt.Errorf("tool %s failed and requested stop", name)
				}
				continue
			}

			r := ToolResult{Name: name, Status: "pass", Duration: dur}
			disp.doneTool(r)
			allResults = append(allResults, r)

			if cacheEnabled && cacheKey != "" {
				updateCacheEntry(tc, cacheKey)
				cacheUpdated = true
			}

			if tool.Restage && tool.PassFilesEnabled() && !checkMode {
				if allFiles {
					// restage suppressed in --all-files mode
				} else {
					if err := addFiles(root, filesToRun); err != nil {
						return fmt.Errorf("tool %s restage failed: %w", name, err)
					}
				}
			}
		}
	}

	closeSpinner()
	disp.finish(allResults, time.Since(hookStart), checkMode)

	if cacheUpdated {
		saveCache(root, tc)
	}

	if failed {
		return errors.New("one or more tools failed")
	}
	return nil
}

// runTUIWave runs a wave of tools concurrently, sending start/done events to disp.
func runTUIWave(root string, names []string, tools map[string]ToolConfig, files []string, exec ExecutionConfig, noCache, checkMode bool, allowedGroups map[string]struct{}, tc toolCache, disp *tuiDisplay) []parallelToolResult {
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
				disp.doneTool(pr.result)
				results[idx] = pr
				return
			}
			if len(allowedGroups) > 0 && tool.Group != "" {
				if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
					pr.result = ToolResult{Name: toolName, Status: "skip"}
					disp.doneTool(pr.result)
					results[idx] = pr
					return
				}
			}

			filesToRun := filterFiles(files, tool)
			pr.filesToRun = filesToRun
			backend := ResolveBackend(root, tool, exec.DefaultBackend)

			// Skip tool if its binary is not available.
			resolvedCmd := resolveCommandForBackend(root, tool, backend)
			if !toolBinaryAvailable(root, resolvedCmd, backend) {
				pr.result = ToolResult{Name: toolName, Status: "skip"}
				disp.doneTool(pr.result)
				results[idx] = pr
				return
			}

			cacheEnabled := !noCache && !checkMode && (tool.Cache || exec.Cache)
			if cacheEnabled {
				if k, err := toolCacheKey(tool, filesToRun); err == nil {
					pr.cacheKey = k
					if isCacheHit(tc, k) {
						pr.result = ToolResult{Name: toolName, Status: "cached"}
						disp.doneTool(pr.result)
						results[idx] = pr
						return
					}
				}
			}

			disp.startTool(toolName)
			effectiveTool := toolConfigForCheck(tool, checkMode)

			var buf strings.Builder
			start := time.Now()
			var runErr error
			if checkMode && tool.CheckFailIfOutput {
				toolOut, err := executeToolCaptureAll(root, effectiveTool, filesToRun, backend, exec)
				buf.WriteString(toolOut)
				runErr = err
				if runErr == nil && strings.TrimSpace(toolOut) != "" {
					runErr = fmt.Errorf("check produced output (check_fail_if_output = true)")
				}
			} else {
				toolOut, err := executeToolCaptured(root, effectiveTool, filesToRun, backend, exec)
				buf.WriteString(toolOut)
				runErr = err
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
				pr.result = ToolResult{Name: toolName, Status: "pass", Duration: dur}
			}
			disp.doneTool(pr.result)
			results[idx] = pr
		}(i, name)
	}

	wg.Wait()
	return results
}
