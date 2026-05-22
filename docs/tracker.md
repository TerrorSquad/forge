# Feature Tracker

Track implementation progress for all booster features.

**Legend:**  `✅ Done` · `🔄 In Progress` · `🔲 Not Started` · `⏸ Deferred`

---

## Core Features (v1)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| [001](specs/feature-001-cli-command-surface.md) | CLI Command Surface | ✅ Done | P0 | `init`, `install`, `uninstall`, `run`, `doctor` implemented |
| [002](specs/feature-002-config-and-policy-model.md) | Config and Policy Model | ✅ Done | P0 | TOML loader, tool config, commit-msg policy implemented |
| [003](specs/feature-003-hook-installation-and-auto-execution.md) | Hook Installation and Auto-Execution | ✅ Done | P0 | Shims written to `.booster/hooks`, `core.hooksPath` set via git config |
| [004](specs/feature-004-pre-commit-runner-engine.md) | Pre-Commit Runner Engine | ✅ Done | P0 | Sequential tool execution, staged-file filtering, restage, on_failure |
| [005](specs/feature-005-commit-message-policy.md) | Commit Message Policy | ✅ Done | P0 | Conventional commit validation, ticket footer append/require |
| [006](specs/feature-006-doctor-observability.md) | Doctor and Observability | ✅ Done | P0 | Binary path, repo root, config, hooksPath, missing binaries reported |
| [007](specs/feature-007-toolchain-with-mise.md) | Toolchain Management with mise | ✅ Done | P0 | `mise.toml` with pinned `go = "1.23.8"` |
| [008](specs/feature-008-env-skip-controls.md) | Environment Variable Skip Controls | ✅ Done | P0 | `SKIP_*`, `HOOKS_ONLY` implemented in runner |
| [009](specs/feature-009-file-filtering-pipeline.md) | File Filtering and Staged-File Pipeline | ✅ Done | P0 | Extension, include/exclude pattern, per-file mode, restage |

## Near-Term Features (v1.x)

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| [010](specs/feature-010-backend-execution-routing.md) | Backend Execution Routing | ✅ Done | P1 | `HostBackend`/`DdevBackend` interfaces, DDEV auto-detect via `ddev status`, per-tool `backend` field, global `[execution] default_backend` |
| [011](specs/feature-011-monorepo-workspace-mode.md) | Monorepo and Workspace Mode | ✅ Done | P1 | `[workspace] members = [...]`, glob expansion, staged-path matching, per-member config fallback |
| [012](specs/feature-012-config-migration.md) | Config Migration Tool | ✅ Done | P1 | `booster migrate [--from FILE] [--to FILE]` reads `.git-hooks.config.json`, emits booster.toml TOML |
| [013](specs/feature-013-presets-and-init.md) | Presets and `booster init` Improvements | ✅ Done | P2 | `--preset` flag + `--list-presets`; built-in presets: `node`, `php`, `php-node`, `go`, `minimal` |
| [014](specs/feature-014-release-distribution.md) | Release and Distribution | ✅ Done | P1 | `.goreleaser.yaml` (multi-platform), CI workflow, release workflow, `install.sh`, `booster version` |

## v2 Features

| # | Feature | Status | Priority | Notes |
|---|---------|--------|----------|-------|
| [015](specs/feature-015-parallel-execution.md) | Parallel Execution | ✅ Done | P2 | `[execution] parallel = true`; `depends_on` DAG; wave-based goroutine execution; buffered per-tool output |
| [016](specs/feature-016-pre-push-runner.md) | Pre-Push Runner | ✅ Done | P1 | Push context parsed from stdin; `BOOSTER_PUSH_REMOTE`/`URL` env vars injected |
| [017](specs/feature-017-colored-output-timing.md) | Colored Output and Timing | ✅ Done | P2 | ANSI color, per-tool duration, summary line |
| [018](specs/feature-018-all-files-flag.md) | `--all-files` Flag | ✅ Done | P2 | `booster run pre-commit --all-files`; uses `git ls-files`; suppresses restage |
| [019](specs/feature-019-per-tool-timeout.md) | Per-Tool Timeout | ✅ Done | P2 | `timeout = "120s"` per tool; `[execution] tool_timeout` global default; context cancellation |
| [020](specs/feature-020-shell-completions.md) | Shell Completions | ✅ Done | P3 | `booster completion bash\|zsh\|fish`; subcommands, flags, hooks, presets |
| [021](specs/feature-021-prepare-commit-msg.md) | `prepare-commit-msg` Hook | ✅ Done | P2 | Ticket prefix from branch; `skip_on_merge`, `skip_if_present` |
| [022](specs/feature-022-doctor-fix.md) | `doctor --fix` | ✅ Done | P2 | Auto-reinstalls shims, sets `core.hooksPath`; `--dry-run` support |
| [023](specs/feature-023-global-user-config.md) | Global User Config | ✅ Done | P2 | `~/.config/booster/config.toml` (XDG); merge semantics — repo wins |
| [024](specs/feature-024-run-cache.md) | Run Cache | ✅ Done | P2 | SHA-256 content-addressed cache in `.booster/cache.json`; per-tool or global `cache = true`; `booster cache clear`; `--no-cache` flag |
| [025](specs/feature-025-remote-presets.md) | Remote Presets | ✅ Done | P3 | `--preset https://...`; HTTPS-only; TOML validation; `--yes`/`CI=true` skip prompt |
| [026](specs/feature-026-check-dry-run.md) | `--check` Dry-Run Mode | ✅ Done | P3 | `booster run pre-commit --check`; `check_args`/`check_fail_if_output` per tool; `would-fail` status; check-mode summary |
| [027](specs/feature-027-post-commit.md) | `post-commit` Hook | ✅ Done | P3 | Informational banner; failures surfaced but commit not aborted; `SKIP_POSTCOMMIT=1` |
| [028](specs/feature-028-tui-progress.md) | TUI Progress | ⏸ Deferred | P4 | Bubble Tea live progress bars; deferred |

---

## Quick Status Summary

| Milestone | Features | Done | Progress |
|-----------|----------|------|----------|
| v1 core   | 001–009  | 9/9  | ✅ Complete |
| v1.x      | 010–014  | 5/5  | ✅ Complete |
| v2        | 015–028  | 13/14 | 🔄 93% |
| **Total** | **28**   | **27** | **96%** |

---

## Decisions Log

| Date       | Decision |
|------------|----------|
| 2026-05-22 | Chose Go over Rust for v1 — faster to ship, orchestration-heavy workload |
| 2026-05-22 | Use `core.hooksPath` (not `.git/hooks`) for clean uninstall and team portability |
| 2026-05-22 | Sequential tool execution in v1; parallel scheduler deferred to v2 |
| 2026-05-22 | Skip variables are runtime-only env vars; never persisted in config |
| 2026-05-22 | TOML-only config in v1; JSON/YAML support deferred |
| 2026-05-22 | Run cache is content-addressed (SHA-256 of tool config + file hashes); atomic writes via temp+rename |
| 2026-05-22 | Global user config merges at execution scalars and per-hook tool maps only; workspace config excluded from merge |
| 2026-05-22 | Remote presets require `https://` only; content validated as TOML before writing; CI env or `--yes` skips prompt |
| 2026-05-22 | Feature 029 (booster publish) removed from scope — not needed |
