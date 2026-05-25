package runner

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TerrorSquad/forge/internal/forge/backend"
	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/git"
	"github.com/TerrorSquad/forge/internal/forge/ui"
)

// IsParallelMode returns true if parallel execution is enabled for this hook.
// Hook-level setting overrides the global [execution] default.
func IsParallelMode(hookCfg config.HookConfig, exec config.ExecutionConfig) bool {
	if hookCfg.Parallel != nil {
		return *hookCfg.Parallel
	}
	return exec.Parallel
}

type parallelToolResult struct {
	name       string
	result     ui.ToolResult
	cacheKey   string
	filesToRun []string
	tool       config.ToolConfig
	err        error
}

func buildDependencyLevels(toolNames []string, tools map[string]config.ToolConfig) [][]string {
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
	for _, name := range toolNames {
		l := levels[name]
		result[l] = append(result[l], name)
	}
	return result
}

func assignLevel(name string, tools map[string]config.ToolConfig, levels map[string]int, visiting map[string]bool) int {
	if l, ok := levels[name]; ok {
		return l
	}
	if visiting[name] {
		return 0
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

func runHookCfgParallel(root, hookName string, hookCfg config.HookConfig, exec config.ExecutionConfig, files []string, opts RunOptions) error {
	allFiles := opts.AllFiles
	noCache := opts.NoCache
	checkMode := opts.CheckMode
	toolNames := applyToolFilter(config.SortedToolNames(hookCfg.Tools), hookCfg.Tools, opts)
	if len(toolNames) == 0 {
		fmt.Fprintf(ui.UI, "%s\n", ui.Dim("no tools configured for "+hookName))
		return nil
	}

	ui.PrintHookHeaderCI(hookName)

	allowedGroups := parseAllowedGroups()
	var allResults []ui.ToolResult
	hookStart := time.Now()
	failed := false

	tc := loadCache(root)
	cacheUpdated := false

	levels := buildDependencyLevels(toolNames, hookCfg.Tools)

	for _, levelNames := range levels {
		if failed {
			break
		}

		waveResults := runToolWave(root, levelNames, hookCfg.Tools, files, exec, noCache, checkMode, allowedGroups, tc)

		for _, pr := range waveResults {
			ui.PrintToolResult(pr.result)
			allResults = append(allResults, pr.result)

			if pr.err != nil {
				isContinue := strings.EqualFold(strings.TrimSpace(pr.tool.OnFailure), "continue")
				if !isContinue {
					failed = true
				}
				if !isContinue && strings.EqualFold(strings.TrimSpace(pr.tool.OnFailure), "stop") {
					if checkMode {
						ui.PrintCheckSummary(allResults, time.Since(hookStart))
					} else {
						ui.PrintSummary(allResults, time.Since(hookStart))
					}
					return fmt.Errorf("tool %s failed and requested stop", pr.name)
				}
			} else if !checkMode && pr.cacheKey != "" {
				updateCacheEntry(tc, pr.cacheKey)
				cacheUpdated = true
			}
		}

		if !checkMode {
			for _, pr := range waveResults {
				if len(pr.tool.StageOutputs) > 0 {
					_ = git.AddFiles(root, pr.tool.StageOutputs)
				}
			}
		}

		if !checkMode && !allFiles {
			for _, pr := range waveResults {
				if pr.err == nil && pr.tool.Restage && pr.tool.PassFilesEnabled() && len(pr.filesToRun) > 0 {
					if err := git.AddFiles(root, pr.filesToRun); err != nil {
						return fmt.Errorf("tool %s restage failed: %w", pr.name, err)
					}
				}
			}
		}
	}

	ui.PrintSummaryCI(allResults, time.Since(hookStart), checkMode)

	if cacheUpdated {
		saveCache(root, tc)
	}

	if failed {
		return errors.New("one or more tools failed")
	}
	return nil
}

func runToolWave(root string, names []string, tools map[string]config.ToolConfig, files []string, exec config.ExecutionConfig, noCache, checkMode bool, allowedGroups map[string]struct{}, tc toolCache) []parallelToolResult {
	results := make([]parallelToolResult, len(names))
	var wg sync.WaitGroup

	for i, name := range names {
		wg.Add(1)
		go func(idx int, toolName string) {
			defer wg.Done()
			tool := tools[toolName]
			pr := parallelToolResult{name: toolName, tool: tool}

			if shouldSkipTool(toolName) {
				pr.result = ui.ToolResult{Name: toolName, Status: "skip"}
				results[idx] = pr
				return
			}
			if shouldSkipGroup(tool.Group) {
				pr.result = ui.ToolResult{Name: toolName, Status: "skip"}
				results[idx] = pr
				return
			}
			if len(allowedGroups) > 0 && tool.Group != "" {
				if _, ok := allowedGroups[strings.ToLower(tool.Group)]; !ok {
					pr.result = ui.ToolResult{Name: toolName, Status: "skip"}
					results[idx] = pr
					return
				}
			}
			if strings.TrimSpace(tool.Command) == "" {
				pr.result = ui.ToolResult{Name: toolName, Status: "fail", Output: "command is required"}
				pr.err = fmt.Errorf("tool %s: command is required", toolName)
				results[idx] = pr
				return
			}

			filesToRun := filterFiles(files, tool)
			pr.filesToRun = filesToRun

			b := backend.ResolveBackend(root, tool, exec.DefaultBackend)

			resolvedCmd := backend.ResolveCommandForBackend(root, tool, b)
			if !backend.ToolBinaryAvailable(root, resolvedCmd, b) {
				pr.result = ui.ToolResult{Name: toolName, Status: "skip"}
				results[idx] = pr
				return
			}

			cacheEnabled := !noCache && (tool.Cache || exec.Cache)
			if cacheEnabled && !checkMode {
				if k, err := toolCacheKey(tool, filesToRun); err == nil {
					pr.cacheKey = k
					if isCacheHit(tc, k) {
						pr.result = ui.ToolResult{Name: toolName, Status: "cached"}
						results[idx] = pr
						return
					}
				}
			}

			effectiveTool := toolConfigForCheck(tool, checkMode)

			var buf bytes.Buffer
			start := time.Now()
			var runErr error
			if checkMode && tool.CheckFailIfOutput {
				runErr = executeToolWithContext(root, effectiveTool, filesToRun, b, exec, &buf)
				if runErr == nil && strings.TrimSpace(buf.String()) != "" {
					runErr = fmt.Errorf("check produced output (check_fail_if_output = true)")
				}
			} else {
				runErr = executeToolWithContext(root, effectiveTool, filesToRun, b, exec, &buf)
			}
			dur := time.Since(start)
			pr.err = runErr

			if runErr != nil {
				status := "fail"
				if checkMode {
					status = "would-fail"
				}
				pr.result = ui.ToolResult{Name: toolName, Status: status, Duration: dur, Output: buf.String()}
			} else {
				passOutput := ""
				if tool.ShowOutput {
					passOutput = buf.String()
				}
				pr.result = ui.ToolResult{Name: toolName, Status: "pass", Duration: dur, Output: passOutput}
			}
			results[idx] = pr
		}(i, name)
	}

	wg.Wait()
	return results
}
