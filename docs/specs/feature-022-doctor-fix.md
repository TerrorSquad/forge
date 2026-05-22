# Feature 022: doctor --fix

## Summary
Extend `booster doctor` with a `--fix` flag that automatically resolves
detected issues instead of just reporting them.

## Motivation
`booster doctor` is diagnostic-only today. Common issues (missing shims, wrong
`core.hooksPath`) are trivially fixable and should not require manual steps.

## Functional Requirements

1. `booster doctor --fix` runs all checks and, for each fixable issue,
   applies the fix automatically.
2. Fixable issues and their remedies:

   | Issue | Fix |
   |-------|-----|
   | Hook shims missing / outdated | Re-run `InstallHooks()` |
   | `core.hooksPath` not set or wrong | Reset via `git config` |
   | `booster.toml` missing | Prompt to run `booster init` (interactive) or skip in CI |

3. Non-fixable issues (missing tool binaries, no git repo) are still
   reported with a clear message and no action taken.
4. Each fix is printed before it is applied:
   ```
   ✗  hook pre-commit: missing  →  installing...  ✓
   ```
5. `--fix` is idempotent: running it twice produces no changes on the second
   run.
6. `--dry-run` combined with `--fix` prints what would be fixed without
   applying changes.

## Non-Functional Requirements
- Must not silently overwrite a customised `core.hooksPath` pointing to a
  different directory. Warn and skip if the current value is neither
  `.booster/hooks` nor unset.
