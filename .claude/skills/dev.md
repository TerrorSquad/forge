# forge dev skill

You are helping develop **forge** — a policy-driven git hook runner written in Go.

## Project layout

| Path | What lives there |
|------|-----------------|
| `internal/forge/runner/` | Hook execution engine (serial + parallel), file filtering, cache |
| `internal/forge/config/` | TOML config loading, migration, presets |
| `internal/forge/git/` | Git helpers (staged files, stash, branch detection) |
| `internal/forge/backend/` | Execution backends: host, ddev, docker |
| `internal/forge/` | CLI (`app.go`), install, doctor, validate |
| `website/` | VitePress docs site (`guide/`, `reference/`) |
| `docs/specs/` | Feature spec documents |
| `forge.toml` | This repo's own hook config (dogfooding) |

## Build & test

```bash
mise exec -- go test ./internal/forge/...   # full test suite (use mise Go, not system Go)
mise exec -- go build -o dist/forge ./cmd/forge
make install                                 # build + cp to ~/.local/bin/forge
forge install                                # regenerate .forge/hooks shims
```

> Always use `mise exec -- go` — the system Go at `/opt/homebrew/bin/go` is a different version and causes compile errors.

## Key invariants

- **File paths from git are always `/`-separated** (`StagedFiles` calls `filepath.ToSlash`). Never compare staged paths using `filepath.Separator`.
- **`toolCacheKey` needs `repoRoot`** so relative file paths resolve correctly, especially in workspace mode where staged files are relative to the member root.
- **Parallel and serial paths must stay in sync** — any logic added to `runHookCfg` (serial) must be mirrored in `runHookCfgParallel` / `runToolWave`, and vice versa.
- **All user-facing output goes through `ui.UI`**, never `fmt.Printf` directly. This ensures CI detection and output redirection work correctly.
- **`exclude_patterns` / `include_patterns` support `**` globs** via `github.com/bmatcuk/doublestar/v4`. Single-level `*` still works as before.

## Where to add tests

| Change area | Test file |
|-------------|-----------|
| File filtering | `runner/filter_test.go` |
| Cache logic | `runner/cache_test.go` |
| Parallel execution | `runner/runner_parallel_test.go` |
| Workspace / monorepo | `runner/workspace_test.go` |
| Commit policy | `runner/policy_test.go` |
| Config parsing | `config/config_test.go` |
| General runner | `runner/runner_test.go` |

## Docs

User-facing behavior changes require updates to:
- `website/reference/config.md` — for new/changed config keys
- `website/guide/workspace.md` — for workspace-mode changes
- `website/guide/hooks.md` — for hook-level behavior changes

## PR / commit workflow

- Use conventional commit messages: `fix(...)`, `feat(...)`, `docs(...)`, `chore(...)`.
- Run `mise exec -- go test ./internal/forge/...` before pushing.
- Update `website/` docs for any user-facing behavior change.
- Open PRs with a clear title; note whether `.forge` shims need to be regenerated (`forge install`).

## Common gotchas

- The installed `~/.local/bin/forge` binary and the `.forge/hooks` shims can get out of sync after source changes. Run `make install && forge install` to rebuild and regenerate both.
- `applyToolFilter` is case-insensitive for tool names (lowercases before lookup). Tool names in `forge.toml` map keys are user-defined and may have any case.
- When both `--only-tools` and `--only-groups` are specified, explicitly named tools bypass the group filter (OR semantics for explicit names).
