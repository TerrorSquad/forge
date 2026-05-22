# Feature 017: Colored Output and Per-Tool Timing

## Summary
Emit ANSI-colored, timed output for each tool so slow tools are immediately
visible and success/failure is clear at a glance.

## Motivation
Current output is plain text. Teams can't tell at a glance whether a hook
passed, failed, or which tool was slow.

## Output Format

```
pre-commit
  ✓  gofmt        12ms
  ✓  govet        340ms
  ✗  golangci-lint 1.2s   exit status 1
    [tool stderr here]
```

- Green `✓` for pass, red `✗` for fail, yellow `~` for skipped.
- Time is shown to the right, auto-scaled (ms / s).
- Tool stderr is indented under the tool line.
- Summary line at the end: `2 passed · 1 failed (1.56s total)`.

## Functional Requirements

1. Colors are disabled automatically when stdout is not a TTY (`os.Stdout`
   is not a terminal) or `NO_COLOR` env is set.
2. `FORCE_COLOR=1` enables colors even when piped.
3. Timing is always printed regardless of color mode.
4. In parallel mode (feature-015), tool output is buffered and flushed
   atomically — the colored block prints in full when the tool exits.
5. No external dependencies — ANSI escape codes only.

## Non-Functional Requirements
- Must work on macOS, Linux, and Windows (Windows: disable colors unless
  `TERM` or `FORCE_COLOR` is set, or VT processing is available).
- Must not break existing `booster run` output parsing in scripts that
  grep for tool names.
