# Feature 018: --all-files Flag

## Summary
Add `booster run pre-commit --all-files` to run configured tools against all
tracked files rather than just staged ones.

## Motivation
- Initial project setup: run all formatters/linters on the whole codebase.
- CI gate: verify the entire repo without needing a staged set.
- After adding a new tool: apply it retroactively.

## Functional Requirements

1. `booster run pre-commit --all-files` collects all files tracked by git
   (`git ls-files`) instead of `git diff --cached --name-only`.
2. All existing file-filtering logic (extensions, include/exclude patterns,
   run_per_file) applies identically.
3. `restage` is **disabled** automatically in `--all-files` mode — no
   side effects on the working tree index.
4. A warning is printed when `--all-files` is used with `restage = true`
   tools, explaining that restage was suppressed.
5. The flag is only valid for `pre-commit`; other hooks reject it with a
   clear error.

## Example

```sh
booster run pre-commit --all-files
booster run pre-commit --all-files --skip-restage   # explicit (same as default)
```

## Non-Functional Requirements
- `git ls-files` output must be scoped to the repo root (relative paths).
- Must respect `.gitignore` (git ls-files already does this).
