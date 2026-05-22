# Feature 038: Staged File Safety (Pre-Commit Stash)

## Summary
Before running `--fix` tools in pre-commit, snapshot the working tree (unstaged
changes), and restore it after — so fixers never accidentally include work-in-
progress changes in the commit.

## Motivation
Tools like `ecs --fix`, `rector`, and `prettier --write` modify files in place.
If a developer has unstaged edits to the same file, the fixer sees the full
file (staged + unstaged), applies fixes, and `git add` re-stages _everything_
including the developer's uncommitted work. This is the classic "oops, I
committed half-finished code" problem.

The standard mitigation is to `git stash -u --keep-index` before running
fixers, then `git stash pop` after. Booster should do this automatically.

## Visual Design

```
pre-commit
  ⬇  stashing unstaged changes...
  ✓  rector              1.2s
  ✓  ecs                 340ms  (restaged 3 files)
  ✓  phpstan             4.2s
  ⬆  restoring unstaged changes...

3 passed · total 5.8s
```

## Functional Requirements

1. Stash is created only when **any** configured tool has `restage = true`
   (i.e. a fixer is present). Pure analysis tools (no restage) don't need it.
2. Stash strategy: `git stash push --keep-index --include-untracked -m
   "booster-pre-commit-safety"`. Only unstaged tracked and untracked files
   are stashed; staged index is left intact.
3. After all tools complete (pass or fail), the stash is always popped. Pop
   failure (e.g. merge conflict due to fixer changing a line the unstaged
   edit also changed) is reported clearly:
   ```
   ⚠ Could not restore stashed changes automatically.
     Run `git stash pop` to restore, then resolve conflicts.
     Stash ref: stash@{0}
   ```
4. If no stash entry is created (nothing to stash), skip silently.
5. `BOOSTER_NO_STASH=1` disables stash behaviour globally.
6. Can be disabled per-hook:
   ```toml
   [hooks.pre-commit]
   safe_stash = false
   ```

## Edge Cases

| Situation | Behaviour |
|-----------|-----------|
| No unstaged changes | Skip stash/pop silently |
| Fixer creates new untracked files | Those files are staged by `restage`; pop restores pre-commit untracked state |
| Stash pop conflict | Warn prominently; leave stash intact for manual resolution |
| Hook fails mid-run | Pop is still attempted (deferred cleanup) |
| Submodules | Exclude from stash (pass `--no-include-untracked` for submodule dirs) |

## Non-Functional Requirements
- Stash/pop adds < 200ms overhead in typical repos (< 500 files).
- Stash is always labelled `booster-pre-commit-safety` for easy identification.
- If a stash already exists with that label from a crashed previous run,
  booster warns and creates a new numbered entry rather than overwriting.

## Out of Scope
- Stash during pre-push (no fixers run in pre-push).
- Partial-file staging (git's `add -p` equivalent); booster operates at whole-
  file granularity.
