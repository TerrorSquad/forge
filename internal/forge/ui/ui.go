package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ANSI color codes
const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
)

// UI is the output writer used by the runner. Swap out in tests.
var UI io.Writer = os.Stdout

// colorEnabled returns true when ANSI colors should be emitted.
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	f, ok := UI.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(f)
}

// isTerminal reports whether f is a character device (TTY).
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func colorize(color, s string) string {
	if !colorEnabled() {
		return s
	}
	return color + s + ansiReset
}

// Bold applies bold formatting.
func Bold(s string) string { return colorize(ansiBold, s) }

// Green applies green color.
func Green(s string) string { return colorize(ansiGreen, s) }

// Red applies red color.
func Red(s string) string { return colorize(ansiRed, s) }

// Yellow applies yellow color.
func Yellow(s string) string { return colorize(ansiYellow, s) }

// Dim applies dim formatting.
func Dim(s string) string { return colorize(ansiDim, s) }

// Cyan applies cyan color.
func Cyan(s string) string { return colorize(ansiCyan, s) }

func fmtDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// ToolResult holds the outcome of a single tool execution.
type ToolResult struct {
	Name     string
	Status   string // "pass", "fail", "skip", "cached"
	Duration time.Duration
	Output   string // captured stderr/stdout on failure
}

// isInteractiveTerminal reports whether the UI writer is a real terminal.
func isInteractiveTerminal() bool {
	f, ok := UI.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(f)
}

// PrintRunning writes a dim "·  name" indicator without a trailing newline.
func PrintRunning(name string) {
	if !isInteractiveTerminal() {
		return
	}
	fmt.Fprintf(UI, "  %s  %s", Dim("·"), Dim(name))
}

// ClearRunning erases the running indicator on the current line.
func ClearRunning() {
	if !isInteractiveTerminal() {
		return
	}
	fmt.Fprintf(UI, "\r\033[2K")
}

// PrintHookHeader prints the hook name banner.
func PrintHookHeader(hookName string) {
	fmt.Fprintf(UI, "\n%s\n", Bold(hookName))
}

// PrintToolResult prints a single tool result line.
func PrintToolResult(r ToolResult) {
	var icon, nameStr, durStr string

	switch r.Status {
	case "pass":
		icon = Green("✓")
		nameStr = r.Name
	case "fail", "would-fail":
		icon = Red("✗")
		nameStr = Red(r.Name)
	case "skip":
		icon = Yellow("~")
		nameStr = Dim(r.Name)
	case "cached":
		icon = Cyan("↩")
		nameStr = Dim(r.Name)
	default:
		icon = " "
		nameStr = r.Name
	}

	if r.Duration > 0 {
		durStr = "  " + Dim(fmtDuration(r.Duration))
	}

	fmt.Fprintf(UI, "  %s  %-24s%s\n", icon, nameStr, durStr)

	if r.Output != "" {
		for _, line := range strings.Split(strings.TrimRight(r.Output, "\n"), "\n") {
			fmt.Fprintf(UI, "     %s\n", Dim(line))
		}
	}
}

// PrintSummary prints the final pass/fail summary line.
func PrintSummary(results []ToolResult, total time.Duration) {
	passed, failed, skipped := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "pass", "cached":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}

	parts := []string{}
	if passed > 0 {
		parts = append(parts, Green(fmt.Sprintf("%d passed", passed)))
	}
	if failed > 0 {
		parts = append(parts, Red(fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, Yellow(fmt.Sprintf("%d skipped", skipped)))
	}
	parts = append(parts, Dim(fmt.Sprintf("total %s", fmtDuration(total))))

	fmt.Fprintf(UI, "\n%s\n", strings.Join(parts, Dim(" · ")))
}

// PrintCheckSummary prints a check-mode "would fail" summary line.
func PrintCheckSummary(results []ToolResult, total time.Duration) {
	passed, wouldFail, skipped := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case "pass", "cached":
			passed++
		case "would-fail", "fail":
			wouldFail++
		case "skip":
			skipped++
		}
	}

	parts := []string{}
	if passed > 0 {
		parts = append(parts, Green(fmt.Sprintf("%d passed", passed)))
	}
	if wouldFail > 0 {
		parts = append(parts, Red(fmt.Sprintf("%d would fail", wouldFail)))
	}
	if skipped > 0 {
		parts = append(parts, Yellow(fmt.Sprintf("%d skipped", skipped)))
	}
	parts = append(parts, Dim(fmt.Sprintf("total %s", fmtDuration(total))))

	fmt.Fprintf(UI, "\nCheck complete: %s\n", strings.Join(parts, Dim(" · ")))
}
