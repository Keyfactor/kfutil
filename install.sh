#!/usr/bin/env bash

# Copyright 2023 The Keyfactor Command Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

: "${BINARY_NAME:=kfutil}"
: "${USE_SUDO:=false}"
: "${VERIFY_CHECKSUM:=true}"

if [ $EUID -ne 0 ] && [ "$USE_SUDO" = "true" ]; then
        : "${KFUTIL_INSTALL_DIR:=/usr/local/bin}"
    else
        : "${KFUTIL_INSTALL_DIR:=${HOME}/.local/bin}"
fi


HAS_CURL="$(type "curl" &>/dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &>/dev/null && echo true || echo false)"
HAS_JQ="$(type "jq" &>/dev/null && echo true || echo false)"
HAS_OPENSSL="$(type "openssl" &>/dev/null && echo true || echo false)"
HAS_UNZIP="$(type "unzip" &>/dev/null && echo true || echo false)"

# Runs the given command as root (detects if we are root already)
runAsRoot() {
    if [ $EUID -ne 0 ] && [ "$USE_SUDO" = "true" ]; then
        sudo "${@}"
    else
        "${@}"
    fi
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ ${#INPUT_ARGUMENTS[@]} -ne 0 ]]; then
      echo "Failed to install $BINARY_NAME with the arguments provided: ${INPUT_ARGUMENTS[*]}"
      usage
    else
      echo "Failed to install $BINARY_NAME"
    fi
    echo -e "For support, go to https://github.com/Keyfactor/kfutil"
  fi
  cleanup
  exit $result
}

# Get host architecture
initArch() {
    ARCH=$(uname -m)
    case $ARCH in
    armv5*) ARCH="armv5" ;;
    armv6*) ARCH="armv6" ;;
    armv7*) ARCH="arm" ;;
    aarch64) ARCH="arm64" ;;
    x86) ARCH="386" ;;
    x86_64) ARCH="amd64" ;;
    i686) ARCH="386" ;;
    i386) ARCH="386" ;;
    esac
}

# Get host OS
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

# Verify that the host OS/Arch is supported
verifySupported() {
    supported_builds=(
        "darwin-amd64"
        "darwin-arm64"
        "linux-386"
        "linux-amd64"
        "linux-arm"
        "linux-arm64"
        "linux-ppc64le"
        "linux-s390x"
        "windows-amd64"
    )

    match_found=false
    for build in "${supported_builds[@]}"; do
        if [[ "${build}" == "${OS}-${ARCH}" ]]; then
            match_found=true
            break
        fi
    done

    if [[ "${match_found}" == "false" ]]; then
        echo "No prebuilt binary for ${OS}-${ARCH}."
        echo "To build from source, go to https://github.com/Keyfactor/kfutil"
        exit 1
    fi

    if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
        echo "Either curl or wget is required"
        exit 1
    fi

    if [ "${HAS_JQ}" != "true" ]; then
        echo "jq is required"
        exit 1
    fi
    
    if [ "${HAS_UNZIP}" != "true" ]; then
        echo "unzip is required"
        exit 1
    fi

    if [ "${VERIFY_CHECKSUM}" == "true" ] && [ "${HAS_OPENSSL}" != "true" ]; then
        echo "In order to verify checksum, openssl must first be installed."
        echo "Please install openssl or set VERIFY_CHECKSUM=false in your environment."
        exit 1
    fi
}

# checkDesiredVersion checks if the desired version is available.
getVersion() {
    local remote_release_url="https://api.github.com/repos/keyfactor/kfutil/releases"
    # Get tag from release URL
    local releases_response=""
    if [ "${HAS_CURL}" == "true" ]; then
        releases_response=$(curl -L --silent --show-error --fail "$remote_release_url" 2>&1 || true)
    elif [ "${HAS_WGET}" == "true" ]; then
        releases_response=$(wget "$remote_release_url" -O - 2>&1 || true)
    fi

    # If VERSION is not set, get latest from GitHub API
    if [ -z "$VERSION" ]; then

        VERSION=$(echo "$releases_response" | jq '[.[] | select(.prerelease == false)] | max_by(.created_at) | .tag_name' | tr -d '"' | tr -d 'v')
        if [ -z "$VERSION" ]; then
            printf "Could not retrieve the latest release tag information from %s: %s\n" "${remote_release_url}" "${releases_response}"
            exit 1
        fi
    else
        # Clean up version if prefixed with 'v'
        VERSION=$(echo "$VERSION" | tr -d 'v')

        # Verify that the version exists as a release before continuing
        if ! echo "$releases_response" | jq  '.[] | select(.tag_name == "v'"$VERSION"'")' >/dev/null; then
            printf "Cannot find release '%s' for %s.\n" "$VERSION" "$remote_release_url"
            exit 1
        else
            echo "kfutil version $VERSION exists"
        fi
    fi
}

