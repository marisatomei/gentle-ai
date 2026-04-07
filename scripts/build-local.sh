#!/usr/bin/env bash
# =============================================================================
# gentle-ai — Local Build & Install Script (macOS / Linux)
# Builds the binary from the current source tree and installs it locally.
#
# Use this when you are working on a local branch and want to test your
# changes with a real versioned binary — without publishing a GitHub Release.
#
# The script:
#   1. Resolves a version string (--version flag, git tag, or fallback)
#   2. Runs `go build` with that version injected via ldflags
#   3. Installs the binary to $INSTALL_DIR (default: ~/.local/bin)
#   4. Verifies the installed binary reports the expected version
#
# Usage:
#   ./scripts/build-local.sh                          # auto version
#   ./scripts/build-local.sh --version 1.2.0-copilot  # explicit version
#   ./scripts/build-local.sh --dir /usr/local/bin      # custom install dir
# =============================================================================

set -euo pipefail

BINARY_NAME="gentle-ai"
MAIN_PKG="./cmd/gentle-ai"
VERSION_VAR="main.version"

# ============================================================================
# Color support
# ============================================================================

setup_colors() {
    if [ -t 1 ] && [ "${TERM:-}" != "dumb" ]; then
        RED='\033[0;31m'
        GREEN='\033[0;32m'
        YELLOW='\033[1;33m'
        BLUE='\033[0;34m'
        CYAN='\033[0;36m'
        BOLD='\033[1m'
        DIM='\033[2m'
        NC='\033[0m'
    else
        RED='' GREEN='' YELLOW='' BLUE='' CYAN='' BOLD='' DIM='' NC=''
    fi
}

info()    { echo -e "${BLUE}[info]${NC}    $*"; }
success() { echo -e "${GREEN}[ok]${NC}      $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}    $*"; }
fatal()   { echo -e "${RED}[error]${NC}   $*" >&2; exit 1; }
step()    { echo -e "\n${CYAN}${BOLD}==>${NC} ${BOLD}$*${NC}"; }

# ============================================================================
# Banner
# ============================================================================

print_banner() {
    echo ""
    echo -e "${CYAN}${BOLD}"
    echo "   ____            _   _              _    ___ "
    echo "  / ___| ___ _ __ | |_| | ___        / \  |_ _|"
    echo " | |  _ / _ \ '_ \| __| |/ _ \_____ / _ \  | | "
    echo " | |_| |  __/ | | | |_| |  __/_____/ ___ \ | | "
    echo "  \____|\___|_| |_|\__|_|\___|    /_/   \_\___|"
    echo -e "${NC}"
    echo -e "  ${DIM}Local Build & Install${NC}"
    echo ""
}

# ============================================================================
# Argument parsing
# ============================================================================

EXPLICIT_VERSION=""
INSTALL_DIR=""

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --version|-v)
                [ $# -lt 2 ] && fatal "--version requires an argument"
                EXPLICIT_VERSION="$2"; shift 2
                ;;
            --dir|-d)
                [ $# -lt 2 ] && fatal "--dir requires an argument"
                INSTALL_DIR="$2"; shift 2
                ;;
            --help|-h)
                echo "Usage: $0 [--version VERSION] [--dir DIR]"
                echo ""
                echo "  --version VERSION   Version string to embed (default: auto from git tag)"
                echo "  --dir DIR           Install directory (default: ~/.local/bin)"
                exit 0
                ;;
            *)
                fatal "Unknown option: $1. Use --help for usage."
                ;;
        esac
    done
}

# ============================================================================
# Resolve version
# ============================================================================

