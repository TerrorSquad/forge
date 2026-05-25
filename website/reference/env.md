# Environment Variables

| Variable | Effect |
|----------|--------|
| `SKIP_PRECOMMIT=1` | Skip the entire `pre-commit` hook |
| `SKIP_COMMITMSG=1` | Skip the entire `commit-msg` hook |
| `SKIP_PREPUSH=1` | Skip the entire `pre-push` hook |
| `SKIP_<TOOL>=1` | Skip a specific tool by its key name (case-insensitive, uppercase) |
| `HOOKS_ONLY=group1,group2` | Run only tools whose `group` field matches one of the values |
| `FORGE_CONFIG=path` | Override the config file location |
| `FORGE_NO_STASH=1` | Disable the pre-commit safety stash of unstaged changes |

## Examples

```sh
# Skip the entire pre-commit hook
SKIP_PRECOMMIT=1 git commit -m "wip"

# Skip only eslint
SKIP_ESLINT=1 git commit -m "style: tweak"

# Run only formatter tools
HOOKS_ONLY=format git commit -m "style: format"

# Use an alternative config
FORGE_CONFIG=configs/strict.toml git commit -m "feat: stricter checks"

# Disable safety stash (faster, but hooks see a mix of staged/unstaged)
FORGE_NO_STASH=1 git commit -m "chore: bulk"
```

## Tool name matching

`SKIP_<TOOL>` matches the **uppercase key** of the tool in `forge.toml`. For example, a tool keyed as `phpcs` is skipped with `SKIP_PHPCS=1`.

## See also

- [Hooks](/guide/hooks)
- [Configuration](/guide/configuration)
