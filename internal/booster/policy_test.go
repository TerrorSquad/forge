package booster

import (
	"os"
	"os/exec"
	"path/filepath"
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
