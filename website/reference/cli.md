# CLI Commands

## forge init

Create `forge.toml` from a built-in preset.

```sh
forge init [--preset NAME] [--force]
```

| Flag | Description |
|------|-------------|
| `--preset NAME` | Use the named preset (`go`, `php`, `node`, `php-node`, `minimal`) |
| `--force` | Overwrite an existing `forge.toml` |
| `--list-presets` | List all available presets and exit |

---

## forge install

Write hook shims to `.forge/hooks/` and set `core.hooksPath`.

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
forge run <hook>
```

```sh
forge run pre-commit
forge run commit-msg
forge run pre-push
```

---

## forge migrate

Convert a legacy `.git-hooks.config.json` to `forge.toml`.

```sh
forge migrate [--from FILE] [--to FILE]
```

| Flag | Default |
|------|---------|
| `--from FILE` | `.git-hooks.config.json` |
| `--to FILE` | `forge.toml` |

---

## forge doctor

Diagnose the forge installation: binary, config, hooks, and tool availability.

```sh
forge doctor
```

Checks:
- forge binary is on PATH
- `forge.toml` is valid
- Hook shims exist for every configured hook
- Configured tools are available on the host (or in the DDEV container)

---

## forge version

Print version, commit hash, and build date.

```sh
forge version
```

```
forge v1.0.0 (abc1234, 2024-01-01T00:00:00Z)
```

---

## forge completion

Generate shell completion scripts.

```sh
forge completion bash   > ~/.bash_completion.d/forge
forge completion zsh    > ~/.zsh/completions/_forge
forge completion fish   > ~/.config/fish/completions/forge.fish
forge completion pwsh   > forge.ps1
```
