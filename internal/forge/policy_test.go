package forge

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyCommitMessagePolicy_ConventionalPass(t *testing.T) {
	dir := t.TempDir()
	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat(runner): add parallel execution\n")

	policy := &CommitMessagePolicy{ConventionalCommits: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApplyCommitMessagePolicy_ConventionalFail(t *testing.T) {
	dir := t.TempDir()
	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "added some stuff\n")

	policy := &CommitMessagePolicy{ConventionalCommits: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err == nil {
		t.Error("expected error for non-conventional commit")
	}
}

func TestApplyCommitMessagePolicy_NilPolicy(t *testing.T) {
	dir := t.TempDir()
	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "anything goes\n")

	// nil policy = no-op
	if err := applyCommitMessagePolicy(dir, nil, msgFile); err != nil {
		t.Errorf("nil policy should be no-op, got: %v", err)
	}
}

func TestApplyCommitMessagePolicy_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "")

	policy := &CommitMessagePolicy{ConventionalCommits: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err == nil {
		t.Error("expected error for empty commit message")
	}
}

func TestApplyCommitMessagePolicy_AppendTicketFooter(t *testing.T) {
	// Create a fake git repo so currentBranch works
	dir := initBareGitRepo(t)

	createBranchInRepo(t, dir, "feat/PRJ-999-my-feature")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: do something\n")

	policy := &CommitMessagePolicy{AppendTicketFooter: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(msgFile)
	content := string(data)
	if !contains(content, "Closes: PRJ-999") {
		t.Errorf("expected 'Closes: PRJ-999' in commit message, got:\n%s", content)
	}
}

func TestApplyCommitMessagePolicy_NoDoubleAppend(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/PRJ-42-thing")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: something\n\nCloses: PRJ-42\n")

	policy := &CommitMessagePolicy{AppendTicketFooter: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(msgFile)
	count := countOccurrences(string(data), "Closes: PRJ-42")
	if count != 1 {
		t.Errorf("expected footer appended once, found %d times", count)
	}
}

func TestApplyCommitMessagePolicy_RequireTicket_NoTicket(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "main")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: no ticket branch\n")

	policy := &CommitMessagePolicy{RequireTicket: true}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err == nil {
		t.Error("expected error when require_ticket=true and branch has no ticket")
	}
}

func TestApplyCommitMessagePolicy_AllConventionalTypes(t *testing.T) {
	types := []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}
	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			dir := t.TempDir()
			msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
			writeFile(t, msgFile, typ+": valid message\n")
			policy := &CommitMessagePolicy{ConventionalCommits: true}
			if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
				t.Errorf("type %q should be valid, got: %v", typ, err)
			}
		})
	}
}

func TestApplyCommitMessagePolicy_ScopeAndBreaking(t *testing.T) {
	cases := []string{
		"feat(scope): with scope",
		"fix!: breaking fix",
		"feat(scope)!: scoped breaking change",
	}
	for _, msg := range cases {
		t.Run(msg, func(t *testing.T) {
			dir := t.TempDir()
			msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
			writeFile(t, msgFile, msg+"\n")
			policy := &CommitMessagePolicy{ConventionalCommits: true}
			if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
				t.Errorf("message %q should be valid, got: %v", msg, err)
			}
		})
	}
}

func TestApplyCommitMessagePolicy_SkippedBranch(t *testing.T) {
	skipped := []string{"main", "master", "development", "develop", "develop/test"}
	for _, branch := range skipped {
		t.Run(branch, func(t *testing.T) {
			dir := initBareGitRepo(t)
			// initBareGitRepo already starts on the default branch (master/main).
			// Only create the branch if it doesn't already exist.
			createBranchInRepoIfNotExists(t, dir, branch)

			msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
			// Non-conventional message — would normally fail conventional_commits check
			writeFile(t, msgFile, "WIP: direct commit to protected branch\n")

			policy := &CommitMessagePolicy{
				ConventionalCommits: true,
				AppendTicketFooter:  true,
				SkippedBranches:     skipped,
			}
			if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
				t.Errorf("branch %q is in skipped list, all validation should be bypassed, got: %v", branch, err)
			}

			// Footer must NOT have been appended
			data, _ := os.ReadFile(msgFile)
			if contains(string(data), "Closes:") {
				t.Errorf("footer must not be appended on skipped branch %q", branch)
			}
		})
	}
}

func TestApplyCommitMessagePolicy_BranchNameValidation_Pass(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "story/PRJ-1234-my-feature")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: valid message\n")

	policy := &CommitMessagePolicy{
		ConventionalCommits: true,
		ValidateBranchName:  true,
		BranchPattern:       `^(feature|fix|chore|story|task|bug|sub-task)/((?:PRJ|ERM)-[0-9]+-[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*|[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*)$`,
	}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
		t.Errorf("valid branch should pass validation, got: %v", err)
	}
}

