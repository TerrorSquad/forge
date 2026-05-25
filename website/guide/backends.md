# Backends (DDEV)

forge can execute tools either on the host machine or inside a DDEV container.

## How it works

When `backend = "ddev"` is set, forge:

1. Reads `.ddev/config.yaml` to find the project `name`.
2. Constructs the container name as `ddev-<name>-web`.
3. Checks that the container is running via `docker inspect`.
4. Runs the tool with `docker exec -i -w /var/www/html ddev-<name>-web <command>`.

Environment variables (e.g., `PATH` expansions) are forwarded explicitly via `-e` flags.

## Auto-detection

forge automatically uses the DDEV backend when:

- `.ddev/config.yaml` exists in the repo root, **and**
- the container is running.

If the container is not running, forge falls back to the host backend and emits a warning.

## Configuration

### Global default

```toml
[execution]
default_backend = "ddev"
```

### Per-tool override

```toml
[hooks.pre-commit.tools.phpcs]
command = "vendor/bin/phpcs"
backend = "ddev"

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
backend = "host"   # always run on host even if default is ddev
```

## Supported values

| Value | Behaviour |
|-------|-----------|
| `host` | Run the command directly on the host machine (default) |
| `ddev` | Run via `docker exec` inside the DDEV web container |

## Troubleshooting

Run `forge doctor` to see which backend is resolved for each tool.

If DDEV is not running, you will see:

```
⚠  backend: ddev container not running — falling back to host
```

Start the container with `ddev start` and re-run.

## See also

- [Configuration](/guide/configuration)
- [forge doctor](/reference/cli#forge-doctor)