# checkkfutilInstalledVersion checks which version of kfutil is installed and
# if it needs to be changed.
checkkfutilInstalledVersion() {
    if [[ -f "${KFUTIL_INSTALL_DIR}/${BINARY_NAME}" ]]; then
        local version
        version=$("${KFUTIL_INSTALL_DIR}/${BINARY_NAME}" version)
        raw_version=$version
        version=${raw_version#*version }
        version=${version%%\%*}
        version=$(echo "$version" | tr -d 'v')
        if [[ "$version" == "$VERSION" ]]; then
            echo "kfutil ${version} is already installed"
            return 0
        else
            echo "Changing from kfutil 'v${version}' to 'v${VERSION}'."
            return 1
        fi
    else
        return 1
    fi
}

# downloadFile downloads the latest binary package and checksums
downloadFile() {
    local download_url
    local base_url
    base_url="https://github.com/Keyfactor/kfutil/releases/download/v${VERSION}"
    KFUTIL_DIST="kfutil_${VERSION}_${OS}_${ARCH}.zip"
    download_url="${base_url}/${KFUTIL_DIST}"
    checksum_url="${base_url}/kfutil_${VERSION}_SHA256SUMS"

    BASE_TEMP_DIR=$(mktemp -dt kfutil-installer-XXXXXX)
    KFUTIL_TMP_FILE="$BASE_TEMP_DIR/$KFUTIL_DIST"
    KFUTIL_SUM_FILE="$BASE_TEMP_DIR/kfutil_${VERSION}_SHA256SUMS"

    echo "Downloading kfutil ${VERSION} ${OS}-${ARCH}"
    if [ "${HAS_CURL}" == "true" ]; then
        curl -SsL "$download_url" -o "$KFUTIL_TMP_FILE"
        curl -SsL "$checksum_url" -o "$KFUTIL_SUM_FILE"
    elif [ "${HAS_WGET}" == "true" ]; then
        wget -q -O "$KFUTIL_TMP_FILE" "$download_url"
        wget -q -O "$KFUTIL_SUM_FILE" "$checksum_url"
    fi
}

# verifyChecksum verifies the SHA256 checksum of the binary package.
verifyChecksum() {
    local sum
    local expected_sum

    printf "Verifying checksum... "
    sum=$(openssl sha1 -sha256 "${KFUTIL_TMP_FILE}" | awk '{print $2}')

    expected_sum=$(grep "${KFUTIL_DIST}" "${KFUTIL_SUM_FILE}" | cut -d ' ' -f1)
    if [ "$sum" != "$expected_sum" ]; then
        echo "SHA sum of ${KFUTIL_TMP_FILE} does not match. Aborting."
        exit 1
    fi
    echo "Done."
}

# installFile installs the kfutil binary.
installFile() {
    local tmp_bin_dir
    tmp_bin_dir="${BASE_TEMP_DIR}/bin"
    mkdir -p "$tmp_bin_dir"
    unzip "$KFUTIL_TMP_FILE" -d "$tmp_bin_dir" >/dev/null
    echo "Preparing to install $BINARY_NAME into ${KFUTIL_INSTALL_DIR}"
    runAsRoot mkdir -p "$KFUTIL_INSTALL_DIR"
    runAsRoot cp "${tmp_bin_dir}/$BINARY_NAME" "$KFUTIL_INSTALL_DIR/$BINARY_NAME"
    echo "$BINARY_NAME installed into $KFUTIL_INSTALL_DIR/$BINARY_NAME"

    testVersion
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
    set +e
    command -v $BINARY_NAME >/dev/null 2>&1
    if [ "$?" = "1" ]; then
        echo "$BINARY_NAME not found. Is $KFUTIL_INSTALL_DIR on your "'$PATH?'
        exit 1
    fi

    local version
    version=$("${KFUTIL_INSTALL_DIR}/${BINARY_NAME}" version)
    raw_version=$version
    version=${raw_version#*version }
    version=${version%%\%*}
    version=$(echo "$version" | tr -d 'v')

    if [[ "$version" == "$VERSION" ]]; then
        echo "$BINARY_NAME $version is installed and available."
    else
        echo "$BINARY_NAME $version is installed, but wanted version $VERSION."
        exit 1
    fi

    set -e
}

cleanup() {
    if [[ -d "${BASE_TEMP_DIR:-}" ]]; then
        rm -rf "$BASE_TEMP_DIR"
    fi
}

usage() {
    echo "Usage: get-kfutil [-v] [-h]"
    echo "  -v      -- kfutil version to install in the form of v0.0.0"
    echo "  -h      -- Print this usage message"
    echo ""
    echo "Or, set the following environment variables:"
    echo "  VERSION -- kfutil version to install in the form of v0.0.0"
}

# Trap if any command in a pipeline exits non-zero
trap "fail_trap" EXIT
set -e

# Parse command line arguments
INPUT_ARGUMENTS=("$@")
set -u
while getopts v:h option
do
    case "${option}"
    in
        v) VERSION=${OPTARG};;
        h) usage && exit 0;;
        *) usage && exit 1;;
    esac
done
set +u

initArch
initOS
verifySupported
getVersion

if ! checkkfutilInstalledVersion; then
    downloadFile
    verifyChecksum
    installFile
fi

cleanup