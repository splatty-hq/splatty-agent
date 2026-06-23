#!/bin/sh
# Install splatty-agent from the latest GitHub release.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/splatty-hq/splatty-agent/master/install.sh | sh
#
# On Linux with systemd, if SPLATTY_DSN is set in the calling env, the script
# also installs and enables a systemd service so the agent survives reboots:
#
#   curl -fsSL https://raw.githubusercontent.com/splatty-hq/splatty-agent/master/install.sh \
#     | sudo SPLATTY_DSN=<hex> sh
#
# Env vars:
#   VERSION       release tag to install (default: latest)
#   INSTALL_DIR   install destination (default: /usr/local/bin)
#   SPLATTY_DSN   if set on Linux, the systemd service is installed and started
#   SPLATTY_URL   server URL (default: https://splatty.k0va1.dev), passed through
#   SPLATTY_HOST, INTERVAL_SECS, DISK_MOUNTS   optional, passed through to the
#                 service env file when present

set -eu

REPO="splatty-hq/splatty-agent"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux|darwin) ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

# pick a runner for commands that need root
if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
elif command -v sudo >/dev/null 2>&1; then
  SUDO="sudo"
else
  SUDO=""
fi

asset="splatty-agent-${os}-${arch}"
if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading $url"
curl -fsSL "$url" -o "$tmp/$asset"
curl -fsSL "${url}.sha256" -o "$tmp/${asset}.sha256"

echo "verifying checksum"
( cd "$tmp" && shasum -a 256 -c "${asset}.sha256" >/dev/null )

dest="${INSTALL_DIR}/splatty-agent"
if [ -w "$INSTALL_DIR" ] || [ ! -e "$INSTALL_DIR" ]; then
  install -m 0755 "$tmp/$asset" "$dest"
else
  $SUDO install -m 0755 "$tmp/$asset" "$dest"
fi
echo "installed $dest"

# systemd service setup — only on Linux, only when the required env is provided,
# and only when systemd is actually in use.
if [ "$os" != "linux" ]; then
  exit 0
fi
if [ -z "${SPLATTY_DSN:-}" ]; then
  cat <<EOF

Skipping service setup (SPLATTY_DSN not set).
To install the systemd service, re-run with it, e.g.:

  curl -fsSL https://raw.githubusercontent.com/${REPO}/master/install.sh \\
    | sudo SPLATTY_DSN=<hex> sh
EOF
  exit 0
fi
if ! command -v systemctl >/dev/null 2>&1 || ! [ -d /run/systemd/system ]; then
  echo "systemd not detected, skipping service setup" >&2
  exit 0
fi

echo "installing systemd service"

# create unprivileged user if missing
if ! id splatty >/dev/null 2>&1; then
  $SUDO useradd --system --no-create-home --shell /usr/sbin/nologin splatty
fi

# write env file with 0640 root:splatty so secrets aren't world-readable
env_file="/etc/splatty-agent.env"
tmp_env="$tmp/splatty-agent.env"
{
  echo "SPLATTY_DSN=${SPLATTY_DSN}"
  [ -n "${SPLATTY_URL:-}" ] && echo "SPLATTY_URL=${SPLATTY_URL}"
  [ -n "${SPLATTY_HOST:-}" ] && echo "SPLATTY_HOST=${SPLATTY_HOST}"
  [ -n "${INTERVAL_SECS:-}" ] && echo "INTERVAL_SECS=${INTERVAL_SECS}"
  [ -n "${DISK_MOUNTS:-}" ] && echo "DISK_MOUNTS=${DISK_MOUNTS}"
} > "$tmp_env"
$SUDO install -m 0640 -o root -g splatty "$tmp_env" "$env_file"

# write the unit file
unit_path="/etc/systemd/system/splatty-agent.service"
$SUDO tee "$unit_path" >/dev/null <<EOF
[Unit]
Description=Splatty host metrics agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=splatty
EnvironmentFile=${env_file}
ExecStart=${dest}
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/proc /sys

[Install]
WantedBy=multi-user.target
EOF

$SUDO systemctl daemon-reload
$SUDO systemctl enable --now splatty-agent.service
echo "service enabled — check status with: systemctl status splatty-agent"
