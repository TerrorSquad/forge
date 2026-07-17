# forge.toml Reference

Full schema reference for `forge.toml`.

## `[execution]`

Repository-wide execution defaults.

```toml
[execution]
default_backend = "ddev"   # "host" | "ddev" | custom Docker container name
parallel        = false    # run tools within a hook concurrently
cache           = false    # skip tools whose inputs are unchanged
tool_timeout    = "60s"    # default per-tool timeout (Go duration)
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `default_backend` | string | *(auto)* | Default backend for all tools. When omitted, forge auto-detects DDEV (uses it if the container is running, else host). Set `"host"` to force host and **disable** auto-detection |
| `parallel` | bool | `false` | Run a hook's tools concurrently (respecting `depends_on`) |
| `cache` | bool | `false` | Enable the run cache — unchanged inputs skip the tool |
| `tool_timeout` | string | — | Default timeout for every tool, as a Go duration (`"30s"`, `"2m"`) |
| `cache_ttl` | string | — | Max age of a cache entry before eviction (Go duration; empty = never) |
| `cache_max_size` | int | `0` | Max cache entries kept; oldest evicted first (`0` = unlimited) |

---

## `[update]`

Controls the `forge update` self-updater.

```toml
[update]
pin_version = "v2.0.0"   # warn when the running binary differs
channel     = "stable"   # "stable" (default) or "rc"
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `pin_version` | string | — | Print a non-blocking warning when the running binary differs from this version |
| `channel` | string | `"stable"` | Release channel to update from: `"stable"` or `"rc"` |

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
enabled  = true
parallel = false   # optional: override [execution].parallel for this hook
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `false` | Whether this hook is active |
| `parallel` | bool | *(inherits `[execution]`)* | Run this hook's tools concurrently |

Supported hook names: `pre-commit`, `commit-msg`, `prepare-commit-msg`, `pre-push`, `post-commit`, `post-merge`, `post-rewrite`.

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
| `backend` | string | global default | `"host"` \| `"ddev"` \| `<container-name>` |
| `extensions` | `[]string` | all | Run only on files with these extensions |
| `include_patterns` | `[]string` | all | Glob allowlist for file paths (supports `**`) |
| `exclude_patterns` | `[]string` | none | Glob blocklist for file paths (supports `**`) |
| `pass_files` | bool | `true` | Append staged file paths as arguments |
| `run_per_file` | bool | `false` | Invoke the tool once per matching file |
| `restage` | bool | `false` | Re-stage modified files after tool runs |
| `on_failure` | string | `""` | Set to `"stop"` to abort remaining tools on error |
| `group` | string | `""` | Used with `HOOKS_ONLY` / `SKIP_GROUP_*` to run a subset of tools |
| `depends_on` | `[]string` | `[]` | Tool names that must run before this one (ordering under `parallel`) |
| `timeout` | string | *(inherits `[execution]`)* | Per-tool timeout as a Go duration (`"30s"`, `"2m"`) |
| `cache` | bool | `false` | Enable the run cache for this tool (or set `[execution] cache`) |
| `env` | table | `{}` | Extra environment variables for the tool process |
| `stage_outputs` | `[]string` | `[]` | Paths to `git add` after the tool runs (for generated files) |
| `show_output` | bool | `false` | Always print the tool's output, even on success |
| `check_args` | `[]string` | *(uses `args`)* | Args to use instead of `args` in `--check` / `forge ci` mode |
| `check_fail_if_output` | bool | `false` | In check mode, treat any output as a failure |

> Tool sections are executed in the order they appear in `forge.toml` (or by `depends_on` when `parallel = true`). `forge validate` rejects unknown or unsupported fields.

### Glob patterns

`include_patterns` and `exclude_patterns` use double-star (`**`) glob syntax:

```toml
exclude_patterns = ["vendor/**/*", "*.generated.php"]
include_patterns = ["src/**/*.php"]
```

- `*` matches any sequence of characters within a single path segment.
- `**` matches zero or more path segments (i.e. any depth of subdirectories).
- Paths are always matched as forward-slash-separated strings regardless of OS.

### Tool name case sensitivity

Tool names in `forge.toml` are matched case-insensitively by `--tool`, `--skip-tool`, `SKIP_<NAME>=1`, and related flags. The name as defined in `forge.toml` is used in output; the comparison is always case-folded.

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
| `ticket_pattern` | string | `([A-Z]+-[0-9]+)` | Regex with capture group matched against the branch name to extract a ticket ID. Default matches Jira/Linear style (e.g. `PRJ-123`). Use `(#[0-9]+)` for GitHub issues. |
| `allowed_types` | array | *(default set)* | Custom list of allowed Conventional Commits types. Overrides the default set (`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`). Only active when `conventional_commits = true`. |

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
