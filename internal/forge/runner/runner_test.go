package runner

import (
	"bytes"
	"errors"
	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/ui"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveToolTimeout_PerTool(t *testing.T) {
	tool := config.ToolConfig{Timeout: "30s"}
	d := resolveToolTimeout(tool, config.ExecutionConfig{})
	if d != 30*time.Second {
		t.Errorf("got %v, want 30s", d)
	}
}

func TestResolveToolTimeout_GlobalFallback(t *testing.T) {
	tool := config.ToolConfig{}
	execCfg := config.ExecutionConfig{ToolTimeout: "60s"}
	d := resolveToolTimeout(tool, execCfg)
	if d != 60*time.Second {
		t.Errorf("got %v, want 60s", d)
	}
}

func TestResolveToolTimeout_PerToolOverridesGlobal(t *testing.T) {
	tool := config.ToolConfig{Timeout: "10s"}
	execCfg := config.ExecutionConfig{ToolTimeout: "300s"}
	d := resolveToolTimeout(tool, execCfg)
	if d != 10*time.Second {
		t.Errorf("got %v, want 10s", d)
	}
}

func TestResolveToolTimeout_Empty(t *testing.T) {
	d := resolveToolTimeout(config.ToolConfig{}, config.ExecutionConfig{})
	if d != 0 {
		t.Errorf("empty timeout should be 0, got %v", d)
	}
}

