package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- test helpers ----------

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	writeFile(t, filepath.Join(dir, "README.md"), "# test\n")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "-c", "commit.gpgsign=false", "commit", "-m", "chore: init")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stageFile(t *testing.T, dir, name, content string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, name), content)
	run(t, dir, "git", "add", name)
}

// ---------- RunGit ----------

func TestRunGit_Status(t *testing.T) {
	dir := initRepo(t)
	out, err := RunGit(dir, "status", "--short")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// clean repo — no output
	if out != "" {
		t.Errorf("expected empty status output, got %q", out)
	}
}

func TestRunGit_InvalidCommand(t *testing.T) {
	dir := initRepo(t)
	_, err := RunGit(dir, "this-command-does-not-exist")
	if err == nil {
		t.Fatal("expected error for invalid git command")
	}
}

func TestRunGit_NotARepo(t *testing.T) {
	_, err := RunGit(t.TempDir(), "status")
	if err == nil {
		t.Fatal("expected error outside a git repo")
	}
}

// ---------- StagedFiles ----------

func TestStagedFiles_Empty(t *testing.T) {
	dir := initRepo(t)
	files, err := StagedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no staged files, got %v", files)
	}
}

func TestStagedFiles_WithStagedFile(t *testing.T) {
	dir := initRepo(t)
	stageFile(t, dir, "src/foo.php", "<?php\n")

	files, err := StagedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "src/foo.php" {
		t.Errorf("expected [src/foo.php], got %v", files)
	}
}

func TestStagedFiles_MultipleStagedFiles(t *testing.T) {
	dir := initRepo(t)
	stageFile(t, dir, "a.go", "package a\n")
	stageFile(t, dir, "b.go", "package b\n")

	files, err := StagedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 staged files, got %v", files)
	}
}

func TestStagedFiles_UnstagedNotIncluded(t *testing.T) {
	dir := initRepo(t)
	// write but do NOT stage
	writeFile(t, filepath.Join(dir, "unstaged.go"), "package x\n")

	files, err := StagedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f, "unstaged") {
			t.Errorf("unstaged file should not appear in StagedFiles: %v", files)
		}
	}
}

// ---------- AllTrackedFiles ----------

func TestAllTrackedFiles_ContainsInitialFile(t *testing.T) {
	dir := initRepo(t)
	files, err := AllTrackedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, f := range files {
		if f == "README.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected README.md in tracked files, got %v", files)
	}
}

func TestAllTrackedFiles_UnstagedFileNotIncluded(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, "new.go"), "package x\n")

	files, err := AllTrackedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range files {
		if strings.Contains(f, "new.go") {
			t.Errorf("untracked file should not be in AllTrackedFiles: %v", files)
		}
	}
}

// ---------- CurrentBranch ----------

func TestCurrentBranch_Default(t *testing.T) {
	dir := initRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// git init defaults to "master" or "main" depending on config
	if branch == "" {
		t.Error("expected a non-empty branch name")
	}
}

func TestCurrentBranch_AfterCheckout(t *testing.T) {
	dir := initRepo(t)
	run(t, dir, "git", "checkout", "-b", "feature/my-branch")

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "feature/my-branch" {
		t.Errorf("got %q, want %q", branch, "feature/my-branch")
	}
}

// ---------- AddFiles ----------

func TestAddFiles_StagesFiles(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, "new.go"), "package x\n")

	if err := AddFiles(dir, []string{"new.go"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files, err := StagedFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "new.go" {
		t.Errorf("expected [new.go] staged, got %v", files)
	}
}

func TestAddFiles_EmptySliceIsNoop(t *testing.T) {
	dir := initRepo(t)
	if err := AddFiles(dir, []string{}); err != nil {
		t.Errorf("empty AddFiles should be a no-op, got error: %v", err)
	}
}

// ---------- LocalHooksPath ----------

func TestLocalHooksPath_UnsetReturnsEmpty(t *testing.T) {
	dir := initRepo(t)
	path, err := LocalHooksPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path when core.hooksPath unset, got %q", path)
	}
}

func TestLocalHooksPath_ReturnsConfiguredPath(t *testing.T) {
	dir := initRepo(t)
	run(t, dir, "git", "config", "core.hooksPath", ".forge/hooks")

	path, err := LocalHooksPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != ".forge/hooks" {
		t.Errorf("got %q, want %q", path, ".forge/hooks")
	}
}

func TestIsSequencerOperation_CleanRepo(t *testing.T) {
	dir := initRepo(t)
	active, err := IsSequencerOperation(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Error("expected no sequencer operation in a clean repo")
	}
}

func TestIsSequencerOperation_DetectsMerge(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, ".git", "MERGE_HEAD"), "abc123")

	active, err := IsSequencerOperation(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected merge operation to be detected")
	}
}

func TestIsSequencerOperation_DetectsCherryPick(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, ".git", "CHERRY_PICK_HEAD"), "abc123")

	active, err := IsSequencerOperation(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected cherry-pick operation to be detected")
	}
}

func TestIsSequencerOperation_DetectsRebase(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, ".git", "REBASE_HEAD"), "abc123")

	active, err := IsSequencerOperation(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected rebase operation to be detected")
	}
}

func TestIsSequencerOperation_DetectsRevert(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, filepath.Join(dir, ".git", "REVERT_HEAD"), "abc123")

	active, err := IsSequencerOperation(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected revert operation to be detected")
	}
}
