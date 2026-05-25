# forge.toml Reference

Full schema reference for `forge.toml`.

## `[execution]`

```toml
[execution]
default_backend = "host"   # "host" | "ddev"
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `default_backend` | string | `"host"` | Default execution backend for all tools |

---

## `[workspace]`

```toml
[workspace]
members = ["apps/*", "packages/*"]
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `members` | `[]string` | `[]` | Glob patterns for workspace members (relative to repo root) |

---

## `[hooks.<name>]`

```toml
[hooks.pre-commit]
enabled = true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Whether this hook is active |

Supported hook names: `pre-commit`, `commit-msg`, `pre-push`.

---

## `[hooks.<name>.tools.<key>]`

```toml
[hooks.pre-commit.tools.gofmt]
command          = "gofmt"
args             = ["-w"]
type             = "system"
backend          = "host"
extensions       = [".go"]
include_patterns = []
exclude_patterns = []
pass_files       = true
run_per_file     = false
restage          = true
on_failure       = ""
group            = ""
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `command` | string | **required** | Binary name or path |
| `args` | `[]string` | `[]` | Extra arguments |
| `type` | string | `"system"` | `"system"` \| `"node"` \| `"php"` — affects binary resolution |
| `backend` | string | global default | `"host"` \| `"ddev"` |
| `extensions` | `[]string` | all | Run only on files with these extensions |
| `include_patterns` | `[]string` | all | Glob allowlist for file paths |
| `exclude_patterns` | `[]string` | none | Glob blocklist for file paths |
| `pass_files` | bool | `true` | Append staged file paths as arguments |
| `run_per_file` | bool | `false` | Invoke the tool once per matching file |
| `restage` | bool | `false` | Re-stage modified files after tool runs |
| `on_failure` | string | `""` | Set to `"stop"` to abort remaining tools on error |
| `group` | string | `""` | Used with `HOOKS_ONLY` to run a subset of tools |

---

## `[hooks.commit-msg.policy]`

```toml
[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = false
require_ticket       = false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `conventional_commits` | bool | `false` | Enforce Conventional Commits prefix |
| `append_ticket_footer` | bool | `false` | Append `Closes: TICKET` footer from branch name |
| `require_ticket` | bool | `false` | Fail if branch has no ticket ID |

---

## Complete example

```toml
[execution]
default_backend = "ddev"

[workspace]
members = ["apps/*"]

[hooks.pre-commit]
enabled = true

[hooks.pre-commit.tools.prettier]
command    = "prettier"
args       = ["--write"]
type       = "node"
extensions = [".ts", ".vue", ".json"]
restage    = true
group      = "format"

[hooks.pre-commit.tools.eslint]
command    = "eslint"
args       = ["--fix"]
type       = "node"
extensions = [".ts", ".vue"]
restage    = true
group      = "lint"

[hooks.pre-commit.tools.phpcs]
command    = "vendor/bin/phpcs"
args       = ["--standard=PSR12"]
type       = "php"
extensions = [".php"]

[hooks.commit-msg]
enabled = true

[hooks.commit-msg.policy]
conventional_commits = true
append_ticket_footer = true
require_ticket       = false
```
