# Feature 032: Per-Branch Hook Profiles

## Summary
Allow different tool sets (or tool configs) to apply depending on the current
git branch, so protected branches (main, release/*) can enforce stricter checks
while feature branches remain fast.

## Motivation
Teams want strict enforcement on `main`/`develop` (full static analysis, all
tests, no `on_failure=continue`) but lightweight hooks on daily feature
branches. Today this requires maintaining separate `booster.toml` files or
overriding per developer. Branch profiles make this a first-class config
concept.

## Config Design

Profiles are defined in `[profiles.<name>]` and matched to branches via
`match` (glob or regex).

```toml
# Default profile applies when no profile matches (existing behaviour).

[profiles.protected]
match = ["main", "master", "develop", "develop/*", "release/*"]

# Override specific tools for this profile.
[profiles.protected.tools.phpstan]
args = ["analyse", "--level=8"]   # stricter level on protected branches

[profiles.protected.tools.tests]
on_failure = "stop"               # was "continue" on feature branches

# Add a tool that only runs on protected branches.
[profiles.protected.tools.infection]
command    = "composer"
args       = ["mutation:changed"]
type       = "php"
pass_files = false
group      = "test"

[profiles.wip]
match       = ["wip/*", "wip-*"]
skip_hooks  = ["pre-push"]       # skip entire hooks for WIP branches
```

## Functional Requirements

1. Profiles are evaluated at hook runtime. The first `match` that matches the
   current branch wins. No match = default profile.
2. `match` supports:
   - Exact strings: `"main"`
   - Glob patterns: `"release/*"`, `"feature-*"`
   - Regex (prefix `r:`): `"r:^(hotfix|release)/"`
3. Profile `tools` section is **merged** with the base config, not replaced.
   Only specified keys are overridden.
4. Profile `skip_hooks` lists hook names whose entire tool list is bypassed.
5. Profile `skip_tools` lists individual tool names to exclude (additive to
   `skip_hooks`).
6. `booster status` shows the active profile and which branches it matches.
7. Profiles have no effect on `booster run` invoked directly (only git
   hook triggers respect them), unless `--profile <name>` is passed.

## Example

```
pre-commit (profile: protected — matches "main")
  ✓  ecs         340ms
  ✓  phpstan     2.1s   (level=8, overridden by profile)
  ✓  psalm       1.8s
  ✓  infection   42.0s  (only on protected)

4 passed · total 46.3s
```

## Out of Scope
- Remote-branch detection (profiles only apply to the checked-out local branch).
- Profile inheritance / `extends`.
