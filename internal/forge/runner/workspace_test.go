package runner

import (
	"os"
	"path/filepath"
	"testing"
)

// Fix 3: pathHasPrefix must use forward-slash separator because StagedFiles
// normalises all paths with filepath.ToSlash on every platform.

func TestPathHasPrefix_ForwardSlash(t *testing.T) {
	cases := []struct {
		file, prefix string
		want         bool
	}{
		{"packages/foo/src/Bar.php", "packages/foo", true},
		{"packages/foobar/src/Baz.php", "packages/foo", false}, // prefix match must stop at separator
		{"packages/foo/deep/nested/File.php", "packages/foo", true},
		{"other/file.php", "packages/foo", false},
		{"packages/foo/file.php", "packages/foo/", true}, // trailing slash in prefix is OK
		{"anything", "", true},                            // empty prefix always matches
	}
	for _, c := range cases {
		got := pathHasPrefix(c.file, c.prefix)
		if got != c.want {
			t.Errorf("pathHasPrefix(%q, %q) = %v, want %v", c.file, c.prefix, got, c.want)
		}
	}
}

func TestPathHasPrefix_NoPrefixCollision(t *testing.T) {
	// "packages/foo" must not match files under "packages/foobar"
	if pathHasPrefix("packages/foobar/main.go", "packages/foo") {
		t.Error("packages/foo should not be a prefix of packages/foobar/main.go")
	}
}

// ---------- isDir ----------

func TestIsDir_Directory(t *testing.T) {
	dir := t.TempDir()
	if !isDir(dir) {
		t.Error("isDir should return true for an actual directory")
	}
}

func TestIsDir_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if isDir(f) {
		t.Error("isDir should return false for a regular file")
	}
}

func TestIsDir_Nonexistent(t *testing.T) {
	if isDir("/nonexistent/path/xyz-forge-test") {
		t.Error("isDir should return false for a nonexistent path")
	}
}

// ---------- matchingMembers ----------

func TestMatchingMembers_MatchesStagedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "foo"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "packages", "bar"), 0755); err != nil {
		t.Fatal(err)
	}

	staged := []string{"packages/foo/main.go", "packages/foo/helper.go"}
	got, err := matchingMembers(dir, []string{"packages/*"}, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "packages/foo" {
		t.Errorf("got %v, want [packages/foo]", got)
	}
}

func TestMatchingMembers_MultipleMembers(t *testing.T) {
	dir := t.TempDir()
	for _, m := range []string{"packages/foo", "packages/bar"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(m)), 0755); err != nil {
			t.Fatal(err)
		}
	}

	staged := []string{"packages/foo/a.go", "packages/bar/b.go"}
	got, err := matchingMembers(dir, []string{"packages/*"}, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("got %v, want 2 members", got)
	}
}

func TestMatchingMembers_DeduplicatesMember(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "foo"), 0755); err != nil {
		t.Fatal(err)
	}

	// Two staged files in the same member — member should appear only once.
	staged := []string{"packages/foo/a.go", "packages/foo/b.go"}
	got, err := matchingMembers(dir, []string{"packages/*"}, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("got %v, want exactly 1 unique member", got)
	}
}

func TestMatchingMembers_IgnoresFilePaths(t *testing.T) {
	dir := t.TempDir()
	// Create a regular file that would match the glob, not a directory.
	if err := os.WriteFile(filepath.Join(dir, "notadir"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	staged := []string{"notadir/file.go"}
	got, err := matchingMembers(dir, []string{"notadir"}, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no matches for file path, got %v", got)
	}
}

func TestMatchingMembers_NoMatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "packages", "foo"), 0755); err != nil {
		t.Fatal(err)
	}

	staged := []string{"src/main.go"}
	got, err := matchingMembers(dir, []string{"packages/*"}, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestMatchingMembers_InvalidPattern(t *testing.T) {
	dir := t.TempDir()
	_, err := matchingMembers(dir, []string{"["}, []string{"foo/bar.go"})
	if err == nil {
		t.Error("expected error for syntactically invalid glob pattern")
	}
}

// ---------- stagedFilesForMember ----------

func initWorkspaceRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@test.com")
	runCmd(t, dir, "git", "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, dir, "git", "add", ".")
	runCmd(t, dir, "git", "-c", "commit.gpgsign=false", "commit", "-m", "init")
	return dir
}

func stageWorkspaceFile(t *testing.T, repoRoot, relPath, content string) {
	t.Helper()
	full := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, repoRoot, "git", "add", relPath)
}

func TestStagedFilesForMember_FiltersByPrefix(t *testing.T) {
	repo := initWorkspaceRepo(t)
	stageWorkspaceFile(t, repo, "packages/foo/main.go", "package main\n")
	stageWorkspaceFile(t, repo, "packages/bar/main.go", "package main\n")

	got, err := stagedFilesForMember(repo, "packages/foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "main.go" {
		t.Errorf("got %v, want [main.go]", got)
	}
}

func TestStagedFilesForMember_NoMatch(t *testing.T) {
	repo := initWorkspaceRepo(t)
	stageWorkspaceFile(t, repo, "packages/bar/main.go", "package main\n")

	got, err := stagedFilesForMember(repo, "packages/foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected no files for non-matching member, got %v", got)
	}
}

func TestStagedFilesForMember_MultipleFiles(t *testing.T) {
	repo := initWorkspaceRepo(t)
	stageWorkspaceFile(t, repo, "packages/foo/a.go", "package a\n")
	stageWorkspaceFile(t, repo, "packages/foo/b.go", "package b\n")
	stageWorkspaceFile(t, repo, "packages/foo/sub/c.go", "package sub\n")

	got, err := stagedFilesForMember(repo, "packages/foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 files, got %v", got)
	}
	// All paths should be relative to the member (prefix stripped).
	for _, f := range got {
		if f == "" {
			t.Error("expected non-empty relative path")
		}
	}
}
