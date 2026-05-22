# Feature 001: CLI Command Surface

## Summary
Provide a single executable (`booster`) with predictable subcommands for repository setup, hook management, and runtime execution.

## Motivation
Teams should not need shell-script glue for every repository. A consistent command surface reduces onboarding friction and enables automation via setup scripts.

## Scope
- `booster init [--force]`
- `booster install`
- `booster uninstall`
- `booster run <hook> [--edit FILE]`
- `booster doctor`

## Functional Requirements
1. CLI returns non-zero on invalid command usage.
2. `help` output must be concise and include examples.
3. `run` command must support hook names passed by git (`pre-commit`, `commit-msg`, `pre-push`).
4. `run commit-msg` must accept the commit message file via flag and positional fallback for git compatibility.

## Non-Functional Requirements
- Fast startup, no network dependency.
- Works in non-interactive CI contexts.
- Clear stderr messaging for errors.

## UX Notes
- Messages should clearly indicate which config file and hook are active.
- Unknown commands should print short help automatically.

## Out of Scope
- Plugin system.
- Background daemon mode.
