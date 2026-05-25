package forge

import (
	"strings"
	"testing"
)

// TestValidateConfig_ValidConfig ensures a well-formed config produces no issues.
func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"eslint": {Command: "eslint", Type: "node", Group: "lint"},
					"prettier": {Command: "prettier", Type: "node", Group: "format",
						DependsOn: []string{"eslint"}},
				},
			},
		},
	}
	if issues := ValidateConfig(cfg); len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

// TestValidateConfig_MissingCommand flags an error when command is empty.
func TestValidateConfig_MissingCommand(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"oops": {Command: ""},
				},
			},
		},
	}
	issues := ValidateConfig(cfg)
	if len(issues) != 1 || issues[0].Level != IssueError {
		t.Errorf("expected 1 error for missing command, got %v", issues)
	}
}

// TestValidateConfig_UnknownOnFailure flags a warning for invalid on_failure.
func TestValidateConfig_UnknownOnFailure(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"tool": {Command: "cmd", OnFailure: "ignore"},
				},
			},
		},
	}
	issues := ValidateConfig(cfg)
	if len(issues) != 1 || issues[0].Level != IssueWarn {
		t.Errorf("expected 1 warning for bad on_failure, got %v", issues)
	}
}

// TestValidateConfig_UnknownDependsOn flags an error for a missing depends_on ref.
func TestValidateConfig_UnknownDependsOn(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"tool": {Command: "cmd", DependsOn: []string{"ghost"}},
				},
			},
		},
	}
	issues := ValidateConfig(cfg)
	if len(issues) != 1 || issues[0].Level != IssueError {
		t.Errorf("expected 1 error for unknown depends_on, got %v", issues)
	}
}

// TestValidateConfig_DependsCycle flags a cycle in depends_on.
func TestValidateConfig_DependsCycle(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"pre-commit": {
				Tools: map[string]ToolConfig{
					"a": {Command: "a", DependsOn: []string{"b"}},
					"b": {Command: "b", DependsOn: []string{"a"}},
				},
			},
		},
	}
	issues := ValidateConfig(cfg)
	found := false
	for _, iss := range issues {
		if iss.Level == IssueError && strings.Contains(iss.Message, "cycle") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a cycle error, got %v", issues)
	}
}

// TestValidateConfig_BadBranchPattern flags an error for an invalid regex.
func TestValidateConfig_BadBranchPattern(t *testing.T) {
	cfg := &Config{
		Hooks: map[string]HookConfig{
			"commit-msg": {
				Policy: &CommitMessagePolicy{
					ValidateBranchName: true,
					BranchPattern:      "[invalid(regex",
				},
			},
		},
	}
	issues := ValidateConfig(cfg)
	if len(issues) != 1 || issues[0].Level != IssueError {
		t.Errorf("expected 1 error for bad branch_pattern, got %v", issues)
	}
}

// TestDetectDependsCycles_NoCycle returns empty for a DAG.
func TestDetectDependsCycles_NoCycle(t *testing.T) {
	tools := map[string]ToolConfig{
		"a": {Command: "a", DependsOn: []string{"b"}},
		"b": {Command: "b", DependsOn: []string{"c"}},
		"c": {Command: "c"},
	}
	if cycles := detectDependsCycles(tools); len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

// TestDetectDependsCycles_DirectCycle detects a direct A→B→A cycle.
func TestDetectDependsCycles_DirectCycle(t *testing.T) {
	tools := map[string]ToolConfig{
		"a": {Command: "a", DependsOn: []string{"b"}},
		"b": {Command: "b", DependsOn: []string{"a"}},
	}
	if cycles := detectDependsCycles(tools); len(cycles) == 0 {
		t.Error("expected a cycle, got none")
	}
}

// TestToolConfig_EnvField ensures Env is correctly parsed as map[string]string.
func TestToolConfig_EnvField(t *testing.T) {
	tool := ToolConfig{
		Command: "phpstan",
		Env:     map[string]string{"PHPSTAN_MEMORY_LIMIT": "512M", "CI": "1"},
	}
	if tool.Env["PHPSTAN_MEMORY_LIMIT"] != "512M" {
		t.Error("expected PHPSTAN_MEMORY_LIMIT=512M")
	}
	if tool.Env["CI"] != "1" {
		t.Error("expected CI=1")
	}
}
