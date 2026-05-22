package booster

import (
	"bytes"
	"fmt"
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
