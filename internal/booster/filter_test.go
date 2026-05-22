package booster

import (
	"testing"
)

func TestFilterFiles_ByExtension(t *testing.T) {
	files := []string{"main.go", "main_test.go", "README.md", "app.ts"}

	tool := ToolConfig{Extensions: []string{".go"}}
	got := filterFiles(files, tool)

	want := []string{"main.go", "main_test.go"}
	assertStringSlice(t, got, want)
}

func TestFilterFiles_NoExtensionFilter(t *testing.T) {
	files := []string{"a.go", "b.ts", "c.php"}
	tool := ToolConfig{}
	got := filterFiles(files, tool)
	assertStringSlice(t, got, files)
}

func TestFilterFiles_IncludePattern(t *testing.T) {
	files := []string{"src/main.go", "cmd/booster/main.go", "docs/README.md"}
	tool := ToolConfig{IncludePatterns: []string{"src/*"}}
	got := filterFiles(files, tool)
	assertStringSlice(t, got, []string{"src/main.go"})
}

func TestFilterFiles_ExcludePattern(t *testing.T) {
	files := []string{"main.go", "main_test.go", "generated.go"}
	tool := ToolConfig{ExcludePatterns: []string{"*_test.go"}}
	got := filterFiles(files, tool)
	assertStringSlice(t, got, []string{"generated.go", "main.go"})
}

func TestFilterFiles_ExtensionAndExclude(t *testing.T) {
	files := []string{"a.go", "b_test.go", "c.ts"}
	tool := ToolConfig{
		Extensions:      []string{".go"},
		ExcludePatterns: []string{"*_test.go"},
	}
	got := filterFiles(files, tool)
	assertStringSlice(t, got, []string{"a.go"})
}

func TestFilterFiles_Empty(t *testing.T) {
	tool := ToolConfig{Extensions: []string{".go"}}
	got := filterFiles(nil, tool)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestFilterFiles_CaseInsensitiveExtension(t *testing.T) {
	files := []string{"Main.GO", "other.ts"}
	tool := ToolConfig{Extensions: []string{".go"}}
	got := filterFiles(files, tool)
	if len(got) != 1 {
		t.Errorf("expected 1 match for case-insensitive ext, got %v", got)
	}
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, len(want) = %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q, want %q", i, got[i], want[i])
		}
	}
}
