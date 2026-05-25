# Feature 003: Hook Installation and Auto-Execution

## Summary
Install managed git hook shims under `.forge/hooks` and configure local `core.hooksPath` so hooks run automatically during git operations.

## Motivation
Users expect a one-command setup where git automatically invokes the hook runner. Manual edits in `.git/hooks` are fragile and not team-friendly.

## Functional Requirements
1. `forge install` must:
   - create `.forge/hooks`
   - write executable shims for `pre-commit`, `commit-msg`, and `pre-push`
   - set local git config `core.hooksPath=.forge/hooks`
2. Shims must call `forge run <hook> "$@"`.
3. If binary is missing from PATH, shims must fail with an actionable error.
4. `forge uninstall` must:
   - remove `.forge/hooks`
   - unset `core.hooksPath` only when it points to `.forge/hooks`

## Security and Safety
- Only local git config should be modified.
- Never overwrite unrelated hook directories.

## UX Notes
- Install/uninstall commands should print exactly what was changed.

## Out of Scope
- Global hook installation.
- Shell-specific generated wrappers.
