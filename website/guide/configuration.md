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

## Tool fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `command` | string | — | Binary to run |
| `args` | `[]string` | `[]` | Arguments (file paths appended unless `pass_files = false`) |
| `type` | string | `system` | `system`, `node`, or `php` — affects binary resolution |
| `backend` | string | global default | `host` or `ddev` — overrides global `[execution] default_backend` |
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
default_backend = "ddev"   # "host" (default) or "ddev"
```

## Config file path

| Priority | Source |
|----------|--------|
| 1 | `FORGE_CONFIG` env var |
| 2 | `forge.toml` in repo root |

## See also

- [Hooks](/guide/hooks)
- [Backends (DDEV)](/guide/backends)
- [forge.toml reference](/reference/config)
