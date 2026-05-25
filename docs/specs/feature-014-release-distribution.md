# Feature 014: Release and Distribution

## Summary
Provide a reproducible, cross-platform release pipeline that lets users install
`forge` in one step via standard package managers or a curl script.

## Distribution Channels

| Channel              | Target users                                   |
|----------------------|------------------------------------------------|
| Homebrew tap         | macOS/Linux developers                         |
| Standalone install   | `curl -fsSL … | sh`                            |
| GitHub Releases      | Manual download, CI pinning                    |
| `go install`         | Go developers                                  |

## Functional Requirements

### Build matrix
- `linux/amd64`, `linux/arm64`
- `darwin/amd64`, `darwin/arm64`
- `windows/amd64`

### Release artifacts
- Single static binary per platform, named `forge-<os>-<arch>[.exe]`
- SHA-256 checksums file `checksums.txt`
- GitHub Release with changelog extracted from CHANGELOG.md

### Homebrew tap
- Repository: `TerrorSquad/homebrew-tap`
- Formula auto-updated on new tag via GitHub Actions

### Standalone installer
```sh
curl -fsSL https://raw.githubusercontent.com/TerrorSquad/forge/main/install.sh | sh
```
- Detects OS/arch.
- Downloads the matching binary from the latest GitHub Release.
- Installs to `~/.local/bin` (or `PREFIX` override).
- Verifies checksum.

## CI/CD
- GitHub Actions workflow triggered on `v*` tags.
- GoReleaser for cross-compilation and artifact packaging.

## Out of Scope
- Scoop, Winget, apt/deb, rpm packaging in v1.
- Signed binaries.
