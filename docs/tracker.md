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
| [010](specs/feature-010-backend-execution-routing.md) | Backend Execution Routing | 🔲 Not Started | P1 | DDEV auto-detect, per-tool backend override |
| [011](specs/feature-011-monorepo-workspace-mode.md) | Monorepo and Workspace Mode | 🔲 Not Started | P1 | `--workspace`, member discovery by staged path |
| [012](specs/feature-012-config-migration.md) | Config Migration Tool | 🔲 Not Started | P1 | `booster migrate` from `.git-hooks.config.json` → `booster.toml` |
| [013](specs/feature-013-presets-and-init.md) | Presets and `booster init` Improvements | 🔲 Not Started | P2 | Built-in presets: `node`, `php`, `php-node`, `go`, `minimal` |
| [014](specs/feature-014-release-distribution.md) | Release and Distribution | 🔲 Not Started | P1 | GoReleaser, Homebrew tap, curl installer |

---

## Quick Status Summary

| Milestone | Features | Done | Progress |
|-----------|----------|------|----------|
| v1 core   | 001–009  | 9/9  | ✅ Complete |
| v1.x      | 010–014  | 0/5  | 🔲 Not started |
| **Total** | **14**   | **9**| **64%** |

---

## Decisions Log

| Date       | Decision |
|------------|----------|
| 2026-05-22 | Chose Go over Rust for v1 — faster to ship, orchestration-heavy workload |
| 2026-05-22 | Use `core.hooksPath` (not `.git/hooks`) for clean uninstall and team portability |
| 2026-05-22 | Sequential tool execution in v1; parallel scheduler deferred to v2 |
| 2026-05-22 | Skip variables are runtime-only env vars; never persisted in config |
| 2026-05-22 | TOML-only config in v1; JSON/YAML support deferred |
