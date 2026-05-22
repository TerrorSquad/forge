# Feature 036: Self-Update (`booster update`)

## Summary
Add a `booster update` command that downloads and installs the latest (or a
specific) release of booster directly, without requiring developers to re-run
the install script or use a package manager.

## Motivation
Keeping booster up to date across a team today requires each developer to
re-run the install script whenever a new version is released. `booster update`
makes upgrades one command and enables version pinning per project.

## CLI Interface

```
booster update                   # upgrade to latest release
booster update --version v1.3.2  # install a specific version
booster update --check           # print latest version without installing
booster update --rollback        # restore the previous version
```

### Output

```
$ booster update
Current version: v1.2.0
Latest version:  v1.3.2

Downloading booster v1.3.2 for linux/amd64... ✓
Verifying checksum (sha256)... ✓
Backing up current binary to ~/.local/bin/booster.v1.2.0.bak
Installing to ~/.local/bin/booster... ✓

booster v1.3.2 installed successfully.
```

```
$ booster update --check
v1.3.2 available (you have v1.2.0)
```

## Functional Requirements

1. Release metadata is fetched from the GitHub Releases API (or configured
   mirror — see `[update]` config below).
2. Binary is verified against SHA256 checksum published alongside each release.
   Installation is aborted if checksum fails.
3. Previous binary is backed up as `<binary>.prev` before replacement, enabling
   `--rollback`.
4. `--rollback` restores the `<binary>.prev` backup.
5. Update respects `BOOSTER_UPDATE_URL` env var for air-gapped / enterprise
   environments pointing to an internal mirror.
6. `booster update --check` is non-destructive and exits 0 if up to date,
   exit 1 if an update is available (useful in CI health checks).

## Version Pinning per Project

```toml
[update]
pin_version = "v1.3.2"   # booster warns when running a different version
channel     = "stable"   # or "rc" for release candidates
```

When `pin_version` is set and the running version differs:
```
⚠ booster v1.2.0 is running but booster.toml pins v1.3.2.
  Run `booster update --version v1.3.2` to upgrade.
```

This warning is non-blocking (never fails a hook).

## Release Metadata Format

```json
{
  "tag_name": "v1.3.2",
  "assets": [
    {
      "name": "booster_linux_amd64",
      "browser_download_url": "https://...",
      "checksum": "sha256:abc123..."
    }
  ]
}
```

## Non-Functional Requirements
- Download is streamed (no full-memory buffer for large binaries).
- Update is atomic: write to temp file, verify, rename.
- No external dependencies beyond the standard library (`net/http`, `crypto/sha256`).

## Out of Scope
- Auto-update on every hook run.
- Updating project-local booster installs (only the user-installed binary).
- Homebrew / apt / snap package management.
