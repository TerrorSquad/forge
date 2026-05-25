# Feature 030: Interactive Failure Recovery

## Summary
When a tool fails in an interactive terminal session, present an inline menu
offering "retry", "skip this tool", "skip remaining tools", or "abort" — so
developers can recover without re-staging files and re-running the hook
manually.

## Motivation
Today, a single tool failure aborts the whole hook. The developer must fix the
issue, re-stage, and re-run. But often the failure is transient (flaky test,
network blip) or intentionally skippable (WIP commit). An interactive recovery
prompt turns a hard stop into a guided decision.

## Visual Design

```
pre-commit
  ✓  ecs              340ms
  ✗  phpstan          4.2s
     src/Foo/Handler/FooHandler.php:42
     Parameter #1 expects string, int given.

  What would you like to do?
  ❯  retry             Re-run phpstan
     skip-tool         Skip phpstan and continue
     skip-remaining    Skip all remaining tools
     abort             Abort the hook (exit 1)
```

Arrow-key navigation, Enter to confirm.

## Functional Requirements

1. Interactive mode activates only when **all** of the following are true:
   - `stdout` is a TTY.
   - `FORGE_NO_INTERACTIVE=1` is NOT set.
   - The tool's `on_failure` is not `"continue"` or `"stop"`.
2. `on_failure = "stop"` still hard-stops without prompting (respects explicit
   config intent).
3. `on_failure = "continue"` still silently continues without prompting.
4. "retry" re-executes the exact same tool with the same file list.
5. "skip-tool" marks the tool as skipped, continues to remaining tools.
6. "skip-remaining" marks all subsequent tools as skipped, exits hook with 0.
7. "abort" exits the hook with exit code 1.
8. In non-interactive mode (CI, pipes), defaults to "abort" behaviour —
   identical to current behaviour.
9. Timeout: if no input within 30 seconds, defaults to "abort".

## Config

```toml
[execution]
interactive = true   # default: true (auto-detected from TTY)
```

Global disable:
```
FORGE_NO_INTERACTIVE=1 git commit -m "..."
```

## Dependencies
- `golang.org/x/term` — TTY detection and raw mode input
- OR `github.com/charmbracelet/bubbletea` if TUI (feature-028) is co-implemented

## Out of Scope
- Editing the commit message inline.
- Selecting which files to exclude from a retry.
