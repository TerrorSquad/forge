package runner

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/TerrorSquad/forge/internal/forge/config"
	"github.com/TerrorSquad/forge/internal/forge/ui"
)

func TestToolConfigForCheck_NormalMode(t *testing.T) {
	tool := config.ToolConfig{
		Command:   "gofmt",
		Args:      []string{"-w"},
		CheckArgs: []string{"-l"},
	}
	result := toolConfigForCheck(tool, false)
	if len(result.Args) != 1 || result.Args[0] != "-w" {
		t.Errorf("expected normal args in normal mode, got %v", result.Args)
	}
}

func TestToolConfigForCheck_CheckModeWithCheckArgs(t *testing.T) {
	tool := config.ToolConfig{
		Command:   "gofmt",
		Args:      []string{"-w"},
		CheckArgs: []string{"-l"},
	}
	result := toolConfigForCheck(tool, true)
	if len(result.Args) != 1 || result.Args[0] != "-l" {
		t.Errorf("expected check_args in check mode, got %v", result.Args)
	}
}

func TestToolConfigForCheck_CheckModeNoCheckArgs(t *testing.T) {
	tool := config.ToolConfig{
		Command: "govet",
		Args:    []string{"./..."},
	}
	result := toolConfigForCheck(tool, true)
	if len(result.Args) != 1 || result.Args[0] != "./..." {
		t.Errorf("expected original args when check_args empty, got %v", result.Args)
	}
}

func TestPrintCheckSummary_AllPass(t *testing.T) {
	var buf bytes.Buffer
	origUI := ui.UI
	ui.UI = &buf
	t.Cleanup(func() { ui.UI = origUI })

	results := []ui.ToolResult{
		{Name: "gofmt", Status: "pass", Duration: 10 * time.Millisecond},
		{Name: "govet", Status: "pass", Duration: 20 * time.Millisecond},
	}
	ui.PrintCheckSummary(results, 30*time.Millisecond)
	out := buf.String()
	if !strings.Contains(out, "Check complete") {
		t.Errorf("expected 'Check complete' in output, got: %s", out)
	}
	if !strings.Contains(out, "2 passed") {
		t.Errorf("expected '2 passed' in output, got: %s", out)
	}
}

func TestPrintCheckSummary_WouldFail(t *testing.T) {
	var buf bytes.Buffer
	origUI := ui.UI
	ui.UI = &buf
	t.Cleanup(func() { ui.UI = origUI })

	results := []ui.ToolResult{
		{Name: "gofmt", Status: "pass", Duration: 10 * time.Millisecond},
		{Name: "govet", Status: "would-fail", Duration: 20 * time.Millisecond},
	}
	ui.PrintCheckSummary(results, 30*time.Millisecond)
	out := buf.String()
	if !strings.Contains(out, "1 would fail") {
		t.Errorf("expected '1 would fail' in output, got: %s", out)
	}
	if !strings.Contains(out, "1 passed") {
		t.Errorf("expected '1 passed' in output, got: %s", out)
	}
}

func TestRunOptions_CheckMode(t *testing.T) {
	opts := RunOptions{CheckMode: true}
	if !opts.CheckMode {
		t.Error("expected CheckMode to be true")
	}
}
