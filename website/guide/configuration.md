# Configuration

forge is configured via `forge.toml` at the repo root (or at `FORGE_CONFIG` if set).

## Minimal example

```toml
[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command    = "gofmt"
args       = ["-w"]
type       = "system"
extensions = [".go"]
restage    = true

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
```

## Tool order

Tool sections are executed in the order they are declared in `forge.toml`. There is no separate `order` field — the declaration order of `[hooks.<hook>.tools.<name>]` blocks determines execution order.

## Tool fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `command` | string | — | Binary to run |
| `args` | `[]string` | `[]` | Arguments (file paths appended unless `pass_files = false`) |
| `type` | string | `system` | `system`, `node`, or `php` — affects binary resolution |
| `backend` | string | global default | `host`, `ddev`, or a Docker container name — overrides global `[execution] default_backend` |
| `extensions` | `[]string` | all | Only run on files with these extensions |
| `include_patterns` | `[]string` | all | Glob allowlist |
| `exclude_patterns` | `[]string` | none | Glob blocklist |
| `pass_files` | bool | `true` | Append staged file paths as args |
| `run_per_file` | bool | `false` | Invoke the tool once per file instead of batch |
| `restage` | bool | `false` | Run `git add` on files after the tool runs |
| `on_failure` | string | — | Set to `stop` to abort remaining tools on failure |
| `group` | string | — | Group name; used with `HOOKS_ONLY` env var |

## Global execution settings

```toml
[execution]
default_backend = "ddev"   # "host" (default), "ddev", or a Docker container name
parallel        = true     # run each hook's tools concurrently
cache           = true     # skip tools whose inputs haven't changed
tool_timeout    = "60s"    # default timeout per tool
```

See the [full reference](/reference/config#execution) for caching and timeout options.

## Config file path

forge loads the first repo config it finds:

| Priority | Source |
|----------|--------|
| 1 | `FORGE_CONFIG` env var (path relative to repo root, or absolute) |
| 2 | `forge.toml` in repo root |

## Global user config

A user-level config at `~/.config/forge/config.toml` (override with `FORGE_GLOBAL_CONFIG`, respects `XDG_CONFIG_HOME`) is merged **underneath** the repo config — the repo's values always win. Use it for personal defaults like `[execution] default_backend` or a shared commit-message policy across all your repos.

## See also

- [Hooks](/guide/hooks)
- [Backends (DDEV)](/guide/backends)
- [forge.toml reference](/reference/config)
