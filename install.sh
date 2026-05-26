#!/usr/bin/env sh
# install.sh — download and install the forge binary
# Usage: curl -fsSL https://raw.githubusercontent.com/TerrorSquad/forge/main/install.sh | sh
# Or: curl -fsSL https://raw.githubusercontent.com/TerrorSquad/forge/main/install.sh | sh -s -- --version v0.2.0
set -eu

REPO="${FORGE_REPO:-TerrorSquad/forge}"
INSTALL_DIR="${FORGE_INSTALL_DIR:-$HOME/.local/bin}"
VERSION=""
FORGE_URL=""

# Parse flags
while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --dir)     INSTALL_DIR="$2"; shift 2 ;;
    --url)     FORGE_URL="$2"; shift 2 ;;
    --repo)    REPO="$2"; shift 2 ;;
    *) echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
 done

# Detect OS / arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Resolve version
if [ -z "$VERSION" ]; then
  echo "Fetching latest release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
fi

VERSION="${VERSION#v}"

echo "Installing forge ${VERSION} (${OS}/${ARCH}) to ${INSTALL_DIR}"

if [ -n "$FORGE_URL" ]; then
  URL="$FORGE_URL"
else
  FILENAME="forge_${VERSION}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

if [ -n "$FORGE_URL" ]; then
  case "$FORGE_URL" in
    file://*)
      FILE_PATH="${FORGE_URL#file://}"
      if [ ! -f "$FILE_PATH" ]; then
        echo "Error: local file not found: $FILE_PATH" >&2
        exit 1
      fi
      cp "$FILE_PATH" "$TMP/forge.tar.gz"
      ;;
    *)
      if [ -f "$FORGE_URL" ]; then
        cp "$FORGE_URL" "$TMP/forge.tar.gz"
      else
        curl -fsSL "$FORGE_URL" -o "$TMP/forge.tar.gz"
      fi
      ;;
  esac
else
  curl -fsSL "$URL" -o "$TMP/forge.tar.gz"
fi

tar -xzf "$TMP/forge.tar.gz" -C "$TMP"

BINARY_PATH="$(find "$TMP" -maxdepth 2 -type f -name forge -print -quit)"
if [ -z "$BINARY_PATH" ]; then
  echo "Error: could not find forge binary in the archive" >&2
  exit 1
fi

DEST="$INSTALL_DIR/forge"
if [ ! -d "$INSTALL_DIR" ]; then
  echo "Creating install directory $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR"
fi

if [ ! -w "$INSTALL_DIR" ]; then
  echo "Error: cannot write to $INSTALL_DIR."
  echo "Run with --dir ~/.local/bin or set FORGE_INSTALL_DIR to a writable path."
  exit 1
fi

mv "$BINARY_PATH" "$DEST"

chmod +x "$DEST"
echo "Installed: $("$DEST" --version 2>/dev/null || echo "$DEST")"
