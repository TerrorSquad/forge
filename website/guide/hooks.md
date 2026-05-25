# Hooks

forge supports three git hook entry points. Each maps directly to a standard git hook.

## Supported hooks

| Hook | Trigger |
|------|---------|
| `pre-commit` | Before a commit is created; receives staged files |
| `commit-msg` | After commit message is written; validates / mutates message |
| `pre-push` | Before a push; can run slower checks (tests, build) |

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

Tools run in the **alphabetical order of their keys**. To force a specific order, prefix keys:

```toml
[hooks.pre-commit.tools.01_gofmt]
[hooks.pre-commit.tools.02_govet]
[hooks.pre-commit.tools.03_golangci]
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

## See also

- [Configuration](/guide/configuration)
- [Commit-message Policy](/guide/commit-policy)
