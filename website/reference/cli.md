# CLI Commands

## forge init

Create `forge.toml` from a built-in preset.

```sh
forge init [--preset NAME] [--force] [--yes]
```

| Flag | Description |
|------|-------------|
| `--preset NAME` | Use the named preset (`go`, `php`, `node`, `php-node`, `minimal`) or an `https://` URL to a remote preset |
| `--list-presets` | List all available presets and exit |
| `--force` | Overwrite an existing `forge.toml` |
| `--yes` | Skip the confirmation prompt (also implied when `CI=true`) |

---

## forge install

Write hook shims to `.forge/hooks/`, install the config schema, and set `core.hooksPath`.

```sh
forge install
```

---

## forge uninstall

Remove hook shims and restore the default `core.hooksPath`.

```sh
forge uninstall
```

---

## forge run

Run a hook manually, outside of a git operation.

```sh
forge run <hook> [flags]
```

| Flag | Description |
|------|-------------|
| `--edit FILE` | Commit-message file (for `commit-msg` / `prepare-commit-msg`) |
| `--all-files` | Run against all tracked files instead of only staged ones (`pre-commit`) |
| `--check` | Dry-run: use `check_args`, suppress restaging, treat output as failure |
| `--no-cache` | Bypass the run cache for this invocation |
| `--tool NAMES` | Only run these tools (comma-separated) |
| `--group NAMES` | Only run tools in these groups (comma-separated) |
| `--skip-tool NAMES` | Skip these tools (comma-separated) |
| `--skip-group NAMES` | Skip these groups (comma-separated) |
| `--verbose` | Show tool selection and command details |

```sh
forge run pre-commit
forge run pre-commit --tool phpstan --all-files
forge run pre-commit --skip-group format
```

---

## forge ci

Shortcut for CI pipelines — runs `pre-commit` in `--all-files --check --no-cache` mode (non-mutating, any tool output counts as failure).

```sh
forge ci
```

---

## forge list

List configured hooks and their tools, with the resolved backend, group, and timeout for each.

```sh
forge list [--hook HOOK] [--json]
```

| Flag | Description |
|------|-------------|
| `--hook HOOK` | Only show the named hook |
| `--json` | Emit JSON instead of the formatted table |

---

## forge validate

Validate `forge.toml` — rejects unknown fields and reports configuration problems.

```sh
forge validate
```

---

## forge migrate

Convert a legacy `.git-hooks.config.json` to `forge.toml`.

```sh
forge migrate [--from FILE] [--to FILE]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--from FILE` | `.git-hooks.config.json` (auto-detected) | Source file |
| `--to FILE` | `-` (stdout) | Output path; `-` prints to stdout |

---

## forge doctor

Diagnose the forge installation: binary, config, hooks, and tool availability.

```sh
forge doctor [--fix] [--dry-run]
```

| Flag | Description |
|------|-------------|
| `--fix` | Automatically fix detected issues (e.g. re-install missing shims) |
| `--dry-run` | Print what would be fixed without changing anything |

Checks:
- forge binary is on `PATH`
- `forge.toml` is valid
- Hook shims exist for every configured hook
- Configured tools are available on the host (or in the DDEV/Docker container)

---

## forge update

Self-update the forge binary from GitHub releases.

```sh
forge update [--check] [--version TAG] [--rollback]
```

| Flag | Description |
|------|-------------|
| `--check` | Print the latest version without installing |
| `--version TAG` | Install a specific version (e.g. `v2.0.0`) |
| `--rollback` | Restore the previous binary backup |

---

## forge cache clear

Clear the run cache for the current repo.

```sh
forge cache clear
```

---

## forge version

Print version, commit hash, and build date.

```sh
forge version
```

```
forge v2.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
```

---

## forge completion

Generate a shell completion script.

```sh
forge completion bash > /etc/bash_completion.d/forge
forge completion zsh  > "${fpath[1]}/_forge"
forge completion fish > ~/.config/fish/completions/forge.fish
```

Supported shells: `bash`, `zsh`, `fish`.
