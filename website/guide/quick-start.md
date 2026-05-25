# Quick Start

## 1. Install forge

See [Installation](/guide/installation).

## 2. Initialise a config

In the root of your git repo:

```sh
forge init --preset go        # Go project
forge init --preset php       # PHP project
forge init --preset node      # Node project
forge init --preset php-node  # PHP + Node monolith
forge init --preset minimal   # Commit-msg policy only
```

This creates `forge.toml`. Review and adjust it.

## 3. Install hook shims

```sh
forge install
```

forge writes thin shell scripts to `.forge/hooks/` and sets git's `core.hooksPath` to that directory.

## 4. Verify

```sh
forge doctor
```

`doctor` checks that the binary is on PATH, the config is valid, and hook shims exist for every configured hook.

## 5. Commit

```sh
git add .
git commit -m "chore: add forge"
```

Your hooks fire. If anything fails, the commit is aborted with a clear error message.

## Skip a tool once

```sh
SKIP_ESLINT=1 git commit -m "wip: quick save"
```

## Skip a hook once

```sh
SKIP_PRECOMMIT=1 git commit -m "wip: quick save"
```

## See also

- [Configuration](/guide/configuration)
- [Hooks](/guide/hooks)
- [Environment Variables](/reference/env)
