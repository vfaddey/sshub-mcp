#!/usr/bin/env bash
# One-liner install: curl -sL https://raw.githubusercontent.com/vfaddey/sshub-mcp/main/install.sh | bash
# Or with version: curl -sL ... | bash -s -- v0.0.2
set -euo pipefail

REPO="${SSHUB_MCP_REPO:-vfaddey/sshub-mcp}"
BASE="https://github.com/${REPO}/releases"

# Version: first arg, or INSTALL_VERSION env, or latest from GitHub
if [[ -n "${1:-}" ]]; then
  VERSION="$1"
elif [[ -n "${INSTALL_VERSION:-}" ]]; then
  VERSION="$INSTALL_VERSION"
else
  VERSION=$(curl -sL -o /dev/null -w '%{url_effective}' "${BASE}/latest" | sed 's|.*/tag/||' || echo "v0.0.2")
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux)  TARBALL="sshub-mcp_linux_${ARCH}.tar.gz" ;;
  darwin) TARBALL="sshub-mcp_darwin_${ARCH}.tar.gz" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

URL="${BASE}/download/${VERSION}/${TARBALL}"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading sshub-mcp ${VERSION} (${OS}/${ARCH}) ..."
curl -sL "$URL" -o "$TMP/archive.tar.gz"
tar xzf "$TMP/archive.tar.gz" -C "$TMP"

if [[ ! -f "$TMP/sshub-mcp" ]]; then
  echo "Archive missing sshub-mcp binary" >&2
  exit 1
fi

if [[ "$OS" == "linux" ]]; then
  echo "Installing to /usr/bin ..."
  sudo install -m0755 "$TMP/sshub-mcp" /usr/bin/sshub-mcp
  echo "Installing systemd user unit ..."
  sudo mkdir -p /lib/systemd/user
  sudo install -m0644 "$TMP/sshub-mcp.service" /lib/systemd/user/
  systemctl --user daemon-reload 2>/dev/null || true
  echo "Enabling and starting service ..."
  systemctl --user enable --now sshub-mcp
else
  if [[ -d /opt/homebrew ]]; then
    PREFIX="/opt/homebrew"
  else
    PREFIX="/usr/local"
    sudo mkdir -p /usr/local/bin
  fi
  BIN_PATH="$PREFIX/bin/sshub-mcp"
  echo "Installing to $BIN_PATH ..."
  sudo mkdir -p "$PREFIX/bin"
  sudo install -m0755 "$TMP/sshub-mcp" "$BIN_PATH"
  PLIST_DEST="$HOME/Library/LaunchAgents/sshub-mcp.plist"
  mkdir -p "$(dirname "$PLIST_DEST")"
  sed "s|__BIN_PATH__|$BIN_PATH|g" "$TMP/sshub-mcp.plist" > "$PLIST_DEST"
  echo "Loading launchd service ..."
  launchctl load "$PLIST_DEST"
fi

echo ""
echo "sshub-mcp installed. Admin UI: http://127.0.0.1:8787/admin/"
