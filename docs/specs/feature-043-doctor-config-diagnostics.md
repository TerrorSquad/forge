# Feature 043: Doctor Config Diagnostics

Improve `forge doctor` so it reports configuration validation issues in addition to hook installation status.

## Summary

`forge doctor` should not only check installation state, but also validate the repo config and report any `forge.toml` issues.

## Acceptance criteria

- `forge doctor` prints config validation issues when the repo config is present.
- `forge doctor` still reports hook shim installation state and backend/tool status.
- Validation issues are shown in a clear and user-friendly format.

## Example

If `forge.toml` includes an unsupported key, `forge doctor` should output:

```
config: /path/to/forge.toml
configured hooks: pre-commit
config validation issues:
✗ [error] pre-commit.gofmt: unknown field "unsupported_field"
```
