# Installation

## Homebrew (macOS / Linux)

```sh
brew install terrorsquad/tap/forge
```

Use the fully-qualified name. Homebrew core has an unrelated `forge`
(ArrayFire visualization), so a bare `brew install forge` installs that one
instead.

## curl installer

```sh
curl -fsSL https://raw.githubusercontent.com/TerrorSquad/forge/main/install.sh | sh
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
forge 2.1.0 (commit: 65a24d9, built: 2026-07-18T00:09:45Z)
```

> **Note:** `terrorsquad/tap/forge-git` is the same binary under the formula's
> previous name, kept so existing installs keep upgrading. New installs should
> use `terrorsquad/tap/forge`.

## Next steps

- [Quick Start](/guide/quick-start)
- [Configuration](/guide/configuration)
