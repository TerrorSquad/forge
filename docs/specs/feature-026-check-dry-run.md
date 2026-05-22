# Feature 026: --check Dry-Run Mode

## Summary
`booster run pre-commit --check` runs all tools in read-only mode: no file
mutations, no restage, exits non-zero if any tool would have reported
failures. Designed for CI gates.

## Motivation
CI pipelines need to verify hook compliance without side effects. Today there
is no way to run booster in CI without potentially modifying files or
restaging them.

## Functional Requirements

1. `--check` flag is accepted by `booster run`.
2. In check mode:
   - `restage` is suppressed for all tools.
   - Tools that mutate files (formatters like `gofmt -w`) should ideally be
     replaced by their check-only equivalent. booster achieves this via a
     per-tool `check_args` override:
     ```toml
     [hooks.pre-commit.tools.gofmt]
     args       = ["-w"]              # normal: write in place
     check_args = ["-l"]              # --check: list unformatted files
     check_fail_if_output = true      # non-empty stdout = failure
     ```
   - When `check_args` is not set, the tool runs with its normal `args`.
3. `check_fail_if_output = true` treats any stdout output as a failure
   (useful for tools that print files needing changes and exit 0).
4. A summary is printed at the end:
   ```
   Check complete: 3 passed · 1 would fail
   ```
5. Exit code mirrors whether all checks passed.

## Example CI Usage

```sh
booster run pre-commit --check --all-files
```

## Out of Scope
- Auto-fix mode (opposite of --check).
- Per-tool `--check` disable (just remove `check_args` if not needed).
