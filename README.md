# booster

A policy-driven git hook runner — fast, portable, no Node.js required.

## Why booster?

Most hook runners (Husky, lint-staged) require Node.js and `package.json`. booster is a single Go binary that works in any project — PHP, Go, Python, or mixed monorepos.

| Feature | booster | Husky | lint-staged |
|---------|---------|-------|-------------|
| Single binary | ✅ | ❌ (needs Node) | ❌ (needs Node) |
| TOML config | ✅ | ❌ | ❌ |
| DDEV backend | ✅ | ❌ | ❌ |
| Monorepo workspace mode | ✅ | ❌ | ✅ |
| Commit-msg policy | ✅ | ❌ | ❌ |
| Staged-file filtering | ✅ | ❌ | ✅ |
| Migration from Husky | ✅ | — | — |

---

## Installation

### curl installer (Linux / macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/TerrorSquad/gobooster/main/install.sh | sh
```

### go install

```sh
go install github.com/TerrorSquad/gobooster/cmd/booster@latest
```

### Manual download

Download a pre-built binary from [Releases](https://github.com/TerrorSquad/gobooster/releases) and place it on your `PATH`.

---

## Quick start

```sh
# 1. Create booster.toml
booster init --preset go       # or: node, php, php-node, minimal

# 2. Install git hook shims
booster install

# 3. Commit — hooks fire automatically
git commit -m "feat: my change"
```

---

## Commands

| Command | Description |
|---------|-------------|
| `booster init [--preset NAME] [--force]` | Create `booster.toml` from a built-in preset |
| `booster init --list-presets` | List available presets |
| `booster install` | Write hook shims to `.booster/hooks`, set `core.hooksPath` |
| `booster uninstall` | Remove shims and restore default `core.hooksPath` |
| `booster run <hook>` | Run a hook manually (e.g. `booster run pre-commit`) |
| `booster migrate [--from FILE] [--to FILE]` | Convert `.git-hooks.config.json` → `booster.toml` |
| `booster doctor` | Diagnose binary, config, hooks, and tool availability |
| `booster version` | Print version, commit, and build date |

---

## Configuration

`booster.toml` lives at the repo root. Copy `booster.toml.example` to get started.

### Minimal example

```toml
[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.gofmt]
command = "gofmt"
args    = ["-w"]
type    = "system"
extensions = [".go"]
restage = true

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
```

### Tool fields

| Field | Type | Description |
|-------|------|-------------|
| `command` | string | Binary to run |
| `args` | []string | Arguments (files appended unless `pass_files = false`) |
| `type` | string | `system`, `node`, `php` — affects binary resolution |
| `backend` | string | `host` or `ddev` (overrides global default) |
| `extensions` | []string | Only run on files with these extensions |
| `include_patterns` | []string | Glob allowlist |
| `exclude_patterns` | []string | Glob blocklist |
| `pass_files` | bool | Pass staged file paths as args (default `true`) |
| `run_per_file` | bool | Invoke tool once per file instead of batch |
| `restage` | bool | `git add` modified files after tool runs |
| `on_failure` | string | `stop` to abort remaining tools on failure |
| `group` | string | Tool group name (used with `HOOKS_ONLY`) |

### Commit-message policy fields

```toml
[hooks.commit-msg.policy]
conventional_commits = true    # enforce feat/fix/chore/... prefix
append_ticket_footer = true    # append "Closes: PRJ-123" from branch name
require_ticket       = false   # fail if branch has no ticket
```

### DDEV backend

```toml
[execution]
default_backend = "ddev"   # route all tools through `ddev exec`
```

Or per tool: `backend = "ddev"`. Auto-detected when `.ddev/config.yaml` exists and `ddev status` reports running.

### Monorepo workspace mode

```toml
[workspace]
members = ["apps/*", "packages/*"]
```

When staged files belong to a member directory, booster runs the hook for that member only (using the member's own `booster.toml` if present, otherwise the root config).

---

## Environment variables

| Variable | Effect |
|----------|--------|
| `SKIP_PRECOMMIT=1` | Skip the entire `pre-commit` hook |
| `SKIP_COMMITMSG=1` | Skip the entire `commit-msg` hook |
| `SKIP_PREPUSH=1` | Skip the entire `pre-push` hook |
| `SKIP_<TOOL>=1` | Skip a specific tool (e.g. `SKIP_ESLINT=1`) |
| `HOOKS_ONLY=format,lint` | Run only tools whose `group` matches |
| `BOOSTER_CONFIG=path` | Override config file location |

---

## Presets

| Preset | Tools |
|--------|-------|
| `node` | prettier + eslint + commit-msg policy |
| `php` | ecs (PHP-CS-Fixer) + commit-msg policy |
| `php-node` | ecs + prettier + eslint + commit-msg policy |
| `go` | gofmt + go vet + commit-msg policy |
| `minimal` | commit-msg policy only |

---

## Migrating from Husky / `.git-hooks.config.json`

```sh
booster migrate --from .git-hooks.config.json --to booster.toml
```

Reads the legacy JSON format and emits an equivalent `booster.toml`. Review and adjust before running `booster install`.

---

## License

[MIT](LICENSE)

## Build

```bash
go build -o booster ./cmd/booster
```

## Quick Start

```bash
# 1) Build or install booster on PATH
go build -o booster ./cmd/booster
mv booster /usr/local/bin/

# 2) In a git repo
booster init
booster install

# 3) Verify
booster doctor
```

After `booster install`, git `core.hooksPath` is set to `.booster/hooks`.
Git automatically executes hook shims there on commit/push.

## Commands

```text
booster init [--force]
booster install
booster uninstall
booster run <hook> [--edit FILE]
booster doctor
```

## Hook Behavior

- `pre-commit`
  - reads staged files (`git diff --cached --name-only --diff-filter=ACMR`)
  - filters files by configured extensions/patterns
  - runs tools in alphabetical order
  - re-stages files for tools with `restage = true`
- `commit-msg`
  - validates conventional commit subject (if enabled)
  - appends `Closes: TICKET` from branch name (if enabled)
- `pre-push`
  - runs configured tools (no staged-file filtering unless tool needs files)

## Environment Variables

- `BOOSTER_CONFIG`: custom config file path
- `HOOKS_ONLY`: comma-separated tool groups (example: `lint,format`)
- `SKIP_PRECOMMIT`, `SKIP_PREPUSH`, `SKIP_COMMITMSG`: skip whole hook
- `SKIP_<TOOL_NAME>`: skip specific tool, normalized to uppercase snake case

Examples:

```bash
HOOKS_ONLY=lint git commit -m "fix: lint only"
SKIP_PHPSTAN=1 git commit -m "chore: bypass phpstan"
```

## Config File (`booster.toml`)

Generated starter config:

```toml
[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.prettier]
command = "prettier"
args = ["--write", "--ignore-unknown"]
extensions = [".js", ".ts", ".json", ".md"]
restage = true
group = "format"

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket = false
```

## Notes

This is a practical v1 prototype aimed at proving the install/run UX and config model.
It intentionally keeps execution simple (host commands, sequential runs, no plugin system yet).
