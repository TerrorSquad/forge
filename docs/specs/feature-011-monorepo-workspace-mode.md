# Feature 011: Monorepo and Workspace Mode

## Summary
Support repositories that contain multiple sub-projects with separate
`booster.toml` files, running only the relevant project hooks based on which
files were staged.

## Motivation
A monorepo with `apps/api`, `apps/frontend`, and `packages/shared` should not
run PHP analysis for a pure JS commit, and vice-versa.

## Functional Requirements

### Discovery
1. Root `booster.toml` may declare workspace members:
```toml
[workspace]
members = ["apps/api", "apps/frontend", "packages/shared"]
```
2. `booster install --workspace` writes shims that call
   `booster run <hook> --workspace`.
3. When `--workspace` is passed, booster:
   a. Collects staged files.
   b. For each member, checks if any staged file is under that member's path.
   c. If yes, runs that member's `booster.toml` in the context of that sub-path.

### Fallback
- If no staged file matches any member, fall back to root config.

### CLI
```text
booster run pre-commit --workspace
booster run pre-commit --project apps/api
```

## Out of Scope (v1)
- Parallel member execution.
- Dependency ordering between members.
