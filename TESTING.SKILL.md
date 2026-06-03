# Forge Testing Skill

## Test strategy

1. Always run `go test ./internal/forge/...` for code changes.
2. Prefer unit tests for small logic changes.
3. Add regression coverage for config, hook behavior, and CLI options.

## Available test files

- `internal/forge/config/config_test.go` — config loading and ordering
- `internal/forge/runner/runner_test.go` — hook execution and filtering
- `internal/forge/runner/runner_parallel_test.go` — parallel execution behavior
- `internal/forge/git/git_test.go` — git environment detection
- `internal/forge/install_test.go` — install/uninstall workflows

## Useful checks

- `go test ./internal/forge/...`
- `go test ./...` when making broader repo-level changes
- `forge doctor` to verify repo-local installation state
- manual `git commit` and `git push` if hook install behavior changes

## Test expectations

- No regressions in hook order or tool filtering
- `forge install` should still create `.forge/hooks`
- `core.hooksPath` must be correctly detected and reported by doctor
- Changes should preserve existing `forge.toml` semantics unless documented otherwise
