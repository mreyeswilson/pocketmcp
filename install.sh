#!/usr/bin/env bash
set -euo pipefail

REPO="${MCP_PB_REPO:-mreyeswilson/pocketmcp}"
BINARY_NAME="pocketmcp"
VERSION_INPUT="${1:-${VERSION:-${MCP_PB_VERSION:-}}}"

resolve_latest_tag() {
  local latest_tag

  latest_tag="$(
    (curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" || true) \
      | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
      | sed -n '1p'
  )"

  if [[ -n "${latest_tag}" ]]; then
    printf '%s\n' "${latest_tag}"
    return
  fi

  latest_tag="$(
    (curl -fsSL "https://api.github.com/repos/${REPO}/tags?per_page=1" || true) \
      | sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
      | sed -n '1p'
  )"

  printf '%s\n' "${latest_tag}"
}

if [[ -n "${VERSION_INPUT}" ]]; then
  TAG="${VERSION_INPUT}"
else
  TAG="$(resolve_latest_tag)"
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
  linux*) ASSET="${BINARY_NAME}" ;;
  darwin*) ASSET="${BINARY_NAME}-macos" ;;
  *)
    echo "Unsupported OS: ${OS}. Supported: Linux/macOS." >&2
    exit 1
    ;;
esac

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"

TMP_FILE="$(mktemp)"
trap 'rm -f "${TMP_FILE}"' EXIT

echo "Downloading ${URL}"
curl -fL "${URL}" -o "${TMP_FILE}"

DEST="${INSTALL_DIR}/${BINARY_NAME}"
mv "${TMP_FILE}" "${DEST}"
chmod +x "${DEST}"

echo "Installed ${BINARY_NAME} ${TAG} to ${DEST}"
echo "Next steps:"
echo "  1) Ensure ${INSTALL_DIR} is in your PATH"
echo "  2) Run: ${BINARY_NAME} serve --url <url> --email <email> --password <password>"
echo "  3) Or run: ${BINARY_NAME} install --client all --url <url> --email <email> --password <password>"
echo "Version selection: latest by default; set VERSION (or MCP_PB_VERSION) or pass the tag as first argument."
