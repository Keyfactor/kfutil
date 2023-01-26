#!/usr/bin/env bash
#gh auth login
export REPO_NAMESPACE=Keyfactor
export REPO_NAME=kfutil
export REPO_PATH="${REPO_NAMESPACE}/${REPO_NAME}"
export TAG=$(gh release list --repo $REPO_PATH --limit 1 | cut -f 1)
export RELEASE_DIR="${REPO_NAME}-${TAG}"
mkdir -p "$RELEASE_DIR"
OS_ARCH=$(uname -m)
OS_NAME=$(uname -s)

# convert to amd64
if [[ "$OS_ARCH" == "x86_64" ]]; then
    OS_ARCH="amd64"
fi
# convert to arm64
if [[ "$OS_ARCH" == "aarch64" ]]; then
    OS_ARCH="arm64"
fi

export file_pattern=$(echo "${OS_NAME}_${OS_ARCH}" | tr '[:upper:]' '[:lower:]')
cd "$RELEASE_DIR" && gh release download "$TAG" --repo $REPO_PATH --pattern "*${file_pattern}*" --clobber
cd .. && zip -r "${RELEASE_DIR}.zip" "$RELEASE_DIR"
cd kfutil-"${TAG}" && ls && unzip "${REPO_NAME}_${TAG#"v"}_${file_pattern}.zip"
mv "${REPO_NAME}" "${HOME}/.local/bin/${REPO_NAME}"
cd .. && rm -rf "${RELEASE_DIR}"
kfutil version