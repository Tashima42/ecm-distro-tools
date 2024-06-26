#!/bin/sh

set -e

TMP_DIR=""
REPO_NAME="ecm-distro-tools"
REPO_URL="https://github.com/tashima42/${REPO_NAME}"
REPO_RELEASE_URL="${REPO_URL}/releases"
INSTALL_DIR="$HOME/.local/bin/ecm-distro-tools"
SUFFIX=""


# setup_arch set arch and suffix fatal if architecture not supported.
setup_arch() {
    case $(uname -m) in
    x86_64|amd64)
        ARCH=amd64
        SUFFIX=$(uname -s | tr '[:upper:]' '[:lower:]')-${ARCH}
        ;;
    aarch64|arm64)
        ARCH=arm64
        SUFFIX=$(uname -s | tr '[:upper:]' '[:lower:]')-${ARCH}
        ;;
    *)
        fatal "unsupported architecture ${ARCH}"
        ;;
    esac
}

# setup_tmp creates a temporary directory and cleans up when done.
setup_tmp() {
    rm -rf /tmp/ecm-distro-tools
    mkdir -p /tmp/ecm-distro-tools
    TMP_DIR=/tmp/ecm-distro-tools
    cleanup() {
        code=$?
        set +e
        trap - EXIT
        rm -rf "${TMP_DIR}"
        exit "$code"
    }
    trap cleanup INT EXIT
}

# install_binaries installs the binaries from the downloaded tar.
install_binaries() {
    cd "${TMP_DIR}"
    wget "${REPO_RELEASE_URL}/download/${RELEASE_VERSION}/release-linux-amd64"
    mkdir -p "${INSTALL_DIR}"

    for f in * ; do
      file_name="${f}"
      if echo "${f}" | grep -q "${SUFFIX}"; then
        file_name=${file_name%"-${SUFFIX}"}
      fi
      cp "${TMP_DIR}/${f}" "${INSTALL_DIR}/${file_name}"
    done
}

{ # main
    RELEASE_VERSION=$1
    if [ -n "${ECM_VERSION}" ]; then
        RELEASE_VERSION=${ECM_VERSION}
    fi

    if [ -z "$RELEASE_VERSION" ]; then 
        RELEASE_VERSION=$(basename "$(curl -Ls -o /dev/null -w %\{url_effective\} https://github.com/tashima42/ecm-distro-tools/releases/latest)")
    fi

    echo "Installing ECM Distro Tools: ${RELEASE_VERSION}"

    setup_tmp
    setup_arch

    install_binaries

    printf "Run command to access tools:\n\nPATH=%s:%s\n\n" "${PATH}" "${INSTALL_DIR}"

    exit 0
}
