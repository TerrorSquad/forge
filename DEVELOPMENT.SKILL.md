# Forge Development Skill

## Goals

- implement small, incremental QoL improvements
- keep hooks and tool execution deterministic
- preserve user-visible config semantics
- avoid breaking existing workflows for installed repos

## Local workflow

1. Create a topic branch for the change.
2. Run `go test ./internal/forge/...` after each meaningful edit.
3. Use `forge doctor` while developing hook-related features.
4. Update config docs and website guides for any user-facing behavior changes.

## Recommended approach

- For config changes: keep `forge.toml` syntax simple and explicit.
- For hook/shim changes: ensure `.forge/hooks` remains the single source of hook scripts.
- For CLI changes: document new flags in `internal/forge/app.go` and website reference pages.
- For order/behavior changes: add regression tests in `internal/forge/config/config_test.go`, `runner/runner_test.go`, or `git/git_test.go` as appropriate.

## Common tasks

- Fixing hook behavior: modify `internal/forge/runner/*.go`
- Config parsing or ordering: modify `internal/forge/config/*.go`
- Installation/doctor behavior: modify `internal/forge/install.go` or `internal/forge/doctor.go`
- Website/docs updates: modify `README.md`, `website/guide/*.md`, or `website/reference/*.md`
