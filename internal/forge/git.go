package forge

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func detectRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not in a git repository: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func stagedFiles(repoRoot string) ([]string, error) {
	out, err := runGit(repoRoot, "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	items := strings.Split(out, "\n")
	res := make([]string, 0, len(items))
	for _, v := range items {
		v = strings.TrimSpace(v)
		if v != "" {
			res = append(res, filepath.ToSlash(v))
		}
	}
	return res, nil
}

func currentBranch(repoRoot string) (string, error) {
	return runGit(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
}

func addFiles(repoRoot string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := []string{"add", "--"}
	args = append(args, files...)
	_, err := runGit(repoRoot, args...)
	return err
}

func localHooksPath(repoRoot string) (string, error) {
	out, err := runGit(repoRoot, "config", "--local", "--get", "core.hooksPath")
	if err != nil {
		return "", nil
	}
	return out, nil
}

// allTrackedFiles returns all files tracked by git in the repo (respects .gitignore).
func allTrackedFiles(repoRoot string) ([]string, error) {
	out, err := runGit(repoRoot, "ls-files")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	items := strings.Split(out, "\n")
	res := make([]string, 0, len(items))
	for _, v := range items {
		v = strings.TrimSpace(v)
		if v != "" {
			res = append(res, filepath.ToSlash(v))
		}
	}
	return res, nil
}

const stashLabel = "booster-pre-commit-safety"

// hasUnstagedChanges returns true when there are unstaged modifications to
// tracked files or untracked files in the working tree.
func hasUnstagedChanges(repoRoot string) (bool, error) {
	// Check modified tracked files not yet staged.
	out, err := runGit(repoRoot, "diff", "--name-only")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(out) != "" {
		return true, nil
	}
	// Check untracked files (non-ignored).
	out, err = runGit(repoRoot, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// stashUnstagedChanges stashes the working tree (unstaged changes + untracked
// files) while keeping the index intact. Returns the stash ref (e.g.
// "stash@{0}"), whether a stash was actually created, and any error.
// A stash is NOT created when there is nothing to stash.
func stashUnstagedChanges(repoRoot string) (stashRef string, created bool, err error) {
	if isTruthy(os.Getenv("FORGE_NO_STASH")) {
		return "", false, nil
	}
	has, err := hasUnstagedChanges(repoRoot)
	if err != nil || !has {
		return "", false, err
	}
	_, err = runGit(repoRoot, "stash", "push", "--keep-index", "--include-untracked",
		"-m", stashLabel)
	if err != nil {
		return "", false, fmt.Errorf("stash failed: %w", err)
	}
	// Confirm a stash entry was created (git exits 0 even with "No local changes").
	out, err := runGit(repoRoot, "stash", "list", "--max-count=1")
	if err != nil || !strings.Contains(out, stashLabel) {
		return "", false, nil
	}
	return "stash@{0}", true, nil
}

// popStash restores the stash created by stashUnstagedChanges.
func popStash(repoRoot string) error {
	_, err := runGit(repoRoot, "stash", "pop", "--index")
	if err != nil {
		return fmt.Errorf(
			"could not restore stashed changes automatically — run 'git stash pop' to restore: %w", err,
		)
	}
	return nil
}
