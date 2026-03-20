#!/usr/bin/env bash
# Cross-compile sshub-mcp and produce tar.gz per platform (Linux/macOS, amd64/arm64).
# Each archive contains: binary (as sshub-mcp), service/plist, install.sh
# Output: dist/sshub-mcp_<os>_<arch>.tar.gz
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
mkdir -p dist
rm -f dist/sshub-mcp_* dist/sshub-mcp-*.tar.gz 2>/dev/null || true

for pair in linux:amd64 linux:arm64 darwin:amd64 darwin:arm64; do
  GOOS=${pair%%:*}
  GOARCH=${pair#*:}
  name="sshub-mcp_${GOOS}_${GOARCH}"
  echo "building $name ..."

  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "dist/$name" ./cmd/sshub-mcp

  STAGE="dist/${name}_stage"
  rm -rf "$STAGE"
  mkdir -p "$STAGE"

  install -m0755 "dist/$name" "$STAGE/sshub-mcp"

  if [[ "$GOOS" == "linux" ]]; then
    cp "$ROOT/packaging/deb/lib/systemd/user/sshub-mcp.service" "$STAGE/"
    cp "$ROOT/packaging/scripts/install-linux.sh" "$STAGE/install.sh"
  else
    cp "$ROOT/packaging/macos/sshub-mcp.plist.in" "$STAGE/sshub-mcp.plist"
    cp "$ROOT/packaging/scripts/install-macos.sh" "$STAGE/install.sh"
  fi
  chmod +x "$STAGE/install.sh"

  (cd dist && tar czvf "${name}.tar.gz" -C "${name}_stage" .)
  rm -rf "$STAGE"
  if [[ "$GOOS" != "linux" ]]; then
    rm -f "dist/$name"
  fi
done

echo "archives in dist/"
