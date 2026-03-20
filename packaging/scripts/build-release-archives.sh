#!/usr/bin/env bash
# Cross-compile sshub-mcp and produce tar.gz per platform (Linux/macOS, amd64/arm64).
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
  (cd dist && tar czvf "${name}.tar.gz" "$name")
done

echo "archives in dist/"
