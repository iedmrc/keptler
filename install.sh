#!/usr/bin/env bash
set -euo pipefail

REPO="keptler/keptler"
OS="$(uname -s)"
case "$OS" in
    Linux) OS=linux ;;
    Darwin) OS=darwin ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH=amd64 ;;
    arm64|aarch64) ARCH=arm64 ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | \
  grep tag_name | head -n1 | cut -d '"' -f4)
TARBALL="keptler_${LATEST}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/${LATEST}/${TARBALL}"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
curl -fsSL "$URL" -o "$tmpdir/$TARBALL"
tar -xzf "$tmpdir/$TARBALL" -C "$tmpdir"

DEST="/usr/local/bin"
if install -m 0755 "$tmpdir/keptler" "$DEST/keptler" 2>/dev/null; then
    echo "Installed keptler to $DEST/keptler"
else
    DEST="$HOME/.local/bin"
    mkdir -p "$DEST"
    install -m 0755 "$tmpdir/keptler" "$DEST/keptler"
    echo "Installed keptler to $DEST/keptler"
fi
