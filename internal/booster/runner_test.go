package booster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveToolTimeout_PerTool(t *testing.T) {
	tool := ToolConfig{Timeout: "30s"}
	d := resolveToolTimeout(tool, ExecutionConfig{})
	if d != 30*time.Second {
		t.Errorf("got %v, want 30s", d)
	}
}

func TestResolveToolTimeout_GlobalFallback(t *testing.T) {
	tool := ToolConfig{}
	execCfg := ExecutionConfig{ToolTimeout: "60s"}
	d := resolveToolTimeout(tool, execCfg)
	if d != 60*time.Second {
		t.Errorf("got %v, want 60s", d)
	}
}

func TestResolveToolTimeout_PerToolOverridesGlobal(t *testing.T) {
	tool := ToolConfig{Timeout: "10s"}
	execCfg := ExecutionConfig{ToolTimeout: "300s"}
	d := resolveToolTimeout(tool, execCfg)
	if d != 10*time.Second {
		t.Errorf("got %v, want 10s", d)
	}
}

func TestResolveToolTimeout_Empty(t *testing.T) {
	d := resolveToolTimeout(ToolConfig{}, ExecutionConfig{})
	if d != 0 {
		t.Errorf("empty timeout should be 0, got %v", d)
	}
}

func TestResolveToolTimeout_InvalidString(t *testing.T) {
	tool := ToolConfig{Timeout: "notaduration"}
	d := resolveToolTimeout(tool, ExecutionConfig{})
	if d != 0 {
		t.Errorf("invalid duration string should yield 0, got %v", d)
	}
}

func TestRunOptions_AllFilesOnlyValidForPreCommit(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	writeFile(t, filepath.Join(dir, "booster.toml"), `
[hooks.commit-msg]
enabled = true
`)

	err := RunHookWithOptions("commit-msg", "", RunOptions{AllFiles: true})
	if err == nil {
		t.Error("expected error when --all-files used with non-pre-commit hook")
	}
	if !strings.Contains(err.Error(), "pre-commit") {
		t.Errorf("expected error to mention pre-commit, got: %v", err)
	}
}

func TestParsePushContext_SingleRef(t *testing.T) {
	input := "refs/heads/main abc123 refs/heads/main def456\n"
	ctx := parsePushContext(strings.NewReader(input))

	if len(ctx.Refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(ctx.Refs))
	}
	ref := ctx.Refs[0]
	if ref.LocalRef != "refs/heads/main" {
		t.Errorf("LocalRef = %q, want refs/heads/main", ref.LocalRef)
	}
	if ref.LocalSHA != "abc123" {
		t.Errorf("LocalSHA = %q, want abc123", ref.LocalSHA)
	}
	if ref.RemoteRef != "refs/heads/main" {
		t.Errorf("RemoteRef = %q, want refs/heads/main", ref.RemoteRef)
	}
	if ref.RemoteSHA != "def456" {
		t.Errorf("RemoteSHA = %q, want def456", ref.RemoteSHA)
	}
}

func TestParsePushContext_MultipleRefs(t *testing.T) {
	input := strings.Join([]string{
		"refs/heads/main abc111 refs/heads/main 0000000000000000000000000000000000000000",
		"refs/heads/feat  abc222 refs/heads/feat  def222",
	}, "\n") + "\n"

	ctx := parsePushContext(strings.NewReader(input))
	if len(ctx.Refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %+v", len(ctx.Refs), ctx.Refs)
	}
}

func TestParsePushContext_Empty(t *testing.T) {
	ctx := parsePushContext(strings.NewReader(""))
	if len(ctx.Refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(ctx.Refs))
	}
}

func TestParsePushContext_SkipsMalformedLines(t *testing.T) {
	input := "only-two-fields something\nrefs/heads/main abc 0refs/heads/main def\n"
	ctx := parsePushContext(strings.NewReader(input))
	// malformed line should be silently skipped; valid line parsed
	if len(ctx.Refs) > 1 {
		t.Errorf("expected at most 1 valid ref, got %d", len(ctx.Refs))
	}
}

func boolPtr(b bool) *bool { return &b }

// TestRunHookCfg_OnFailureContinue verifies that a tool with on_failure=continue
// does not cause the hook to return a non-zero exit code.
// Regression test for: push blocked despite all failing tools having on_failure=continue.
func TestRunHookCfg_OnFailureContinue(t *testing.T) {
	dir := initBareGitRepo(t)

	tool := ToolConfig{
		Command:   "false", // /usr/bin/false — always exits 1
		Type:      "system",
		PassFiles: boolPtr(false),
		OnFailure: "continue",
	}
	cfg := HookConfig{
		Enabled: boolPtr(true),
		Tools:   map[string]ToolConfig{"always-fails": tool},
	}

	err := runHookCfg(dir, "pre-push", "", cfg, ExecutionConfig{}, nil, RunOptions{})
	if err != nil {
		t.Errorf("on_failure=continue must not block the hook, got: %v", err)
	}
}

