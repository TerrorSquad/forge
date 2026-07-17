package runner

import (
	"path/filepath"
	"testing"

	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/git"
)

// TestPrePushHonorsRunOptions is a regression test: the pre-push path used to
// drop RunOptions (--tool/--skip-tool/--check/... were silently ignored).
func TestPrePushHonorsRunOptions(t *testing.T) {
	dir := initBareGitRepo(t)

	tool := config.ToolConfig{
		Command:   "false", // always exits 1
		Type:      "system",
		PassFiles: boolPtr(false),
	}
	cfg := config.HookConfig{
		Enabled: boolPtr(true),
		Tools:   map[string]config.ToolConfig{"always-fails": tool},
	}

	// Sanity: without a skip the failing tool must fail the hook.
	if err := runHookCfgWithPushContext(dir, "pre-push", cfg, config.ExecutionConfig{}, PushContext{}, RunOptions{}); err == nil {
		t.Fatal("expected failure when the tool runs and fails")
	}

	// With --skip-tool the opts must be honored, so the hook passes.
	err := runHookCfgWithPushContext(dir, "pre-push", cfg, config.ExecutionConfig{}, PushContext{}, RunOptions{SkipTools: []string{"always-fails"}})
	if err != nil {
		t.Errorf("pre-push must honor RunOptions.SkipTools, got: %v", err)
	}
}

// TestStageOutputsOnlyOnSuccess verifies stage_outputs are added to the index
// when the tool succeeds, but not when it fails (mirrors restage semantics).
func TestStageOutputsOnlyOnSuccess(t *testing.T) {
	t.Run("failure does not stage", func(t *testing.T) {
		dir := initBareGitRepo(t)
		writeFile(t, filepath.Join(dir, "out.txt"), "generated\n")

		tool := config.ToolConfig{
			Command:      "false", // fails
			Type:         "system",
			PassFiles:    boolPtr(false),
			StageOutputs: []string{"out.txt"},
		}
		cfg := config.HookConfig{Enabled: boolPtr(true), Tools: map[string]config.ToolConfig{"gen": tool}}
		_ = runHookCfgWithPushContext(dir, "pre-push", cfg, config.ExecutionConfig{}, PushContext{}, RunOptions{})

		if staged := stagedSet(t, dir); staged["out.txt"] {
			t.Error("out.txt must NOT be staged when the tool fails")
		}
	})

	t.Run("success stages", func(t *testing.T) {
		dir := initBareGitRepo(t)
		writeFile(t, filepath.Join(dir, "out.txt"), "generated\n")

		tool := config.ToolConfig{
			Command:      "true", // succeeds
			Type:         "system",
			PassFiles:    boolPtr(false),
			StageOutputs: []string{"out.txt"},
		}
		cfg := config.HookConfig{Enabled: boolPtr(true), Tools: map[string]config.ToolConfig{"gen": tool}}
		if err := runHookCfgWithPushContext(dir, "pre-push", cfg, config.ExecutionConfig{}, PushContext{}, RunOptions{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if staged := stagedSet(t, dir); !staged["out.txt"] {
			t.Error("out.txt must be staged when the tool succeeds")
		}
	})
}

func stagedSet(t *testing.T, dir string) map[string]bool {
	t.Helper()
	files, err := git.StagedFiles(dir)
	if err != nil {
		t.Fatalf("StagedFiles: %v", err)
	}
	set := map[string]bool{}
	for _, f := range files {
		set[f] = true
	}
	return set
}

// TestParseAllowedGroups covers the HOOKS_ONLY parsing after the strings.Split
// simplification (trimming, case-folding, empty entries).
func TestParseAllowedGroups(t *testing.T) {
	t.Setenv("HOOKS_ONLY", " Format , ,LINT ")
	got := parseAllowedGroups()
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %v", got)
	}
	if _, ok := got["format"]; !ok {
		t.Error("expected lowercased 'format'")
	}
	if _, ok := got["lint"]; !ok {
		t.Error("expected lowercased 'lint'")
	}
}