func TestResolveToolTimeout_InvalidString(t *testing.T) {
	tool := config.ToolConfig{Timeout: "notaduration"}
	d := resolveToolTimeout(tool, config.ExecutionConfig{})
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

	writeFile(t, filepath.Join(dir, "forge.toml"), `
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

func TestRunHookWithOptions_SkipsDuringGitSequencerOperation(t *testing.T) {
	dir := initBareGitRepo(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	writeFile(t, filepath.Join(dir, ".git", "MERGE_HEAD"), "abc123")

	err := RunHookWithOptions("pre-commit", "", RunOptions{})
	if err == nil {
		t.Fatal("expected hook skip error")
	}
	if !errors.Is(err, config.ErrHookSkipped) {
		t.Fatalf("expected ErrHookSkipped, got %v", err)
	}
}

func TestRunHookWithOptions_SkipsDuringGitSequencerOperation_Rebase(t *testing.T) {
	dir := initBareGitRepo(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	writeFile(t, filepath.Join(dir, ".git", "REBASE_HEAD"), "abc123")

	err := RunHookWithOptions("commit-msg", "", RunOptions{})
	if err == nil {
		t.Fatal("expected hook skip error")
	}
	if !errors.Is(err, config.ErrHookSkipped) {
		t.Fatalf("expected ErrHookSkipped, got %v", err)
	}
}

func TestRunHookWithOptions_SkipsDuringGitSequencerOperation_Revert(t *testing.T) {
	dir := initBareGitRepo(t)
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	writeFile(t, filepath.Join(dir, ".git", "REVERT_HEAD"), "abc123")

	err := RunHookWithOptions("pre-push", "", RunOptions{})
	if err == nil {
		t.Fatal("expected hook skip error")
	}
	if !errors.Is(err, config.ErrHookSkipped) {
		t.Fatalf("expected ErrHookSkipped, got %v", err)
	}
}

func TestRunHookCfg_PreCommitSkipsPassFilesFalseWhenNoMatchingStagedFiles(t *testing.T) {
	dir := initBareGitRepo(t)

	cfg := config.HookConfig{
		Enabled: boolPtr(true),
		Tools: map[string]config.ToolConfig{
			"vue-tsc": {
				Command:    "echo",
				Type:       "system",
				Extensions: []string{".ts", ".tsx", ".vue"},
				PassFiles:  boolPtr(false),
			},
		},
	}

	var buf bytes.Buffer
	ui.UI = &buf
	t.Cleanup(func() { ui.UI = os.Stdout })

	err := runHookCfg(dir, "pre-commit", "", cfg, config.ExecutionConfig{}, []string{"package.json"}, RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "skip") {
		t.Fatalf("expected tool to be skipped when no staged files match extensions, got output: %q", buf.String())
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

	tool := config.ToolConfig{
		Command:   "false", // /usr/bin/false — always exits 1
		Type:      "system",
		PassFiles: boolPtr(false),
		OnFailure: "continue",
	}
	cfg := config.HookConfig{
		Enabled: boolPtr(true),
		Tools:   map[string]config.ToolConfig{"always-fails": tool},
	}

	err := runHookCfg(dir, "pre-push", "", cfg, config.ExecutionConfig{}, nil, RunOptions{})
	if err != nil {
		t.Errorf("on_failure=continue must not block the hook, got: %v", err)
	}
}

// TestRunHookCfg_DefaultFailureFails verifies that a failing tool without
// on_failure=continue causes the hook to return an error.
func TestRunHookCfg_DefaultFailureFails(t *testing.T) {
	dir := initBareGitRepo(t)

	tool := config.ToolConfig{
		Command:   "false",
		Type:      "system",
		PassFiles: boolPtr(false),
		// no OnFailure → default behaviour: fail the hook
	}
	cfg := config.HookConfig{
		Enabled: boolPtr(true),
		Tools:   map[string]config.ToolConfig{"always-fails": tool},
	}

	err := runHookCfg(dir, "pre-push", "", cfg, config.ExecutionConfig{}, nil, RunOptions{})
	if err == nil {
		t.Error("expected error when tool fails without on_failure=continue")
	}
}

// TestApplyToolFilter_OnlyTools checks that --tool filters to just the named tools.
func TestApplyToolFilter_OnlyTools(t *testing.T) {
	tools := map[string]config.ToolConfig{
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
	tools := map[string]config.ToolConfig{
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
	tools := map[string]config.ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
		"psalm":   {Group: "analysis"},
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

func TestApplyToolFilter_SkipGroups(t *testing.T) {
	tools := map[string]config.ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
		"psalm":   {Group: "analysis"},
	}
	names := []string{"ecs", "phpstan", "psalm"}
	opts := RunOptions{SkipGroups: []string{"analysis"}}
	got := applyToolFilter(names, tools, opts)
	for _, n := range got {
		if n == "phpstan" || n == "psalm" {
			t.Errorf("tools in analysis group should have been skipped, got %s", n)
		}
	}
	if len(got) != 1 || got[0] != "ecs" {
		t.Errorf("expected [ecs], got %v", got)
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

// Fix 4: --skip-tools and --only-tools comparisons are case-insensitive.
func TestApplyToolFilter_SkipToolsCaseInsensitive(t *testing.T) {
	tools := map[string]config.ToolConfig{
		"PHPLint": {Group: "lint"},
		"phpstan": {Group: "analysis"},
	}
	names := []string{"PHPLint", "phpstan"}

	got := applyToolFilter(names, tools, RunOptions{SkipTools: []string{"phplint"}})
	for _, n := range got {
		if n == "PHPLint" {
			t.Error("PHPLint should be skipped by lowercase 'phplint'")
		}
	}
}

func TestApplyToolFilter_OnlyToolsCaseInsensitive(t *testing.T) {
	tools := map[string]config.ToolConfig{
		"PHPLint": {Group: "lint"},
		"phpstan": {Group: "analysis"},
	}
	names := []string{"PHPLint", "phpstan"}

	got := applyToolFilter(names, tools, RunOptions{OnlyTools: []string{"PHPLINT"}})
	if len(got) != 1 || got[0] != "PHPLint" {
		t.Errorf("expected [PHPLint] from uppercase 'PHPLINT', got %v", got)
	}
}

// Fix 5: an explicit --only-tools entry runs even when --only-groups is also set
// and the tool belongs to a different group.
func TestApplyToolFilter_OnlyToolsOverridesGroupFilter(t *testing.T) {
	tools := map[string]config.ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
		"psalm":   {Group: "analysis"},
	}
	names := []string{"ecs", "phpstan", "psalm"}

	// ecs is in group "format" but is explicitly named; it should still run.
	opts := RunOptions{OnlyTools: []string{"ecs"}, OnlyGroups: []string{"analysis"}}
	got := applyToolFilter(names, tools, opts)

	found := false
	for _, n := range got {
		if n == "ecs" {
			found = true
		}
	}
	if !found {
		t.Errorf("explicitly named tool 'ecs' should run even though its group 'format' is not in --only-groups; got %v", got)
	}
}

// Fix 5: when only --only-groups is set (no --only-tools), group filter works normally.
func TestApplyToolFilter_OnlyGroupsWithoutOnlyTools(t *testing.T) {
	tools := map[string]config.ToolConfig{
		"ecs":     {Group: "format"},
		"phpstan": {Group: "analysis"},
	}
	names := []string{"ecs", "phpstan"}
	got := applyToolFilter(names, tools, RunOptions{OnlyGroups: []string{"analysis"}})
	if len(got) != 1 || got[0] != "phpstan" {
		t.Errorf("expected [phpstan], got %v", got)
	}
}
