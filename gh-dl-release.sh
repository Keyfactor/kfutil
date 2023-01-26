#!/usr/bin/env bash
#gh auth login
set -e
export REPO_NAMESPACE=Keyfactor
export REPO_NAME=kfutil
export REPO_PATH="${REPO_NAMESPACE}/${REPO_NAME}"
export TAG=$(gh release list --repo $REPO_PATH --limit 1 | cut -f 1)
export RELEASE_DIR="${REPO_NAME}-${TAG}"
export OS_ARCH=$(uname -m)
export OS_NAME=$(uname -s)

function check_installed_binary() {
  bin_installed=$(which "$1")
  if [[ -z "$bin_installed" ]]; then
    echo "'$1' is not installed, unable to continue. Please install '$1' and try again. For more information, see https://github.com/Keyfactor/kfutil/blob/main/README.md#prerequisites ."
    exit 1
  fi
}

# convert to amd64
if [[ "$OS_ARCH" == "x86_64" ]]; then
  export OS_ARCH="amd64"
fi
# convert to arm64
if [[ "$OS_ARCH" == "aarch64" ]]; then
  export OS_ARCH="arm64"
fi

export FILE_PATTERN=$(echo "${OS_NAME}_${OS_ARCH}" | tr '[:upper:]' '[:lower:]')

# check deps
check_installed_binary "gh"
check_installed_binary "zip"
check_installed_binary "unzip"

# download release
mkdir -p "$RELEASE_DIR" || true
cd "$RELEASE_DIR" && gh release download "$TAG" --repo $REPO_PATH --pattern "*${FILE_PATTERN}*" --clobber

# unzip release
cd .. && zip -r "${RELEASE_DIR}.zip" "$RELEASE_DIR"
cd kfutil-"${TAG}" && ls && unzip "${REPO_NAME}_${TAG#"v"}_${FILE_PATTERN}.zip"

# move binary to $HOME/.local/bin
mkdir -p "${HOME}/.local/bin/" && \
  mv "${REPO_NAME}" "${HOME}/.local/bin/${REPO_NAME}"

# cleanup
cd .. && rm -rf "${RELEASE_DIR}" && rm -f "${REPO_NAME}-v${TAG}.zip"

# test
check_installed_binary "${REPO_NAME}"
kfutil version
