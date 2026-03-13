#!/usr/bin/env bash
set -euo pipefail

REPO="${MCP_PB_REPO:-mreyeswilson/pocketmcp}"
BINARY_NAME="pocketmcp"
VERSION_INPUT="${1:-${VERSION:-${MCP_PB_VERSION:-}}}"

ensure_path_in_shell_profile() {
  local profile_path="$1"
  local export_line="export PATH=\"${INSTALL_DIR}:\$PATH\""

  touch "${profile_path}"
  if grep -Fqs "${export_line}" "${profile_path}"; then
    return
  fi

  {
    printf '\n# Added by pocketmcp installer\n'
    printf '%s\n' "${export_line}"
  } >> "${profile_path}"
}

ensure_install_dir_on_path() {
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*)
      echo "${INSTALL_DIR} is already available in PATH for this session."
      return
      ;;
  esac

  local shell_name profile_path
  shell_name="$(basename "${SHELL:-}")"

  case "${shell_name}" in
    bash)
      if [[ "${OS}" == darwin* ]]; then
        profile_path="${HOME}/.bash_profile"
      else
        profile_path="${HOME}/.bashrc"
      fi
      ;;
    zsh)
      profile_path="${HOME}/.zshrc"
      ;;
    fish)
      profile_path="${HOME}/.config/fish/config.fish"
      mkdir -p "$(dirname "${profile_path}")"
      touch "${profile_path}"
      if grep -Fqs "${INSTALL_DIR}" "${profile_path}"; then
        :
      else
        {
          printf '\n# Added by pocketmcp installer\n'
          printf 'fish_add_path "%s"\n' "${INSTALL_DIR}"
        } >> "${profile_path}"
      fi
      export PATH="${INSTALL_DIR}:${PATH}"
      echo "Added ${INSTALL_DIR} to PATH in ${profile_path}"
      return
      ;;
    *)
      if [[ "${OS}" == darwin* ]]; then
        profile_path="${HOME}/.zprofile"
      else
        profile_path="${HOME}/.profile"
      fi
      ;;
  esac

  ensure_path_in_shell_profile "${profile_path}"
  export PATH="${INSTALL_DIR}:${PATH}"
  echo "Added ${INSTALL_DIR} to PATH in ${profile_path}"
}

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
ensure_install_dir_on_path

echo "Installed ${BINARY_NAME} ${TAG} to ${DEST}"
echo "Next steps:"
echo "  1) Open a new shell if your current session doesn't see ${INSTALL_DIR} in PATH yet"
echo "  2) Run: ${BINARY_NAME} mcp --url <url> --email <email> --password <password>"
echo "  3) Or run: ${BINARY_NAME} setup --client all --url <url> --email <email> --password <password>"
echo "Version selection: latest by default; set VERSION (or MCP_PB_VERSION) or pass the tag as first argument."
