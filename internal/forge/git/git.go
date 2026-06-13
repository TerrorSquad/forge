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

// IsSequencerOperation returns true when Git is in a merge/rebase/cherry-pick/revert/bisect/am operation.
func IsSequencerOperation(repoRoot string) (bool, error) {
	gitDir, err := RunGit(repoRoot, "rev-parse", "--git-dir")
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoRoot, gitDir)
	}

	paths := []string{
		"MERGE_HEAD",
		"CHERRY_PICK_HEAD",
		"REVERT_HEAD",
		"BISECT_LOG",
		"REBASE_HEAD",
		"AM_HEAD",
		"sequencer",
		"rebase-apply",
		"rebase-merge",
	}

	for _, p := range paths {
		if exists(filepath.Join(gitDir, p)) {
			return true, nil
		}
	}
	return false, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
