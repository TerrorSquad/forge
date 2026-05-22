# Feature 023: Global User Config

## Summary
Support a user-level config at `~/.config/booster/config.toml` that provides
personal defaults, merged under the repo config.

## Motivation
Some preferences (e.g. always disable ticket-footer in personal projects,
prefer a specific backend) shouldn't be committed. A user config lets
developers set them once globally without touching repo configs.

## Merge Semantics

1. Load order: global user config → repo `booster.toml` (repo wins on
   conflict).
2. Merging is shallow at the `[execution]` and `[hooks.<name>.policy]` level:
   - Scalar fields: repo value overrides global.
   - Tool maps: repo tools are merged with global tools; repo-defined tool
     takes precedence if the same name appears in both.
3. `[workspace]` and per-tool settings from global config are NOT merged
   into repo config (too project-specific).

## Config Location

| Platform | Path |
|----------|------|
| Linux/macOS | `$XDG_CONFIG_HOME/booster/config.toml` or `~/.config/booster/config.toml` |
| Windows | `%APPDATA%\booster\config.toml` |

Override via `BOOSTER_GLOBAL_CONFIG` env var.

## Example Global Config

```toml
[execution]
default_backend = "host"
tool_timeout    = "60s"

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = false    # personal preference
```

## Functional Requirements

1. Global config is silently ignored if it does not exist.
2. `booster doctor` reports the global config path and whether it is present.
3. Invalid global config produces a warning but does not fail the hook.

## Out of Scope
- Team-shared config outside the repo.
- Config inheritance chains deeper than two levels.
