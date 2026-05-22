package booster

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
	// ModeCharDevice set => TTY
	return (info.Mode() & os.ModeCharDevice) != 0
}

func colorize(color, s string) string {
	if !colorEnabled() {
		return s
	}
	return color + s + ansiReset
}

func bold(s string) string   { return colorize(ansiBold, s) }
func green(s string) string  { return colorize(ansiGreen, s) }
func red(s string) string    { return colorize(ansiRed, s) }
func yellow(s string) string { return colorize(ansiYellow, s) }
func dim(s string) string    { return colorize(ansiDim, s) }
func cyan(s string) string   { return colorize(ansiCyan, s) }

// fmtDuration formats a duration as a human-readable string scaled to the magnitude.
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

// PrintHookHeader prints the hook name banner.
func PrintHookHeader(hookName string) {
	fmt.Fprintf(UI, "\n%s\n", bold(hookName))
}

// PrintToolResult prints a single tool result line.
func PrintToolResult(r ToolResult) {
	var icon, nameStr, durStr string

	switch r.Status {
	case "pass":
		icon = green("✓")
		nameStr = r.Name
	case "fail":
		icon = red("✗")
		nameStr = red(r.Name)
	case "skip":
		icon = yellow("~")
		nameStr = dim(r.Name)
	case "cached":
		icon = cyan("↩")
		nameStr = dim(r.Name)
	default:
		icon = " "
		nameStr = r.Name
	}

	if r.Duration > 0 {
		durStr = "  " + dim(fmtDuration(r.Duration))
	}

	fmt.Fprintf(UI, "  %s  %-24s%s\n", icon, nameStr, durStr)

	if r.Output != "" {
		for _, line := range strings.Split(strings.TrimRight(r.Output, "\n"), "\n") {
			fmt.Fprintf(UI, "     %s\n", dim(line))
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
		parts = append(parts, green(fmt.Sprintf("%d passed", passed)))
	}
	if failed > 0 {
		parts = append(parts, red(fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, yellow(fmt.Sprintf("%d skipped", skipped)))
	}
	parts = append(parts, dim(fmt.Sprintf("total %s", fmtDuration(total))))

	fmt.Fprintf(UI, "\n%s\n", strings.Join(parts, dim(" · ")))
}
