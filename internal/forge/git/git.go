package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunGit runs a git command in repoRoot and returns combined stdout.
func RunGit(repoRoot string, args ...string) (string, error) {
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

// DetectRepoRoot returns the absolute path of the git repository root.
func DetectRepoRoot() (string, error) {
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

// StagedFiles returns the list of staged files (ACMR filter).
func StagedFiles(repoRoot string) ([]string, error) {
	out, err := RunGit(repoRoot, "diff", "--cached", "--name-only", "--diff-filter=ACMR")
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

// CurrentBranch returns the current branch name.
func CurrentBranch(repoRoot string) (string, error) {
	return RunGit(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
}

// AddFiles stages the given file paths.
func AddFiles(repoRoot string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := []string{"add", "--"}
	args = append(args, files...)
	_, err := RunGit(repoRoot, args...)
	return err
}

// LocalHooksPath returns the configured core.hooksPath, or "" if unset.
func LocalHooksPath(repoRoot string) (string, error) {
	out, err := RunGit(repoRoot, "config", "--local", "--get", "core.hooksPath")
	if err != nil {
		return "", nil
	}
	return out, nil
}

// AllTrackedFiles returns all files tracked by git in the repo.
func AllTrackedFiles(repoRoot string) ([]string, error) {
	out, err := RunGit(repoRoot, "ls-files")
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

const stashLabel = "forge-pre-commit-safety"

// HasUnstagedChanges returns true when there are unstaged modifications or untracked files.
func HasUnstagedChanges(repoRoot string) (bool, error) {
	out, err := RunGit(repoRoot, "diff", "--name-only")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(out) != "" {
		return true, nil
	}
	out, err = RunGit(repoRoot, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// StashUnstagedChanges stashes working-tree changes while keeping the index.
// Returns the stash ref, whether a stash was created, and any error.
func StashUnstagedChanges(repoRoot string) (stashRef string, created bool, err error) {
	if isTruthy(os.Getenv("FORGE_NO_STASH")) {
		return "", false, nil
	}
	has, err := HasUnstagedChanges(repoRoot)
	if err != nil || !has {
		return "", false, err
	}
	_, err = RunGit(repoRoot, "stash", "push", "--keep-index", "--include-untracked",
		"-m", stashLabel)
	if err != nil {
		return "", false, fmt.Errorf("stash failed: %w", err)
	}
	out, err := RunGit(repoRoot, "stash", "list", "--max-count=1")
	if err != nil || !strings.Contains(out, stashLabel) {
		return "", false, nil
	}
	return "stash@{0}", true, nil
}

// PopStash restores the stash created by StashUnstagedChanges.
func PopStash(repoRoot string) error {
	_, err := RunGit(repoRoot, "stash", "pop", "--index")
	if err != nil {
		return fmt.Errorf(
			"could not restore stashed changes automatically — run 'git stash pop' to restore: %w", err,
		)
	}
	return nil
}

func isTruthy(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
