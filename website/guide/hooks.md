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

## See also

- [Configuration](/guide/configuration)
- [Commit-message Policy](/guide/commit-policy)
