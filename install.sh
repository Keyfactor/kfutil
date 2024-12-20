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

# Use parameter expansion to provide default values.
: "${BINARY_NAME:=kfutil}"
: "${USE_SUDO:=false}"
: "${VERIFY_CHECKSUM:=true}"
: "${INSTALL_DIR:=${HOME}/.local/bin}"

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
        echo ""
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
    OS=$(echo $(uname) | tr '[:upper:]' '[:lower:]')

    case "$OS" in
    # Minimalist GNU for Windows
    mingw* | cygwin*) OS='windows' ;;
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
        echo "Either curl or wget is required."
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
        echo "Latest release version is $VERSION"
    else
        # Clean up version if prefixed with 'v'
        VERSION=$(echo "$VERSION" | tr -d 'v')

        # Verify that the version exists as a release before continuing
        if ! echo "$releases_response" | jq '.[] | select(.tag_name == "v'"$VERSION"'")' >/dev/null; then
            printf "Cannot find release '%s' for %s.\n" "$VERSION" "$remote_release_url"
            exit 1
        else
            echo "$BINARY_NAME version $VERSION exists"
        fi
    fi
}

# checkBinaryInstalledVersion checks which version of kfutil is installed and
# if it needs to be changed.
checkBinaryInstalledVersion() {
    if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        local version
        version=$("${INSTALL_DIR}/${BINARY_NAME}" version)
        raw_version=$version
        version=${raw_version#*version }
        version=${version%%\%*}
        version=$(echo "$version" | tr -d 'v')
        if [[ "$version" == "$VERSION" ]]; then
            echo "kfutil ${version} is already installed in ${INSTALL_DIR}/${BINARY_NAME}"
            return 0
        else
            echo "Changing from ${BINARY_NAME} 'v${version}' to 'v${VERSION}'."
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
    base_url="https://github.com/Keyfactor/${BINARY_NAME}/releases/download/${VERSION}"
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
    if [ "${VERIFY_CHECKSUM}" == "false" ]; then
        echo "Skipping checksum verification"
        return 0
    fi

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
    echo "Preparing to install $BINARY_NAME into ${INSTALL_DIR}"
    runAsRoot mkdir -p "$INSTALL_DIR"
    runAsRoot cp "${tmp_bin_dir}/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    echo "$BINARY_NAME installed into $INSTALL_DIR/$BINARY_NAME"

    testVersion
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
    set +e
    command -v $BINARY_NAME >/dev/null 2>&1
    if [ "$?" = "1" ]; then
        echo "$BINARY_NAME not found. Is $INSTALL_DIR in your "'$PATH?'
        exit 1
    fi

    local version
    version=$("${INSTALL_DIR}/${BINARY_NAME}" version)
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

uninstall_fail_trap() {
    result=$?

    if [[ "$result" -ne "0" ]]; then
        echo "Failed to uninstall $BINARY_NAME."

        echo "You may need to use 'sudo' to uninstall. Refer to the usage:"
        usage
    fi

    exit $result
}

uninstall() {
    set +e
    if ! current_install_dir=$(which $BINARY_NAME); then
        echo "$BINARY_NAME is not installed"
        exit 0
    fi
    trap uninstall_fail_trap EXIT
    set -e

    printf "Uninstalling %s from %s... " "$BINARY_NAME" "${current_install_dir}"

    # Uninstall binary
    runAsRoot rm -f "$current_install_dir"

    set +e
    command -v $BINARY_NAME >/dev/null 2>&1
    if [ "$?" != "1" ]; then
        echo "$BINARY_NAME is still installed. Uninstallation failed."
        exit 1
    fi
    set -e

    echo "Done."
}

usage() {
    echo "Usage: $0 [-v] [-d] [-h]"
    echo "  -v      -- kfutil version to install in the form of v0.0.0"
    echo "  -d      -- The install directory for kfutil. Defaults to ${HOME}/.local/bin"
    echo "  -h      -- Print this usage message"
    echo ""
    echo "Or, set the following environment variables:"
    echo "  USE_SUDO           -- Whether to use sudo or not. Defaults to false."
    echo "  VERSION            -- kfutil version to install in the form of v0.0.0"
    echo "  INSTALL_DIR        -- The install directory for kfutil. Defaults to ${HOME}/.local/bin"
    echo "  BINARY_NAME        -- The name of the binary to install. Defaults to kfutil"
    echo "  VERIFY_CHECKSUM    -- Whether or not to verify the downloaded binary checksum. Defaults to true."
    echo ""
    echo "Uninstall kfutil:"
    echo "  $0 --uninstall"
    echo ""
    echo "Examples:"
    echo "  Install the latest stable release into ${HOME}/.local/bin:"
    echo "    $0"
    echo "  Install a specific version of kfutil into /usr/local/bin:"
    echo "    USE_SUDO=true VERSION=v1.2.0 INSTALL_DIR=/usr/local/bin $0"
    echo "  or"
    echo "    sudo $0 -v v1.2.0 -d /usr/local/bin"
}

# Trap if any command in a pipeline exits non-zero
trap "fail_trap" EXIT
set -e

# Parse command line arguments
INPUT_ARGUMENTS=("$@")
set -u
# If INPUT_ARGUMENTS contains --uninstall, uninstall kfutil and exit.
if [[ ${#INPUT_ARGUMENTS[@]} -gt 0 && " ${INPUT_ARGUMENTS[*]} " == *" --uninstall "* ]]; then
    uninstall
    exit 0
fi

while getopts v:d:h option; do
    case "${option}" in
    v) VERSION=${OPTARG} && echo "Setting target version to ${VERSION}" ;;
    d) INSTALL_DIR=${OPTARG} && echo "Setting install directory to ${INSTALL_DIR}" ;;
    h) usage && exit 0 ;;
    *) usage && exit 1 ;;
    esac
done
set +u

initArch
initOS
verifySupported
getVersion

if ! checkBinaryInstalledVersion; then
    downloadFile
    verifyChecksum
    installFile
fi

cleanup
