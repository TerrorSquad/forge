#!/usr/bin/env sh
# install.sh — download and install the booster binary
# Usage: curl -fsSL https://raw.githubusercontent.com/TerrorSquad/gobooster/main/install.sh | sh
# Or: curl -fsSL https://raw.githubusercontent.com/TerrorSquad/gobooster/main/install.sh | sh -s -- --version v0.2.0
set -eu

REPO="TerrorSquad/gobooster"
INSTALL_DIR="${BOOSTER_INSTALL_DIR:-/usr/local/bin}"
VERSION=""

# Parse flags
while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --dir)     INSTALL_DIR="$2"; shift 2 ;;
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

echo "Installing booster ${VERSION} (${OS}/${ARCH}) to ${INSTALL_DIR}"

FILENAME="booster_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/booster.tar.gz"
tar -xzf "$TMP/booster.tar.gz" -C "$TMP"

# May need sudo for system dirs
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/booster" "$INSTALL_DIR/booster"
else
  echo "Need elevated privileges to write to $INSTALL_DIR"
  sudo mv "$TMP/booster" "$INSTALL_DIR/booster"
fi

chmod +x "$INSTALL_DIR/booster"
echo "Installed: $("$INSTALL_DIR/booster" --version 2>/dev/null || echo "$INSTALL_DIR/booster")"
