#!/bin/bash
set -e

FALLBACK_TAG="v0.1.0"
OWNER="sirrryasir"
REPO="archon"

echo "Checking the latest release of Archon..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
  LATEST_TAG=$FALLBACK_TAG
fi

echo "Latest release found: ${LATEST_TAG}"

OS_UNAME=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS_UNAME" in
  darwin*)  OS="darwin" ;;
  linux*)   OS="linux" ;;
  *)        echo "Unsupported OS: $OS_UNAME"; exit 1 ;;
esac

ARCH_UNAME=$(uname -m)
case "$ARCH_UNAME" in
  x86_64)   ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *)        echo "Unsupported Architecture: $ARCH_UNAME"; exit 1 ;;
esac

FILENAME="archon-${OS}-${ARCH}"
URL="https://github.com/sirrryasir/archon/releases/download/${LATEST_TAG}/${FILENAME}"

echo "Downloading ${FILENAME} from GitHub Releases..."
TEMP_DIR=$(mktemp -d)
TEMP_FILE="${TEMP_DIR}/archon"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TEMP_FILE"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TEMP_FILE" "$URL"
else
  echo "Error: Neither curl nor wget was found. Please install one of them."
  exit 1
fi

chmod +x "$TEMP_FILE"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Prompting for sudo to install to ${INSTALL_DIR}..."
  sudo mv "$TEMP_FILE" "${INSTALL_DIR}/archon"
else
  mv "$TEMP_FILE" "${INSTALL_DIR}/archon"
fi

rm -rf "$TEMP_DIR"

echo "Archon has been successfully installed to ${INSTALL_DIR}/archon!"
echo "Run 'archon doctor' to verify your installation."
