package booster

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOptions_AllFilesOnlyValidForPreCommit(t *testing.T) {
	dir := initBareGitRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	writeFile(t, filepath.Join(dir, "booster.toml"), `
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
