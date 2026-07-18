# Hooks

Each hook maps directly to a standard git hook. `forge install` writes a shim for every supported hook.

## Supported hooks

| Hook | Trigger |
|------|---------|
| `pre-commit` | Before a commit is created; receives staged files |
| `commit-msg` | After the commit message is written; validates / mutates the message |
| `prepare-commit-msg` | Before the editor opens; can pre-fill the message (e.g. ticket prefix) |
| `pre-push` | Before a push; good for slower checks (tests, build) |
| `post-commit` | After a commit completes; non-blocking |
| `post-merge` | After a merge (e.g. `git pull`); non-blocking |
| `post-rewrite` | After history is rewritten (`rebase`, `commit --amend`); non-blocking |

`pre-commit`, `pre-push`, and the `post-*` hooks run configured tools. `commit-msg` and `prepare-commit-msg` additionally apply the [commit-message policy](/guide/commit-policy).

## Enabling a hook

```toml
[hooks.pre-commit]
enabled = true

[hooks.commit-msg]
enabled = true

[hooks.pre-push]
enabled = true
```

## Adding tools

Each hook has a `tools` table. Keys are arbitrary tool names.

```toml
[hooks.pre-commit.tools.phpcs]
command    = "vendor/bin/phpcs"
args       = ["--standard=PSR12"]
type       = "php"
extensions = [".php"]
```

## Execution order

Tools run in the order they are declared in `forge.toml`.

```toml
[hooks.pre-commit.tools.gofmt]
command = "gofmt"

[hooks.pre-commit.tools.golangci]
command = "golangci-lint"
```

## Staged file filtering

For `pre-commit`, forge automatically passes only the staged files matching the tool's `extensions` and patterns. Tools with `pass_files = false` receive no file arguments.

## Restaging

When a tool modifies files (e.g., a formatter), set `restage = true` to automatically re-add those files to the index:

```toml
[hooks.pre-commit.tools.prettier]
command = "prettier"
args    = ["--write"]
restage = true
```

## Stopping on failure

By default, forge continues running remaining tools even if one fails. To abort on first failure:

```toml
[hooks.pre-commit.tools.critical-lint]
command    = "golangci-lint"
args       = ["run"]
on_failure = "stop"
```

## Running hooks manually

```sh
forge run pre-commit
forge run commit-msg
forge run pre-push
```

## Skipping hooks

```sh
SKIP_PRECOMMIT=1 git commit -m "..."
SKIP_COMMITMSG=1 git commit -m "..."
SKIP_PREPUSH=1   git push
```

## Running only specific tool groups

```toml
[hooks.pre-commit.tools.prettier]
group = "format"

[hooks.pre-commit.tools.eslint]
group = "lint"
```

```sh
HOOKS_ONLY=format git commit -m "..."
```

## Missing tools fail the hook

Before running anything, forge checks that every enabled tool can actually run. If any can't, the whole hook aborts **before** a single tool runs — so a mistyped command or an uninstalled linter can never silently pass as "all checks green".

```
run failed: cannot run hook, some tools are unavailable:
  - eslint — binary not found: node_modules/.bin/eslint
install/start them, set SKIP_<TOOL>=1 to skip, or disable the tool in forge.toml
```

You then have three ways forward:

- **Install** the tool (or start its container).
- **Skip it for one run**: `SKIP_ESLINT=1 git commit …`.
- **Disable it**: comment out or remove the tool's block in `forge.toml`.

What's checked:

- **Host tools** — the binary must resolve on `PATH` (or as a `vendor/bin` / `node_modules` path).
- **DDEV / Docker tools** — the container must be **running**. forge does **not** start it for you (a commit shouldn't spin up infrastructure); it fails fast with `container "…" is not running` so you can `ddev start` and retry.

Tools you've already skipped (`SKIP_*`, `--skip-tool`, a non-matching `HOOKS_ONLY` group) are exempt. Run `forge doctor` to see availability without triggering a commit.

## See also

- [Configuration](/guide/configuration)
- [Commit-message Policy](/guide/commit-policy)
