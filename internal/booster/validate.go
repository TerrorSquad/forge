package booster

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// IssueLevel indicates severity of a validation finding.
type IssueLevel string

const (
	IssueError IssueLevel = "error"
	IssueWarn  IssueLevel = "warn"
)

// ValidationIssue is a single finding from config validation.
type ValidationIssue struct {
	Level   IssueLevel
	Hook    string
	Tool    string // empty for hook-level issues
	Message string
}

func (v ValidationIssue) String() string {
	loc := v.Hook
	if v.Tool != "" {
		loc = fmt.Sprintf("%s.%s", v.Hook, v.Tool)
	}
	return fmt.Sprintf("[%s] %s: %s", v.Level, loc, v.Message)
}

// ValidateConfig validates cfg and returns any issues found.
// Issues with level "error" indicate mis-configurations that will cause
// runtime failures. Issues with level "warn" are best-practice violations.
// The function never fails fatally — callers decide whether to abort.
func ValidateConfig(cfg *Config) []ValidationIssue {
	var issues []ValidationIssue

	// Sort hook names for deterministic output.
	hookNames := make([]string, 0, len(cfg.Hooks))
	for name := range cfg.Hooks {
		hookNames = append(hookNames, name)
	}
	sort.Strings(hookNames)

	for _, hookName := range hookNames {
		hookCfg := cfg.Hooks[hookName]

		for _, toolName := range sortedToolNames(hookCfg.Tools) {
			tool := hookCfg.Tools[toolName]

			// Command is required.
			if strings.TrimSpace(tool.Command) == "" {
				issues = append(issues, ValidationIssue{
					Level: IssueError, Hook: hookName, Tool: toolName,
					Message: "command is required",
				})
			}

			// on_failure must be a known value.
			if tool.OnFailure != "" {
				switch strings.ToLower(strings.TrimSpace(tool.OnFailure)) {
				case "stop", "continue":
					// valid
				default:
					issues = append(issues, ValidationIssue{
						Level: IssueWarn, Hook: hookName, Tool: toolName,
						Message: fmt.Sprintf("on_failure %q is not recognized; use \"stop\" or \"continue\"", tool.OnFailure),
					})
				}
			}

			// type must be a known value when set.
			if tool.Type != "" {
				switch strings.ToLower(tool.Type) {
				case "php", "node":
					// valid
				default:
					issues = append(issues, ValidationIssue{
						Level: IssueWarn, Hook: hookName, Tool: toolName,
						Message: fmt.Sprintf("type %q is not recognized; use \"php\" or \"node\"", tool.Type),
					})
				}
			}

			// backend must be a known value when set.
			if tool.Backend != "" {
				switch strings.ToLower(tool.Backend) {
				case "host", "ddev":
					// valid
				default:
					issues = append(issues, ValidationIssue{
						Level: IssueWarn, Hook: hookName, Tool: toolName,
						Message: fmt.Sprintf("backend %q is not recognized; use \"host\" or \"ddev\"", tool.Backend),
					})
				}
			}

			// depends_on must reference tools that exist in the same hook.
			for _, dep := range tool.DependsOn {
				if _, ok := hookCfg.Tools[dep]; !ok {
					issues = append(issues, ValidationIssue{
						Level: IssueError, Hook: hookName, Tool: toolName,
						Message: fmt.Sprintf("depends_on references unknown tool %q (not defined in hooks.%s)", dep, hookName),
					})
				}
			}
		}

		// Detect depends_on cycles.
		if cycles := detectDependsCycles(hookCfg.Tools); len(cycles) > 0 {
			for _, cycle := range cycles {
				issues = append(issues, ValidationIssue{
					Level:   IssueError,
					Hook:    hookName,
					Message: "depends_on cycle detected: " + strings.Join(cycle, " → "),
				})
			}
		}

		// Branch pattern must compile when validate_branch_name is true.
		if p := hookCfg.Policy; p != nil && p.ValidateBranchName && p.BranchPattern != "" {
			if _, err := regexp.Compile(p.BranchPattern); err != nil {
				issues = append(issues, ValidationIssue{
					Level:   IssueError,
					Hook:    hookName,
					Message: fmt.Sprintf("branch_pattern %q does not compile: %v", p.BranchPattern, err),
				})
			}
		}
	}

	return issues
}

// detectDependsCycles returns all cycles found in the depends_on graph.
// Each cycle is represented as the ordered list of tool names forming the loop.
func detectDependsCycles(tools map[string]ToolConfig) [][]string {
	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := map[string]int{}
	var cycles [][]string

	var dfs func(name string, path []string)
	dfs = func(name string, path []string) {
		switch state[name] {
		case done:
			return
		case inStack:
			for i, p := range path {
				if p == name {
					cycle := make([]string, len(path[i:]), len(path[i:])+1)
					copy(cycle, path[i:])
					cycles = append(cycles, append(cycle, name))
					return
				}
			}
			return
		}
		state[name] = inStack
		if tool, ok := tools[name]; ok {
			for _, dep := range tool.DependsOn {
				dfs(dep, append(path, name))
			}
		}
		state[name] = done
	}

	for _, name := range sortedToolNames(tools) {
		if state[name] == unvisited {
			dfs(name, nil)
		}
	}
	return cycles
}

// PrintValidationIssues writes issues to UI in a human-readable format.
// Returns true if any errors (not just warnings) were found.
func PrintValidationIssues(issues []ValidationIssue) bool {
	hasError := false
	for _, issue := range issues {
		switch issue.Level {
		case IssueError:
			hasError = true
			fmt.Fprintf(UI, "%s %s\n", red("✗"), issue.String())
		case IssueWarn:
			fmt.Fprintf(UI, "%s %s\n", yellow("⚠"), issue.String())
		}
	}
	return hasError
}
