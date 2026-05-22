# Feature 039: Config Inheritance (`extends`)

## Summary
Allow a `booster.toml` to inherit from a base config — a local path or a
remote URL — so organizations can publish a shared standard config and
individual projects override only what they need.

## Motivation
In a multi-repo organization, every project starts by copy-pasting the same
booster.toml. When the organization wants to add a new required tool (e.g.
a security scanner), someone has to update every repo manually. With `extends`,
the shared config is the source of truth and projects can opt in to overrides
without forking.

## Config Design

```toml
# Inherit from an organization-shared config (fetched and cached locally).
extends = "https://raw.githubusercontent.com/myorg/standards/main/booster.toml"

# Or inherit from a local path (monorepo use case).
extends = "../../shared/booster.toml"

# Override only what differs from the base.
[hooks.pre-commit.tools.phpstan]
args = ["analyse", "--level=8"]   # stricter than the org default of level=5
```

### Multiple Inheritance (ordered, last wins)
```toml
extends = [
  "https://cdn.myorg.com/booster/base.toml",
  "./local-overrides.toml"
]
```

## Merge Semantics

The resolved config is the result of deep-merging all inherited configs in
order, with the local `booster.toml` having the highest priority.

| Config key | Merge behaviour |
|------------|----------------|
| `[hooks.<hook>.tools.<name>]` | Shallow merge: local keys override parent keys |
| `[hooks.<hook>.tools.<name>].args` | **Replace**, not append (override args explicitly) |
| `[hooks.<hook>.tools.<name>].extensions` | Replace |
| Tool present only in parent | Inherited as-is |
| Tool set to `enabled = false` locally | Removes the tool from the chain |

### Disabling an inherited tool
```toml
[hooks.pre-commit.tools.infection]
enabled = false   # this project opts out of mutation testing
```

## Remote Config Caching

- Remote URLs are fetched on `booster install` and cached in
  `.booster/extends-cache/<hash>.toml`.
- Cache TTL is 24 hours by default. `booster update-extends` refreshes
  manually.
- If the remote is unreachable and a cache exists, the cached version is used
  with a warning.
- If the remote is unreachable and no cache exists, booster aborts with a
  clear error.

```toml
[extends_config]
ttl          = "24h"
offline_mode = "warn"   # or "fail" — what to do when remote is unreachable
```

## Security

- Remote URLs must be HTTPS (no HTTP unless `allow_insecure = true`).
- Content is stored as-is; booster does NOT execute remote config, only merges
  TOML structure.
- `booster doctor --extends` prints the resolved merged config for review.
- SHA256 pinning:
  ```toml
  extends = { url = "https://...", sha256 = "abc123" }
  ```

## `booster doctor --extends`

Shows the full resolved config after inheritance:
```
$ booster doctor --extends

Resolved config (3 layers):
  1. https://cdn.myorg.com/booster/base.toml  (cached 2h ago)
  2. ./local-overrides.toml
  3. ./booster.toml                           (this project)

pre-commit tools (resolved):
  ecs          → base.toml            args: ["check", "--fix"]
  phpstan      → booster.toml         args: ["analyse", "--level=8"]  (overrides base level=5)
  psalm        → base.toml
  infection    → base.toml            [DISABLED by booster.toml]
```

## Out of Scope
- Diamond inheritance resolution (circular `extends` is an error).
- Overriding `[execution]` or `[workspace]` from remote (security boundary).
- Private registry authentication for remote configs.
