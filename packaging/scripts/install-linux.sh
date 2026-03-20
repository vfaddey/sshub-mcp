#!/usr/bin/env bash
# Install sshub-mcp binary and systemd user unit.
# Run from the extracted tarball directory. Uses sudo for /usr/bin and /lib/systemd/user.
set -euo pipefail

cd "$(dirname "$0")"
BIN="$(pwd)/sshub-mcp"

if [[ ! -f "$BIN" ]]; then
  echo "sshub-mcp binary not found in $(pwd)" >&2
  exit 1
fi

echo "Installing sshub-mcp to /usr/bin ..."
sudo install -m0755 "$BIN" /usr/bin/sshub-mcp

echo "Installing systemd user unit ..."
sudo mkdir -p /lib/systemd/user
sudo install -m0644 sshub-mcp.service /lib/systemd/user/

echo "Reloading systemd ..."
systemctl --user daemon-reload 2>/dev/null || true

echo ""
echo "Done. Enable and start the service:"
echo "  systemctl --user enable --now sshub-mcp"
echo ""
echo "Admin UI: http://127.0.0.1:8787/admin/"
