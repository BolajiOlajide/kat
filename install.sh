#!/bin/bash
set -e

# Define variables
VERSION="v1.0.0"
DOWNLOAD_URL="https://github.com/BolajiOlajide/kat/releases/download/$VERSION/kat_$VERSION.tar.gz"
INSTALL_DIR="/usr/local/bin"
OS=$(uname | tr '[:upper:]' '[:lower:]')

# Set download URL based on operating system
if [ "$OS" == "darwin" ]; then
  DOWNLOAD_URL="https://github.com/BolajiOlajide/kat/releases/download/$VERSION/kat_${VERSION}_darwin_amd64.tar.gz"
elif [ "$OS" == "linux" ]; then
  DOWNLOAD_URL="https://github.com/BolajiOlajide/kat/releases/download/$VERSION/kat_${VERSION}_linux_amd64.tar.gz"
else
  echo "Unsupported operating system: $OS"
  exit 1
fi

# Download and extract binary
curl -sL $DOWNLOAD_URL | tar xz -C $INSTALL_DIR kat

echo "kat $VERSION has been installed to $INSTALL_DIR"
