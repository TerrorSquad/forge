package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// initRepoWithBranch creates a bare git repo on the given branch.
func initRepoWithBranch(t *testing.T, branch string) string {
	t.Helper()
	dir := initBareGitRepo(t)
	createBranchInRepoIfNotExists(t, dir, branch)
	return dir
}
