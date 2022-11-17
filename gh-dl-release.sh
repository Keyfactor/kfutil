#!/usr/bin/env bash
#gh auth login
export REPO_NAMESPACE=Keyfactor
export REPO_NAME=kfutil
export REPO_PATH="${REPO_NAMESPACE}/${REPO_NAME}"
export TAG=$(gh release list --repo $REPO_PATH --limit 1 | cut -f 1)
export RELEASE_DIR="${REPO_NAME}-${TAG}"
mkdir -p "$RELEASE_DIR"
cd "$RELEASE_DIR" && gh release download "$TAG" --repo $REPO_PATH --clobber
cd .. && zip -r "${RELEASE_DIR}.zip" "$RELEASE_DIR"