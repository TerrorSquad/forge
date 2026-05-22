# gobooster (prototype)

`booster` is a policy-driven Git hook runner.

This prototype ships the end-to-end local workflow:

- bootstrap config: `booster init`
- install Git hook shims automatically: `booster install`
- run hooks through config: `booster run <hook>`
- inspect setup: `booster doctor`

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
