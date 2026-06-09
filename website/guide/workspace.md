# Workspace / Monorepo

forge supports monorepos via a `[workspace]` configuration block.

## How it works

When staged files match a configured workspace member, forge:

1. Resolves the member's own `forge.toml` (if it exists).
2. Falls back to the root `forge.toml` if no member config is found.
3. Runs the hook in the context of that member directory.

## Configuration

```toml
[workspace]
members = ["apps/*", "packages/*"]
```

Glob patterns are resolved relative to the repo root.

## Example structure

```
repo/
├── forge.toml            ← root config
├── apps/
│   ├── api/
│   │   └── forge.toml   ← member-specific config (optional)
│   └── web/
└── packages/
    └── ui/
```

If you commit a file in `apps/api/`, forge uses `apps/api/forge.toml` (if present) or falls back to the root config.

## Per-member configs

Member `forge.toml` files follow the same schema as the root config. They can define different tools, backends, or policies for that member.

## File path handling

Staged file paths are always represented with forward slashes (`/`) internally, regardless of the operating system. Workspace member patterns in `forge.toml` are normalized the same way, so member matching works correctly on Windows and Unix alike.

## Disabling workspace mode

Remove the `[workspace]` block or set `members = []`.

## See also

- [Configuration](/guide/configuration)
- [Hooks](/guide/hooks)
