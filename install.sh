#!/bin/bash
set -e

check_cmd() {
  command -v "$1" >/dev/null 2>&1
}

need_cmd() {
  if ! check_cmd "$1"; then
    err "need '$1' (command not found)"
  fi
}

need_cmd curl
need_cmd uname
need_cmd grep
need_cmd sed

# Define variables
# If VERSION is not provided, fetch the latest version from GitHub API
if [ -z "$VERSION" ]; then
  echo "No version specified, fetching latest release..."
  VERSION=$(curl -s https://api.github.com/repos/BolajiOlajide/kat/releases/latest | grep '"tag_name":' | sed -E 's/.*"tag_name": ?"([^"]+)".*/\1/')
  echo "Latest version: '$VERSION'"
fi

# Allow users to override the installation directory
KAT_INSTALL_DIR=${KAT_INSTALL_DIR:-"/usr/local/bin"}
OS=$(uname | tr '[:upper:]' '[:lower:]')

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64|x86-64|x64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac
echo "Detected architecture: $ARCH"

# Set download URL based on operating system and architecture
if [ "$OS" == "darwin" ]; then
  DOWNLOAD_URL="https://github.com/BolajiOlajide/kat/releases/download/$VERSION/kat_darwin_${ARCH}.tar.gz"
elif [ "$OS" == "linux" ]; then
  DOWNLOAD_URL="https://github.com/BolajiOlajide/kat/releases/download/$VERSION/kat_linux_${ARCH}.tar.gz"
else
  echo "Unsupported operating system: $OS"
  exit 1
fi

# Verify binary exists before downloading
if ! curl --output /dev/null --silent --head --fail "$DOWNLOAD_URL"; then
  echo "Error: Binary for $OS on $ARCH architecture not found for version $VERSION"
  echo "Please check available releases at: https://github.com/BolajiOlajide/kat/releases"
  exit 1
fi

# Download and extract binary
echo "Downloading 'kat $VERSION' for $OS ($ARCH)..."
curl -sL "$DOWNLOAD_URL" | tar xz -C "$KAT_INSTALL_DIR" kat

echo "kat $VERSION has been installed in '$KAT_INSTALL_DIR'"
