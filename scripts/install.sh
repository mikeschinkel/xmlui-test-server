#!/bin/bash
#
# Install script for xmlui-test-server prebuilt binaries
# Usage: install.sh [macos|linux]
#

set -e

# Get script directory and source variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/shared.sh"

usage() {
    echo "Usage: $0 [macos|linux]"
    echo ""
    echo "Download and install prebuilt xmlui-test-server binary for specified platform."
    exit 1
}

install_macos() {
    # Detect architecture
    case "$ARCH" in
        arm64)
            local binary_name="xmlui-test-server-mac-arm"
            log "Downloading and installing prebuilt macOS ARM binary..."
            ;;
        amd64)
            local binary_name="xmlui-test-server-mac-intel"
            log "Downloading and installing prebuilt macOS Intel binary..."
            ;;
        *)
            error "Unsupported macOS architecture: $ARCH"
            ;;
    esac
    
    ensure_bin_dir
    
    local url="https://github.com/JonUdell/xmlui-test-server/releases/download/v0.0.1/${binary_name}.tar.gz"
    local binary_path="$BIN_DIR/$binary_name"
    
    # Download and extract to bin directory
    curl -L "$url" | tar -xz -C "$BIN_DIR"
    
    # Set permissions
    chmod +x "$binary_path"
    
    # Remove quarantine attribute (macOS security feature)
    if command -v xattr >/dev/null 2>&1; then
        xattr -d com.apple.quarantine "$binary_path" 2>/dev/null || true
    fi
    
    log "Installation complete! Binary: $binary_path"
}

install_linux() {
    error "Linux prebuilt binaries not yet available. Use 'make build' or 'make build-ext-linux' instead."
}

main() {
    local platform="${1:-}"
    
    if [[ -z "$platform" ]]; then
        usage
    fi
    
    case "$platform" in
        macos)
            install_macos
            ;;
        linux)
            install_linux
            ;;
        *)
            echo "Error: Unknown platform '$platform'"
            usage
            ;;
    esac
}

# Run main function
main "$@"