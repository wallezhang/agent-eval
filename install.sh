#!/usr/bin/env bash
# Copyright 2024 agent-eval authors
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

set -euo pipefail

REPO="wallezhang/agent-eval"
BINARY="agent-eval"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

info() { printf "\033[1;34m==>\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33mWarning:\033[0m %s\n" "$*"; }
error() { printf "\033[1;31mError:\033[0m %s\n" "$*" >&2; exit 1; }

# Detect OS
detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux)  echo "linux" ;;
        Darwin) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*)
            error "Windows is not supported by this installer. Please download the .exe binary manually from https://github.com/${REPO}/releases"
            ;;
        *) error "Unsupported operating system: $os" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) error "Unsupported architecture: $arch" ;;
    esac
}

# Get latest release tag from GitHub API
get_latest_version() {
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    local response

    if command -v curl &>/dev/null; then
        response=$(curl -fsSL "$url" 2>/dev/null) || error "Failed to fetch latest release info. Check your network connection."
    elif command -v wget &>/dev/null; then
        response=$(wget -qO- "$url" 2>/dev/null) || error "Failed to fetch latest release info. Check your network connection."
    else
        error "Either curl or wget is required."
    fi

    echo "$response" | grep '"tag_name"' | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/'
}

# Download a file
download() {
    local url="$1" dest="$2"
    if command -v curl &>/dev/null; then
        curl -fsSL -o "$dest" "$url"
    elif command -v wget &>/dev/null; then
        wget -qO "$dest" "$url"
    fi
}

# Verify SHA256 checksum
verify_checksum() {
    local file="$1" expected="$2"
    local actual

    if command -v sha256sum &>/dev/null; then
        actual=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum &>/dev/null; then
        actual=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        return 0
    fi

    if [ "$actual" != "$expected" ]; then
        error "Checksum verification failed.\n  Expected: $expected\n  Got:      $actual"
    fi
}

main() {
    local os arch version

    os=$(detect_os)
    arch=$(detect_arch)

    info "Detected platform: ${os}/${arch}"

    info "Fetching latest release..."
    version=$(get_latest_version)

    if [ -z "$version" ]; then
        error "Could not determine latest version."
    fi

    info "Latest version: ${version}"

    # Check if already installed and up to date
    if command -v "$BINARY" &>/dev/null; then
        local current
        current=$("$BINARY" --version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+[^ ]*' || echo "")
        if [ "$current" = "$version" ]; then
            info "AgentEval ${version} is already installed. Nothing to do."
            exit 0
        fi
        if [ -n "$current" ]; then
            info "Upgrading from ${current} to ${version}..."
        fi
    fi

    local filename="${BINARY}-${version}-${os}-${arch}"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading ${filename}..."
    download "$download_url" "${tmpdir}/${filename}" || error "Failed to download binary. The release may not include a binary for ${os}/${arch}."

    info "Downloading checksums..."
    if download "$checksums_url" "${tmpdir}/checksums.txt" 2>/dev/null; then
        local expected_checksum
        expected_checksum=$(grep "${filename}" "${tmpdir}/checksums.txt" | awk '{print $1}')
        if [ -n "$expected_checksum" ]; then
            info "Verifying checksum..."
            verify_checksum "${tmpdir}/${filename}" "$expected_checksum"
        else
            warn "No checksum found for ${filename} in checksums.txt. Skipping verification."
        fi
    else
        warn "Could not download checksums.txt. Skipping checksum verification."
    fi

    info "Installing to ${INSTALL_DIR}/${BINARY}..."
    chmod +x "${tmpdir}/${filename}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmpdir}/${filename}" "${INSTALL_DIR}/${BINARY}"
    else
        sudo mv "${tmpdir}/${filename}" "${INSTALL_DIR}/${BINARY}"
    fi

    info "AgentEval ${version} installed successfully!"
    info "Run 'agent-eval --help' to get started."
}

main
