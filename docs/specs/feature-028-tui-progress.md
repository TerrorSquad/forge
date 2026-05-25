# Feature 028: TUI Progress View

## Summary
Display a real-time, interactive terminal progress view during hook execution
using the [bubbletea](https://github.com/charmbracelet/bubbletea) framework.

## Motivation
Parallel execution (feature-015) makes plain sequential output misleading.
A TUI view shows each tool as a live spinner that resolves to ✓ or ✗ when
done, giving instant feedback on which tools are still running.

## Visual Design

```
pre-commit
  ⠸  gofmt          running...
  ✓  govet           340ms
  ✗  golangci-lint   1.2s
     [exit status 1]
     ./pkg/foo.go:12:3: error message

2 passed · 1 failed · total 1.2s
```

- Each tool row updates in place.
- Spinner animates while the tool is running.
- Completed rows replace spinner with ✓ (green) or ✗ (red).
- Failed tool output is shown inline below its row.
- Summary line appears after all tools finish.

## Functional Requirements

1. TUI mode is activated automatically when stdout is a TTY.
2. When stdout is NOT a TTY (CI, pipes), plain text output is used (feature-017).
3. `FORGE_NO_TUI=1` forces plain text mode.
4. TUI is disabled in sequential mode if only one tool is configured
   (no value in spinning for a single tool — just print it plainly).
5. TUI handles terminal resize gracefully.

## Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — styling (optional, for color theming)

## Out of Scope
- Mouse interaction.
- Scrollback beyond the current hook run.
- Persistent history view.
