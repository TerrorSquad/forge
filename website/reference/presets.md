# Presets

forge ships built-in presets for common project types. Use them with `forge init --preset NAME`.

## Available presets

| Preset | Tools configured |
|--------|-----------------|
| `go` | `gofmt`, `go vet`, commit-msg policy |
| `php` | `ecs` (PHP-CS-Fixer via ECS), commit-msg policy |
| `node` | `prettier`, `eslint`, commit-msg policy |
| `php-node` | `ecs`, `prettier`, `eslint`, commit-msg policy |
| `minimal` | commit-msg policy only |

## List presets

```sh
forge init --list-presets
```

## go

```sh
forge init --preset go
```

Generated `forge.toml` includes:

- `gofmt -w` on `.go` files with `restage = true`
- `go vet ./...` with `pass_files = false`
- Conventional Commits policy

## php

```sh
forge init --preset php
```

Includes:

- `vendor/bin/ecs check --fix` on `.php` files with `restage = true`
- Conventional Commits policy

## node

```sh
forge init --preset node
```

Includes:

- `prettier --write` on `.ts/.js/.vue/.json/.css/.scss` with `restage = true`
- `eslint --fix` on `.ts/.js/.vue` with `restage = true`
- Conventional Commits policy

## php-node

```sh
forge init --preset php-node
```

Combines `php` and `node` presets.

## minimal

```sh
forge init --preset minimal
```

Includes only the `commit-msg` hook with Conventional Commits policy.

## See also

- [Quick Start](/guide/quick-start)
- [Configuration](/guide/configuration)
