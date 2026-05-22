# Feature 010: Backend Execution Routing

## Summary
Route tool execution to the correct environment (host, DDEV, Docker, devcontainer)
based on per-tool or global configuration.

## Motivation
PHP mono-repos typically run their toolchain inside a DDEV container while JS
tools live on the host. A single hook runner must handle both without shell
gymnastics.

## Supported Backends (v1+)

| Backend        | Description                                             |
|----------------|---------------------------------------------------------|
| `host`         | Run binary directly on the host PATH (default)          |
| `ddev`         | `ddev exec <cmd>` — auto-detected from `.ddev/config.yaml` |
| `docker`       | `docker exec <container> <cmd>`                         |
| `devcontainer` | `devcontainer exec -- <cmd>`                            |

## Functional Requirements

### v1 (host only, DDEV detection)
1. Default backend is `host`.
2. When `backend = "ddev"` (or DDEV is auto-detected), wrap command with
   `ddev exec --`.
3. DDEV auto-detection: look for `.ddev/config.yaml` in repo root; check that
   `ddev status` exits 0.
4. Binary availability check must use the resolved backend (e.g. for DDEV,
   check `vendor/bin/<tool>` exists inside the project, not on host).

### Global override
```toml
[execution]
default_backend = "ddev"
```

### Per-tool override
```toml
[hooks.pre-commit.tools.phpstan]
backend = "ddev"
```

## Out of Scope (v1)
- Docker and devcontainer backends.
- Automatic container startup.
