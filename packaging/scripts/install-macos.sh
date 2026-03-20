#!/usr/bin/env bash
# Install sshub-mcp binary and launchd plist.
# Run from the extracted tarball directory. Uses sudo for /usr/local/bin.
set -euo pipefail

cd "$(dirname "$0")"
BIN="$(pwd)/sshub-mcp"

if [[ ! -f "$BIN" ]]; then
  echo "sshub-mcp binary not found in $(pwd)" >&2
  exit 1
fi

if [[ -d /opt/homebrew ]]; then
  PREFIX="/opt/homebrew"
elif [[ -d /usr/local ]]; then
  PREFIX="/usr/local"
else
  sudo mkdir -p /usr/local/bin
  PREFIX="/usr/local"
fi

BIN_PATH="$PREFIX/bin/sshub-mcp"
echo "Installing sshub-mcp to $BIN_PATH ..."
sudo mkdir -p "$PREFIX/bin"
sudo install -m0755 "$BIN" "$BIN_PATH"

PLIST_DEST="$HOME/Library/LaunchAgents/sshub-mcp.plist"
mkdir -p "$(dirname "$PLIST_DEST")"
sed "s|__BIN_PATH__|$BIN_PATH|g" sshub-mcp.plist > "$PLIST_DEST"

echo "Installed launchd plist to $PLIST_DEST"
echo ""
echo "Load and start the service:"
echo "  launchctl load $PLIST_DEST"
echo ""
echo "Admin UI: http://127.0.0.1:8787/admin/"
