# Feature 013: Presets and `booster init` Improvements

## Summary
Extend `booster init` with built-in presets that generate sensible starter
configs for common project types.

## Motivation
An empty config is more friction than a well-chosen default. Presets let teams
be productive in under a minute without reading documentation.

## Available Presets

| Preset         | Contents                                               |
|----------------|--------------------------------------------------------|
| `node`         | Prettier, ESLint, commitlint-style commit-msg policy   |
| `php`          | PHP Syntax Check, ECS, Rector, PHPStan, Psalm, Deptrac |
| `php-node`     | All from both `php` and `node` presets                 |
| `minimal`      | Prettier + conventional commit policy only             |
| `go`           | gofmt, golangci-lint, conventional commit policy       |

## CLI
```text
booster init [--preset <name>] [--force]
booster init --list-presets
```

## Functional Requirements
1. Without `--preset`, generate the existing default (Prettier + ESLint + commit-msg policy).
2. `--list-presets` prints the available preset names and a one-line description.
3. Generated config must include comments explaining each section.
4. `--force` overwrites an existing `booster.toml`.

## Out of Scope
- Remote/custom preset URLs.
- Interactive wizard mode.
