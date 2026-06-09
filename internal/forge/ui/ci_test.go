package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// ---------- isTruthy ----------

func TestIsTruthy(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"yes", true},
		{"on", true},
		{"TRUE", true},
		{"  1  ", true},
		{"", false},
		{"0", false},
		{"false", false},
		{"no", false},
		{"off", false},
	}
	for _, c := range cases {
		if got := isTruthy(c.in); got != c.want {
			t.Errorf("isTruthy(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// ---------- ciMode ----------

func TestCiMode_ForgeOutputOverride(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "github")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	if got := ciMode(); got != "github" {
		t.Errorf("ciMode() = %q, want %q", got, "github")
	}
}

func TestCiMode_ForgeOutputLowercased(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "GITHUB")
	t.Setenv("GITHUB_ACTIONS", "")
	if got := ciMode(); got != "github" {
		t.Errorf("ciMode() = %q, want %q", got, "github")
	}
}

func TestCiMode_GitHubActions(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITLAB_CI", "")
	if got := ciMode(); got != "github" {
		t.Errorf("ciMode() = %q, want %q", got, "github")
	}
}

func TestCiMode_GitLabCI(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "true")
	if got := ciMode(); got != "gitlab" {
		t.Errorf("ciMode() = %q, want %q", got, "gitlab")
	}
}

func TestCiMode_None(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	if got := ciMode(); got != "" {
		t.Errorf("ciMode() = %q, want empty", got)
	}
}

// ---------- IsCI ----------

func TestIsCI_CIEnvVar(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	if !IsCI() {
		t.Error("IsCI() should be true when CI=true")
	}
}

func TestIsCI_GitHubActions(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITLAB_CI", "")
	if !IsCI() {
		t.Error("IsCI() should be true when GITHUB_ACTIONS=true")
	}
}

func TestIsCI_None(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	if IsCI() {
		t.Error("IsCI() should be false when no CI env vars are set")
	}
}

// ---------- escapeGHAValue / escapeGHAParam ----------