func TestApplyCommitMessagePolicy_BranchNameValidation_Fail(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "INVALID_BRANCH_NAME")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: valid message\n")

	policy := &CommitMessagePolicy{
		ConventionalCommits: true,
		ValidateBranchName:  true,
		BranchPattern:       `^(feature|fix|chore|story|task|bug|sub-task)/((?:PRJ|ERM)-[0-9]+-[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*|[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*)$`,
	}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err == nil {
		t.Error("invalid branch name should fail validation")
	}
}

func TestApplyCommitMessagePolicy_FooterLabel(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/PRJ-777-something")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "feat: custom label test\n")

	policy := &CommitMessagePolicy{
		AppendTicketFooter: true,
		FooterLabel:        "Fixes",
	}
	if err := applyCommitMessagePolicy(dir, policy, msgFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(msgFile)
	if !contains(string(data), "Fixes: PRJ-777") {
		t.Errorf("expected custom label 'Fixes: PRJ-777', got:\n%s", string(data))
	}
}

// helpers

func initBareGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@test.com")
	runCmd(t, dir, "git", "config", "user.name", "Test")
	// Create an initial commit so branches can be created
	writeFile(t, filepath.Join(dir, "README.md"), "# test\n")
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "-c", "commit.gpgsign=false", "commit", "-m", "chore: init")
	return dir
}

func createBranchInRepo(t *testing.T, dir, branch string) {
	t.Helper()
	runCmd(t, dir, "git", "checkout", "-b", branch)
}

func createBranchInRepoIfNotExists(t *testing.T, dir, branch string) {
	t.Helper()
	// Check current branch first; if it already matches, nothing to do.
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == branch {
		return
	}
	runCmd(t, dir, "git", "checkout", "-b", branch)
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cmd %s %v failed: %v\n%s", name, args, err, out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func countOccurrences(s, sub string) int {
	count := 0
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			count++
		}
	}
	return count
}

func TestApplyPrepareCommitMsgPolicy_PrependTicket(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/PRJ-42-my-feature")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "add widget\n")

	policy := &CommitMessagePolicy{PrependTicket: true}
	if err := applyPrepareCommitMsgPolicy(dir, policy, msgFile, "message"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(msgFile)
	if !contains(string(content), "PRJ-42: ") {
		t.Errorf("expected ticket prefix, got: %q", string(content))
	}
}

func TestApplyPrepareCommitMsgPolicy_SkipOnMerge(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/PRJ-42-my-feature")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	original := "Merge branch 'main'\n"
	writeFile(t, msgFile, original)

	policy := &CommitMessagePolicy{PrependTicket: true, SkipOnMerge: true}
	if err := applyPrepareCommitMsgPolicy(dir, policy, msgFile, "merge"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(msgFile)
	if string(content) != original {
		t.Errorf("expected file unchanged for merge, got: %q", string(content))
	}
}

func TestApplyPrepareCommitMsgPolicy_SkipIfPresent(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/PRJ-42-my-feature")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	original := "PRJ-42: already present\n"
	writeFile(t, msgFile, original)

	policy := &CommitMessagePolicy{PrependTicket: true, SkipIfPresent: true}
	if err := applyPrepareCommitMsgPolicy(dir, policy, msgFile, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(msgFile)
	if string(content) != original {
		t.Errorf("expected file unchanged when ticket already present, got: %q", string(content))
	}
}

func TestApplyPrepareCommitMsgPolicy_NilPolicy(t *testing.T) {
	dir := t.TempDir()
	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	writeFile(t, msgFile, "some message\n")

	if err := applyPrepareCommitMsgPolicy(dir, nil, msgFile, ""); err != nil {
		t.Errorf("nil policy should be no-op, got: %v", err)
	}
}

func TestApplyPrepareCommitMsgPolicy_NoTicketInBranch(t *testing.T) {
	dir := initBareGitRepo(t)
	createBranchInRepo(t, dir, "feat/no-ticket-here")

	msgFile := filepath.Join(dir, "COMMIT_EDITMSG")
	original := "some message\n"
	writeFile(t, msgFile, original)

	policy := &CommitMessagePolicy{PrependTicket: true}
	if err := applyPrepareCommitMsgPolicy(dir, policy, msgFile, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(msgFile)
	if string(content) != original {
		t.Errorf("file should be unchanged when no ticket in branch, got: %q", string(content))
	}
}
