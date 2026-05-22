package booster

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// tuiRow holds the render state for a single tool row.
type tuiRow struct {
	name     string
	status   string // "pending", "running", "pass", "fail", "would-fail", "skip", "cached"
	duration time.Duration
	output   string
}

// tuiDisplay is a lock-protected, ANSI in-place terminal renderer.
// It replaces PrintToolResult / PrintSummary when TUI mode is active.
type tuiDisplay struct {
	mu           sync.Mutex
	hookName     string
	tools        []tuiRow
	spinner      int
	linesPrinted int  // how many lines we've printed so we can overwrite them
	started      bool // first render has happened
}

// newTUIDisplay creates a display pre-populated with all tool names as "pending".
func newTUIDisplay(hookName string, toolNames []string) *tuiDisplay {
	rows := make([]tuiRow, len(toolNames))
	for i, n := range toolNames {
		rows[i] = tuiRow{name: n, status: "pending"}
	}
	return &tuiDisplay{hookName: hookName, tools: rows}
}

func (d *tuiDisplay) startTool(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.tools {
		if d.tools[i].name == name {
			d.tools[i].status = "running"
		}
	}
	d.render()
}

func (d *tuiDisplay) doneTool(r ToolResult) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.tools {
		if d.tools[i].name == r.Name {
			d.tools[i].status = r.Status
			d.tools[i].duration = r.Duration
			d.tools[i].output = r.Output
		}
	}
	d.render()
}

// finish prints the final summary after all tools have completed.
func (d *tuiDisplay) finish(results []ToolResult, elapsed time.Duration, checkMode bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.render()
	if checkMode {
		PrintCheckSummary(results, elapsed)
	} else {
		PrintSummary(results, elapsed)
	}
}

// render overwrites the previously printed rows with the current state.
// Must be called with d.mu held.
func (d *tuiDisplay) render() {
	frame := spinnerFrames[d.spinner%len(spinnerFrames)]

	var lines []string
	// header line
	if !d.started {
		lines = append(lines, "\n"+bold(d.hookName))
	}

	for _, row := range d.tools {
		lines = append(lines, formatTUIRow(row, frame))
		if row.output != "" {
			for _, l := range strings.Split(strings.TrimRight(row.output, "\n"), "\n") {
				if l != "" {
					lines = append(lines, "     "+dim(l))
				}
			}
		}
	}

	if d.started && d.linesPrinted > 0 {
		// Move cursor up and overwrite
		fmt.Fprintf(UI, "\033[%dA", d.linesPrinted)
		for range d.linesPrinted {
			fmt.Fprintf(UI, "\033[2K\033[1G\n")
		}
		fmt.Fprintf(UI, "\033[%dA", d.linesPrinted)
	}

	for _, l := range lines {
		fmt.Fprintf(UI, "%s\n", l)
	}

	if !d.started {
		d.started = true
	}
	d.linesPrinted = len(lines)
}

func formatTUIRow(row tuiRow, spinnerFrame string) string {
	var icon, nameStr, durStr string
	if row.duration > 0 {
		durStr = "  " + dim(fmtDuration(row.duration))
	}
	switch row.status {
	case "pending":
		icon = dim("·")
		nameStr = dim(row.name)
	case "running":
		icon = cyan(spinnerFrame)
		nameStr = row.name
	case "pass":
		icon = green("✓")
		nameStr = row.name
	case "fail", "would-fail":
		icon = red("✗")
		nameStr = red(row.name)
	case "skip":
		icon = yellow("~")
		nameStr = dim(row.name)
	case "cached":
		icon = cyan("↩")
		nameStr = dim(row.name)
	default:
		icon = " "
		nameStr = row.name
	}
	return fmt.Sprintf("  %s  %-24s%s", icon, nameStr, durStr)
}

// spinnerLoop ticks the spinner animation until stop is closed.
func (d *tuiDisplay) spinnerLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			d.mu.Lock()
			d.spinner++
			d.render()
			d.mu.Unlock()
		}
	}
}

// isTerminalOutput reports whether stdout is an interactive terminal.
func isTerminalOutput() bool {
	if os.Getenv("BOOSTER_NO_TUI") != "" {
		return false
	}
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// isTUIMode returns true when there are 2+ tools and we're attached to a TTY.
func isTUIMode(numTools int) bool {
	return numTools >= 2 && isTerminalOutput()
}
