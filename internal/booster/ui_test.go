package booster

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintToolResult_Success(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	result := ToolResult{
		Name:     "gofmt",
		Status:   "pass",
		Duration: 50 * time.Millisecond,
	}
	PrintToolResult(result)

	out := buf.String()
	if !strings.Contains(out, "gofmt") {
		t.Errorf("expected tool name in output, got: %s", out)
	}
	if !strings.Contains(out, "50ms") {
		t.Errorf("expected duration in output, got: %s", out)
	}
}

func TestPrintToolResult_Failure(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	result := ToolResult{
		Name:   "eslint",
		Status: "fail",
		Output: "error: semicolon expected",
	}
	PrintToolResult(result)

	out := buf.String()
	if !strings.Contains(out, "eslint") {
		t.Errorf("expected tool name, got: %s", out)
	}
	if !strings.Contains(out, "error: semicolon expected") {
		t.Errorf("expected error output, got: %s", out)
	}
}

func TestPrintToolResult_Skipped(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	result := ToolResult{Name: "prettier", Status: "skip"}
	PrintToolResult(result)

	out := buf.String()
	if !strings.Contains(out, "prettier") {
		t.Errorf("expected tool name, got: %s", out)
	}
}

func TestPrintSummary_AllPass(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	results := []ToolResult{
		{Name: "gofmt", Status: "pass"},
		{Name: "govet", Status: "pass"},
	}
	PrintSummary(results, 100*time.Millisecond)

	out := buf.String()
	if !strings.Contains(out, "2 passed") {
		t.Errorf("expected '2 passed', got: %s", out)
	}
}

func TestPrintSummary_MixedResults(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	results := []ToolResult{
		{Name: "gofmt", Status: "pass"},
		{Name: "eslint", Status: "fail"},
		{Name: "prettier", Status: "skip"},
	}
	PrintSummary(results, 200*time.Millisecond)

	out := buf.String()
	if !strings.Contains(out, "1 passed") {
		t.Errorf("expected '1 passed' in: %s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected '1 failed' in: %s", out)
	}
}

func TestPrintHookHeader(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	PrintHookHeader("pre-commit")
	out := buf.String()
	if !strings.Contains(out, "pre-commit") {
		t.Errorf("expected hook name in header, got: %s", out)
	}
}

func TestFmtDuration_Scales(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Nanosecond, "500ns"},
		{1500 * time.Nanosecond, "1µs"},
		{1500 * time.Microsecond, "1ms"},
		{2500 * time.Millisecond, "2.50s"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got := fmtDuration(c.d)
			if got != c.want {
				t.Errorf("fmtDuration(%v) = %q, want %q", c.d, got, c.want)
			}
		})
	}
}
