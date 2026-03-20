#!/usr/bin/env bash
# Build a .deb from a pre-built Linux sshub-mcp binary.
# Usage: build-deb.sh <version> <amd64|arm64> <path-to-binary>
# Requires: dpkg-deb (dpkg-dev on Debian/Ubuntu)
set -euo pipefail

VERSION="${1:?version (e.g. 1.2.3)}"
ARCH="${2:?amd64 or arm64}"
BIN="${3:?path to linux sshub-mcp binary}"

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
case "$ARCH" in
  amd64) DEB_ARCH=amd64 ;;
  arm64) DEB_ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH (use amd64 or arm64)" >&2; exit 1 ;;
esac

if [[ ! -f "$BIN" ]]; then
  echo "binary not found: $BIN" >&2
  exit 1
fi

STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

mkdir -p "$STAGE/DEBIAN" "$STAGE/usr/bin" "$STAGE/lib/systemd/user"

sed "s/@VERSION@/${VERSION}/g; s/@ARCH@/${DEB_ARCH}/g" \
  "$ROOT/packaging/deb/DEBIAN/control.in" >"$STAGE/DEBIAN/control"

install -m0755 "$BIN" "$STAGE/usr/bin/sshub-mcp"
cp "$ROOT/packaging/deb/lib/systemd/user/sshub-mcp.service" "$STAGE/lib/systemd/user/"

OUT="$ROOT/dist/sshub-mcp_${VERSION}_${DEB_ARCH}.deb"
mkdir -p "$(dirname "$OUT")"
dpkg-deb --root-owner-group --build "$STAGE" "$OUT"
echo "wrote $OUT"
