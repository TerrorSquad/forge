# Feature 025: Remote Presets

## Summary
Allow `booster init --preset <url>` to fetch a `booster.toml` from a remote
URL (raw GitHub, GitLab, any HTTPS endpoint) and use it as the initial config.

## Motivation
Organisations want a single source of truth for hook standards. Rather than
copy-pasting configs into each repo, teams can point to a shared preset URL.

## Functional Requirements

1. `booster init --preset https://example.com/presets/php.toml` fetches the
   URL and writes it to `booster.toml`.
2. Supported URL schemes: `https://` only (no `http://`, no filesystem URLs
   via URL syntax).
3. Before writing, booster displays the fetched content and the source URL,
   then prompts for confirmation:
   ```
   Fetching: https://example.com/presets/php.toml
   --- preview ---
   [hooks.pre-commit]
   ...
   ---------------
   Write to booster.toml? [y/N]
   ```
4. In non-interactive mode (`CI=true` or `--yes` flag), write without
   prompting.
5. Redirect limit: max 3 redirects; fail with an error if exceeded.
6. Timeout: 10s fetch timeout.
7. The written file includes a comment header:
   ```toml
   # Fetched from https://example.com/presets/php.toml on 2026-05-22
   ```
8. `--force` overwrites an existing `booster.toml` without the check.

## Security
- Only `https://` is allowed to prevent cleartext credential leakage.
- The fetched content is validated as valid TOML before writing.
- Content is not executed — it is written as-is to disk.

## Out of Scope
- Git-ref-based preset URLs (e.g. `github:org/repo#ref`).
- Preset versioning / lockfile.
- Authenticated endpoints.