func TestEscapeGHAValue(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"%", "%25"},
		{"\r", "%0D"},
		{"\n", "%0A"},
		{"hello%world\nfoo\rbar", "hello%25world%0Afoo%0Dbar"},
		{"plain", "plain"},
	}
	for _, c := range cases {
		if got := escapeGHAValue(c.in); got != c.want {
			t.Errorf("escapeGHAValue(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeGHAParam(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{":", "%3A"},
		{",", "%2C"},
		{"src/foo.php", "src/foo.php"},
		{"src:foo,bar", "src%3Afoo%2Cbar"},
	}
	for _, c := range cases {
		if got := escapeGHAParam(c.in); got != c.want {
			t.Errorf("escapeGHAParam(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ---------- parseErrorAnnotation ----------

func TestParseErrorAnnotation_PHPStyle(t *testing.T) {
	line := "src/Foo.php:42 - Something went wrong"
	got := parseErrorAnnotation(line)
	if !strings.Contains(got, "file=src/Foo.php") {
		t.Errorf("expected file annotation, got %q", got)
	}
	if !strings.Contains(got, "line=42") {
		t.Errorf("expected line annotation, got %q", got)
	}
	if !strings.Contains(got, "::error") {
		t.Errorf("expected error level, got %q", got)
	}
}

func TestParseErrorAnnotation_GoStyle(t *testing.T) {
	line := "main.go:10:5: undefined: foo"
	got := parseErrorAnnotation(line)
	if !strings.Contains(got, "file=main.go") {
		t.Errorf("expected file annotation, got %q", got)
	}
	if !strings.Contains(got, "line=10") {
		t.Errorf("expected line annotation, got %q", got)
	}
	if !strings.Contains(got, "col=5") {
		t.Errorf("expected col annotation, got %q", got)
	}
}

func TestParseErrorAnnotation_GenericStyle(t *testing.T) {
	line := "some/file.ts:3: parse error"
	got := parseErrorAnnotation(line)
	if got == "" {
		t.Errorf("expected annotation for generic style, got empty")
	}
	if !strings.Contains(got, "line=3") {
		t.Errorf("expected line=3 in annotation, got %q", got)
	}
}

func TestParseErrorAnnotation_NoMatch(t *testing.T) {
	got := parseErrorAnnotation("this is just a plain message with no file or line info")
	if got != "" {
		t.Errorf("expected empty annotation, got %q", got)
	}
}

func TestParseErrorAnnotation_ErrorLevel(t *testing.T) {
	line := "ERROR: lint - src/main.php:42:1 - Something broke"
	got := parseErrorAnnotation(line)
	if got == "" {
		t.Errorf("expected annotation for ERROR level line, got empty")
	}
	if !strings.Contains(got, "::error") {
		t.Errorf("expected error level, got %q", got)
	}
}

func TestParseErrorAnnotation_WarningLevel(t *testing.T) {
	line := "WARNING: lint - src/main.php:42:1 - Some message"
	got := parseErrorAnnotation(line)
	if got == "" {
		t.Errorf("expected annotation for WARNING level line, got empty")
	}
	if !strings.Contains(got, "::warning") {
		t.Errorf("expected warning level, got %q", got)
	}
}

// ---------- ghaEmitAnnotations ----------

func TestGhaEmitAnnotations_Empty(t *testing.T) {
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	ghaEmitAnnotations("", "mytool")
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty input, got %q", buf.String())
	}
}

func TestGhaEmitAnnotations_WhitespaceOnly(t *testing.T) {
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	ghaEmitAnnotations("   \n\n  ", "mytool")
	if buf.Len() != 0 {
		t.Errorf("expected no output for whitespace-only input, got %q", buf.String())
	}
}

func TestGhaEmitAnnotations_PlainLineFallback(t *testing.T) {
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	ghaEmitAnnotations("something went wrong", "mytool")
	out := buf.String()
	if !strings.Contains(out, "::error title=mytool::") {
		t.Errorf("expected fallback annotation with tool name, got %q", out)
	}
}

func TestGhaEmitAnnotations_ParsedLine(t *testing.T) {
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	ghaEmitAnnotations("main.go:5:3: undefined var", "gotool")
	out := buf.String()
	if !strings.Contains(out, "file=main.go") {
		t.Errorf("expected parsed file annotation, got %q", out)
	}
}

func TestGhaEmitAnnotations_SkipsBlankLines(t *testing.T) {
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	ghaEmitAnnotations("main.go:1: undefined foo\n\nsecond.go:2: unused import", "tool")
	out := buf.String()
	// Blank line in the middle should be skipped — expect exactly 2 output lines.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 annotation lines (blank skipped), got %d: %q", len(lines), out)
	}
}

// ---------- PrintHookHeaderCI ----------

func TestPrintHookHeaderCI_GitHub(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "github")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	PrintHookHeaderCI("pre-commit")
	out := buf.String()
	if !strings.Contains(out, "::group::") {
		t.Errorf("expected GHA group annotation, got %q", out)
	}
	if !strings.Contains(out, "pre-commit") {
		t.Errorf("expected hook name in output, got %q", out)
	}
}

func TestPrintHookHeaderCI_NonCI(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	PrintHookHeaderCI("pre-commit")
	out := buf.String()
	if strings.Contains(out, "::group::") {
		t.Errorf("expected no GHA annotation in non-CI mode, got %q", out)
	}
	if !strings.Contains(out, "pre-commit") {
		t.Errorf("expected hook name in output, got %q", out)
	}
}

// ---------- PrintToolResultCI ----------

func TestPrintToolResultCI_GitHub_Fail_EmitsAnnotations(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "github")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	r := ToolResult{
		Name:   "eslint",
		Status: "fail",
		Output: "main.go:1:1: undefined foo",
	}
	PrintToolResultCI(r, "eslint")
	out := buf.String()
	if !strings.Contains(out, "eslint") {
		t.Errorf("expected tool name in output, got %q", out)
	}
	if !strings.Contains(out, "::error") {
		t.Errorf("expected GHA annotation for fail, got %q", out)
	}
}

func TestPrintToolResultCI_GitHub_Pass_NoAnnotations(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "github")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	r := ToolResult{Name: "gofmt", Status: "pass"}
	PrintToolResultCI(r, "gofmt")
	out := buf.String()
	if strings.Contains(out, "::error") {
		t.Errorf("expected no GHA annotation for pass, got %q", out)
	}
}

func TestPrintToolResultCI_NonCI_NoAnnotations(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	r := ToolResult{Name: "eslint", Status: "fail", Output: "error here"}
	PrintToolResultCI(r, "eslint")
	out := buf.String()
	if strings.Contains(out, "::error") {
		t.Errorf("expected no GHA annotation in non-CI mode, got %q", out)
	}
}

// ---------- PrintSummaryCI ----------

func TestPrintSummaryCI_GitHub_ClosesGroup(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "github")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	PrintSummaryCI([]ToolResult{{Name: "gofmt", Status: "pass"}}, 100*time.Millisecond, false)
	out := buf.String()
	if !strings.Contains(out, "::endgroup::") {
		t.Errorf("expected ::endgroup:: in GitHub CI mode, got %q", out)
	}
}

func TestPrintSummaryCI_NonCI_NoEndGroup(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	PrintSummaryCI([]ToolResult{{Name: "gofmt", Status: "pass"}}, 100*time.Millisecond, false)
	out := buf.String()
	if strings.Contains(out, "::endgroup::") {
		t.Errorf("expected no ::endgroup:: in non-CI mode, got %q", out)
	}
}

func TestPrintSummaryCI_CheckMode(t *testing.T) {
	t.Setenv("FORGE_OUTPUT", "")
	t.Setenv("GITHUB_ACTIONS", "")
	var buf bytes.Buffer
	orig := UI
	UI = &buf
	t.Cleanup(func() { UI = orig })

	PrintSummaryCI([]ToolResult{{Name: "gofmt", Status: "pass"}}, 50*time.Millisecond, true)
	// Should not panic and should produce some output
	if buf.Len() == 0 {
		t.Error("expected non-empty output from PrintSummaryCI in check mode")
	}
}
