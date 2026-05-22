package booster

import (
	"testing"
)

func TestIsParallelMode_GlobalDefault(t *testing.T) {
	hookCfg := HookConfig{}
	exec := ExecutionConfig{Parallel: true}
	if !isParallelMode(hookCfg, exec) {
		t.Error("expected parallel=true from global default")
	}
}

func TestIsParallelMode_GlobalDisabled(t *testing.T) {
	hookCfg := HookConfig{}
	exec := ExecutionConfig{Parallel: false}
	if isParallelMode(hookCfg, exec) {
		t.Error("expected parallel=false when global disabled")
	}
}

func TestIsParallelMode_HookOverrideTrue(t *testing.T) {
	bTrue := true
	hookCfg := HookConfig{Parallel: &bTrue}
	exec := ExecutionConfig{Parallel: false}
	if !isParallelMode(hookCfg, exec) {
		t.Error("expected hook-level true to override global false")
	}
}

func TestIsParallelMode_HookOverrideFalse(t *testing.T) {
	bFalse := false
	hookCfg := HookConfig{Parallel: &bFalse}
	exec := ExecutionConfig{Parallel: true}
	if isParallelMode(hookCfg, exec) {
		t.Error("expected hook-level false to override global true")
	}
}

func TestBuildDependencyLevels_NoDepends(t *testing.T) {
	tools := map[string]ToolConfig{
		"gofmt": {Command: "gofmt"},
		"govet": {Command: "go vet"},
	}
	names := []string{"gofmt", "govet"}
	levels := buildDependencyLevels(names, tools)
	if len(levels) != 1 {
		t.Errorf("expected 1 level for independent tools, got %d", len(levels))
	}
	if len(levels[0]) != 2 {
		t.Errorf("expected 2 tools in level 0, got %d", len(levels[0]))
	}
}

func TestBuildDependencyLevels_WithDepends(t *testing.T) {
	tools := map[string]ToolConfig{
		"gofmt":         {Command: "gofmt"},
		"golangci-lint": {Command: "golangci-lint", DependsOn: []string{"gofmt"}},
	}
	names := []string{"gofmt", "golangci-lint"}
	levels := buildDependencyLevels(names, tools)
	if len(levels) != 2 {
		t.Errorf("expected 2 levels, got %d", len(levels))
	}
	if levels[0][0] != "gofmt" {
		t.Errorf("expected gofmt in level 0, got %v", levels[0])
	}
	if levels[1][0] != "golangci-lint" {
		t.Errorf("expected golangci-lint in level 1, got %v", levels[1])
	}
}

func TestBuildDependencyLevels_MultiLevel(t *testing.T) {
	tools := map[string]ToolConfig{
		"a": {Command: "a"},
		"b": {Command: "b", DependsOn: []string{"a"}},
		"c": {Command: "c", DependsOn: []string{"b"}},
	}
	names := []string{"a", "b", "c"}
	levels := buildDependencyLevels(names, tools)
	if len(levels) != 3 {
		t.Errorf("expected 3 levels for chain a->b->c, got %d", len(levels))
	}
}

func TestBuildDependencyLevels_MissingDep(t *testing.T) {
	// If a dep is not in the tools map, it should not panic
	tools := map[string]ToolConfig{
		"a": {Command: "a", DependsOn: []string{"nonexistent"}},
	}
	names := []string{"a"}
	levels := buildDependencyLevels(names, tools)
	if len(levels) == 0 {
		t.Error("expected at least 1 level")
	}
}

func TestBuildDependencyLevels_Cycle(t *testing.T) {
	// Cycle detection should not panic or infinite-loop
	tools := map[string]ToolConfig{
		"a": {Command: "a", DependsOn: []string{"b"}},
		"b": {Command: "b", DependsOn: []string{"a"}},
	}
	names := []string{"a", "b"}
	levels := buildDependencyLevels(names, tools)
	if len(levels) == 0 {
		t.Error("expected at least 1 level even with cycle")
	}
}
