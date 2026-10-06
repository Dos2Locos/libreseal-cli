#!/usr/bin/env sh
# Build the LibreSeal CLI from this checkout and install it as `libreseal`.
#
# Usage: ./scripts/install-from-source.sh [INSTALL_DIR]
#   INSTALL_DIR defaults to /usr/local/bin (uses sudo when not writable),
#   or set PREFIX=$HOME/.local to install into $HOME/.local/bin.
# Requires Go (see src/go.mod for the minimum version). No network access is
# needed beyond fetching Go modules.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${1:-${PREFIX:+$PREFIX/bin}}"
DEST="${DEST:-/usr/local/bin}"

command -v go >/dev/null 2>&1 || { echo "Go is required: https://go.dev/dl/" >&2; exit 1; }

VERSION="$(git -C "$ROOT" describe --tags --always --dirty 2>/dev/null || echo dev)"
VERSION="${VERSION#v}"
OUT="$(mktemp -d)"
trap 'rm -rf "$OUT"' EXIT

echo "Building libreseal $VERSION..."
(cd "$ROOT/src" && CGO_ENABLED=0 go build -trimpath \
  -ldflags "-s -w -X github.com/phasehq/cli/pkg/version.Version=$VERSION" \
  -o "$OUT/libreseal" .)

mkdir -p "$DEST" 2>/dev/null || true
if [ -w "$DEST" ]; then
  install -m 0755 "$OUT/libreseal" "$DEST/libreseal"
else
  sudo install -m 0755 "$OUT/libreseal" "$DEST/libreseal"
fi

echo "Installed $DEST/libreseal"
"$DEST/libreseal" --version
