# Feature 024: Run Cache

## Summary
Skip tools whose inputs (config + staged file contents) have not changed
since the last passing run, reducing repeated commit overhead.

## Motivation
Running gofmt on 200 unchanged files adds latency. A content-addressed cache
makes hooks near-instant for no-op commits.

## Cache Design

- Cache file: `.booster/cache.json` (gitignored by booster install).
- Key per tool: `sha256(tool_config_json + sorted(file_path:file_hash) list)`.
- Value: `{ "passed": true, "timestamp": "..." }`.
- On cache hit: tool is skipped with `- gofmt: cached (12ms saved)`.
- On cache miss or tool failure: run normally; update cache only on success.

## Functional Requirements

1. Cache is enabled per-tool:
   ```toml
   [hooks.pre-commit.tools.gofmt]
   cache = true
   ```
   Or globally:
   ```toml
   [execution]
   cache = true
   ```
2. Cache is invalidated when:
   - Any staged file matching the tool's filter changes.
   - The tool's config changes (command, args, extensions, patterns).
   - The tool exits non-zero.
3. `booster run pre-commit --no-cache` bypasses cache for a single run.
4. `booster cache clear` deletes `.booster/cache.json`.
5. Corrupted cache file is silently deleted and recreated (never fail a hook
   due to cache errors).

## Non-Functional Requirements
- Cache file must not be committed: `booster install` adds `.booster/cache.json`
  to `.git/info/exclude`.
- Cache reads/writes must be atomic (write to temp file, rename).

## Out of Scope
- Remote/shared cache.
- Cross-machine cache portability.
