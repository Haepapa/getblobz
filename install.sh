#!/bin/bash
# getblobz installation script
# This script downloads and installs the latest version of getblobz

set -e

# Configuration
REPO="haepapa/getblobz"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac
    
    info "Detected platform: $OS-$ARCH"
}

# Get latest version if not specified
get_version() {
    if [ "$VERSION" = "latest" ]; then
        info "Fetching latest version..."
        VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | \
            grep '"tag_name":' | \
            sed -E 's/.*"([^"]+)".*/\1/')
        
        if [ -z "$VERSION" ]; then
            error "Failed to fetch latest version"
        fi
        info "Latest version: $VERSION"
    fi
}

# Download binary
download_binary() {
    local ext=""
    [ "$OS" = "windows" ] && ext=".exe"
    
    BINARY_NAME="getblobz-${VERSION}-${OS}-${ARCH}${ext}"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME"
    
    info "Downloading from: $DOWNLOAD_URL"
    
    if ! curl -L -f -o "getblobz${ext}" "$DOWNLOAD_URL"; then
        error "Failed to download binary. Please check version and platform support."
    fi
    
    info "Downloaded successfully"
}

# Verify checksum
verify_checksum() {
    local ext=""
    [ "$OS" = "windows" ] && ext=".exe"
    
    CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}.sha256"
    
    info "Downloading checksum..."
    if curl -L -f -o "getblobz.sha256" "$CHECKSUM_URL" 2>/dev/null; then
        info "Verifying checksum..."
        if command -v sha256sum >/dev/null 2>&1; then
            if ! sha256sum -c "getblobz.sha256" >/dev/null 2>&1; then
                warn "Checksum verification failed"
                rm -f "getblobz${ext}" "getblobz.sha256"
                error "Binary checksum does not match"
            fi
            info "Checksum verified"
        else
            warn "sha256sum not found, skipping verification"
        fi
        rm -f "getblobz.sha256"
    else
        warn "Checksum file not available, skipping verification"
    fi
}

# Install binary
install_binary() {
    local ext=""
    [ "$OS" = "windows" ] && ext=".exe"
    
    chmod +x "getblobz${ext}"
    
    # Check if install directory is writable
    if [ -w "$INSTALL_DIR" ]; then
        info "Installing to $INSTALL_DIR..."
        mv "getblobz${ext}" "$INSTALL_DIR/getblobz${ext}"
    else
        info "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mv "getblobz${ext}" "$INSTALL_DIR/getblobz${ext}"
    fi
    
    info "Installation complete!"
}

# Verify installation
verify_installation() {
    local ext=""
    [ "$OS" = "windows" ] && ext=".exe"
    
    if command -v "getblobz${ext}" >/dev/null 2>&1; then
        info "Verification successful!"
        echo ""
        "getblobz${ext}" --version
        echo ""
        info "Run 'getblobz --help' to get started"
    else
        warn "getblobz is installed but not in PATH"
        warn "Add $INSTALL_DIR to your PATH to use getblobz"
    fi
}

# Main installation flow
main() {
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║           getblobz Installation Script                   ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""
    
    detect_platform
    get_version
    download_binary
    verify_checksum
    install_binary
    verify_installation
    
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║           ✓ Installation Complete                        ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
}

# Handle script arguments
while [ $# -gt 0 ]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -h|--help)
            echo "getblobz installation script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION    Install specific version (default: latest)"
            echo "  -d, --dir DIRECTORY      Installation directory (default: /usr/local/bin)"
            echo "  -h, --help               Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  VERSION                  Version to install"
            echo "  INSTALL_DIR              Installation directory"
            echo ""
            echo "Examples:"
            echo "  $0                                 # Install latest version"
            echo "  $0 --version v1.0.0                # Install specific version"
            echo "  $0 --dir ~/.local/bin              # Install to custom directory"
            echo ""
            exit 0
            ;;
        *)
            error "Unknown option: $1. Use -h for help."
            ;;
    esac
done

# Run installation
main
