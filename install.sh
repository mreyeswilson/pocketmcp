#!/usr/bin/env bash
set -euo pipefail

REPO="${MCP_PB_REPO:-mreyeswilson/pocketmcp}"
BINARY_NAME="pocketmcp"

if [[ -n "${MCP_PB_VERSION:-}" ]]; then
  TAG="${MCP_PB_VERSION}"
else
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
fi

if [[ -z "${TAG}" ]]; then
  echo "Failed to resolve release tag from ${REPO}." >&2
  exit 1
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64|amd64) ARCH="x86_64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}. Supported: x86_64." >&2
    exit 1
    ;;
esac

case "${OS}" in
  linux*) TARGET="x86_64-unknown-linux-gnu" ;;
  darwin*) TARGET="x86_64-apple-darwin" ;;
  mingw*|msys*|cygwin*) TARGET="x86_64-pc-windows-msvc" ;;
  *)
    echo "Unsupported OS: ${OS}. Supported: Linux/macOS/Windows." >&2
    exit 1
    ;;
esac

EXT=""
if [[ "${TARGET}" == "x86_64-pc-windows-msvc" ]]; then
  EXT=".exe"
fi

ASSET="${BINARY_NAME}-${TAG}-${TARGET}${EXT}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"

TMP_FILE="$(mktemp)"
trap 'rm -f "${TMP_FILE}"' EXIT

echo "Downloading ${URL}"
curl -fL "${URL}" -o "${TMP_FILE}"

DEST="${INSTALL_DIR}/${BINARY_NAME}${EXT}"
mv "${TMP_FILE}" "${DEST}"
chmod +x "${DEST}"

echo "Installed ${BINARY_NAME} ${TAG} to ${DEST}"
echo "Next steps:"
echo "  1) Ensure ${INSTALL_DIR} is in your PATH"
echo "  2) Run: ${BINARY_NAME} serve --url <url> --email <email> --password <password>"
echo "  3) Or run: ${BINARY_NAME} install --client all --url <url> --email <email> --password <password>"
