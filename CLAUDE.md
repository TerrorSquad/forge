# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`forge` is a single-binary, policy-driven git hook runner written in Go — a Node-free alternative to Husky/lint-staged/lefthook. Config lives in `forge.toml`; `forge install` sets git's `core.hooksPath` to `.forge/hooks`, whose shims shell out to `forge run <hook>`.

## Commands

```sh
make build          # -> dist/forge (CGO off, -s -w -trimpath)
make install        # build + copy to ~/.local/bin/forge
make test           # go test -race -coverprofile=coverage.out ./...
make coverage       # go tool cover -func on the last coverage.out
go build -o forge ./cmd/forge

go test ./internal/forge/runner/ -run TestName   # single test
```

Toolchain is pinned in `mise.toml` (go 1.23.8, node 24, pnpm 11). `node`/`pnpm` are only for `website/`, not the binary.

## Architecture

See [docs/architecture.md](docs/architecture.md) for the request-pipeline diagram and package map.

Entry: `cmd/forge/main.go` → `forge.Run(args)` in `internal/forge/app.go`, a flat switch that dispatches every subcommand (`init install uninstall doctor completion cache list ci validate run migrate update`). New subcommand = new case here.

Packages under `internal/forge/`:

- **config** — TOML load/parse/merge and starter presets. `Config` → `Hooks[name]` → `Tools[name]`. Global config (`~/.config/forge/`) is merged under the repo config. `migrate.go` converts legacy Husky `.git-hooks.config.json`; `remote_preset.go` fetches `https://` presets for `init --preset URL`.
- **runner** — the core. `RunHookWithOptions` loads config, resolves workspace member, filters staged files, then executes tools sequentially (`runner.go`) or concurrently (`runner_parallel.go`). Also holds commit-msg policy enforcement (conventional-commit + ticket footer) and the run cache (`cache.go`, keyed on file+tool hashes; `--no-cache`/`cache clear` bypass it).
- **backend** — execution abstraction behind the `Backend` interface: `HostBackend`, `DdevBackend`, `DockerBackend`. `ResolveBackend` picks per-tool `backend` → `[execution].default_backend` → auto-detect (ddev if `.ddev/config.yaml` + container running via `docker inspect`). `ResolveCommandForBackend` rewrites the command for the container path.
- **git** — repo-root detection and staged-file listing (`git diff --cached --name-only --diff-filter=ACMR`).
- **install** — writes hook shims + embedded JSON schema to `.forge/`, sets `core.hooksPath`, adds entries to `.git/info/exclude`. `supportedHooks` in `install.go` is the authoritative hook list (pre-commit, commit-msg, pre-push, prepare-commit-msg, post-commit, post-merge, post-rewrite).
- **ui** — colored terminal output (`ui.UI` writer, `ui.Green/Dim/...`); `ci.go` is the plain-output variant.
- **update** — self-update from GitHub releases (`forge update`, `--check`, `--version`, `--rollback`).

## Non-obvious things

- **Tool order matters and Go maps don't preserve it.** `config.parseHookToolOrder` regex-scans the raw TOML for `[hooks.X.tools.Y]` sections so tools run in declaration order. If you touch config loading, keep `OrderedToolNames()` fed by that, not by map iteration.
- **The binary configures its own hooks via `forge.toml`** at the repo root (gofmt + go vet). Editing Go here triggers those on commit.
- **`forge.schema.json`** is embedded via `//go:embed` (`schema_embed.go`) and written out on install for editor autocompletion. Regenerate/update the file in `internal/forge/schema/` if config fields change.
- **`forge ci`** ≡ `run pre-commit --all-files --check --no-cache` — the CI-friendly, non-mutating path (`--check` uses `check_args`, suppresses restage, treats any output as failure).
- Skips are env-driven: `SKIP_<TOOL>=1`, `SKIP_GROUP_<GROUP>=1`, `SKIP_PRECOMMIT/COMMITMSG/PREPUSH=1`, `HOOKS_ONLY=group,...`, `FORGE_CONFIG=path`.

## Release

Automated via release-please (`release-please-config.json`) + goreleaser (`.goreleaser.yaml`). Commits must be conventional-commit format (the repo enforces its own commit-msg policy). Version/Commit/Date are injected into `app.go` vars at build time via ldflags.
