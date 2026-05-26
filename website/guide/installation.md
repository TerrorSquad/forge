# Installation

## Homebrew (macOS / Linux)

```sh
brew tap TerrorSquad/tap
brew install forge-git
```

## curl installer

```sh
curl -fsSL https://raw.githubusercontent.com/TerrorSquad/forge/master/install.sh | sh
```

The script downloads the latest release binary for your OS/arch and installs it to `/usr/local/bin` (or `~/.local/bin` if `/usr/local/bin` is not writable).

## go install

Requires Go 1.23+.

```sh
go install github.com/TerrorSquad/forge/cmd/forge@latest
```

## Manual download

1. Go to the [Releases](https://github.com/TerrorSquad/forge/releases) page.
2. Download the archive for your platform.
3. Extract and move the `forge` binary to a directory on your `PATH`.

## Verify

```sh
forge version
```

```
forge v1.0.0 (abc1234, 2024-01-01)
```

> **Note:** The Homebrew formula is named `forge-git` to avoid conflicts with an existing `forge` package. Install with `brew install forge-git` but the binary is named `forge`.

## Next steps

- [Quick Start](/guide/quick-start)
- [Configuration](/guide/configuration)