// TestRunHookCfg_DefaultFailureFails verifies that a failing tool without
// on_failure=continue causes the hook to return an error.
func TestRunHookCfg_DefaultFailureFails(t *testing.T) {
	dir := initBareGitRepo(t)

	tool := ToolConfig{
		Command:   "false",
		Type:      "system",
		PassFiles: boolPtr(false),
		// no OnFailure → default behaviour: fail the hook
	}
	cfg := HookConfig{
		Enabled: boolPtr(true),
		Tools:   map[string]ToolConfig{"always-fails": tool},
	}

	err := runHookCfg(dir, "pre-push", "", cfg, ExecutionConfig{}, nil, RunOptions{})
	if err == nil {
		t.Error("expected error when tool fails without on_failure=continue")
	}
}

// TestApplyToolFilter_OnlyTools checks that --tool filters to just the named tools.
func TestApplyToolFilter_OnlyTools(t *testing.T) {
	tools := map[string]ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
		"psalm":   {Group: "analysis"},
	}
	names := []string{"ecs", "phpstan", "psalm"}
	opts := RunOptions{OnlyTools: []string{"phpstan"}}
	got := applyToolFilter(names, tools, opts)
	if len(got) != 1 || got[0] != "phpstan" {
		t.Errorf("expected [phpstan], got %v", got)
	}
}

// TestApplyToolFilter_OnlyGroups checks that --group filters by group name.
func TestApplyToolFilter_OnlyGroups(t *testing.T) {
	tools := map[string]ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
		"psalm":   {Group: "analysis"},
	}
	names := []string{"ecs", "phpstan", "psalm"}
	opts := RunOptions{OnlyGroups: []string{"analysis"}}
	got := applyToolFilter(names, tools, opts)
	if len(got) != 2 {
		t.Errorf("expected 2 tools, got %v", got)
	}
}

// TestApplyToolFilter_SkipTools checks that --skip-tool excludes named tools.
func TestApplyToolFilter_SkipTools(t *testing.T) {
	tools := map[string]ToolConfig{
		"ecs":     {},
		"phpstan": {},
		"psalm":   {},
	}
	names := []string{"ecs", "phpstan", "psalm"}
	opts := RunOptions{SkipTools: []string{"psalm"}}
	got := applyToolFilter(names, tools, opts)
	for _, n := range got {
		if n == "psalm" {
			t.Error("psalm should have been skipped")
		}
	}
	if len(got) != 2 {
		t.Errorf("expected 2 tools, got %v", got)
	}
}

// TestShouldSkipGroup checks SKIP_GROUP_* env var behaviour.
func TestShouldSkipGroup(t *testing.T) {
	t.Setenv("SKIP_GROUP_ANALYSIS", "1")
	if !shouldSkipGroup("analysis") {
		t.Error("expected analysis group to be skipped")
	}
	if shouldSkipGroup("format") {
		t.Error("format group should not be skipped")
	}
}

// TestSafeStashEnabled_AutoDetect checks that safe_stash is auto-enabled when
// any tool has restage=true.
func TestSafeStashEnabled_AutoDetect(t *testing.T) {
	cfg := HookConfig{
		Tools: map[string]ToolConfig{
			"ecs":     {Restage: true},
			"phpstan": {Restage: false},
		},
	}
	if !safeStashEnabled(cfg) {
		t.Error("expected safe stash to be auto-enabled when a tool has restage=true")
	}
}

// TestSafeStashEnabled_ExplicitFalse checks that safe_stash=false opts out.
func TestSafeStashEnabled_ExplicitFalse(t *testing.T) {
	f := false
	cfg := HookConfig{
		SafeStash: &f,
		Tools: map[string]ToolConfig{
			"ecs": {Restage: true},
		},
	}
	if safeStashEnabled(cfg) {
		t.Error("expected safe stash to be disabled when safe_stash=false")
	}
}

// TestSafeStashEnabled_NoRestageTools checks safe_stash is off when no fixers.
func TestSafeStashEnabled_NoRestageTools(t *testing.T) {
	cfg := HookConfig{
		Tools: map[string]ToolConfig{
			"phpstan": {Restage: false},
			"psalm":   {Restage: false},
		},
	}
	if safeStashEnabled(cfg) {
		t.Error("expected safe stash to be off when no tools have restage=true")
	}
}
