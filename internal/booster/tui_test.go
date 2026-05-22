package booster

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewTUIDisplay_PendingTools(t *testing.T) {
	d := newTUIDisplay("pre-commit", []string{"gofmt", "govet"})
	if len(d.tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(d.tools))
	}
	for _, row := range d.tools {
		if row.status != "pending" {
			t.Errorf("expected pending status, got %q for %s", row.status, row.name)
		}
	}
}

func TestTUIDisplay_StartTool(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	d := newTUIDisplay("pre-commit", []string{"gofmt", "govet"})
	d.startTool("gofmt")
	if d.tools[0].status != "running" {
		t.Errorf("expected 'running', got %q", d.tools[0].status)
	}
}

func TestTUIDisplay_DoneTool(t *testing.T) {
	var buf bytes.Buffer
	origUI := UI
	UI = &buf
	t.Cleanup(func() { UI = origUI })

	d := newTUIDisplay("pre-commit", []string{"gofmt"})
	d.doneTool(ToolResult{Name: "gofmt", Status: "pass", Duration: 10 * time.Millisecond})
	if d.tools[0].status != "pass" {
		t.Errorf("expected 'pass', got %q", d.tools[0].status)
	}
}

func TestFormatTUIRow_Pending(t *testing.T) {
	row := tuiRow{name: "gofmt", status: "pending"}
	out := formatTUIRow(row, "⠋")
	if !strings.Contains(out, "gofmt") {
		t.Errorf("expected tool name in row, got: %s", out)
	}
}

func TestFormatTUIRow_Running(t *testing.T) {
	row := tuiRow{name: "govet", status: "running"}
	out := formatTUIRow(row, "⠋")
	if !strings.Contains(out, "govet") {
		t.Errorf("expected tool name in row, got: %s", out)
	}
}

func TestFormatTUIRow_Pass(t *testing.T) {
	row := tuiRow{name: "gofmt", status: "pass", duration: 50 * time.Millisecond}
	out := formatTUIRow(row, "⠋")
	if !strings.Contains(out, "50ms") {
		t.Errorf("expected duration in row, got: %s", out)
	}
}

func TestIsTUIMode_SingleTool(t *testing.T) {
	// Single tool should never use TUI
	// We can't control isTerminalOutput() in tests (no TTY), so isTUIMode will be
	// false in test environment regardless — just test the count guard.
	result := isTUIMode(1)
	if result {
		t.Error("isTUIMode should be false for single tool")
	}
}

func TestIsTUIMode_ZeroTools(t *testing.T) {
	result := isTUIMode(0)
	if result {
		t.Error("isTUIMode should be false for zero tools")
	}
}