resolve_version() {
    if [ -n "$EXPLICIT_VERSION" ]; then
        # Strip leading 'v' for consistency
        BUILD_VERSION="${EXPLICIT_VERSION#v}"
        info "Using explicit version: $BUILD_VERSION"
        return
    fi

    # Try the nearest git tag on the current branch
    if command -v git &>/dev/null && git rev-parse --git-dir &>/dev/null 2>&1; then
        local tag
        tag="$(git describe --tags --abbrev=0 2>/dev/null || true)"
        if [ -n "$tag" ]; then
            local commits_since short_hash
            commits_since="$(git rev-list "${tag}..HEAD" --count 2>/dev/null || echo 0)"
            short_hash="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
            BUILD_VERSION="${tag#v}"
            if [ "$commits_since" -gt 0 ]; then
                BUILD_VERSION="${BUILD_VERSION}-dev.${commits_since}+${short_hash}"
            fi
            info "Resolved version from git: $BUILD_VERSION"
            return
        fi
    fi

    # Fallback
    BUILD_VERSION="0.0.0-local"
    warn "Could not resolve git version — using fallback: $BUILD_VERSION"
}

# ============================================================================
# Prerequisites
# ============================================================================

check_prerequisites() {
    step "Checking prerequisites"

    if ! command -v go &>/dev/null; then
        fatal "Go is not installed or not in PATH. Install it from https://golang.org/dl/"
    fi

    local go_version
    go_version="$(go version)"
    success "Go found: $go_version"
}

# ============================================================================
# Build
# ============================================================================

build_binary() {
    local out_path="$1"

    step "Building $BINARY_NAME"
    info "Version : $BUILD_VERSION"
    info "Output  : $out_path"

    local ldflags="-s -w -X ${VERSION_VAR}=${BUILD_VERSION}"

    if ! go build -ldflags "$ldflags" -o "$out_path" "$MAIN_PKG"; then
        fatal "go build failed"
    fi

    local size_kb
    size_kb=$(( $(wc -c < "$out_path") / 1024 ))
    success "Built successfully (${size_kb} KB)"
}

# ============================================================================
# Install
# ============================================================================

install_binary() {
    local src_path="$1"

    step "Installing binary"

    # Resolve install dir
    if [ -z "$INSTALL_DIR" ]; then
        if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="${HOME}/.local/bin"
        fi
    fi

    mkdir -p "$INSTALL_DIR"

    local dest_path="${INSTALL_DIR}/${BINARY_NAME}"

    if cp "$src_path" "$dest_path" 2>/dev/null; then
        chmod +x "$dest_path"
    elif command -v sudo &>/dev/null; then
        warn "Permission denied. Trying with sudo..."
        sudo cp "$src_path" "$dest_path"
        sudo chmod +x "$dest_path"
    else
        fatal "Cannot write to ${INSTALL_DIR}. Use --dir to specify a writable directory."
    fi

    success "Installed to: $dest_path"

    # PATH advisory
    if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
        warn "${INSTALL_DIR} is not in your PATH."
        echo ""
        echo -e "  ${DIM}Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):${NC}"
        echo -e "  ${DIM}export PATH=\"\$PATH:${INSTALL_DIR}\"${NC}"
        echo ""
    fi
}

# ============================================================================
# Verify
# ============================================================================

verify_installation() {
    step "Verifying installation"

    hash -r 2>/dev/null || true

    local binary_path
    if command -v "$BINARY_NAME" &>/dev/null; then
        binary_path="$(command -v "$BINARY_NAME")"
    elif [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Could not locate installed binary — you may need to reload your shell."
        return 0
    fi

    local version_output
    version_output="$("$binary_path" version 2>&1 || true)"
    success "$BINARY_NAME version output: $version_output"

    if [[ "$version_output" != *"$BUILD_VERSION"* ]]; then
        warn "Reported version does not match expected '$BUILD_VERSION' — check ldflags"
    fi
}

# ============================================================================
# Main
# ============================================================================

main() {
    setup_colors
    parse_args "$@"
    print_banner

    check_prerequisites

    step "Resolving version"
    resolve_version

    local tmp_binary
    tmp_binary="$(mktemp /tmp/${BINARY_NAME}-local-build.XXXXXX)"
    trap 'rm -f "$tmp_binary"' EXIT

    build_binary "$tmp_binary"
    install_binary "$tmp_binary"
    verify_installation

    echo ""
    echo -e "${GREEN}${BOLD}Done! Local build ${BUILD_VERSION} is installed.${NC}"
    echo -e "Run: ${CYAN}${BINARY_NAME}${NC}"
    echo ""
}

main "$@"
