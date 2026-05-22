# Feature 004: Pre-Commit Runner Engine

## Summary
Execute configured tools for the pre-commit hook using staged-file awareness and deterministic tool order.

## Motivation
Hook runners must avoid running expensive checks on unrelated files while still preserving formatting and fail-fast policy behavior.

## Functional Requirements
1. For `pre-commit`, collect staged files from:
   - `git diff --cached --name-only --diff-filter=ACMR`
2. For each tool, filter files by:
   - `extensions`
   - `include_patterns`
   - `exclude_patterns`
3. Support execution modes:
   - once with all files
   - per-file (`run_per_file`)
   - without file args (`pass_files = false`)
4. Respect `on_failure` behavior:
   - `stop`: abort remaining tools
   - default/other: continue and fail at end
5. Restage modified files when `restage = true`.

## Determinism
- Tool map order must be normalized before execution.
- Runner output should state tool status (`running`, `skipped`, `failed`).

## Environment Controls
- Hook-level skip variables (`SKIP_PRECOMMIT`, etc.)
- Tool-level skip variables (`SKIP_<TOOL_NAME>`)
- Group filtering via `HOOKS_ONLY`

## Out of Scope
- Parallel execution scheduler.
- Native toolchain installation.
