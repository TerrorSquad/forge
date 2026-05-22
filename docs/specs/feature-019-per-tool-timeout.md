# Feature 019: Per-Tool Timeout

## Summary
Kill a tool that exceeds a configured duration, preventing hooks from hanging
indefinitely.

## Motivation
Network-dependent tools (e.g. type-checkers that fetch remote types) or
runaway processes can block commits/pushes forever. A per-tool timeout makes
the overall hook duration predictable.

## Functional Requirements

1. Timeout is configured per tool (duration string):
   ```toml
   [hooks.pre-commit.tools.phpstan]
   timeout = "120s"
   ```
2. Supported units: `s`, `m`, `ms` (standard Go duration strings).
3. When a tool exceeds its timeout, the process tree is killed (`SIGKILL`
   after `SIGTERM` with 2s grace period on Unix).
4. Exit message: `tool phpstan timed out after 120s`.
5. A timeout is treated as a tool failure — same `on_failure` logic applies.
6. Global default timeout:
   ```toml
   [execution]
   tool_timeout = "300s"    # default: no timeout (0 = unlimited)
   ```
7. Per-tool timeout overrides the global default.

## Non-Functional Requirements
- Uses `exec.CommandContext` with `context.WithTimeout`.
- Must work on all platforms (no POSIX-only signals).
- Zero / empty timeout = no limit (backward compatible).
