# Feature 041: Config Validation

Validate `forge.toml` strictly and surface syntax or unsupported field errors early.

## Summary

`forge` should reject invalid or unknown configuration fields during config load.
This prevents silent misconfiguration and keeps rules strict for repository hooks.

## Acceptance criteria

- `forge validate` fails when `forge.toml` contains unsupported tool options or invalid top-level keys.
- `forge doctor` reports config validation issues when the configuration file is present.
- `forge` uses strict TOML decoding for repository config.

## Examples

```toml
[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
unsupported_field = true
```

`forge validate` should report that `unsupported_field` is not recognized.
