# Environment Variables

| Variable | Effect |
|----------|--------|
| `SKIP_PRECOMMIT=1` | Skip the entire `pre-commit` hook |
| `SKIP_COMMITMSG=1` | Skip the entire `commit-msg` hook |
| `SKIP_PREPUSH=1` | Skip the entire `pre-push` hook |
| `SKIP_<TOOL>=1` | Skip a specific tool by its key name (case-insensitive, uppercase) |
| `SKIP_GROUP_<GROUP>=1` | Skip every tool in a group (e.g. `SKIP_GROUP_ANALYSIS=1`) |
| `HOOKS_ONLY=group1,group2` | Run only tools whose `group` field matches one of the values |
| `FORGE_CONFIG=path` | Override the repo config file location |
| `FORGE_GLOBAL_CONFIG=path` | Override the global user config location (default `~/.config/forge/config.toml`) |

## Examples

```sh
# Skip the entire pre-commit hook
SKIP_PRECOMMIT=1 git commit -m "wip"

# Skip only eslint
SKIP_ESLINT=1 git commit -m "style: tweak"

# Skip everything in the "analysis" group
SKIP_GROUP_ANALYSIS=1 git commit -m "wip"

# Run only formatter tools
HOOKS_ONLY=format git commit -m "style: format"

# Use an alternative config
FORGE_CONFIG=configs/strict.toml git commit -m "feat: stricter checks"
```

## Tool name matching

`SKIP_<TOOL>` matches the **uppercase key** of the tool in `forge.toml`. For example, a tool keyed as `phpcs` is skipped with `SKIP_PHPCS=1`.

## See also

- [Hooks](/guide/hooks)
- [Configuration](/guide/configuration)
