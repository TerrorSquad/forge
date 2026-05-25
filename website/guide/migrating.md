# Migrating from Husky

forge ships a `migrate` command that converts a legacy `.git-hooks.config.json` to `forge.toml`.

## Run the migration

```sh
forge migrate --from .git-hooks.config.json --to forge.toml
```

Review the generated `forge.toml` before committing — the migrator covers common patterns but may not capture every edge case.

## Then install hooks

```sh
forge install
```

## Remove the old hook runner

```sh
# Remove Husky
npm uninstall husky
rm -rf .husky

# Remove legacy config (after verifying forge.toml is correct)
rm .git-hooks.config.json
```

## Differences to be aware of

| Behaviour | `.git-hooks.config.json` | forge.toml |
|-----------|--------------------------|------------|
| Tool ordering | Array order | Alphabetical key order |
| Skip env vars | `SKIP_PRECOMMIT=1` | Same |
| DDEV detection | Manual wrappers | `backend = "ddev"` |
| Commit-msg policy | Not supported | `[hooks.commit-msg.policy]` |

## See also

- [Configuration](/guide/configuration)
- [Hooks](/guide/hooks)
- [CLI: forge migrate](/reference/cli#forge-migrate)
